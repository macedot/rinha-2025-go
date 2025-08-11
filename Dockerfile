# Stage 1: Build the Go application
FROM golang:1-alpine AS builder
RUN apk update && rm -rf /var/cache/apk/*
RUN apk add --no-cache dumb-init
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3
RUN go build -ldflags="-s -w" -o app .

# Stage 2: Create a minimal runtime image
FROM scratch
COPY --from=builder ["/usr/bin/dumb-init", "/usr/bin/dumb-init"]
COPY --from=builder ["/build/app", "/"]
EXPOSE 5000
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/app"]
