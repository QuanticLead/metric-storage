# Build stage
FROM golang:1.26-alpine AS builder

# Install build dependencies for confluent-kafka-go
RUN apk add --no-cache build-base librdkafka-dev

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -tags musl -o metric-storage .

# Final stage
FROM alpine:3.19

# Install runtime dependencies for confluent-kafka-go
RUN apk add --no-cache librdkafka

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/metric-storage /app/metric-storage

# Expose gRPC port
EXPOSE 4317

# Run the server
ENTRYPOINT ["/app/metric-storage", "-listenAddr", "0.0.0.0:4317"]
