package main

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
)

type BatchJob struct {
	Ctx    context.Context
	Gauges []GaugeRow
	Sums   []SumRow
}

type dash0MetricsServiceServer struct {
	addr  string
	store MetricsStore

	colmetricspb.UnimplementedMetricsServiceServer
}

func newServer(addr string, store MetricsStore) colmetricspb.MetricsServiceServer {
	return &dash0MetricsServiceServer{addr: addr, store: store}
}

func (m *dash0MetricsServiceServer) Export(ctx context.Context, request *colmetricspb.ExportMetricsServiceRequest) (*colmetricspb.ExportMetricsServiceResponse, error) {
	ctx, span := tracer.Start(ctx, "Export")
	defer span.End()

	slog.DebugContext(ctx, "Received ExportMetricsServiceRequest")
	metricsReceivedCounter.Add(ctx, 1)

	if m.store != nil {
		rm := request.GetResourceMetrics()
		job := BatchJob{
			Ctx:    ctx,
			Gauges: MapGaugeRows(rm),
			Sums:   MapSumRows(rm),
		}

		if len(job.Gauges) > 0 || len(job.Sums) > 0 {
			if ingestionChannel != nil {
				ingestionChannel <- job
			} else {
				// Fallback to synchronous insert if channel is not initialized (e.g. in tests)
				if len(job.Gauges) > 0 {
					if err := m.store.InsertGauge(ctx, job.Gauges); err != nil {
						slog.ErrorContext(ctx, "Synchronous InsertGauge failed, attempting Kafka fallback", slog.Int("batch_size", len(job.Gauges)), "error", err)
						if fallbackErr := PublishFallbackMetrics(ctx, "gauge", job.Gauges); fallbackErr != nil {
							span.RecordError(err)
							span.SetStatus(codes.Error, err.Error())
							return nil, err
						}
					}
				}
				if len(job.Sums) > 0 {
					if err := m.store.InsertSum(ctx, job.Sums); err != nil {
						slog.ErrorContext(ctx, "Synchronous InsertSum failed, attempting Kafka fallback", slog.Int("batch_size", len(job.Sums)), "error", err)
						if fallbackErr := PublishFallbackMetrics(ctx, "sum", job.Sums); fallbackErr != nil {
							span.RecordError(err)
							span.SetStatus(codes.Error, err.Error())
							return nil, err
						}
					}
				}
			}
		}
	}

	return &colmetricspb.ExportMetricsServiceResponse{}, nil
}
