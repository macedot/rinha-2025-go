# Performance Upgrade Plan: Rinha 2025 Go

## TL;DR

> **Objective**: Comprehensive performance optimization of Go payment processor to reduce p99 latency below 0.45ms
> 
> **Deliverables**: 
> - Fixed JSON typo
> - Consolidated Redis pool (2→1)
> - Pipelined Redis operations
> - Blocking queue (no polling)
> - Reduced goroutine overhead
> - Memory pooling for hot paths
> - Tuned FastHTTP configuration
> 
> **Estimated Effort**: Large (8 incremental tasks)
> **Parallel Execution**: NO - Sequential with validation gates
> **Critical Path**: Fix typo → Shared Redis → Pipelining → Queue optimization → Goroutine reduction → Pooling → Config tuning

---

## Context

### Original Request
Comprehensive performance revision of the payment processing application with conservative, incremental testing approach. Each change must be validated with `make build`, `make up`, `make k6` before proceeding.

### Current Architecture
- **Language**: Go 1.24.6
- **HTTP**: FastHTTP (valyala/fasthttp)
- **Database**: Redis via Unix sockets
- **Load Balancer**: HAProxy with 4 backend instances
- **Resources**: 0.25 CPU / 50MB RAM per API container
- **Target**: p99 latency ~0.45ms (already excellent baseline)

### Identified Bottlenecks
1. **Two Redis clients** wasting 400 connections
2. **Sequential Redis ops** (2 round trips where 1 suffices)
3. **Polling queue** with 1s sleep adding latency
4. **Excessive goroutines** (3+ per payment)
5. **Memory allocations** in hot paths
6. **JSON typo** in model (TtotalFee)

---

## Work Objectives

### Core Objective
Optimize payment processing pipeline to reduce latency and resource usage while maintaining correctness and stability.

### Concrete Deliverables
- Fixed `TtotalFee` → `TotalFee` in `internal/models/summary.go`
- Shared Redis client between database and queue layers
- Pipelined Redis operations in `SavePayment`
- Blocking queue pattern replacing sleep polling
- Reduced goroutine spawning in payment flow
- Memory pooling for JSON marshaling buffers
- Tuned FastHTTP client configuration

### Definition of Done
- [ ] All tasks complete with passing `make k6` benchmarks
- [ ] No regression in p99 latency
- [ ] Docker builds successfully (`make build`)
- [ ] Stack starts cleanly (`make up`)

### Must Have
- Each change tested independently before next
- Backward compatibility maintained
- Error handling preserved
- Logging preserved for observability

### Must NOT Have (Guardrails)
- NO breaking API changes
- NO removal of existing error handling
- NO changes to HAProxy or Redis configuration
- NO changes to environment variable interface
- NO concurrent execution of risky changes

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: YES (k6 load tests in `test/`)
- **Automated tests**: NO (integration testing only)
- **Framework**: None (k6 for load testing)

### Agent-Executed QA Scenarios (MANDATORY)

Each task includes validation via Docker build and k6 benchmark:

```
Scenario: Validate optimization with k6 load test
  Tool: Bash (make commands)
  Preconditions: Docker daemon running, k6 installed
  Steps:
    1. Run: make build
    2. Assert: Build completes with no errors
    3. Run: make up
    4. Wait: 5 seconds for services to start
    5. Run: make k6
    6. Assert: k6 test completes successfully
    7. Assert: No errors in container logs (docker logs rinha-api01)
  Expected Result: Services start, load test passes
  Evidence: k6 output, container logs
```

---

## Execution Strategy

### Sequential Execution (NO Parallelism)

Due to conservative approach, each task must complete and be validated before the next begins:

