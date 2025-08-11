# https://github.com/andersongomes001/rinha-2025

DOCKER_USER := macedot
APP_NAME := rinha-2025
IMAGE_NAME := $(DOCKER_USER)/$(APP_NAME)
VERSION := $(shell git rev-parse --short HEAD)
ENV_FILES := $(wildcard .env.*)

all : $(ENV_FILES) dev run build push prod
.PHONY : all

$(ENV_FILES): FORCE
	@echo "🚀 Running $@"
	docker compose down && \
	docker compose --env-file "$@" up

docker:
	docker compose down && \
	docker compose build

run:
	docker compose down && \
	docker compose up

image:
	@echo "🐳 Build da imagem Docker..."
	docker build -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest .

push:
	@echo "🔐 Enviando imagens..."
	docker push $(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_NAME):latest

prod:
	docker compose down && \
	docker compose -f ./docker-compose-latest.yml up

FORCE: ;