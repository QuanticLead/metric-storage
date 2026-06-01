# OTLP Metric Storage (Go)

## Development Process (OpenSpec)

This project was built iteratively using an AI-assisted workflow based on OpenSpec. Instead of trying to write the entire application in one go, the development was broken down into a series of focused, verifiable changes (which you can see archived in the `openspec/changes/archive` directory):

1. **Benchmarking Baseline:** First, we built a k6-based benchmark suite running in Docker to establish a performance baseline for the existing ingestion endpoint.
2. **Splitting Metadata:** We tackled the core schema change—separating the primary gauge and sum metrics from their metadata to optimize ClickHouse storage.
3. **Splitting Remaining Metadata:** We followed up by extending the metadata split to cover the remaining metric types.
4. **LRU Cache:** We added an in-memory cache to reduce the load on ClickHouse during ingestion and skip database writes for known metrics.
5. **Async Ingestion:** We moved from synchronous blocking writes to an asynchronous worker pool to handle higher throughput.
6. **Kafka Fallback:** We introduced Kafka to catch failed database inserts, improving overall reliability and preventing data loss.
7. **Observability:** We added OpenTelemetry instrumentation to monitor the service's internal health and performance.
8. **Tuning Benchmarks:** Finally, we refined and tuned our benchmark reports to validate that all our optimizations met the required throughput goals.

Each step was proposed, reviewed, and applied as an independent specification. This allowed us to safely evolve the architecture from a basic synchronous endpoint into a high-throughput, resilient ingestion pipeline.

### Running & Testing

**Start Infrastructure Dependencies:**
```shell
docker-compose up -d
```

**Run Unit Tests:**
```shell
go test -v ./...
```

**Run Integration Tests:**
```shell
go test -tags=integration -v ./...
```

**Run Ingestion Benchmark:**
```shell
make benchmark-run
```

---

## Problem Statement

### Introduction
This take-home assignment is designed to give you an opportunity to demonstrate your skills and experience in
building a small backend application. We expect you to spend 3-4 hours on this assignment (using AI coding agents).
If you find yourself spending more time than that, please stop and submit what you have. We are not looking for a
complete solution, but rather a demonstration of your skills and experience.

To submit your solution, please create a public GitHub repository and send us the link. Please include a `README.md` file
with instructions on how to run your application.

### Overview
The goal of this assignment is to build a simple backend application that receives [metric datapoints](https://opentelemetry.io/docs/concepts/signals/metrics/)
on a gRPC endpoint and processes them, before storing in ClickHouse.
Current state is that we have a gRPC endpoint for receiving metrics, and Gauge and Sum type get correctly converted to
records and inserted into ClickHouse. This is tested with both unit- and integration-tests.

What we're looking for is to extract meta-data about the metrics into a separate table, which will then act as a 'lookup'
table, and that actual data-points just get stored as value + timestamp and with a reference to the lookup table.

Think about and keep in mind the following things:
- How to do the reference between tables?
- How to efficiently store the meta-data in ClickHouse?
- All data should be stored in such a way that full table scans are never needed, under the assumption data always gets queried for a specific time-frame
- Other than time-frame, there are no other mandatory filters for querying
- While you can assume cardinality of the metrics is 'low', e.g. Resources (Attributes) are likely to change over time 

Your solution should take into account high throughput, both in number of messages and the number of metrics / data-points per message.

Feel free to use the existing scaffoling in this folder. Of course, you can also change anything else as you see fit.

### Technology Constraints
- Your Go program should compile using standard Go SDK, and be compatible with Go 1.26.
- Use any additional libraries you want and need.

### Notes
- As this assignment is for the role of a Staff / Senior Product Engineer, we expect you to pay some attention to maintainability and operability of the solution. For example:
  - Consistent terminology usage
  - Validation of the behaviour
  - Include signals / events to help in debugging
- Assume that this application will be deployed to production. Build it accordingly.

### Usage

Build the application:
```shell
go build ./...
```

Run the application:
```shell
go run ./...
```

Run tests
```shell
go test ./...
```

### References

- [OpenTelemetry Metrics](https://opentelemetry.io/docs/concepts/signals/metrics/)
- [OpenTelemetry Protocol (OTLP)](https://github.com/open-telemetry/opentelemetry-proto)