```
Phase 1: Foundation
└── Task 1: Fix JSON typo (lowest risk)
    └── Validate → Proceed

Phase 2: Redis Consolidation  
└── Task 2: Share Redis client
    └── Validate → Proceed

Phase 3: Redis Optimization
└── Task 3: Add pipelining
    └── Validate → Proceed

Phase 4: Queue Optimization
└── Task 4: Blocking queue pattern
    └── Validate → Proceed

Phase 5: Concurrency Optimization
└── Task 5: Reduce goroutines
    └── Validate → Proceed

Phase 6: Memory Optimization
└── Task 6: Memory pooling
    └── Validate → Proceed

Phase 7: Configuration Tuning
└── Task 7: FastHTTP tuning
    └── Validate → Proceed

Phase 8: Connection Reliability
└── Task 8: Redis timeouts
    └── Validate → FINAL
```

### Rollback Strategy
Each task maintains ability to revert via git:
- Commit after each successful validation
- Tag commits for easy rollback
- Document any breaking changes in commit message

---

## TODOs

### Task 1: Fix JSON Tag Typo

**What to do**:
- Change `TtotalFee` to `TotalFee` in `internal/models/summary.go:16`
- This fixes a serialization bug where the field won't match expected JSON key

**Must NOT do**:
- Do NOT change any other field names
- Do NOT change struct tags on other fields
- Do NOT modify business logic

**Recommended Agent Profile**:
- **Category**: `quick`
- **Reason**: Single-line fix, minimal scope, no logic changes
- **Skills**: None required

**Parallelization**:
- **Can Run In Parallel**: NO (first task)
- **Sequential**: Must complete before Task 2
- **Blocks**: Task 2
- **Blocked By**: None

**References**:
- `internal/models/summary.go:16` - Field definition with typo
- Pattern: Standard Go struct tags

**Acceptance Criteria**:
- [ ] `TtotalFee` changed to `TotalFee`
- [ ] `make build` succeeds
- [ ] `make up` starts services
- [ ] `make k6` passes

**Agent-Executed QA Scenario**:
```
Scenario: Typo fix validation
  Tool: Bash
  Steps:
    1. grep -n "TotalFee" internal/models/summary.go
    2. Assert: Line 16 shows `TotalFee` not `TtotalFee`
    3. Run: make build
    4. Assert: Exit code 0, no errors
    5. Run: make up
    6. Wait: 5s
    7. Run: docker ps | grep rinha-api
    8. Assert: 4 api containers running
    9. Run: make k6
    10. Assert: k6 completes with success metrics
  Expected Result: Build, deploy, and load test all pass
```

**Commit**: YES
- Message: `fix(models): correct JSON tag typo TtotalFee -> TotalFee`
- Files: `internal/models/summary.go`

---

### Task 2: Consolidate Redis Clients

**What to do**:
- Modify `internal/services/queue.go` to accept existing Redis client instead of creating new one
- Update `internal/services/payment.go` NewPaymentWorker to pass shared Redis client
- Update `cmd/rinha/main.go` to create single Redis client and pass to both components
- Remove duplicate connection pool (400 → 200 connections)

**Must NOT do**:
- Do NOT change Redis operations logic
- Do NOT modify pool size configuration
- Do NOT break existing queue functionality
- Do NOT change the Queue interface

**Recommended Agent Profile**:
- **Category**: `unspecified-medium`
- **Reason**: Requires understanding dependency injection and careful interface changes
- **Skills**: None required

**Parallelization**:
- **Can Run In Parallel**: NO
- **Sequential**: After Task 1
- **Blocks**: Task 3
- **Blocked By**: Task 1

**References**:
- `internal/database/redis.go:23-35` - Redis client creation
- `internal/services/queue.go:20-35` - Queue creates its own client
- `cmd/rinha/main.go:11-26` - Main initialization
- Pattern: Dependency injection of shared resources

**Acceptance Criteria**:
- [ ] Only one `redis.NewClient` call in entire codebase
- [ ] Queue uses passed-in Redis client
- [ ] All existing tests pass
- [ ] `make build` succeeds
- [ ] `make up` starts services
- [ ] `make k6` passes

