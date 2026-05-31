# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o metric-storage .

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/metric-storage /app/metric-storage

# Expose gRPC port
EXPOSE 4317

# Run the server
ENTRYPOINT ["/app/metric-storage", "-listenAddr", "0.0.0.0:4317"]
