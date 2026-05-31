## Why

Currently, the application uses an unbounded in-memory cache (`sync.Map`) to track successfully inserted or resolved `MetricID` values. In a long-running production service with high metric cardinality or dynamically changing attributes (e.g., container IDs, dynamic span attributes mapped to metrics), this unbounded cache will grow indefinitely, eventually causing the Go service to run out of memory (OOM). 

Replacing the unbounded cache with a bounded Least Recently Used (LRU) cache prevents memory exhaustion while retaining the benefits of cache-based metadata deduplication.

## What Changes

- Replace the unbounded `sync.Map` with a thread-safe LRU cache (e.g., using `hashicorp/golang-lru/v2` or similar).
- Make the LRU cache capacity configurable via environment variables/configuration.
- Implement telemetry (metrics) for cache hits, misses, and current size.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `metric-storage`: Refine the **In-Memory Metadata Cache** requirement to specify that the cache must be bounded (e.g., utilizing an LRU eviction strategy) to prevent memory exhaustion, with configurable size constraints.

## Impact

- **Codebase**: `clickhouse_client.go` will be updated to use the new LRU cache implementation instead of `sync.Map`.
- **Configuration**: `config.go` will be updated to load cache configuration properties (e.g., `METRIC_METADATA_CACHE_SIZE`).
- **Dependencies**: Add `github.com/hashicorp/golang-lru/v2` to `go.mod`.
