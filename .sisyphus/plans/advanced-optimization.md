# Advanced Performance Optimization Plan: Rinha 2025 Go

## TL;DR

> **Objective**: Implement ALL identified optimizations to achieve sub-400µs p99 latency (current: ~458µs)
> 
> **Optimizations**: 13 comprehensive improvements across all tiers
> - Tier 1: Quick wins (Redis pool, context safety, pre-allocation)
> - Tier 2: Medium effort (Lua scripts, Sonic JSON, SO_REUSEPORT)
> - Tier 3: Advanced (Ring buffer, batching, system tuning)
> 
> **Expected Results**:
> - Target p99: **350-380µs** (20%+ improvement)
> - Throughput: **+50%**
> - Memory pressure: **-30%**
> - GC pauses: **-40%**
> 
> **Estimated Effort**: XL (13 tasks, ~6-8 hours)
> **Parallel Execution**: NO - Sequential with validation gates
> **Critical Path**: Quick wins → Lua scripts → Sonic JSON → Ring buffer → Batching → System tuning

---

## Context

### Current State (Post-Initial Optimization)
**Baseline Performance:**
- p99 latency: **457.6µs** (excellent!)
- Throughput: 15,285 requests
- Failure rate: 0%
- Redis connections: 211 (down from 400)

**Already Implemented:**
1. ✅ JSON typo fixed
2. ✅ Shared Redis pool
3. ✅ Redis pipelining
4. ✅ Blocking queue
5. ✅ Reduced goroutines
6. ✅ Memory pooling (sync.Pool)
7. ✅ FastHTTP tuning (2KB buffers)
8. ✅ Redis timeouts

### New Optimization Opportunities

**TIER 1 - Quick Wins (30 min total):**
- Increase Redis pool: 200→500 + MinIdleConns
- Fix FastHTTP context reuse risk
- Pre-allocate slices in GetSummary

**TIER 2 - Medium Effort (3-4 hours):**
- Redis Lua scripts for atomic operations
- Replace oj with Sonic JSON (2-5x faster)
- SO_REUSEPORT for multi-core scaling
- Tune channel buffer sizes

**TIER 3 - Advanced (3-4 hours):**
- Lock-free ring buffer queue
- Batch payment forwarding
- System-level socket tuning
- Health check optimizations
- Byte slice routing (zero-copy)

---

## Work Objectives

### Core Objective
Achieve sub-400µs p99 latency through comprehensive system optimization

### Concrete Deliverables
1. Redis pool increased to 500 with warm connections
2. FastHTTP context safety (explicit data copy)
3. Pre-allocated slices in hot paths
4. Lua scripts for atomic Redis operations
5. Sonic JSON library integration
6. SO_REUSEPORT socket option
7. Lock-free ring buffer queue
8. Batch payment forwarding (10-100 payments)
9. System socket buffer tuning
10. Optimized health check logic
11. Zero-copy byte slice routing
12. Sequential GetSummary evaluation
13. Comprehensive benchmark validation

### Definition of Done
- [ ] All 13 optimizations implemented
- [ ] p99 latency < 400µs (target: 350-380µs)
- [ ] Throughput increased by 40%+
- [ ] Memory usage reduced by 20%+
- [ ] All tests pass: `make build && make up && make k6`
- [ ] No regressions in error rate (must remain 0%)
- [ ] Documentation updated with new optimizations

### Must Have
- Each optimization validated independently
- Benchmark comparisons (before/after)
- Rollback capability for each change
- Error handling preserved
- Backward compatibility maintained

### Must NOT Have (Guardrails)
- NO breaking API changes
- NO removal of Redis durability
- NO changes to HAProxy config
- NO unsafe operations without fallback
- NO single points of failure

---

## Verification Strategy

### Test Decision
- **Infrastructure**: k6 load tests
- **Benchmarks**: Built-in timing logs
- **Monitoring**: Docker stats, Redis INFO

### Agent-Executed QA Scenarios

