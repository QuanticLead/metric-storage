MODULE := dash0.com/otlp-log-processor-backend

.PHONY: build run test test-integration test-all fmt vet lint tidy clean benchmark-run benchmark-clean

build:
	go build ./...

run:
	go run .

test:
	go test -count=1 ./...

test-integration:
	go test -tags integration -count=1 -v ./...

test-all: test test-integration

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: vet
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed, skipping (go install honnef.co/go/tools/cmd/staticcheck@latest)"; \
	fi

tidy:
	go mod tidy

clean:
	go clean ./...

benchmark-run:
	docker compose down -v
	docker compose up -d --build clickhouse server
	docker compose up telemetrygen-gauge telemetrygen-sum
	./benchmark/analyze.sh

benchmark-clean:
	docker compose down -v
