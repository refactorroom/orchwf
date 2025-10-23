# Lesson 5: Async and Parallel Execution

## Overview

Learn how to increase throughput by running independent steps concurrently.

## Marking Steps Async

Steps with `WithAsync(true)` run in parallel when their dependencies are satisfied.

```go
// After fetch completes, these three steps can run in parallel
process, _ := orchwf.NewStepBuilder("process", "Process", processExecutor).
    WithDependencies("fetch").
    WithAsync(true).
    Build()

enrich, _ := orchwf.NewStepBuilder("enrich", "Enrich", enrichExecutor).
    WithDependencies("fetch").
    WithAsync(true).
    Build()

analyze, _ := orchwf.NewStepBuilder("analyze", "Analyze", analyzeExecutor).
    WithDependencies("fetch").
    WithAsync(true).
    Build()
```

## Aggregating Results

Downstream steps can depend on multiple upstream async steps:

```go
aggregate, _ := orchwf.NewStepBuilder("aggregate", "Aggregate", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    // access outputs by step id
    p := input["process"].(map[string]interface{})
    e := input["enrich"].(map[string]interface{})
    a := input["analyze"].(map[string]interface{})
    return map[string]interface{}{"summary": map[string]interface{}{"p": p, "e": e, "a": a}}, nil
}).
    WithDependencies("process", "enrich", "analyze").
    Build()
```

## Async Workers

Control background concurrency at the orchestrator level:

```go
// Increase worker pool for async-heavy workloads
orchestrator := orchwf.NewOrchestratorWithAsyncWorkers(stateManager, 20)
```

## Design Tips

- Prefer async for I/O-bound, independent steps
- Keep outputs compact to reduce fan-in memory
- Set per-step timeouts to avoid tail latency
- Use retry policies suitable for each async branch

## Example Flow

1. `fetch` (sync) → gathers input
2. `process`, `enrich`, `analyze` (async in parallel)
3. `aggregate` (sync) depends on all three

This pattern minimizes wall-clock time: total ≈ max(async branches) + aggregation time.

## Next Steps
Next, we'll cover compensation (sagas) to safely undo side-effects on failures.