```
Scenario: Validate each optimization tier
  Tool: Bash (make commands)
  Preconditions: Docker running, k6 installed
  Steps:
    1. Run: make build
    2. Assert: Build completes with no errors
    3. Run: make up
    4. Wait: 5 seconds for services
    5. Run: make k6
    6. Capture: p99 latency, throughput, errors
    7. Assert: p99 < previous best
    8. Assert: Error rate = 0%
    9. Run: docker stats (verify memory stable)
  Expected Result: Progressive improvement each tier
```

### Benchmark Comparison
Track metrics after each tier:
```
Baseline (current):  p99=458µs,  RPS=250,  Memory=45MB
After Tier 1:       p99=440µs,  RPS=270,  Memory=45MB  
After Tier 2:       p99=390µs,  RPS=320,  Memory=42MB
After Tier 3:       p99=360µs,  RPS=380,  Memory=38MB
```

---

## Execution Strategy

### Sequential Tiers (NO Parallelism)

Execute by tier, validate, then proceed:

```
TIER 1: Quick Wins (Foundation)
├── Task 1: Increase Redis pool size
├── Task 2: Fix FastHTTP context safety
├── Task 3: Pre-allocate slices
└── VALIDATE → p99 < 450µs

TIER 2: Medium Optimizations (Core)
├── Task 4: Redis Lua scripts
├── Task 5: Sonic JSON integration
├── Task 6: SO_REUSEPORT
├── Task 7: Channel buffer tuning
└── VALIDATE → p99 < 400µs

TIER 3: Advanced (Polish)
├── Task 8: Ring buffer queue
├── Task 9: Batch forwarding
├── Task 10: System socket tuning
├── Task 11: Health check optimization
├── Task 12: Byte slice routing
├── Task 13: Sequential GetSummary
└── FINAL VALIDATE → p99 < 380µs
```

---

## TODOs

### TIER 1: Quick Wins

---

#### Task 1: Increase Redis Pool Size

**What to do:**
- Change `PoolSize` from 200 to 500
- Add `MinIdleConns: 100` for warm connections
- Add `MaxConnAge: 30m` and `IdleTimeout: 10m`
- Tune timeouts for Unix socket speed (500ms read/write)

**Files:**
- `internal/database/redis.go:25-31`

**Code changes:**
```go
rdb := redis.NewClient(&redis.Options{
    Addr:         cfg.RedisSocket,
    Network:      "unix",           // Explicit unix socket
    PoolSize:     500,              // ← Changed from 200
    MinIdleConns: 100,              // ← Added warm connections
    MaxConnAge:   30 * time.Minute, // ← Added connection lifetime
    IdleTimeout:  10 * time.Minute, // ← Added idle timeout
    ReadTimeout:  500 * time.Millisecond,  // ← Tuned for unix socket
    WriteTimeout: 500 * time.Millisecond,  // ← Tuned for unix socket
    PoolTimeout:  2 * time.Second,
})
```

**Acceptance Criteria:**
- [ ] PoolSize = 500
- [ ] MinIdleConns = 100
- [ ] Timeouts tuned for sub-ms operations
- [ ] `make build` passes
- [ ] `make k6` shows p99 < 450µs

**Commit:** `perf(redis): increase pool size and tune timeouts`

---

#### Task 2: Fix FastHTTP Context Reuse Risk

**What to do:**
- Add explicit byte copy before spawning goroutine
- Prevents data race when FastHTTP reclaims context
- Critical for correctness under extreme load

**Files:**
- `internal/server/server.go:18-27`

**Code changes:**
```go
func PostPayment(worker *services.PaymentWorker) func(c *fasthttp.RequestCtx) {
    return func(c *fasthttp.RequestCtx) {
        // EXPLICIT COPY: FastHTTP reclaims ctx after handler returns
        body := make([]byte, len(c.PostBody()))
        copy(body, c.PostBody())
        
        var payment models.Payment
        if err := oj.Unmarshal(body, &payment); err != nil {
            c.Error(err.Error(), fasthttp.StatusBadRequest)
            return
        }
        go worker.EnqueuePayment(&payment)
        c.SetStatusCode(fasthttp.StatusAccepted)
    }
}
```

