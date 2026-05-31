## ADDED Requirements

### Requirement: Trace Instrumentation for Export
The server gRPC Export handler SHALL create an active trace span for each incoming OTLP Export request.

#### Scenario: Export tracing
- **WHEN** an OTLP export request is handled by the Metrics Service
- **THEN** a trace span is created, mapping to the export operation.

### Requirement: Trace Context Propagation for Background Workers
The background workers SHALL extract trace propagation context from the queue job and create active child trace spans without deadlines when processing metric job batches from the channel.

#### Scenario: Background worker tracing
- **WHEN** a background worker pulls a metric job from the channel
- **THEN** it starts a trace span specifying worker information and job sizes, linked as a child span of the original request trace.

### Requirement: ClickHouse Transaction Query Tracing
All database insertions performed by the ClickHouse store client SHALL execute under the active trace context, propagating tracing details to ClickHouse.

#### Scenario: Traced ClickHouse write
- **WHEN** a background worker performs a ClickHouse insert using the active worker trace context
- **THEN** the ClickHouse driver creates child spans representing the query operations.

### Requirement: Ingestion Metric Tracking
The system SHALL track metrics for active worker counts, metadata cache performance, and observe queue sizes via standard OpenTelemetry metrics.

#### Scenario: Metric reporting
- **WHEN** the server is running and metrics are processed or queued
- **THEN**:
  - `com.dash0.homeexercise.workers.active` UpDownCounter tracks active worker processing counts.
  - `com.dash0.homeexercise.ingestion.queue_length` observable gauge returns the current ingestion queue size.
  - `com.dash0.homeexercise.metadata_cache.hits` and `com.dash0.homeexercise.metadata_cache.misses` counters track in-memory metadata de-duplication cache operations.
