## ADDED Requirements

### Requirement: Kafka Ingestion Fallback
When a metric ingestion batch fails to be committed to ClickHouse, the system SHALL publish the failed metric data points to a designated Kafka fallback topic to prevent data loss.

#### Scenario: Fallback publishing on ClickHouse gauge insert error
- **WHEN** ingestion of a batch of Gauge metrics fails due to a ClickHouse write error
- **THEN** the system serializes the batch to JSON format and publishes it to the configured Kafka topic.

#### Scenario: Fallback publishing on ClickHouse sum insert error
- **WHEN** ingestion of a batch of Sum metrics fails due to a ClickHouse write error
- **THEN** the system serializes the batch to JSON format and publishes it to the configured Kafka topic.
