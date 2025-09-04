# Makefile for rinha-2025-go

# Variables
APP_NAME := rinha-2025-go
APP_PORT = 9999
CLIENT_COMPOSE_FILE = ./test/payment-processor/docker-compose.yml
COMPOSE_FILE = ./build/docker-compose.yml
CURL = curl -s -w "\nHTTP Status: %{http_code}\n"
DOCKER_USER := macedot
IMAGE_NAME := $(DOCKER_USER)/$(APP_NAME)
VERSION := $(shell git rev-parse --short HEAD)

# Default target
.PHONY: all
all: build up logs

# Build the Docker images
.PHONY: build
build:
	@echo "Building Docker images..."
	docker-compose -f $(COMPOSE_FILE) build

# Start the services
.PHONY: up
up:
	@echo "Starting services..."
	docker-compose -f $(COMPOSE_FILE) up -d

# Stop the services
.PHONY: down
down:
	@echo "Stopping services..."
	docker-compose -f $(COMPOSE_FILE) down

# Remove containers, images, and volumes
.PHONY: clean
clean:
	@echo "Cleaning up containers, images, and volumes..."
	docker-compose -f $(COMPOSE_FILE) down --rmi all -v
	rm -rf payments/*
	rm -rf prometheus/*

# View logs for all services
.PHONY: logs
logs:
	@echo "Displaying logs..."
	docker-compose -f $(COMPOSE_FILE) logs -f

# Build the Docker image
.PHONY: image
image:
	@echo "🐳 Build da imagem Docker..."
	docker build -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest .

# Push the Docker image to Docker Hub
.PHONY: push
push:
	@echo "🔐 Enviando imagens..."
	docker push $(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_NAME):latest

# Test the POST /payments endpoint
.PHONY: test-payment
test-payment:
	@echo "Testing POST /payments..."
	$(CURL) -X POST http://localhost:$(APP_PORT)/payments \
		-H "Content-Type: application/json" \
		-d '{"correlationId":"4a7901b8-7d26-4d9d-aa19-4dc1c7cf60b3","amount": 19.90}'

# Test the GET /payments-summary endpoint with optional from/to parameters
.PHONY: test-stats
test-stats:
	@echo "Testing GET /payments-summary..."
	$(CURL) "http://localhost:$(APP_PORT)/payments-summary?from=2020-08-14T00:00:00Z&to=2035-08-14T23:59:59Z"

# Test the GET /payments-summary endpoint without parameters
.PHONY: test-stats-no-params
test-stats-no-params:
	@echo "Testing GET /payments-summary without parameters..."
	$(CURL) http://localhost:$(APP_PORT)/payments-summary

# Test the POST /purge endpoint
.PHONY: test-purge
test-purge:
	@echo "Testing POST /purge..."
	$(CURL) -X POST http://localhost:$(APP_PORT)/purge-payments

# Run all tests
.PHONY: test
test: test-stats test-stats-no-params test-purge test-metrics
	@echo "Skipping test-payment as it requires a running payment-client server"

# Start the client services
.PHONY: client
client:
	@echo "Starting CLIENT services..."
	docker-compose -f $(CLIENT_COMPOSE_FILE) up -d

k6:
	K6_WEB_DASHBOARD_OPEN=false \
	K6_WEB_DASHBOARD=true \
	K6_WEB_DASHBOARD_EXPORT='./test/report.html' \
	K6_WEB_DASHBOARD_PERIOD=2s \
	SUMMARY_FILE=./test/summary/partial-result.json \
	k6 run ./test/rinha.js

k6-1k:
	K6_WEB_DASHBOARD_OPEN=false \
	K6_WEB_DASHBOARD=true \
	K6_WEB_DASHBOARD_EXPORT='./test/report-1k.html' \
	K6_WEB_DASHBOARD_PERIOD=2s \
	SUMMARY_FILE=./test/summary/partial-result-1k.json \
	k6 run -e MAX_REQUESTS=1000 ./test/rinha.js

k6-5k:
	K6_WEB_DASHBOARD_OPEN=false \
	K6_WEB_DASHBOARD=true \
	K6_WEB_DASHBOARD_EXPORT='./test/report-5k.html' \
	K6_WEB_DASHBOARD_PERIOD=2s \
	SUMMARY_FILE=./test/summary/partial-result-5k.json \
	k6 run -e MAX_REQUESTS=5000 ./test/rinha.js

k6-final:
	K6_WEB_DASHBOARD_OPEN=false \
	K6_WEB_DASHBOARD=true \
	K6_WEB_DASHBOARD_EXPORT='./test/report-final.html' \
	K6_WEB_DASHBOARD_PERIOD=2s \
	SUMMARY_FILE=./test/summary/final-result.json \
	k6 run ./test/rinha-final.js
