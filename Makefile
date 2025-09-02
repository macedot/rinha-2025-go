# https://github.com/andersongomes001/rinha-2025

DOCKER_USER := macedot
APP_NAME := rinha-2025
IMAGE_NAME := $(DOCKER_USER)/$(APP_NAME)
VERSION := $(shell git rev-parse --short HEAD)

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

FORCE: ;

.PHONY : docker run build push prod

ci:
	K6_WEB_DASHBOARD_OPEN=false \
	K6_WEB_DASHBOARD=true \
	K6_WEB_DASHBOARD_EXPORT='report.html' \
	K6_WEB_DASHBOARD_PERIOD=2s \
	k6 run ./test/rinha.js

ci-m:
	K6_WEB_DASHBOARD_OPEN=false \
	K6_WEB_DASHBOARD=true \
	K6_WEB_DASHBOARD_EXPORT='report.html' \
	K6_WEB_DASHBOARD_PERIOD=2s \
	k6 run -e MAX_REQUESTS=4000 ./test/rinha-final.js

ci-f:
	K6_WEB_DASHBOARD_OPEN=false \
	K6_WEB_DASHBOARD=true \
	K6_WEB_DASHBOARD_EXPORT='report.html' \
	K6_WEB_DASHBOARD_PERIOD=2s \
	k6 run ./test/rinha-final.js
