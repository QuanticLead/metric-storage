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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
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
	tracer   trace.Tracer  = otel.Tracer(name)
	logger                 = otelslog.NewLogger(name)
	metricsReceivedCounter metric.Int64Counter

	workersActive       metric.Int64UpDownCounter
	queueLengthGauge    metric.Int64ObservableGauge
	metadataCacheHits   metric.Int64Counter
	metadataCacheMisses    metric.Int64Counter
	ingestionErrorsCounter metric.Int64Counter
	batchSizeHistogram     metric.Int64Histogram
)

func init() {
	var err error
	metricsReceivedCounter, err = meter.Int64Counter("com.dash0.homeexercise.metrics.received",
		metric.WithDescription("The number of metrics received by otlp-metrics-processor-backend"),
		metric.WithUnit("{metric}"))
	if err != nil {
		panic(err)
	}

	workersActive, err = meter.Int64UpDownCounter("com.dash0.homeexercise.workers.active",
		metric.WithDescription("The number of background workers actively processing jobs"),
		metric.WithUnit("{worker}"))
	if err != nil {
		panic(err)
	}

	queueLengthGauge, err = meter.Int64ObservableGauge("com.dash0.homeexercise.ingestion.queue_length",
		metric.WithDescription("The current ingestion queue size"),
		metric.WithUnit("{job}"),
		metric.WithInt64Callback(func(ctx context.Context, obs metric.Int64Observer) error {
			if ingestionChannel != nil {
				obs.Observe(int64(len(ingestionChannel)))
			}
			return nil
		}))
	if err != nil {
		panic(err)
	}

	metadataCacheHits, err = meter.Int64Counter("com.dash0.homeexercise.metadata_cache.hits",
		metric.WithDescription("The number of metadata cache hits"),
		metric.WithUnit("{hit}"))
	if err != nil {
		panic(err)
	}

	metadataCacheMisses, err = meter.Int64Counter("com.dash0.homeexercise.metadata_cache.misses",
		metric.WithDescription("The number of metadata cache misses"),
		metric.WithUnit("{miss}"))
	if err != nil {
		panic(err)
	}

	ingestionErrorsCounter, err = meter.Int64Counter("com.dash0.homeexercise.ingestion.errors",
		metric.WithDescription("The number of failed ingestion attempts"),
		metric.WithUnit("{error}"))
	if err != nil {
		panic(err)
	}

	batchSizeHistogram, err = meter.Int64Histogram("com.dash0.homeexercise.ingestion.batch_size",
		metric.WithDescription("The size of the batches processed by workers"),
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

	// Initialize Kafka fallback.
	if err := InitKafka(cfg); err != nil {
		return fmt.Errorf("initializing kafka fallback: %w", err)
	}
	defer CloseKafka()

	var store MetricsStore
	if cfg.ClickHouseAddr != "" {
		slog.Info("Connecting to ClickHouse", slog.String("addr", cfg.ClickHouseAddr), slog.String("db", cfg.ClickHouseDB))
		clickhouseStore, err := NewClickHouseMetricsStore(context.Background(), cfg.ClickHouseAddr, cfg.ClickHouseDB, cfg.ClickHouseUser, cfg.ClickHousePassword, cfg.MetricMetadataCacheSize)
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
			slog.DebugContext(context.Background(), "Background worker started", slog.Int("id", workerID))
			for job := range ingestionChannel {
				func(job BatchJob) {
					spanCtx := trace.SpanContextFromContext(job.Ctx)
					ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)
					workerCtx, span := tracer.Start(ctx, "WorkerInsertBatch", trace.WithAttributes(
						attribute.Int("worker.id", workerID),
						attribute.Int("gauges.count", len(job.Gauges)),
						attribute.Int("sums.count", len(job.Sums)),
					))
					defer span.End()

					workersActive.Add(workerCtx, 1)
					defer workersActive.Add(workerCtx, -1)

					if store != nil {
						if len(job.Gauges) > 0 {
							batchSizeHistogram.Record(workerCtx, int64(len(job.Gauges)), metric.WithAttributes(attribute.String("type", "gauge")))
							if err := store.InsertGauge(workerCtx, job.Gauges); err != nil {
								ingestionErrorsCounter.Add(workerCtx, 1, metric.WithAttributes(attribute.String("destination", "clickhouse"), attribute.String("type", "gauge")))
								span.RecordError(err)
								span.SetStatus(codes.Error, err.Error())
								slog.ErrorContext(workerCtx, "Worker failed to insert gauges, attempting Kafka fallback", slog.Int("id", workerID), slog.Int("batch_size", len(job.Gauges)), "error", err)
								if fallbackErr := PublishFallbackMetrics(workerCtx, "gauge", job.Gauges); fallbackErr != nil {
									ingestionErrorsCounter.Add(workerCtx, 1, metric.WithAttributes(attribute.String("destination", "kafka"), attribute.String("type", "gauge")))
									slog.ErrorContext(workerCtx, "Kafka fallback for gauges failed", slog.Int("id", workerID), slog.Int("batch_size", len(job.Gauges)), "error", fallbackErr)
								}
							}
						}
						if len(job.Sums) > 0 {
							batchSizeHistogram.Record(workerCtx, int64(len(job.Sums)), metric.WithAttributes(attribute.String("type", "sum")))
							if err := store.InsertSum(workerCtx, job.Sums); err != nil {
								ingestionErrorsCounter.Add(workerCtx, 1, metric.WithAttributes(attribute.String("destination", "clickhouse"), attribute.String("type", "sum")))
								span.RecordError(err)
								span.SetStatus(codes.Error, err.Error())
								slog.ErrorContext(workerCtx, "Worker failed to insert sums, attempting Kafka fallback", slog.Int("id", workerID), slog.Int("batch_size", len(job.Sums)), "error", err)
								if fallbackErr := PublishFallbackMetrics(workerCtx, "sum", job.Sums); fallbackErr != nil {
									ingestionErrorsCounter.Add(workerCtx, 1, metric.WithAttributes(attribute.String("destination", "kafka"), attribute.String("type", "sum")))
									slog.ErrorContext(workerCtx, "Kafka fallback for sums failed", slog.Int("id", workerID), slog.Int("batch_size", len(job.Sums)), "error", fallbackErr)
								}
							}
						}
					}
				}(job)
			}
			slog.DebugContext(context.Background(), "Background worker stopped", slog.Int("id", workerID))
		}(i)
	}
}
