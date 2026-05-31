## Why

Currently, the gRPC `Export` service handler synchronously maps and inserts incoming OTLP metric payloads to ClickHouse. Under high load, this increases client response times and couples endpoint availability to database latency. Furthermore, the application lacks a centralized configuration system, with flags and env variables parsed in an ad-hoc manner in `server.go`.

## What Changes

* **Centralized Configuration System**: Create a `Config` structure to manage all application parameters (gRPC settings, ClickHouse connection details, ingestion channel size, and background worker count) with sensible defaults, parsed from environment variables and command-line flags.
* **Asynchronous Ingestion Channel**: Introduce a configurable buffered Go channel to hold incoming Gauge and Sum data-point rows, enabling the gRPC server to respond immediately to clients once payload processing and mapping succeeds.
* **Background Worker Pool**: Spawn a configurable number of background worker goroutines to pull data-point batches from the channel and execute bulk inserts into ClickHouse.

## Capabilities

### New Capabilities

* `async-ingestion`: The OTLP Metrics Store SHALL decouple gRPC request handling from database writes via a configurable buffered channel and background worker pool.

### Modified Capabilities

*(None)*

## Impact

* `config.go` (new file for config parsing)
* `server.go` (main entrypoint, service config, and worker initialization)
* `metrics_service.go` (asynchronous write queueing)
* `clickhouse_client.go` (handling concurrent inserts from worker pool)
