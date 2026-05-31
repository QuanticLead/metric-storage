## Context

Currently, the Go OTLP metric storage server is only tested using unit tests and integration tests via `testcontainers-go`. There is no performance benchmarking setup to verify how the server scales under load or to establish a baseline for throughput and latency. We want a reproducible performance test environment using Docker Compose and telemetrygen.

## Goals / Non-Goals

**Goals:**
- Containerize the Go server with a Dockerfile.
- Configure a Docker Compose setup to run the Go server, ClickHouse, and telemetrygen.
- Configure the Go server to dynamically connect to a ClickHouse instance when environment variables are supplied.
- Provide a load testing setup using telemetrygen that generates OTLP metrics over gRPC without needing reflection.
- Provide an analysis script to pull both telemetrygen run statistics (client-side) and ClickHouse DB query statistics (server/storage-side) to build a comprehensive baseline.
- Create helper commands (e.g., via `Makefile`) to run and teardown the benchmark easily.

**Non-Goals:**
- Deploying the containerized application to Kubernetes or cloud providers.
- Benchmarking other signals (traces, logs).
- Implementing advanced auth/security features for the benchmark.

## Decisions

### 1. Load Simulation Tool
- **Option A**: Use k6 with statically loaded protobuf definitions.
- **Option B (Chosen)**: Use `telemetrygen` (from `open-telemetry/opentelemetry-collector-contrib`).
- **Rationale**: `telemetrygen` is the standard tool maintained by the OpenTelemetry project. It generates compliant gRPC OTLP load natively without requiring custom Javascript scripts, protobuf maintenance, or runtime reflection.

### 2. Go Server ClickHouse Connectivity
- **Option A**: Build a separate benchmark executable.
- **Option B (Chosen)**: Update `server.go` to connect to ClickHouse if `CLICKHOUSE_ADDR` environment variable is defined.
- **Rationale**: Modifying the existing server enables running it in production-like containers easily without code duplication, keeping the codebase unified.

### 3. Telemetry Analysis and Baseline Verification
- **Option A**: Parse Go server console stdout logs only.
- **Option B**: Query ClickHouse db insertion records only.
- **Option C (Chosen)**: Parse `telemetrygen` exit logs (for client-side throughput/latency/error rates) AND query ClickHouse database (for server-side ingestion count/rate).
- **Rationale**: Combining both sources gives a complete end-to-end view of the system performance, validating that metrics sent by the client were successfully processed and written to disk.

### 4. Docker Compose Orchestration
- **Services**:
  - `clickhouse`: `clickhouse/clickhouse-server:26.2` (matching integration tests).
  - `server`: Local build of Go app.
  - `telemetrygen`: Runs a short-lived load-generation task against `server:4317` after the server starts.
- **Synchronization**: `server` waits for `clickhouse` healthcheck to pass, and `telemetrygen` runs after `server` port `4317` is active.

## Risks / Trade-offs

- **[Risk]** ClickHouse takes longer to boot up than the Go server expects.
  - *Mitigation*: Configure a robust ClickHouse healthcheck in Docker Compose using `clickhouse-client --query "SELECT 1"` and make the server service depend on it with `service_healthy`.
- **[Risk]** `telemetrygen` finishes running quickly and exits, requiring a script to capture logs before the container shuts down.
  - *Mitigation*: Run Docker Compose with run/up flags that capture the stdout logs of the run, or write a script that runs the services and collects logs post-execution.
