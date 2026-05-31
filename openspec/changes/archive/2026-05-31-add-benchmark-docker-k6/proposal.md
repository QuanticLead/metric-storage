## Why

We currently lack a standard, reproducible way to measure the performance (throughput, latency, resource usage) of the OTLP Metric Storage server. To optimize the storage engine and database schema (especially as we move towards extracting metadata into a lookup table), we need a performance baseline. Running a benchmark via Docker Compose and telemetrygen (the official OpenTelemetry load simulator) will allow us to establish a reliable baseline under simulated load.

## What Changes

- Add a `Dockerfile` for containerizing the Go OTLP Metric Storage application.
- Add a `docker-compose.yaml` configuration to orchestrate the Go application and a ClickHouse database server.
- Add a `telemetrygen` load generator container to simulate realistic metric ingestion over gRPC (no reflection needed).
- Add configuration parameters (via environment variables or flags) to the Go server so it can connect to ClickHouse when running in the container.
- Add an analysis script (e.g., Python/Bash) that queries ClickHouse and parses `telemetrygen` metrics to calculate end-to-end ingestion throughput, latency, and request success rates.
- Add helper scripts or Makefile targets to run and analyze the benchmark results.

## Capabilities

### New Capabilities
- `performance-benchmark`: Defines the performance testing suite, load profiles, target ingestion types, and success criteria for verifying server scalability.

### Modified Capabilities

## Impact

- **Configuration & Deployment**: Adds Docker and Docker Compose files to the project.
- **Go Server App**: Extends flags/environment variables to support ClickHouse connectivity configurations in production/standalone runs.
- **CI/CD & Developer Experience**: Adds automated benchmarking and performance reporting capabilities.