**Agent-Executed QA Scenario**:
```
Scenario: Shared Redis client validation
  Tool: Bash
  Steps:
    1. grep -n "redis.NewClient" internal/**/*.go
    2. Assert: Only 1 occurrence (in database/redis.go)
    3. Run: make build
    4. Assert: Exit code 0
    5. Run: make up
    6. Wait: 5s
    7. Run: docker exec rinha-redis redis-cli -s /sockets/redis.sock INFO clients
    8. Assert: connected_clients shows reasonable number (~200, not 400)
    9. Run: make k6
    10. Assert: Load test passes
  Expected Result: Single Redis client, correct connection count
```

**Commit**: YES
- Message: `refactor(redis): share single client between database and queue`
- Files: `internal/services/queue.go`, `internal/services/payment.go`, `cmd/rinha/main.go`

---

### Task 3: Add Redis Pipelining

**What to do**:
- Modify `internal/database/redis.go` `SavePayment` method
- Use `redis.Pipeline` to batch HSet and ZAdd into single round trip
- Maintain same transactional semantics (both succeed or both fail implied by independent ops)

**Must NOT do**:
- Do NOT use MULTI/EXEC (overkill for this use case)
- Do NOT change data structure or key naming
- Do NOT change error handling behavior
- Do NOT modify other Redis operations yet

**Recommended Agent Profile**:
- **Category**: `unspecified-medium`
- **Reason**: Requires understanding Redis pipelining and error handling
- **Skills**: None required

**Parallelization**:
- **Can Run In Parallel**: NO
- **Sequential**: After Task 2
- **Blocks**: Task 4
- **Blocked By**: Task 2

**References**:
- `internal/database/redis.go:42-51` - Current SavePayment implementation
- go-redis docs: Pipeline for batching commands
- Pattern: Redis pipelining for atomic-like batches

**Acceptance Criteria**:
- [ ] SavePayment uses Pipeline
- [ ] Both HSet and ZAdd in same pipeline
- [ ] Error handling preserved
- [ ] `make build` succeeds
- [ ] `make up` starts services
- [ ] `make k6` passes with maintained or improved latency

**Agent-Executed QA Scenario**:
```
Scenario: Redis pipelining validation
  Tool: Bash
  Steps:
    1. Check: internal/database/redis.go uses pipe := r.Rdb.Pipeline()
    2. Run: make build
    3. Assert: Build succeeds
    4. Run: make up
    5. Wait: 5s
    6. Run: make k6
    7. Capture: k6 latency metrics
    8. Assert: p99 latency maintained or improved
    9. Run: docker logs rinha-api01 2>&1 | grep -i error
    10. Assert: No payment save errors
  Expected Result: Payments saved correctly, latency maintained
```

**Commit**: YES
- Message: `perf(redis): pipeline SavePayment operations`
- Files: `internal/database/redis.go`

---

### Task 4: Implement Blocking Queue

**What to do**:
- Replace polling loop in `internal/services/payment.go` `ProcessQueue`
- Use channel-based blocking instead of sleep polling
- Modify `EnqueuePayment` to send to channel
- Keep Redis queue as persistence layer, use channel for signaling

**Must NOT do**:
- Do NOT remove Redis queue entirely (needed for durability)
- Do NOT change queue order semantics (FIFO)
- Do NOT remove retry logic for failed payments
- Do NOT introduce deadlocks

**Recommended Agent Profile**:
- **Category**: `ultrabrain`
- **Reason**: Complex concurrency refactoring with channel synchronization
- **Skills**: None required

**Parallelization**:
- **Can Run In Parallel**: NO
- **Sequential**: After Task 3
- **Blocks**: Task 5
- **Blocked By**: Task 3

**References**:
- `internal/services/payment.go:52-63` - Current polling ProcessQueue
- `internal/services/payment.go:48-50` - EnqueuePayment
- `internal/services/queue.go` - Redis queue implementation
- Pattern: Channel-based worker pool with Redis backing

