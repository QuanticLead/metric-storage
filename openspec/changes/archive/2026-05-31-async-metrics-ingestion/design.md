## Context

The gRPC server currently maps and inserts OTLP metric payloads to ClickHouse synchronously within the client request context. To support higher ingestion throughput, lower client response times, and isolate API endpoints from ClickHouse latency, we need to decouple ingestion using a configurable Go channel and background worker pool.

## Goals / Non-Goals

**Goals:**
* Centralize all application configurations (gRPC ports, database details, worker count, channel size) into a single, clean `Config` parsing module.
* Implement a buffered channel to hold mapped metric batches.
* Implement a worker pool that consumes from the channel and batch-writes to ClickHouse concurrently using background context.
* Support graceful shutdown where all workers complete processing the queue before exiting.

**Non-Goals:**
* Implementing persistent disk-backed queuing (e.g. SQLite, disk-buffer).

## Decisions

* ** central Configuration system**:
  Create `config.go` defining a `Config` struct. It will load values from environment variables and support fallback command-line flags.
* **Batch Ingestion Job**:
  Define a struct `BatchJob` to group mapped data-point slices:
  ```go
  type BatchJob struct {
      Gauges []GaugeRow
      Sums   []SumRow
  }
  ```
  A single channel of type `chan BatchJob` will hold incoming tasks.
* **Worker Pool Execution**:
  On startup, spawn `WorkerCount` background worker goroutines. The workers will pull `BatchJob` tasks from the channel and write them to ClickHouse. The database writes will use a background context (`context.Background()`) rather than the short-lived gRPC request context to prevent write cancellations.
* **Graceful Shutdown**:
  When shutting down, the server will close the ingestion channel, waiting for all background workers to finish draining it via a `sync.WaitGroup` before terminating.

## Risks / Trade-offs

* **[Risk]** Memory footprint grows if ClickHouse ingestion lags and the channel fills up. → **[Mitigation]** The channel is bounded by a configurable size. If the channel is full, the gRPC handler will block, acting as backpressure.
