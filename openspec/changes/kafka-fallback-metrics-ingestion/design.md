## Context

If ClickHouse is unavailable or encounters a write error, the system currently drops the metric data points after logging the error. This change introduces a fallback publishing mechanism to a Kafka topic so that another service can process them.

## Goals / Non-Goals

**Goals:**
- Add Kafka client library support using `github.com/confluentinc/confluent-kafka-go/v2`.
- Configure Kafka connection details via environment variables (`KAFKA_BROKERS` and `KAFKA_FALLBACK_TOPIC`).
- Intercept ClickHouse gauge/sum insertion errors in `metrics_service.go` and `server.go` (workers).
- Serialize failed metric data points to JSON and publish them to the Kafka fallback topic.
- Add an integration test simulating ClickHouse failures and verifying that metrics are successfully published to Kafka.

**Non-Goals:**
- Retrying ClickHouse writes within this service.
- Reading or consuming the fallback topic within this service (this is handled by another service).

## Decisions

### 1. Kafka Client Library
- **Option A (Chosen)**: `github.com/confluentinc/confluent-kafka-go/v2` (requires cgo and native `librdkafka`).
- **Option B**: `segmentio/kafka-go`.
- **Rationale**: The confluent-kafka-go library is the official, high-performance client wrapper around the robust, industry-standard librdkafka client. It is chosen for its superior reliability and full compatibility with enterprise Kafka deployments.

### 2. Message Format
- **Decision**: Serialize the raw `GaugeRow` and `SumRow` batches to JSON.
- **Rationale**: JSON is human-readable, easy to debug, and universally supported by any secondary service that will reprocess the fallback queue.

### 3. Kafka Producer Lifecycle
- **Decision**: Initialize a single global/shared `*kafka.Producer` during server startup if `KAFKA_BROKERS` is specified. Reuse this producer for all fallback attempts, and close it cleanly on server shutdown.
- **Rationale**: Maintaining a single `*kafka.Producer` instance allows connection reuse, internal buffering, and optimal throughput, preventing connection leaks.

## Risks / Trade-offs

- **[Risk]** Kafka failure during fallback.
  - *Mitigation*: If Kafka itself is unavailable or errors out during a fallback attempt, the system will log a critical error. The priority is to avoid blocking the main ingestion path when both ClickHouse and Kafka are down.
- **[Risk]** Serialization overhead.
  - *Mitigation*: JSON serialization is fast enough for the fallback path, which is only activated on failure (exceptional cases).