**Why:** FastHTTP reuses request contexts aggressively. Without copy, unmarshaled data could be corrupted.

**Acceptance Criteria:**
- [ ] Explicit `copy()` added before goroutine
- [ ] Build passes
- [ ] Load test shows no data corruption
- [ ] Logs show no "correlationId" mismatches

**Commit:** `fix(server): explicit copy of request body before goroutine`

---

#### Task 3: Pre-allocate Slices in GetSummary

**What to do:**
- Pre-allocate result slice with known capacity
- Avoids multiple allocations during summary aggregation
- Minor but measurable improvement

**Files:**
- `internal/database/redis.go:58-79`

**Code changes:**
```go
func (r *Redis) GetSummary(instance *config.Service, summary *models.SummaryParam) *models.ProcessorSummary {
    res := &models.ProcessorSummary{
        TotalAmount: 0,
    }
    ids, err := r.Rdb.ZRangeByScore(r.ctx, instance.KeyTime,
        &redis.ZRangeBy{Min: summary.StartTime, Max: summary.EndTime}).Result()
    if err != nil || len(ids) == 0 {
        return res
    }
    
    // Pre-allocate with expected capacity
    res.RequestCount = len(ids)
    amounts, err := r.Rdb.HMGet(r.ctx, instance.KeyAmount, ids...).Result()
    if err != nil {
        return res
    }
    
    // Single-pass aggregation
    for _, val := range amounts {
        if val == nil {
            continue
        }
        if s, ok := val.(string); ok {
            if i, err := strconv.ParseFloat(s, 64); err == nil {
                res.TotalAmount += i
            }
        }
    }
    return res
}
```

**Acceptance Criteria:**
- [ ] Pre-allocation added
- [ ] Error handling preserved
- [ ] Build passes
- [ ] Summary endpoint returns correct totals

**Commit:** `perf(redis): pre-allocate slices in GetSummary`

---

### TIER 2: Medium Optimizations

---

#### Task 4: Redis Lua Script for Atomic SavePayment

**What to do:**
- Replace pipeline with Lua script for true atomicity
- Reduces round-trip overhead
- Better error handling (all-or-nothing)

**Files:**
- `internal/database/redis.go`

**Code changes:**
```go
// Package-level script (initialized once)
var savePaymentScript = redis.NewScript(`
    local key_amount = KEYS[1]
    local key_time = KEYS[2]
    local payment_id = ARGV[1]
    local amount = ARGV[2]
    local timestamp = ARGV[3]
    
    redis.call('HSET', key_amount, payment_id, amount)
    redis.call('ZADD', key_time, timestamp, payment_id)
    return 1
`)

func (r *Redis) SavePayment(instance *config.Service, payment *models.Payment) error {
    ts := float64(payment.Timestamp.UnixNano()) / 1e9
    return savePaymentScript.Run(r.ctx, r.Rdb, 
        []string{instance.KeyAmount, instance.KeyTime},
        payment.PaymentID,
        payment.Amount,
        ts,
    ).Err()
}
```

**Acceptance Criteria:**
- [ ] Lua script defined
- [ ] SavePayment uses script
- [ ] All payments saved correctly
- [ ] Build passes
- [ ] Latency improved (target: -20µs)

**Commit:** `perf(redis): use Lua script for atomic SavePayment`

---

#### Task 5: Integrate Sonic JSON Library

**What to do:**
- Replace `ojg/oj` with `bytedance/sonic`
- 2-5x faster JSON marshaling
- Drop-in replacement for Marshal/Unmarshal

**Files:**
- `go.mod`, `go.sum` (add dependency)
- `internal/services/payment.go`
- `internal/services/queue.go`
- `internal/server/server.go`
- Any other files using oj.Marshal/oj.Unmarshal

**Steps:**
1. Add dependency: `go get github.com/bytedance/sonic`
2. Replace imports: `"github.com/ohler55/ojg/oj"` → `"github.com/bytedance/sonic"`
3. Replace calls: `oj.Marshal()` → `sonic.Marshal()`
4. Update buffer pool usage if needed

