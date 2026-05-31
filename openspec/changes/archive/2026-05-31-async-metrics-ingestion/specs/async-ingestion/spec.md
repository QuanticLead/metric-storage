## ADDED Requirements

### Requirement: Central Configuration loading
The system SHALL load all service configuration parameters (including gRPC listen address, max receive message size, ClickHouse connection parameters, buffered channel size, and worker pool size) from environment variables (or fallback flags) during startup.

#### Scenario: Configuration initialization
- **WHEN** the server starts with environment variables or default flag options
- **THEN** the configuration is parsed into a centralized configuration structure with sensible defaults.

### Requirement: Asynchronous Metrics Queueing
The gRPC server handler SHALL convert incoming resource metrics into row formats and queue them onto a thread-safe buffered Go channel. The gRPC server MUST return a success response to the OTLP client immediately once the items are successfully queued.

#### Scenario: Decoupled gRPC response
- **WHEN** OTLP metric payloads are exported to the gRPC service
- **THEN** they are successfully mapped, queued on the channel, and the gRPC call completes successfully without waiting for ClickHouse writes.

### Requirement: Background Worker Processing
The system SHALL spawn a configurable number of background worker goroutines on startup that read from the ingestion channel and execute batch inserts of Gauge and Sum rows to ClickHouse.

#### Scenario: Asynchronous database write
- **WHEN** rows are queued on the channel by the gRPC handler
- **THEN** the background workers extract them from the channel and insert them to ClickHouse in batches.
