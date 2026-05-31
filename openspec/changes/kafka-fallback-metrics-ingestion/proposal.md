## Why

Currently, if ClickHouse is down or encounters an insertion error, metric data points are logged but dropped. To prevent data loss under database pressure or outages, failed metric ingestion attempts must fallback to publishing the data points into a Kafka topic for offline reprocessing.

## What Changes

- Introduce a Kafka fallback mechanism that intercepts ClickHouse write failures (both synchronous and async worker paths).
- Serialize failed Gauge and Sum metric batches to JSON and publish them to a configured Kafka fallback topic.
- Add configuration parameters for Kafka brokers and the fallback topic.
- Equip the system to gracefully handle scenarios where Kafka itself is unavailable (log error as critical fallback failure).
- Add integration tests using `testcontainers-go` to verify the Kafka fallback write path when ClickHouse writes fail.

## Capabilities

### New Capabilities

*(None)*

### Modified Capabilities

- `metric-storage`: Add the requirement to forward failed metric data points to a Kafka fallback topic on ClickHouse write failure.

## Impact

- **Dependencies**: Add `github.com/confluentinc/confluent-kafka-go/v2` library to `go.mod`.
- **Configuration**: New configuration variables for `KAFKA_BROKERS` and `KAFKA_FALLBACK_TOPIC`.
- **Ingestion Path**: Modified synchronous and worker write loops to dispatch failed batches to Kafka.
- **Testing**: Added a test suite that simulates database write failure and asserts proper Kafka message production.