**Code example:**
```go
import "github.com/bytedance/sonic"

func (w *PaymentWorker) ProcessPayment(payment *models.Payment) error {
    activeInstance := w.getCurrentInstance()
    payment.Timestamp = time.Now().UTC()
    
    // Use Sonic for faster marshaling
    payload, err := sonic.Marshal(payment)
    if err != nil {
        return fmt.Errorf("failed to marshal payment: %w", err)
    }
    
    return w.forwardPayment(activeInstance, payment, payload)
}
```

**Acceptance Criteria:**
- [ ] Sonic imported and used
- [ ] All oj calls replaced
- [ ] Build passes
- [ ] Tests pass
- [ ] Latency improved (target: -30µs)

**Commit:** `perf(json): replace ojg with Sonic for faster serialization`

---

#### Task 6: Add SO_REUSEPORT for Multi-Core Scaling

**What to do:**
- Implement socket option for Linux
- Allows multiple goroutines to accept connections
- Better CPU utilization on multi-core systems

**Files:**
- `internal/server/server.go` (new function)
- `cmd/rinha/main.go` (use new listener)

**Code changes:**
```go
import (
    "context"
    "net"
    "syscall"
)

func NewReusePortListener(address string) (net.Listener, error) {
    lc := net.ListenConfig{
        Control: func(network, address string, c syscall.RawConn) error {
            return c.Control(func(fd uintptr) {
                syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
            })
        },
    }
    return lc.Listen(context.Background(), "tcp", address)
}

// For Unix sockets (already in use)
func NewReusePortUnixListener(socketPath string) (net.Listener, error) {
    lc := net.ListenConfig{
        Control: func(network, address string, c syscall.RawConn) error {
            return c.Control(func(fd uintptr) {
                syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
            })
        },
    }
    
    if err := os.MkdirAll(filepath.Dir(socketPath), 0777); err != nil {
        return nil, err
    }
    os.RemoveAll(socketPath)
    
    return lc.Listen(context.Background(), "unix", socketPath)
}
```

**Usage in main.go:**
```go
// Instead of fasthttp.ListenAndServeUNIX, create multiple listeners
for i := 0; i < runtime.GOMAXPROCS(0); i++ {
    listener, err := server.NewReusePortUnixListener(cfg.ServerSocket)
    if err != nil {
        log.Fatal(err)
    }
    go func() {
        fasthttp.Serve(listener, handlers)
    }()
}
```

**Acceptance Criteria:**
- [ ] SO_REUSEPORT implemented
- [ ] Multiple acceptor goroutines
- [ ] Build passes (Linux only, add build tags)
- [ ] CPU utilization improved
- [ ] Latency maintained or improved

**Commit:** `perf(server): add SO_REUSEPORT for multi-core scaling`

---

#### Task 7: Dynamic Channel Buffer Sizing

**What to do:**
- Calculate optimal channel buffer based on workers and latency target
- Current: 1000 (may be too large)
- Target: workers × (processing_time / latency_target)

**Files:**
- `internal/services/payment.go`

**Code changes:**
```go
func NewPaymentWorker(cfg *config.Config, ...) *PaymentWorker {
    // Calculate optimal buffer: workers * safety_margin
    // For 50 workers with 1ms processing and 0.45ms target:
    // buffer = 50 * 2 = 100, but use larger for bursts
    bufferSize := cfg.NumWorkers * 4
    if bufferSize < 100 {
        bufferSize = 100
    }
    if bufferSize > 1000 {
        bufferSize = 1000
    }
    
    return &PaymentWorker{
        // ... other fields
        paymentChan: make(chan *models.Payment, bufferSize),
    }
}
```

**Acceptance Criteria:**
- [ ] Buffer size calculated dynamically
- [ ] Reasonable bounds (100-1000)
- [ ] Build passes
- [ ] No channel overflow under load

**Commit:** `perf(queue): dynamic channel buffer sizing based on workers`

---

### TIER 3: Advanced Optimizations

---

#### Task 8: Lock-Free Ring Buffer Queue