**Acceptance Criteria**:
- [ ] ProcessQueue uses channel receive instead of sleep polling
- [ ] EnqueuePayment sends to channel
- [ ] No time.Sleep in queue processing loop
- [ ] Retry logic preserved
- [ ] `make build` succeeds
- [ ] `make up` starts services
- [ ] `make k6` passes

**Agent-Executed QA Scenario**:
```
Scenario: Blocking queue validation
  Tool: Bash
  Steps:
    1. Check: internal/services/payment.go uses chan *models.Payment
    2. Check: No time.Sleep in ProcessQueue loop
    3. Run: make build
    4. Assert: Build succeeds
    5. Run: make up
    6. Wait: 5s
    7. Run: make k6
    8. Assert: All payments processed
    9. Run: docker logs rinha-api01 | grep -c "error"
    10. Assert: Error count is 0 or minimal
  Expected Result: Queue processes without polling delays
```

**Commit**: YES
- Message: `perf(queue): implement blocking channel-based queue`
- Files: `internal/services/payment.go`, `internal/services/queue.go`

---

### Task 5: Reduce Goroutine Spawning

**What to do**:
- Remove goroutine from `EnqueuePayment` (line 49)
- Evaluate if goroutines in `ProcessPayment` (lines 77-89) are necessary
- Consolidate concurrent work where beneficial, remove where not
- Consider synchronous instance lookup if health check is fast

**Must NOT do**:
- Do NOT remove worker goroutines (the `for range cfg.NumWorkers` loop)
- Do NOT block the HTTP handler thread
- Do NOT break the concurrent processing model
- Do NOT introduce race conditions

**Recommended Agent Profile**:
- **Category**: `ultrabrain`
- **Reason**: Critical concurrency changes affecting performance and correctness
- **Skills**: None required

**Parallelization**:
- **Can Run In Parallel**: NO
- **Sequential**: After Task 4
- **Blocks**: Task 6
- **Blocked By**: Task 4

**References**:
- `internal/services/payment.go:48-50` - EnqueuePayment spawns goroutine
- `internal/services/payment.go:74-91` - ProcessPayment with 2 goroutines
- `internal/services/payment.go:65-72` - getCurrentInstance with polling
- Pattern: Efficient goroutine usage, avoid spawning where unnecessary

**Acceptance Criteria**:
- [ ] EnqueuePayment is synchronous (no `go` keyword)
- [ ] ProcessPayment goroutines justified or removed
- [ ] HTTP handlers remain non-blocking
- [ ] `make build` succeeds
- [ ] `make up` starts services
- [ ] `make k6` passes
- [ ] Goroutine count reduced (check with `go tool pprof` if available)

**Agent-Executed QA Scenario**:
```
Scenario: Goroutine optimization validation
  Tool: Bash
  Steps:
    1. Check: internal/services/payment.go EnqueuePayment has no 'go' keyword
    2. Run: make build
    3. Assert: Build succeeds
    4. Run: make up
    5. Wait: 5s
    6. Run: docker exec rinha-api01 ps aux | grep -c "app"
    7. Assert: Main process running
    8. Run: make k6
    9. During test: docker exec rinha-api01 ps aux | wc -l
    10. Assert: Reasonable goroutine count (not excessively high)
    11. Assert: k6 completes successfully
  Expected Result: Fewer goroutines, same or better performance
```

**Commit**: YES
- Message: `perf(concurrency): reduce unnecessary goroutine spawning`
- Files: `internal/services/payment.go`

---

### Task 6: Add Memory Pooling

**What to do**:
- Add `sync.Pool` for JSON marshaling buffers in hot paths
- Pool byte slices used in `oj.Marshal` calls
- Focus on `internal/services/payment.go` and `internal/services/queue.go`
- Benchmark to verify improvement

