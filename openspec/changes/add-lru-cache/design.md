## Context

Currently, the `ClickHouseMetricsStore` uses a global `sync.Map` (`metadataCache`) to track which `MetricID`s have been processed. While thread-safe and highly efficient for read operations, `sync.Map` does not support eviction or size bounds. In high-cardinality environments, this cache will grow without bounds, resulting in eventual memory exhaustion (OOM).

## Goals / Non-Goals

**Goals:**
- Bound the maximum memory usage of the metadata cache.
- Implement Least Recently Used (LRU) eviction to keep the most active metric streams cached.
- Ensure the cache is thread-safe and performs well under heavy concurrent write/read operations.
- Make the maximum capacity of the cache configurable.

**Non-Goals:**
- Persist the cache across application restarts.
- Share/distribute the cache across multiple instances of the metric-storage service.

## Decisions

- **Use `github.com/hashicorp/golang-lru/v2`**: We will introduce this library as a dependency. It is the standard thread-safe LRU cache implementation in the Go ecosystem and supports generics. We will use the thread-safe `lru.Cache`.
- **Add Configurable Capacity**: We will introduce a configuration setting `MetricMetadataCacheSize` (bound to environment variable `METRICS_CACHE_SIZE`) with a sensible default of `50000` entries.
- **Integrate with Existing OTel Metrics**: We will preserve the existing OTel counters for cache hits and misses, ensuring that LRU cache hits/misses are tracked seamlessly.

## Risks / Trade-offs

- **Risk**: Lock contention on the LRU cache mutex under high concurrency, as LRU cache operations (even reads) modify a doubly-linked list to update access order.
- **Mitigation**: A default size of 50,000 is small enough that CPU operations within the cache lock are negligible. If lock contention becomes a bottleneck in profiling, we can migrate to a sharded cache design.
