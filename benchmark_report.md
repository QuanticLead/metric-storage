# Benchmark Ingestion Report

## Run 1: Baseline Solution (Wide Tables)

### Client Ingestion Stats
* **Gauge Metrics Sent:** 8,100
* **Sum Metrics Sent:** 8,100
* **Total Metrics Sent:** 16,200
* **Duration:** 10s

### Database (ClickHouse) Storage Stats
* **Gauges Stored:** 8,100
* **Sums Stored:** 8,100
* **Histograms Stored:** 0
* **Total Records Stored:** 16,200

### Table Storage Utilization
| table | rows | compressed | uncompressed |
|:-|-:|:-|:-|
| otel_metrics_gauge | 8100 | 52.30 KiB | 853.49 KiB |
| otel_metrics_sum | 8100 | 52.54 KiB | 956.31 KiB |

### Performance Summary
* **Ingestion Success Rate:** 100%
* **Database Write Rate:** 1,620 records/second

## Run 2: Split Metadata Solution (Cached Lookups)

### Client Ingestion Stats
* **Gauge Metrics Sent:** 8,000
* **Sum Metrics Sent:** 8,000
* **Total Metrics Sent:** 16,000
* **Duration:** 10s

### Database (ClickHouse) Storage Stats
* **Gauges Stored:** 8,000
* **Sums Stored:** 8,000
* **Histograms Stored:** 0
* **Metadata Stored:** 858
* **Total Records Stored:** 16,000

### Table Storage Utilization
| table | rows | compressed | uncompressed |
|:-|-:|:-|:-|
| otel_metrics_gauge | 8000 | 69.50 KiB | 250.09 KiB |
| otel_metrics_metadata | 858 | 17.12 KiB | 90.61 KiB |
| otel_metrics_sum | 8000 | 69.71 KiB | 351.61 KiB |

### Performance Summary
* **Ingestion Success Rate:** 100%
* **Database Write Rate:** 1,600 records/second

## Run 3: Split Metadata + Kafka Fallback (CGO build)

### Client Ingestion Stats
* **Gauge Metrics Sent:** 15,283
* **Sum Metrics Sent:** 15,248
* **Total Metrics Sent:** 30,531
* **Duration:** 10s

### Database (ClickHouse) Storage Stats
* **Gauges Stored:** 15,283
* **Sums Stored:** 15,248
* **Histograms Stored:** 0
* **Metadata Stored:** 1,769
* **Total Records Stored:** 30,531

### Table Storage Utilization
| table | rows | compressed | uncompressed |
|:-|-:|:-|:-|
| otel_metrics_gauge | 15283 | 130.17 KiB | 477.66 KiB |
| otel_metrics_metadata | 1769 | 36.55 KiB | 187.90 KiB |
| otel_metrics_sum | 15248 | 127.72 KiB | 670.10 KiB |

### Performance Summary
* **Ingestion Success Rate:** 100%
* **Database Write Rate:** 3,053 records/second