**Must NOT do**:
- Do NOT pool payment structs (different lifetimes)
- Do NOT change API signatures
- Do NOT introduce memory leaks (always Put back to pool)
- Do NOT optimize prematurely - measure first

**Recommended Agent Profile**:
- **Category**: `ultrabrain`
- **Reason**: Requires understanding sync.Pool semantics and careful lifecycle management
- **Skills**: None required

**Parallelization**:
- **Can Run In Parallel**: NO
- **Sequential**: After Task 5
- **Blocks**: Task 7
- **Blocked By**: Task 5

**References**:
- `internal/services/payment.go:87` - oj.Marshal in ProcessPayment
- `internal/services/queue.go:38` - oj.Marshal in Enqueue
- `sync.Pool` - Go standard library for object pooling
- Pattern: sync.Pool for buffer reuse in hot paths

**Acceptance Criteria**:
- [ ] sync.Pool defined for byte slices
- [ ] Pool used in oj.Marshal operations
- [ ] Proper Get/Put lifecycle
- [ ] `make build` succeeds
- [ ] `make up` starts services
- [ ] `make k6` passes
- [ ] No memory leaks (stable memory usage)

**Agent-Executed QA Scenario**:
```
Scenario: Memory pooling validation
  Tool: Bash
  Steps:
    1. Check: internal/services/payment.go has var bufferPool = sync.Pool{...}
    2. Check: bufferPool.Get() and Put() used correctly
    3. Run: make build
    4. Assert: Build succeeds
    5. Run: make up
    6. Wait: 5s
    7. Run: docker stats rinha-api01 --no-stream
    8. Record: Memory usage baseline
    9. Run: make k6
    10. During test: docker stats rinha-api01 (in another terminal)
    11. Assert: Memory usage remains stable
    12. Assert: k6 completes successfully
  Expected Result: Stable memory, reduced GC pressure
```

**Commit**: YES
- Message: `perf(memory): add sync.Pool for JSON buffer reuse`
- Files: `internal/services/payment.go`, `internal/services/queue.go`

---

### Task 7: Tune FastHTTP Configuration

**What to do**:
- Review and optimize `pkg/http/http.go` FastHTTP client config
- Right-size buffers based on actual payment payload sizes
- Tune connection pool settings
- Evaluate if `MaxConnsPerHost: 4096` is appropriate

**Must NOT do**:
- Do NOT disable keep-alive (would hurt performance)
- Do NOT reduce timeouts below safe thresholds
- Do NOT change TCP dialer configuration unnecessarily
- Do NOT break existing connectivity

**Recommended Agent Profile**:
- **Category**: `unspecified-low`
- **Reason**: Configuration tuning, minimal code changes
- **Skills**: None required

**Parallelization**:
- **Can Run In Parallel**: NO
- **Sequential**: After Task 6
- **Blocks**: Task 8
- **Blocked By**: Task 6

**References**:
- `pkg/http/http.go:9-25` - Current FastHTTP client configuration
- FastHTTP docs: Client configuration options
- Pattern: Right-sizing resource limits

**Acceptance Criteria**:
- [ ] Buffer sizes tuned for payment payload (~200-500 bytes)
- [ ] MaxConnsPerHost right-sized (likely 100-500)
- [ ] Read/Write timeouts appropriate
- [ ] `make build` succeeds
- [ ] `make up` starts services
- [ ] `make k6` passes

**Agent-Executed QA Scenario**:
```
Scenario: FastHTTP tuning validation
  Tool: Bash
  Steps:
    1. Check: pkg/http/http.go has optimized values
    2. Run: make build
    3. Assert: Build succeeds
    4. Run: make up
    5. Wait: 5s
    6. Run: docker exec rinha-haproxy ss -tunap | grep :8080
    7. Assert: Connection count reasonable
    8. Run: make k6
    9. Capture: k6 latency percentiles
    10. Assert: p99 latency maintained or improved
  Expected Result: Optimized connection usage, good latency
```

