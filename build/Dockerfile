FROM golang:1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
# ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v2 GOGC=off GOMEMLIMIT=50MiB
# curl -o rinha.pprof "http://10.4.2.250:8081/debug/pprof/profile?seconds=70"
RUN go build -o /app/rinha ./cmd/rinha/main.go
FROM scratch
WORKDIR /app
COPY --from=builder /app/rinha .
EXPOSE 9999 8888
CMD ["./rinha"]
