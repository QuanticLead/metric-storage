//go:build integration

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/testcontainers/testcontainers-go"
	redpanda "github.com/testcontainers/testcontainers-go/modules/redpanda"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type brokenStore struct{}

func (b *brokenStore) CreateTables(ctx context.Context) error { return nil }
func (b *brokenStore) InsertGauge(ctx context.Context, rows []GaugeRow) error {
	return fmt.Errorf("clickhouse simulated connection failure")
}
func (b *brokenStore) InsertSum(ctx context.Context, rows []SumRow) error {
	return fmt.Errorf("clickhouse simulated connection failure")
}
func (b *brokenStore) Close() error { return nil }

func setupRedpanda(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()

	ctr, err := redpanda.RunContainer(ctx,
		testcontainers.WithImage("docker.redpanda.com/redpandadata/redpanda:v24.1.1"),
		redpanda.WithAutoCreateTopics(),
	)
	if err != nil {
		t.Fatalf("starting redpanda container: %v", err)
	}

	broker, err := ctr.KafkaSeedBroker(ctx)
	if err != nil {
		t.Fatalf("getting kafka seed broker: %v", err)
	}

	cleanup := func() {
		if err := ctr.Terminate(ctx); err != nil {
			t.Logf("terminating redpanda container: %v", err)
		}
	}

	return broker, cleanup
}

func TestKafkaFallback(t *testing.T) {
	broker, cleanup := setupRedpanda(t)
	defer cleanup()

	ctx := context.Background()
	topic := "test-fallback-topic"

	cfg := &Config{
		KafkaBrokers:       broker,
		KafkaFallbackTopic: topic,
	}

	if err := InitKafka(cfg); err != nil {
		t.Fatalf("initializing kafka: %v", err)
	}
	defer CloseKafka()

	// Initialize gRPC server with the broken store.
	lis := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	store := &brokenStore{}
	colmetricspb.RegisterMetricsServiceServer(grpcServer, newServer("bufconn", store))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()
	defer grpcServer.Stop()

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("connecting to grpc server: %v", err)
	}
	defer conn.Close()

	client := colmetricspb.NewMetricsServiceClient(conn)

	// Send a gauge metric via gRPC.
	now := uint64(time.Now().UnixNano())
	_, err = client.Export(ctx, &colmetricspb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{
			{
				ScopeMetrics: []*metricspb.ScopeMetrics{
					{
						Metrics: []*metricspb.Metric{
							{
								Name: "fallback.gauge",
								Data: &metricspb.Metric_Gauge{
									Gauge: &metricspb.Gauge{
										DataPoints: []*metricspb.NumberDataPoint{
											{
												TimeUnixNano: now,
												Value:        &metricspb.NumberDataPoint_AsDouble{AsDouble: 100.5},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("exporting metrics via grpc: %v", err)
	}

	FlushKafka(5000)

	// Create Kafka consumer to verify fallback message.
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          "test-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		t.Fatalf("creating kafka consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Subscribe(topic, nil); err != nil {
		t.Fatalf("subscribing to topic: %v", err)
	}

	msg, err := consumer.ReadMessage(15 * time.Second)
	if err != nil {
		t.Fatalf("reading message from kafka: %v", err)
	}

	var payload struct {
		Type    string     `json:"type"`
		Metrics []GaugeRow `json:"metrics"`
	}

	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		t.Fatalf("unmarshaling kafka message: %v", err)
	}

	if payload.Type != "gauge" {
		t.Errorf("expected type 'gauge', got '%s'", payload.Type)
	}

	if len(payload.Metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(payload.Metrics))
	}

	if payload.Metrics[0].MetricName != "fallback.gauge" {
		t.Errorf("expected metric name 'fallback.gauge', got '%s'", payload.Metrics[0].MetricName)
	}

	if payload.Metrics[0].Value != 100.5 {
		t.Errorf("expected value 100.5, got %f", payload.Metrics[0].Value)
	}
}
