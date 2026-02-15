# Conservative Re-Application Plan: Individual Optimization Testing

## Strategy
Apply ONE optimization at a time, validate with full k6 suite, commit only if 0% failures.

## Execution Order (Safest First)

### Phase 1: Safety & Correctness
1. **FastHTTP Context Copy** (Task 2) - Critical bug fix
   - Risk: Low (correctness improvement)
   - Expected: May add slight latency but prevents data corruption
   
2. **Pre-allocated Slices** (Task 3) - Memory optimization
   - Risk: Very low
   - Expected: Minor improvement, no functional change

### Phase 2: Configuration Tuning  
3. **Dynamic Buffer Sizing** (Task 7) - Tuning adjustment
   - Risk: Low
   - Expected: Better memory usage

4. **Byte Slice Routing** (Task 12) - Zero-allocation routing
   - Risk: Low
   - Expected: Minor allocation reduction

### Phase 3: Algorithm Changes (High Risk)
5. **Sequential GetSummary** (Task 13) - Only if benchmarks show improvement
   - Risk: Medium
   - Test: Compare parallel vs sequential directly

6. **Health Check Optimization** (Task 11) - Startup improvement
   - Risk: Low
   - Expected: Faster startup

### Phase 4: Advanced (Skip for now)
- Redis Pool 500 (caused issues - needs investigation)
- Lua Scripts (caused issues - needs investigation)
- Ring Buffer (complex - defer to later)
- Batch Forwarding (requires upstream changes)

## Testing Protocol

For EACH optimization:
```bash
# 1. Apply single change
git checkout -b test/optimization-name

# 2. Build
make build
# Assert: Exit 0

# 3. Deploy  
make up
sleep 5

# 4. Test (run 3 times to verify consistency)
make k6
# Assert: http_req_failed = 0.00% (all 3 runs)
# Assert: payments_inconsistency = 0

# 5. If passed:
git add .
git commit -m "perf(...): optimization description"
git checkout main
git merge test/optimization-name

# 6. If failed:
git checkout main
git branch -D test/optimization-name
# Log failure, skip this optimization
```

## Success Criteria
- Each optimization must pass 3 consecutive k6 runs with 0% failures
- p99 latency must not regress more than 10%
- No payment inconsistencies allowed

## Rollback Strategy
Each optimization on separate branch - instant rollback by switching back to main.
