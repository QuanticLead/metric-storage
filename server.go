package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	listenAddr            = flag.String("listenAddr", "localhost:4317", "The listen address")
	maxReceiveMessageSize = flag.Int("maxReceiveMessageSize", 16777216, "The max message size in bytes the server can receive")
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
	slog.SetDefault(logger)
	logger.Info("Starting application")

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(context.Background())
	if err != nil {
		return
	}

	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	flag.Parse()

	// Parse ClickHouse configuration
	chAddr := os.Getenv("CLICKHOUSE_ADDR")
	var store MetricsStore
	if chAddr != "" {
		chDB := os.Getenv("CLICKHOUSE_DB")
		if chDB == "" {
			chDB = "default"
		}
		chUser := os.Getenv("CLICKHOUSE_USER")
		if chUser == "" {
			chUser = "default"
		}
		chPassword := os.Getenv("CLICKHOUSE_PASSWORD")

		slog.Info("Connecting to ClickHouse", slog.String("addr", chAddr), slog.String("db", chDB))
		clickhouseStore, err := NewClickHouseMetricsStore(context.Background(), chAddr, chDB, chUser, chPassword)
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

	slog.Debug("Starting listener", slog.String("listenAddr", *listenAddr))
	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.MaxRecvMsgSize(*maxReceiveMessageSize),
		grpc.Creds(insecure.NewCredentials()),
	)
	colmetricspb.RegisterMetricsServiceServer(grpcServer, newServer(*listenAddr, store))

	slog.Debug("Starting gRPC server")

	return grpcServer.Serve(listener)
}