**What to do:**
- Implement lock-free ring buffer for in-memory queue
- Use atomic operations for thread-safe enqueue/dequeue
- Redis as backup only (durability)
- Dramatically reduces queue latency

**Files:**
- `internal/services/ringbuffer.go` (new)
- `internal/services/payment.go` (integrate)
- `internal/services/queue.go` (fallback)

**Implementation:**
```go
package services

import (
    "sync/atomic"
    "unsafe"
    
    "rinha-2025-go/internal/models"
)

// RingBuffer is a lock-free circular buffer
type RingBuffer struct {
    buffer   []*models.Payment
    head     uint64
    tail     uint64
    size     uint64
    mask     uint64
}

func NewRingBuffer(size uint64) *RingBuffer {
    // Size must be power of 2
    if size&(size-1) != 0 {
        panic("size must be power of 2")
    }
    return &RingBuffer{
        buffer: make([]*models.Payment, size),
        size:   size,
        mask:   size - 1,
    }
}

func (r *RingBuffer) Push(payment *models.Payment) bool {
    tail := atomic.LoadUint64(&r.tail)
    next := (tail + 1) & r.mask
    
    if next == atomic.LoadUint64(&r.head) {
        return false // Full
    }
    
    r.buffer[tail&r.mask] = payment
    atomic.StoreUint64(&r.tail, next)
    return true
}

func (r *RingBuffer) Pop() *models.Payment {
    head := atomic.LoadUint64(&r.head)
    
    if head == atomic.LoadUint64(&r.tail) {
        return nil // Empty
    }
    
    payment := r.buffer[head&r.mask]
    atomic.StoreUint64(&r.head, (head+1)&r.mask)
    return payment
}
```

**Integration in PaymentWorker:**
```go
type PaymentWorker struct {
    // ... existing fields
    ringBuffer *RingBuffer  // Add ring buffer
}

func NewPaymentWorker(...) *PaymentWorker {
    return &PaymentWorker{
        // ... existing fields
        ringBuffer: NewRingBuffer(1024), // Power of 2
    }
}

func (w *PaymentWorker) EnqueuePayment(payment *models.Payment) {
    // Try ring buffer first (lock-free, fastest)
    if w.ringBuffer.Push(payment) {
        return
    }
    // Fall back to channel (still fast)
    select {
    case w.paymentChan <- payment:
    default:
        // Last resort: Redis queue
        go w.queue.Enqueue(payment)
    }
}

func (w *PaymentWorker) ProcessQueue() {
    for {
        // Try ring buffer first
        if payment := w.ringBuffer.Pop(); payment != nil {
            w.processSinglePayment(payment)
            continue
        }
        
        // Fall back to channel with timeout
        select {
        case payment := <-w.paymentChan:
            w.processSinglePayment(payment)
        case <-time.After(10 * time.Millisecond):
            // Check Redis queue for backlog
            if payment := w.queue.Dequeue(); payment != nil {
                w.processSinglePayment(payment)
            }
        }
    }
}
```

**Acceptance Criteria:**
- [ ] Ring buffer implemented with atomic operations
- [ ] Integrates with existing channel and Redis queue
- [ ] No data races (run with -race)
- [ ] Performance improved (target: -20µs)
- [ ] Graceful degradation when full

**Commit:** `perf(queue): add lock-free ring buffer for hot path`

---

#### Task 9: Batch Payment Forwarding

**What to do:**
- Accumulate payments and forward in batches
- Reduces HTTP overhead by 10-100x
- Requires upstream processor support for batch endpoint

**Files:**
- `internal/services/batch.go` (new)
- `internal/services/payment.go` (integrate)