**Commit**: YES
- Message: `perf(http): tune FastHTTP client configuration`
- Files: `pkg/http/http.go`

---

### Task 8: Add Redis Connection Timeouts

**What to do**:
- Add explicit timeouts to Redis client configuration in `internal/database/redis.go`
- Add ReadTimeout, WriteTimeout, PoolTimeout
- Set reasonable defaults (5s read, 5s write, 10s pool)
- Make configurable via environment variables

**Must NOT do**:
- Do NOT set timeouts too low (would cause failures)
- Do NOT change existing connection pool size
- Do NOT modify other Redis settings
- Do NOT break existing functionality

**Recommended Agent Profile**:
- **Category**: `unspecified-low`
- **Reason**: Configuration addition, straightforward implementation
- **Skills**: None required

**Parallelization**:
- **Can Run In Parallel**: NO
- **Sequential**: After Task 7 (final task)
- **Blocks**: None
- **Blocked By**: Task 7

**References**:
- `internal/database/redis.go:25-28` - Redis client options
- go-redis docs: Timeout configuration
- Pattern: Defensive timeout configuration

**Acceptance Criteria**:
- [ ] ReadTimeout configured
- [ ] WriteTimeout configured
- [ ] PoolTimeout configured
- [ ] Optional env var configuration
- [ ] `make build` succeeds
- [ ] `make up` starts services
- [ ] `make k6` passes

**Agent-Executed QA Scenario**:
```
Scenario: Redis timeouts validation
  Tool: Bash
  Steps:
    1. Check: internal/database/redis.go has timeout fields
    2. Run: make build
    3. Assert: Build succeeds
    4. Run: make up
    5. Wait: 5s
    6. Run: docker exec rinha-redis redis-cli -s /sockets/redis.sock INFO stats
    7. Assert: Redis responsive
    8. Run: make k6
    9. Assert: No timeout errors in logs
    10. Assert: k6 completes successfully
  Expected Result: Redis operations complete without hanging
```

**Commit**: YES
- Message: `fix(redis): add connection timeouts for reliability`
- Files: `internal/database/redis.go`, `internal/config/config.go`

---

## Commit Strategy

| After Task | Message | Files |
|------------|---------|-------|
| 1 | `fix(models): correct JSON tag typo TtotalFee -> TotalFee` | `internal/models/summary.go` |
| 2 | `refactor(redis): share single client between database and queue` | `internal/services/queue.go`, `internal/services/payment.go`, `cmd/rinha/main.go` |
| 3 | `perf(redis): pipeline SavePayment operations` | `internal/database/redis.go` |
| 4 | `perf(queue): implement blocking channel-based queue` | `internal/services/payment.go`, `internal/services/queue.go` |
| 5 | `perf(concurrency): reduce unnecessary goroutine spawning` | `internal/services/payment.go` |
| 6 | `perf(memory): add sync.Pool for JSON buffer reuse` | `internal/services/payment.go`, `internal/services/queue.go` |
| 7 | `perf(http): tune FastHTTP client configuration` | `pkg/http/http.go` |
| 8 | `fix(redis): add connection timeouts for reliability` | `internal/database/redis.go`, `internal/config/config.go` |

---

## Success Criteria

### Verification Commands
```bash
# Build validation
make build
# Expected: Clean build, exit code 0

# Deployment validation  
make up
# Expected: All services start, 4 API containers healthy

# Performance validation
make k6
# Expected: Load test passes, p99 latency ≤ 0.45ms (maintained or improved)

# Error check
docker logs rinha-api01 2>&1 | grep -i error
# Expected: Minimal or no errors
```

### Final Checklist
- [ ] All 8 tasks complete
- [ ] All commits pushed
- [ ] k6 benchmarks pass for each task
- [ ] No regressions in functionality
- [ ] p99 latency maintained or improved
- [ ] Resource usage (CPU/Memory) stable or improved
