## 1. Trace Instrumentation

- [x] 1.1 Declare package-level `tracer` in `server.go` and import `go.opentelemetry.io/otel/trace` and `go.opentelemetry.io/otel/attribute`.
- [x] 1.2 Add `Ctx context.Context` field to `BatchJob` struct. Update the gRPC `Export` handler in `metrics_service.go` to wrap mapping and channel queueing in an `"Export"` trace span, and assign `job.Ctx = ctx`.
- [x] 1.3 Update background workers in `server.go` to extract `SpanContext` from `job.Ctx`, start a `"WorkerInsertBatch"` trace span using a detached context base, and update log calls (`slog.Debug`, `slog.Error`) to use context-aware variants (`slog.DebugContext`, `slog.ErrorContext`).
- [x] 1.4 Pass the worker trace context to `store.InsertGauge` and `store.InsertSum` invocations in the background workers.

## 2. Ingestion Metric Tracking

- [x] 2.1 Declare performance metrics in `server.go`:
  - `workersActive` (UpDownCounter)
  - `queueLengthGauge` (Observable Gauge)
  - `metadataCacheHits` (Counter)
  - `metadataCacheMisses` (Counter)
- [x] 2.2 Initialize the metrics inside the OpenTelemetry meter registry on startup. Update `queueLengthGauge` to safely verify `ingestionChannel != nil`.
- [x] 2.3 Update background workers in `server.go` to increment `workersActive` on job start and decrement it on completion.
- [x] 2.4 Update `checkAndInsertMetadata` in `clickhouse_client.go` to increment `metadataCacheHits` on cache hits and `metadataCacheMisses` on cache misses.

## 3. Tests & Verification

- [x] 3.1 Run unit and integration tests to verify all tests compile and pass successfully with the updated tracing and metrics.
