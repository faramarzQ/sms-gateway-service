.PHONY: help api classifier consumer swagger docker-up docker-down fmt tidy test clean

help:
	@echo "Available commands:"
	@echo "  make api           Run HTTP API"
	@echo "  make classifier    Run Traffic Classifier"
	@echo "  make consumer      Run Message Consumer"
	@echo "  make swagger       Generate Swagger documentation"
	@echo "  make docker-up     Start infrastructure"
	@echo "  make docker-down   Stop infrastructure"
	@echo "  make fmt           Format Go code"
	@echo "  make tidy          Run go mod tidy"
	@echo "  make test          Run tests"

api:
	go run ./cmd/api

classifier:
	go run ./cmd/traffic_classifier

consumer:
	go run ./cmd/message_consumer

swagger:
	swag init -g cmd/api/main.go
	@echo ""
	@echo "Swagger documentation generated successfully."
	@echo "Run 'make api' and open:"
	@echo "http://localhost:8080/swagger/index.html"

docker-up:
	docker compose -f deployments/docker-compose.yml up -d

docker-down:
	docker compose -f deployments/docker-compose.yml down

fmt:
	go fmt ./...

tidy:
	go mod tidy

test:
	go test ./...

clean:
	go clean