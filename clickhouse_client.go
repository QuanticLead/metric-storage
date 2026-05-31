package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/zeebo/xxh3"
)

// GaugeRow represents a single gauge data point for ClickHouse insertion.
type GaugeRow struct {
	ResourceAttributes    map[string]string
	ResourceSchemaUrl     string
	ScopeName             string
	ScopeVersion          string
	ScopeAttributes       map[string]string
	ScopeDroppedAttrCount uint32
	ScopeSchemaUrl        string
	ServiceName           string
	MetricName            string
	MetricDescription     string
	MetricUnit            string
	Attributes            map[string]string
	StartTimeUnix         time.Time
	TimeUnix              time.Time
	Value                 float64
	Flags                 uint32
}

// SumRow represents a single sum data point for ClickHouse insertion.
type SumRow struct {
	GaugeRow
	AggregationTemporality int32
	IsMonotonic            bool
}

// MetricsStore defines the interface for storing metrics in ClickHouse.
type MetricsStore interface {
	CreateTables(ctx context.Context) error
	InsertGauge(ctx context.Context, rows []GaugeRow) error
	InsertSum(ctx context.Context, rows []SumRow) error
	Close() error
}

// ClickHouseMetricsStore implements MetricsStore using a ClickHouse connection.
type ClickHouseMetricsStore struct {
	conn driver.Conn
}

// In-memory cache for registered MetricIDs
var metadataCache sync.Map // map[uuid.UUID]bool

// Serialize map deterministically by sorting keys
func writeSortedMap(buf *bytes.Buffer, m map[string]string) {
	if len(m) == 0 {
		return
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString(m[k])
	}
}

func computeMetricID(r *GaugeRow) uuid.UUID {
	var buf bytes.Buffer
	buf.WriteString(r.ServiceName)
	buf.WriteString(r.MetricName)
	buf.WriteString(r.MetricUnit)
	buf.WriteString(r.MetricDescription)
	buf.WriteString(r.ResourceSchemaUrl)
	buf.WriteString(r.ScopeName)
	buf.WriteString(r.ScopeVersion)
	buf.WriteString(r.ScopeSchemaUrl)
	_ = binary.Write(&buf, binary.BigEndian, r.ScopeDroppedAttrCount)

	writeSortedMap(&buf, r.ResourceAttributes)
	writeSortedMap(&buf, r.ScopeAttributes)
	writeSortedMap(&buf, r.Attributes)

	hash := xxh3.Hash128(buf.Bytes())
	var b [16]byte
	binary.BigEndian.PutUint64(b[0:8], hash.Hi)
	binary.BigEndian.PutUint64(b[8:16], hash.Lo)
	return uuid.UUID(b)
}

// NewClickHouseMetricsStore creates a new ClickHouseMetricsStore connected to the given address.
func NewClickHouseMetricsStore(ctx context.Context, addr string, database string, username string, password string) (*ClickHouseMetricsStore, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("opening clickhouse connection: %w", err)
	}
	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("pinging clickhouse: %w", err)
	}
	return &ClickHouseMetricsStore{conn: conn}, nil
}

// CreateTables executes DDL for all metric tables.
func (s *ClickHouseMetricsStore) CreateTables(ctx context.Context) error {
	ddls := []string{
		createMetadataTableSQL,
		createGaugeTableSQL,
		createSumTableSQL,
		createHistogramTableSQL,
		createExponentialHistogramTableSQL,
		createSummaryTableSQL,
	}
	for _, ddl := range ddls {
		if err := s.conn.Exec(ctx, ddl); err != nil {
			return fmt.Errorf("creating table: %w", err)
		}
	}
	return nil
}

// checkAndInsertMetadata checks the cache for the MetricID of each row and batch inserts new metadata.
func (s *ClickHouseMetricsStore) checkAndInsertMetadata(ctx context.Context, rows []GaugeRow) error {
	var newMetadataRows []GaugeRow
	var newMetadataIDs []uuid.UUID

	for _, r := range rows {
		id := computeMetricID(&r)
		if _, cached := metadataCache.Load(id); !cached {
			newMetadataRows = append(newMetadataRows, r)
			newMetadataIDs = append(newMetadataIDs, id)
			// Optimistically set to prevent parallel goroutines from duplicate queuing
			metadataCache.Store(id, true)
		}
	}

	if len(newMetadataRows) > 0 {
		batch, err := s.conn.PrepareBatch(ctx, "INSERT INTO otel_metrics_metadata")
		if err != nil {
			// Roll back the cache store if prepare fails
			for _, id := range newMetadataIDs {
				metadataCache.Delete(id)
			}
			return fmt.Errorf("preparing metadata batch: %w", err)
		}
		for i, r := range newMetadataRows {
			id := newMetadataIDs[i]
			if err := batch.Append(
				id,
				r.ResourceAttributes,
				r.ResourceSchemaUrl,
				r.ScopeName,
				r.ScopeVersion,
				r.ScopeAttributes,
				r.ScopeDroppedAttrCount,
				r.ScopeSchemaUrl,
				r.ServiceName,
				r.MetricName,
				r.MetricDescription,
				r.MetricUnit,
				r.Attributes,
			); err != nil {
				for _, cleanId := range newMetadataIDs {
					metadataCache.Delete(cleanId)
				}
				return fmt.Errorf("appending metadata row: %w", err)
			}
		}
		if err := batch.Send(); err != nil {
			for _, cleanId := range newMetadataIDs {
				metadataCache.Delete(cleanId)
			}
			return fmt.Errorf("sending metadata batch: %w", err)
		}
	}
	return nil
}

// InsertGauge batch-inserts gauge rows.
func (s *ClickHouseMetricsStore) InsertGauge(ctx context.Context, rows []GaugeRow) error {
	if err := s.checkAndInsertMetadata(ctx, rows); err != nil {
		return err
	}

	batch, err := s.conn.PrepareBatch(ctx, "INSERT INTO otel_metrics_gauge")
	if err != nil {
		return fmt.Errorf("preparing gauge batch: %w", err)
	}
	for _, r := range rows {
		id := computeMetricID(&r)
		if err := batch.Append(
			id,
			r.StartTimeUnix,
			r.TimeUnix,
			r.Value,
			r.Flags,
		); err != nil {
			return fmt.Errorf("appending gauge row: %w", err)
		}
	}
	return batch.Send()
}

// InsertSum batch-inserts sum rows.
func (s *ClickHouseMetricsStore) InsertSum(ctx context.Context, rows []SumRow) error {
	gaugeRows := make([]GaugeRow, len(rows))
	for i, r := range rows {
		gaugeRows[i] = r.GaugeRow
	}
	if err := s.checkAndInsertMetadata(ctx, gaugeRows); err != nil {
		return err
	}

	batch, err := s.conn.PrepareBatch(ctx, "INSERT INTO otel_metrics_sum")
	if err != nil {
		return fmt.Errorf("preparing sum batch: %w", err)
	}
	for _, r := range rows {
		id := computeMetricID(&r.GaugeRow)
		if err := batch.Append(
			id,
			r.StartTimeUnix,
			r.TimeUnix,
			r.Value,
			r.Flags,
			r.AggregationTemporality,
			r.IsMonotonic,
		); err != nil {
			return fmt.Errorf("appending sum row: %w", err)
		}
	}
	return batch.Send()
}

// Close closes the underlying ClickHouse connection.
func (s *ClickHouseMetricsStore) Close() error {
	return s.conn.Close()
}
