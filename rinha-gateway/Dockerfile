FROM golang:1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v2 GOGC=off GOMEMLIMIT=45MiB
RUN go build -ldflags='-s -w' -o rinha-gateway .
FROM scratch
WORKDIR /app
COPY --from=builder /app/rinha-gateway .
EXPOSE 9999
CMD ["./rinha-gateway"]
