## 1. Setup and Dependencies

- [x] 1.1 Add `github.com/hashicorp/golang-lru/v2` dependency to `go.mod`
- [x] 1.2 Run `go mod tidy` to download and verify the new dependency

## 2. Configuration

- [x] 2.1 Add `MetricMetadataCacheSize` configuration parameter in `config.go`
- [x] 2.2 Bind `MetricMetadataCacheSize` to environment variable `METRICS_CACHE_SIZE` with a default value of `50000`

## 3. Implementation of LRU Cache

- [x] 3.1 Initialize `lru.Cache[uuid.UUID, bool]` during initialization in `NewClickHouseMetricsStore`
- [x] 3.2 Refactor the global `metadataCache` variable to reside within `ClickHouseMetricsStore` rather than being a package-level global
- [x] 3.3 Update `checkAndInsertMetadata`, `InsertGauge`, and `InsertSum` in `clickhouse_client.go` to query and update the new LRU cache
- [x] 3.4 Ensure the OTel hit and miss counters (`metadataCacheHits`, `metadataCacheMisses`) are incremented appropriately based on LRU lookup results
- [x] 3.5 Ensure failed batch insertions delete metadata from the LRU cache to allow retry attempts

## 4. Verification and Testing

- [x] 4.1 Write unit tests in `clickhouse_client_test.go` or similar to verify cache eviction under capacity limits
- [x] 4.2 Run tests (`go test ./...`) to ensure ClickHouse and Kafka fallback pipelines continue to operate correctly with the LRU cache
