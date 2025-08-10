# https://github.com/andersongomes001/rinha-2025

DOCKER_USER := macedot
APP_NAME := rinha-2025
IMAGE_NAME := $(DOCKER_USER)/$(APP_NAME)
VERSION := $(shell git rev-parse --short HEAD)

dev:
	docker compose down && \
	docker compose build && \
	docker compose up

run:
	docker compose down && \
	docker compose up

build:
	@echo "🐳 Build da imagem Docker..."
	docker build -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest .

push:
	@echo "🔐 Enviando imagens..."
	docker push $(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_NAME):latest

prod:
	docker compose down && \
	docker compose -f ./docker-compose-latest.yml up

.PHONY: dev build push prod
