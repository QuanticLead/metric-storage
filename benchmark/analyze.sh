#!/usr/bin/env bash
set -euo pipefail

# Give ClickHouse a moment to process any remaining batch buffers
sleep 3

echo "=================================================="
echo "          BENCHMARK ANALYSIS REPORT               "
echo "=================================================="
echo ""

# Get telemetrygen-gauge logs
LOGS_GAUGE=$(docker compose logs telemetrygen-gauge 2>&1 || true)
# Get telemetrygen-sum logs
LOGS_SUM=$(docker compose logs telemetrygen-sum 2>&1 || true)

extract_sent() {
    local logs="$1"
    local sent=$(echo "$logs" | grep -oE '"metrics":\s*[0-9]+' | head -n 1 | awk -F: '{print $2}' | tr -d ' ' || echo "0")
    if [ "$sent" = "0" ] || [ -z "$sent" ]; then
        sent=$(echo "$logs" | grep -oE 'metrics generated.*' | grep -oE '[0-9]+' | head -n 1 || echo "0")
    fi
    echo "$sent"
}

SENT_GAUGE=$(extract_sent "$LOGS_GAUGE")
SENT_SUM=$(extract_sent "$LOGS_SUM")
METRICS_SENT=$((SENT_GAUGE + SENT_SUM))

echo "Client (telemetrygen) Statistics:"
echo "--------------------------------"
echo "Gauge Metrics Sent:  $SENT_GAUGE"
echo "Sum Metrics Sent:    $SENT_SUM"
echo "Total Metrics Sent:  $METRICS_SENT"
echo ""

# Query ClickHouse
echo "Database (ClickHouse) Statistics:"
echo "--------------------------------"

# Run query helper
query_clickhouse() {
    docker compose exec -T clickhouse clickhouse-client --user default --password test --database default --query "$1" 2>/dev/null || echo "0"
}

GAUGE_COUNT=$(query_clickhouse "SELECT count() FROM otel_metrics_gauge")
SUM_COUNT=$(query_clickhouse "SELECT count() FROM otel_metrics_sum")
HISTOGRAM_COUNT=$(query_clickhouse "SELECT count() FROM otel_metrics_histogram")
METADATA_COUNT=$(query_clickhouse "SELECT count() FROM otel_metrics_metadata")
TOTAL_DB_RECORDS=$((GAUGE_COUNT + SUM_COUNT + HISTOGRAM_COUNT))

echo "Gauges Stored:     $GAUGE_COUNT"
echo "Sums Stored:       $SUM_COUNT"
echo "Histograms Stored: $HISTOGRAM_COUNT"
echo "Metadata Stored:   $METADATA_COUNT"
echo "Total Stored:      $TOTAL_DB_RECORDS"
echo ""

# Calculate stats
echo "Performance Summary:"
echo "--------------------"
if [ "$METRICS_SENT" -gt 0 ]; then
    SUCCESS_PERCENT=$(( 100 * TOTAL_DB_RECORDS / METRICS_SENT ))
    echo "Ingestion Success Rate: $SUCCESS_PERCENT%"
else
    echo "Ingestion Success Rate: N/A"
fi

# Query DB execution duration
DB_DURATION_SEC=$(query_clickhouse "SELECT dateDiff('second', min(TimeUnix), max(TimeUnix)) FROM (SELECT TimeUnix FROM otel_metrics_gauge UNION ALL SELECT TimeUnix FROM otel_metrics_sum UNION ALL SELECT TimeUnix FROM otel_metrics_histogram)")
# Fallback to a minimum of 1s to prevent division by zero
if [ -z "$DB_DURATION_SEC" ] || [ "$DB_DURATION_SEC" -le 0 ]; then
    DB_DURATION_SEC=1
fi

DB_RATE=$(( TOTAL_DB_RECORDS / DB_DURATION_SEC ))
echo "Database Write Rate:    $DB_RATE records/second"
echo "=================================================="
