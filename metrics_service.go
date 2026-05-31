package main

import (
	"context"
	"log/slog"

	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
)

type BatchJob struct {
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
	slog.DebugContext(ctx, "Received ExportMetricsServiceRequest")
	metricsReceivedCounter.Add(ctx, 1)

	if m.store != nil {
		rm := request.GetResourceMetrics()
		job := BatchJob{
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
						return nil, err
					}
				}
				if len(job.Sums) > 0 {
					if err := m.store.InsertSum(ctx, job.Sums); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return &colmetricspb.ExportMetricsServiceResponse{}, nil
}
