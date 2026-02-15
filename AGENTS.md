# AGENTS.md - Rinha 2025 Go Development Guide

## Project Overview

- **Language**: Go 1.24.6
- **Framework**: FastHTTP (valyala/fasthttp)
- **Database**: Redis (Unix sockets)
- **Build**: Docker + Makefile
- **Entry**: `cmd/rinha/main.go`

## Build & Test Commands

```bash
# Build
make build              # Build all Docker images
make image              # Build single image
make all                # Full stack: down → build → up → logs

# Run
make up                 # Start services
make down               # Stop services
make clean              # Clean containers, images, volumes

# Test
go test ./...           # Run all Go tests
go test -run TestName ./pkg/...  # Run single test
go test -v -cover ./... # Run with coverage and verbose output

# Functional tests
make test               # All integration tests
make test-payment       # POST /payments
make test-stats         # GET /payments-summary
make test-purge         # POST /purge-payments

# Load tests
make k6                 # Default workload
make k6-1k              # 1,000 requests
make k6-5k              # 5,000 requests
make k6-final           # Final benchmark

# Lint/Format
go fmt ./...            # Format code
go vet ./...            # Vet code
go mod tidy             # Tidy modules

# Development
make client             # Start payment client (x86_64)
make client-arm         # Start payment client (ARM64)
```

## Project Structure

```
cmd/rinha/              # Application entry
internal/
  config/               # Config singleton with env vars
  database/             # Redis client
  models/               # Data structures
  server/               # HTTP handlers (FastHTTP)
  services/             # Business logic
pkg/
  http/                 # HTTP utilities
  utils/                # General utilities
build/                  # Docker files
config/                 # HAProxy, Redis configs
test/                   # k6 load tests
```

## Code Style Guidelines

### Import Organization (Standard Go)

**3 groups, blank line between**:
1. Standard library
2. Third-party imports
3. Local project imports (`rinha-2025-go/*`)

```go
import (
    "context"
    "fmt"
    "log"
    
    "github.com/ohler55/ojg/oj"
    "github.com/valyala/fasthttp"
    
    "rinha-2025-go/internal/config"
    "rinha-2025-go/internal/models"
)
```

### Naming Conventions

- **Exported**: PascalCase (`PaymentWorker`, `ProcessPayment`)
- **Unexported**: camelCase (`paymentWorker`, `processPayment`)
- **Structs**: PascalCase, descriptive
- **Constants**: PascalCase for exports

### Error Handling

```go
// Return wrapped errors
if err := w.redis.SavePayment(instance, payment); err != nil {
    return fmt.Errorf("failed to save payment: %w", err)
}

// Log operational errors
if _, _, err := fasthttp.Post(nil, url, nil); err != nil {
    log.Print("error:", err)
    return err
}

// Fatal only in initialization
workers, err := strconv.Atoi(utils.GetEnvOr("NUM_WORKERS", "50"))
if err != nil {
    log.Fatal("error parsing NUM_WORKERS:", err)
}
```

### Concurrency Patterns

```go
var wg sync.WaitGroup

wg.Add(1)
go func() {
    defer wg.Done()
    // work here
}()

wg.Wait()
```

### Resource Management

```go
redis := database.NewRedisClient(cfg)
defer redis.Close()

// FastHTTP lifecycle
req := fasthttp.AcquireRequest()
resp := fasthttp.AcquireResponse()
defer fasthttp.ReleaseRequest(req)
defer fasthttp.ReleaseResponse(resp)
```

### HTTP Handlers (FastHTTP)

```go
func PostPayment(worker *services.PaymentWorker) func(c *fasthttp.RequestCtx) {
    return func(c *fasthttp.RequestCtx) {
        var payment models.Payment
        if err := oj.Unmarshal(c.PostBody(), &payment); err != nil {
            c.Error(err.Error(), fasthttp.StatusBadRequest)
            return
        }
        go worker.EnqueuePayment(&payment)
        c.SetStatusCode(fasthttp.StatusAccepted)
    }
}
```

### JSON Serialization

Use `github.com/ohler55/ojg/oj` for performance:

```go
// Unmarshal
var payment models.Payment
if err := oj.Unmarshal(data, &payment); err != nil {
    return err
}

// Marshal
payload, err := oj.Marshal(payment)
```

### Struct Tags

```go
type Payment struct {
    PaymentID string    `json:"correlationId" binding:"required"`
    Amount    float64   `json:"amount" binding:"required,ge=0"`
    Timestamp time.Time `json:"requestedAt"`
}
```

### Configuration

```go
// utils
func GetEnvOr(key string, defaultValue string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return defaultValue
}

// config singleton
func (c *Config) Init() *Config {
    c.RedisSocket = utils.GetEnvOr("REDIS_SOCKET", "/sockets/redis.sock")
    // ...
    return c
}
```

## Testing Patterns

No unit tests currently exist. Testing is integration-based.

When adding tests:
- Place in same package: `<filename>_test.go`
- Use standard `testing` package
- Table-driven pattern:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {name: "valid", input: "test", wantErr: false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/valyala/fasthttp` | Fast HTTP server |
| `github.com/redis/go-redis/v9` | Redis client |
| `github.com/ohler55/ojg` | High-performance JSON |
| `github.com/joho/godotenv` | Environment loading |

## Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `DEFAULT_URL` | `http://payment-processor-default:8080` | Primary processor |
| `FALLBACK_URL` | `http://payment-processor-fallback:8080` | Backup processor |
| `REDIS_SOCKET` | `/sockets/redis.sock` | Redis socket |
| `SERVER_SOCKET` | `` | Server socket (empty = TCP :9999) |
| `NUM_WORKERS` | `50` | Worker goroutines |
| `GOMAXPROCS` | `3` | OS threads |

## Repository Rules

- **No modifying**: `.cursor/`, `.cursorrules`, `.github/copilot-instructions.md`
- **No linting config** currently present (optional to add `.golangci.yml`)
- **Commit style**: Clear, descriptive messages; atomic changes
- **Branch strategy**: `main` for production, feature branches via PR

## Performance Targets

Target: p99 latency ~0.45ms (Ryzen 9 5900X)

Optimization checklist:
- [ ] FastHTTP connection pooling
- [ ] Redis connection pooling
- [ ] Unix sockets for local services
- [ ] Minimize allocations in hot paths
- [ ] Proper NUM_WORKERS/GOMAXPROCS tuning

## References

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [FastHTTP](https://github.com/valyala/fasthttp)
- [Redis Go Client](https://redis.io/docs/latest/develop/connect/clients/go/)
