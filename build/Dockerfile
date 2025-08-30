FROM golang:1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
# ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v2 GOGC=off GOMEMLIMIT=50MiB
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v2
# RUN go build -pgo=storage.pprof -ldflags='-s -w' -o /app/storage ./cmd/storage/main.go
RUN go build -ldflags='-s -w' -o /app/api ./cmd/api/main.go
FROM scratch
WORKDIR /app
COPY --from=builder /app/api .
EXPOSE 9999 8888
CMD [ "/app/api" ]
