## 1. Go Server Enhancements

- [x] 1.1 Implement parsing of ClickHouse configuration environment variables (`CLICKHOUSE_ADDR`, `CLICKHOUSE_DB`, `CLICKHOUSE_USER`, `CLICKHOUSE_PASSWORD`) in `server.go`.
- [x] 1.2 Implement ClickHouse client initialization and table creation logic on startup in `server.go` when environment variables are set, and pass the store connection to the gRPC metrics service.

## 2. Docker & Telemetrygen Setup

- [x] 2.1 Create a `Dockerfile` for compiling and packaging the Go metric storage application.
- [x] 2.2 Create a `docker-compose.yaml` to orchestrate ClickHouse, the Go server, and the `telemetrygen` load generator task.
- [x] 2.3 Configure the `telemetrygen` service parameters (e.g. rate, duration, endpoint, payload) in `docker-compose.yaml`.
- [x] 2.4 Create a Python/Bash analysis script at `benchmark/analyze.sh` or `benchmark/analyze.py` to scrape telemetrygen log output and query ClickHouse database for total inserted metrics.

## 3. Tooling & Verification

- [x] 3.1 Update the `Makefile` to include `benchmark-run` and `benchmark-clean` targets.
- [x] 3.2 Run and verify the benchmark, ensuring that the containers orchestrate correctly, telemetrygen executes the load test successfully, and the analysis script accurately displays both client-side and server-side metrics.
