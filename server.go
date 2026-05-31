package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"sync"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ingestionChannel chan BatchJob
	workerWG         sync.WaitGroup
)

const name = "dash0.com/otlp-log-processor-backend"

var (
	meter                  = otel.Meter(name)
	logger                 = otelslog.NewLogger(name)
	metricsReceivedCounter metric.Int64Counter
)

func init() {
	var err error
	metricsReceivedCounter, err = meter.Int64Counter("com.dash0.homeexercise.metrics.received",
		metric.WithDescription("The number of metrics received by otlp-metrics-processor-backend"),
		metric.WithUnit("{metric}"))
	if err != nil {
		panic(err)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	cfg := LoadConfig()

	slog.SetDefault(logger)
	logger.Info("Starting application", slog.Any("config", cfg))

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(context.Background())
	if err != nil {
		return
	}

	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	var store MetricsStore
	if cfg.ClickHouseAddr != "" {
		slog.Info("Connecting to ClickHouse", slog.String("addr", cfg.ClickHouseAddr), slog.String("db", cfg.ClickHouseDB))
		clickhouseStore, err := NewClickHouseMetricsStore(context.Background(), cfg.ClickHouseAddr, cfg.ClickHouseDB, cfg.ClickHouseUser, cfg.ClickHousePassword)
		if err != nil {
			return fmt.Errorf("connecting to clickhouse: %w", err)
		}
		defer clickhouseStore.Close()

		slog.Info("Creating ClickHouse tables if not exist")
		if err := clickhouseStore.CreateTables(context.Background()); err != nil {
			return fmt.Errorf("creating clickhouse tables: %w", err)
		}
		store = clickhouseStore
	}

	// Initialize buffered channel
	ingestionChannel = make(chan BatchJob, cfg.ChannelSize)

	// Defer graceful channel close and worker drain
	defer func() {
		close(ingestionChannel)
		slog.Info("Ingestion channel closed, draining pending jobs...")
		workerWG.Wait()
		slog.Info("All background workers stopped successfully")
	}()

	// Start workers
	startWorkers(store, cfg.WorkerCount)

	slog.Debug("Starting listener", slog.String("listenAddr", cfg.ListenAddr))
	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.MaxRecvMsgSize(cfg.MaxReceiveMessageSize),
		grpc.Creds(insecure.NewCredentials()),
	)
	colmetricspb.RegisterMetricsServiceServer(grpcServer, newServer(cfg.ListenAddr, store))

	slog.Debug("Starting gRPC server")

	return grpcServer.Serve(listener)
}

func startWorkers(store MetricsStore, workerCount int) {
	for i := 0; i < workerCount; i++ {
		workerWG.Add(1)
		go func(workerID int) {
			defer workerWG.Done()
			slog.Debug("Background worker started", slog.Int("id", workerID))
			for job := range ingestionChannel {
				if store != nil {
					if len(job.Gauges) > 0 {
						if err := store.InsertGauge(context.Background(), job.Gauges); err != nil {
							slog.Error("Worker failed to insert gauges", slog.Int("id", workerID), "error", err)
						}
					}
					if len(job.Sums) > 0 {
						if err := store.InsertSum(context.Background(), job.Sums); err != nil {
							slog.Error("Worker failed to insert sums", slog.Int("id", workerID), "error", err)
						}
					}
				}
			}
			slog.Debug("Background worker stopped", slog.Int("id", workerID))
		}(i)
	}
}
