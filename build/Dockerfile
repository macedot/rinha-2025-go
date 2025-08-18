FROM golang:1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v2 GOGC=off GOMEMLIMIT=50MiB
# ENV CGO_ENABLED=0 GOMEMLIMIT=50MiB
# RUN go build -pgo=gateway.pprof -ldflags='-s -w' -o /app/gateway ./cmd/gateway/main.go
# RUN go build -pgo=service.pprof -ldflags='-s -w' -o /app/service ./cmd/service/main.go
# RUN go build -pgo=storage.pprof -ldflags='-s -w' -o /app/storage ./cmd/storage/main.go
RUN go build -ldflags='-s -w' -o /app/gateway ./cmd/gateway/main.go
RUN go build -ldflags='-s -w' -o /app/service ./cmd/service/main.go
RUN go build -ldflags='-s -w' -o /app/storage ./cmd/storage/main.go
FROM scratch
WORKDIR /app
COPY --from=builder /app/gateway .
COPY --from=builder /app/service .
COPY --from=builder /app/storage .
EXPOSE 9999 8888
