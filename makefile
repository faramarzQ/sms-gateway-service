.PHONY: \
	help api classifier consumer swagger \
	fmt tidy lint lint-fix test check clean

API_IMAGE=sms-gateway-api
CLASSIFIER_IMAGE=sms-gateway-traffic-classifier
CONSUMER_IMAGE=sms-gateway-message-consumer

help:
	@echo "Available commands:"
	@echo ""
	@echo "Applications:"
	@echo "  make api                      Run HTTP API"
	@echo "  make classifier               Run Traffic Classifier"
	@echo "  make consumer                 Run Message Consumer"
	@echo ""
	@echo "Documentation:"
	@echo "  make swagger                  Generate Swagger documentation"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up                Start infrastructure"
	@echo "  make docker-down              Stop infrastructure"
	@echo "  make docker-logs              Show infrastructure logs"
	@echo "  make docker-build-api         Build API Docker image"
	@echo "  make docker-build-classifier  Build Traffic Classifier Docker image"
	@echo "  make docker-build-consumer    Build Message Consumer Docker image"
	@echo "  make docker-build-all         Build all Docker images"
	@echo ""
	@echo "Development:"
	@echo "  make fmt                      Format Go code"
	@echo "  make tidy                     Run go mod tidy"
	@echo "  make lint                     Run golangci-lint"
	@echo "  make lint-fix                 Run golangci-lint with automatic fixes"
	@echo "  make test                     Run tests"
	@echo "  make check                    Run fmt, lint and tests"
	@echo "  make clean                    Clean Go build cache"

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

docker-logs:
	docker compose -f deployments/docker-compose.yml logs -f

fmt:
	go fmt ./...

tidy:
	go mod tidy

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

test:
	go test ./...

check: fmt lint test

clean:
	go clean

docker-build-api:
	docker build \
		-f deployments/docker/api.Dockerfile \
		-t $(API_IMAGE) .

docker-build-classifier:
	docker build \
		-f deployments/docker/traffic-classifier.Dockerfile \
		-t $(CLASSIFIER_IMAGE) .

docker-build-consumer:
	docker build \
		-f deployments/docker/message-consumer.Dockerfile \
		-t $(CONSUMER_IMAGE) .

docker-build-all: docker-build-api docker-build-classifier docker-build-consumer