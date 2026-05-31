## Context

The application initializes the OpenTelemetry SDK (via `otel.go`) and instruments incoming gRPC routes. However, internal operations (like the asynchronous worker batch inserts, the ingestion queue, and ClickHouse TCP calls) are uninstrumented. This leaves a gap in observability, preventing us from tracing a payload's lifecycle from gRPC to ClickHouse disk, and monitoring queue backpressure metrics.

## Goals / Non-Goals

**Goals:**
* Add custom trace span instrumentation to the gRPC `Export` handler.
* Add trace spans to background workers to track batch consumption, using a detached context to avoid write cancellations.
* Propagate trace context to ClickHouse client insertions.
* Implement performance metrics:
  * `com.dash0.homeexercise.ingestion.queue_length`: An observable gauge measuring queue capacity buffer.
  * `com.dash0.homeexercise.workers.active`: An UpDownCounter tracking concurrent active workers.
  * `com.dash0.homeexercise.metadata_cache.hits` and `com.dash0.homeexercise.metadata_cache.misses`: Counters tracking in-memory metadata de-duplication cache performance.

**Non-Goals:**
* Introducing new spans for unit tests where the tracer provider is mock/stdout.

## Decisions

* **Global Tracer & Meter**:
  Define a package-level tracer in `server.go` initialized via `otel.Tracer("dash0.com/otlp-log-processor-backend")`.
* **Span Instrumentation & Propagation**:
  * gRPC `Export` handler will wrap mapping and queueing in an `"Export"` span.
  * We will add a `Ctx context.Context` field to the `BatchJob` structure.
  * Worker goroutines will extract the parent `SpanContext` from the job context using `trace.SpanContextFromContext(job.Ctx)`, inject it into a clean `context.Background()`, and start a `"WorkerInsertBatch"` span. This creates a child span linked to the request trace while shielding database queries from HTTP/gRPC deadline cancellations.
  * All log statements in background workers will use `slog.ErrorContext` or `slog.DebugContext` passing the active worker span context to correlate logs and trace spans.
* **Additional Metrics**:
  * Declare `workersActive` as an UpDownCounter, incremented at the start of a worker job and decremented at completion.
  * Declare `metadataCacheHits` and `metadataCacheMisses` as direct counters, incremented in `checkAndInsertMetadata` according to cache results.
  * Register `queueLengthGauge` as a callback-driven observable gauge reading `len(ingestionChannel)`.

## Risks / Trade-offs

*(None)*
