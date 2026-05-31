## ADDED Requirements

### Requirement: Performance Benchmarking Suite
The performance benchmarking suite SHALL orchestrate the Go metrics processor server and a ClickHouse database using Docker Compose, and generate load using a `telemetrygen` container. The `telemetrygen` container MUST send gRPC OTLP Export requests containing Gauge and Sum metrics.

#### Scenario: Running benchmark successfully
- **WHEN** the benchmark orchestration is started via Docker Compose and `telemetrygen` executes the OTLP gRPC load simulation
- **THEN** the OTLP server ingests Gauge and Sum metric data, inserts them into ClickHouse, and the suite aggregates both `telemetrygen` client stats and ClickHouse query stats to report 100% request success rate, end-to-end ingestion throughput, and database insertion rate.

### Requirement: OTLP Server ClickHouse Connection Configuration
The Go OTLP metric storage server SHALL accept database connection configurations (including database address, database name, username, and password) via environment variables or command line flags, and initialize the ClickHouse connection on startup.

#### Scenario: Startup with ClickHouse environment variables
- **WHEN** the Go server is launched with environment variables `CLICKHOUSE_ADDR`, `CLICKHOUSE_DB`, `CLICKHOUSE_USER`, and `CLICKHOUSE_PASSWORD`
- **THEN** the server connects to the ClickHouse server on the given address, automatically creates the required metrics tables, and is ready to process and store OTLP gRPC metrics.
