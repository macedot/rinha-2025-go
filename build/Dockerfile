FROM golang:1-alpine AS builder
RUN apk update && rm -rf /var/cache/apk/*
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o app ./cmd/rinha/

FROM scratch
COPY --from=builder ["/build/app", "/"]
EXPOSE 9999 8888 7777
ENTRYPOINT ["/app"]