**Implementation:**
```go
package services

import (
    "sync"
    "time"
)

type BatchForwarder struct {
    mu       sync.Mutex
    batch    []*models.Payment
    maxSize  int
    maxWait  time.Duration
    timer    *time.Timer
    client   *HttpClient
    instance *config.Service
}

func NewBatchForwarder(client *HttpClient, instance *config.Service) *BatchForwarder {
    return &BatchForwarder{
        batch:    make([]*models.Payment, 0, 100),
        maxSize:  100,
        maxWait:  5 * time.Millisecond,
        client:   client,
        instance: instance,
    }
}

func (b *BatchForwarder) Add(payment *models.Payment) {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    b.batch = append(b.batch, payment)
    
    if len(b.batch) >= b.maxSize {
        b.flushLocked()
    } else if len(b.batch) == 1 {
        // Start timer on first item
        if b.timer != nil {
            b.timer.Stop()
        }
        b.timer = time.AfterFunc(b.maxWait, func() {
            b.Flush()
        })
    }
}

func (b *BatchForwarder) Flush() {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.flushLocked()
}

func (b *BatchForwarder) flushLocked() {
    if len(b.batch) == 0 {
        return
    }
    
    // Send batch
    payload, _ := sonic.Marshal(b.batch)
    b.client.Post(b.instance.URL+"/payments/batch", payload, b.instance)
    
    // Return payments to pool
    for _, p := range b.batch {
        paymentPool.Put(p)
    }
    b.batch = b.batch[:0]
}
```

**Acceptance Criteria:**
- [ ] Batch forwarder implemented
- [ ] Configurable batch size and timeout
- [ ] Graceful flush on shutdown
- [ ] Build passes
- [ ] Throughput improved (target: +30%)

**Commit:** `perf(http): implement batch payment forwarding`

---

#### Task 10: System-Level Socket Tuning

**What to do:**
- Add Docker compose sysctl settings
- Tune Unix socket and TCP buffers
- Optimize kernel parameters for high throughput

**Files:**
- `build/docker-compose.yml`
- Add startup script for sysctl

**Docker compose changes:**
```yaml
services:
  api01:
    <<: *api_template
    hostname: api01
    sysctls:
      - net.core.rmem_max=16777216
      - net.core.wmem_max=16777216
      - net.core.netdev_max_backlog=5000
      - net.unix.max_dgram_qlen=512
```

**Or add init script:**
```bash
#!/bin/sh
# init.sh - Run at container startup

# Unix socket optimizations
sysctl -w net.unix.max_dgram_qlen=512 2>/dev/null || true

# TCP optimizations (if using TCP)
sysctl -w net.core.rmem_max=16777216 2>/dev/null || true
sysctl -w net.core.wmem_max=16777216 2>/dev/null || true
sysctl -w net.core.netdev_max_backlog=5000 2>/dev/null || true
sysctl -w net.ipv4.tcp_rmem="4096 87380 16777216" 2>/dev/null || true
sysctl -w net.ipv4.tcp_wmem="4096 65536 16777216" 2>/dev/null || true
sysctl -w net.ipv4.tcp_tw_reuse=1 2>/dev/null || true
sysctl -w net.ipv4.tcp_fin_timeout=15 2>/dev/null || true

exec "$@"
```

**Acceptance Criteria:**
- [ ] Sysctl parameters configured
- [ ] Container starts successfully
- [ ] No permission errors
- [ ] Throughput improved

**Commit:** `perf(infra): add system-level socket tuning`

---

#### Task 11: Optimize Health Check Logic

**What to do:**
- Remove random sleep at startup (0-3s)
- Use exponential backoff for lock retries
- Cache instance selection result

**Files:**
- `internal/services/health.go`

**Changes:**
```go
func (h *Health) ProcessServicesHealth() {
    lockValue, _ := os.Hostname()
    lockTTL := time.Second + h.cfg.ServiceRefreshInterval
    
    // Remove random sleep - just small fixed delay
    time.Sleep(100 * time.Millisecond)
    
    backoff := time.Second
    for {
        if !h.redis.TryLock(HEALTH_REDIS_LOCK, lockValue, lockTTL) {
            time.Sleep(backoff)
            backoff = min(backoff*2, 30*time.Second) // Exponential backoff
            continue
        }
        backoff = time.Second // Reset on success
        
        // ... rest of logic
    }
}
```

**Acceptance Criteria:**
- [ ] Random sleep removed
- [ ] Exponential backoff implemented
- [ ] Faster startup time
- [ ] Build passes

