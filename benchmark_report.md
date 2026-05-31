# Benchmark Ingestion Report

## Run 1: Baseline Solution (Wide Tables)

### Client Ingestion Stats
* **Gauge Metrics Sent:** 9,034
* **Sum Metrics Sent:** 8,982
* **Total Metrics Sent:** 18,016
* **Duration:** 10s

### Database (ClickHouse) Storage Stats
* **Gauges Stored:** 9,034
* **Sums Stored:** 8,982
* **Histograms Stored:** 0
* **Total Records Stored:** 18,016

### Performance Summary
* **Ingestion Success Rate:** 100%
* **Database Write Rate:** 1,801 records/second

## Run 2: Split Metadata Solution (Cached Lookups)

### Client Ingestion Stats
* **Gauge Metrics Sent:** 9,000
* **Sum Metrics Sent:** 9,004
* **Total Metrics Sent:** 18,004
* **Duration:** 10s

### Database (ClickHouse) Storage Stats
* **Gauges Stored:** 9,000
* **Sums Stored:** 9,004
* **Histograms Stored:** 0
* **Metadata Stored:** 1
* **Total Records Stored:** 18,004

### Performance Summary
* **Ingestion Success Rate:** 100%
* **Database Write Rate:** 1,800 records/second

