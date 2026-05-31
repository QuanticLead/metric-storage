## 1. Centralized Configuration

- [x] 1.1 Create `config.go` containing a centralized `Config` struct that loads options (listen address, max message size, clickhouse address, db, user, password, channel size, and worker count) from environment variables and command-line flags.
- [x] 1.2 Update `server.go` to parse the new configuration and initialize the ClickHouse connection using the loaded parameters.

## 2. Decoupled Ingestion

- [x] 2.1 Define the `BatchJob` structure and initialize a buffered Go channel on server startup based on the configured channel size.
- [x] 2.2 Implement the worker loop logic that pulls `BatchJob`s from the channel and inserts them using background context.
- [x] 2.3 Update `server.go` to spawn the configured number of background workers on startup, and implement graceful shutdown.
- [x] 2.4 Update the `Export` method in `metrics_service.go` to process resource metrics, map them into rows, and push them to the ingestion channel asynchronously.

## 3. Tests & Verification

- [x] 3.1 Adjust `integration_test.go` and `server_test.go` to instantiate and test the asynchronous server structure (and wait for data propagation where needed).
- [x] 3.2 Run the unit and integration test suites to verify that the asynchronous pipeline works correctly.