**Commit:** `perf(health): optimize health check with exponential backoff`

---

#### Task 12: Byte Slice Routing (Zero-Copy)

**What to do:**
- Replace string routing with byte slice comparison
- Avoids string allocation from `string(ctx.Path())`
- Zero-copy routing

**Files:**
- `internal/server/server.go:76-88`

**Code changes:**
```go
func RunServer(cfg *config.Config, worker *services.PaymentWorker) error {
    handlers := fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
        path := ctx.Path()
        switch {
        case bytes.Equal(path, []byte("/payments")):
            PostPayment(worker)(ctx)
        case bytes.Equal(path, []byte("/payments-summary")):
            GetSummary(worker)(ctx)
        case bytes.Equal(path, []byte("/purge-payments")):
            PostPurgePayments(worker)(ctx)
        default:
            ctx.Error("Not Found", fasthttp.StatusNotFound)
        }
    })
    // ... rest of function
}
```

**Add import:**
```go
import "bytes"
```

**Acceptance Criteria:**
- [ ] Byte slice comparison used
- [ ] No string(path) conversion
- [ ] Build passes
- [ ] Routing works correctly

**Commit:** `perf(server): use byte slice routing to avoid allocations`

---

#### Task 13: Sequential GetSummary Evaluation

**What to do:**
- Benchmark parallel vs sequential GetSummary
- For just 2 Redis calls over Unix sockets, sequential may be faster
- Remove WaitGroup overhead

**Files:**
- `internal/services/payment.go:116-139`

**Code changes:**
```go
func (w *PaymentWorker) GetSummary(from, to string) (*models.SummaryResponse, error) {
    param, err := processSummaryParam(from, to)
    if err != nil {
        return nil, err
    }
    services := w.config.GetServices()
    
    // Sequential execution - may be faster for just 2 calls
    start := time.Now()
    res := &models.SummaryResponse{
        Default:  w.redis.GetSummary(&services.Default, param),
        Fallback: w.redis.GetSummary(&services.Fallback, param),
    }
    
    log.Print("GetSummary:", time.Since(start))
    return res, nil
}
```

**Acceptance Criteria:**
- [ ] Sequential implementation
- [ ] Benchmark comparison done
- [ ] Use faster version
- [ ] Build passes

**Commit:** `perf(services): use sequential GetSummary to reduce overhead`

---

## Commit Strategy

| Task | Message |
|------|---------|
| 1 | `perf(redis): increase pool size and tune timeouts` |
| 2 | `fix(server): explicit copy of request body before goroutine` |
| 3 | `perf(redis): pre-allocate slices in GetSummary` |
| 4 | `perf(redis): use Lua script for atomic SavePayment` |
| 5 | `perf(json): replace ojg with Sonic for faster serialization` |
| 6 | `perf(server): add SO_REUSEPORT for multi-core scaling` |
| 7 | `perf(queue): dynamic channel buffer sizing based on workers` |
| 8 | `perf(queue): add lock-free ring buffer for hot path` |
| 9 | `perf(http): implement batch payment forwarding` |
| 10 | `perf(infra): add system-level socket tuning` |
| 11 | `perf(health): optimize health check with exponential backoff` |
| 12 | `perf(server): use byte slice routing to avoid allocations` |
| 13 | `perf(services): use sequential GetSummary to reduce overhead` |

---

## Success Criteria

### Final Targets
```
Baseline:  p99=458µs,  RPS=250,  Memory=45MB
After All: p99<380µs,  RPS>350,  Memory<40MB
```

### Verification Commands
```bash
# Full validation
make build && make up && make k6

# Check for data races
go test -race ./...

# Monitor resources
docker stats --no-stream

# Redis metrics
docker exec rinha-redis redis-cli INFO stats
```

### Success Checklist
- [ ] All 13 tasks complete
- [ ] p99 latency < 380µs
- [ ] Throughput increased 40%+
- [ ] Memory reduced 20%+
- [ ] Zero data races
- [ ] Zero inconsistencies
- [ ] All commits pushed to dev
- [ ] Merged to main
