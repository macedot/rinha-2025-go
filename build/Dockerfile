FROM golang:1-alpine AS builder
RUN apk update && rm -rf /var/cache/apk/*
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o api ./cmd/api/
RUN go build -ldflags="-s -w" -o worker ./cmd/worker/

FROM scratch
COPY --from=builder ["/build/api", "/"]
COPY --from=builder ["/build/worker", "/"]
