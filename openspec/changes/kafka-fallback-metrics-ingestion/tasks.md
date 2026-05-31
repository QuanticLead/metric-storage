## 1. Setup & Configuration

- [x] 1.1 Add `github.com/confluentinc/confluent-kafka-go/v2` dependency to `go.mod`.
- [x] 1.2 Update `config.go` to parse environment variables `KAFKA_BROKERS` and `KAFKA_FALLBACK_TOPIC`.
- [x] 1.3 Add a Kafka client module/initializer in `kafka.go` that sets up a thread-safe `*kafka.Producer`.

## 2. Ingestion Fallback Implementation

- [x] 2.1 Update synchronous fallback write path in `metrics_service.go` to publish failed gauge/sum inserts to Kafka.
- [x] 2.2 Update asynchronous worker write path in `server.go` to publish failed gauge/sum inserts to Kafka.

## 3. Verification & Integration Testing

- [x] 3.1 Create `kafka_fallback_test.go` integration test using `testcontainers-go` to run a Redpanda/Kafka container and mock/cause ClickHouse failures, asserting that failed metrics are published to Kafka in JSON format.
- [x] 3.2 Ensure all tests pass using `go test -tags integration -v ./...`.
