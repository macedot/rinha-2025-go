# Makefile for rinha-2025-rpc

# Variables
APP_NAME := rinha-2025
APP_PORT = 9999
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
