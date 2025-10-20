# Lesson 4: Retries and Timeouts

## Overview

In this lesson, you'll learn how to make workflows resilient using retry policies and timeouts.

## Retry Policies

### Why retries?
- Network calls can fail transiently (timeouts, connection resets)
- External services may be temporarily unavailable
- Retrying with backoff improves reliability without overwhelming services

### Configuring retries

```go
// Aggressive retry for flaky network endpoints
step, _ := orchwf.NewStepBuilder("call_api", "Call External API", apiExecutor).
    WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
        WithMaxAttempts(5).                // up to 5 attempts
        WithInitialInterval(1 * time.Second). // start with 1s
        WithMultiplier(2.0).               // exponential backoff
        WithMaxInterval(30 * time.Second). // cap the backoff
        WithRetryableErrors("timeout", "connection_refused"). // only retry on these patterns
        Build()).
    Build()
```

### Choosing values
- MaxAttempts: 3–5 for most cases; higher for crucial but idempotent operations
- InitialInterval: 500ms–2s depending on latency
- Multiplier: 1.5–2.0 for exponential backoff
- MaxInterval: keep within user-facing SLAs
- RetryableErrors: restrict to transient errors to avoid retry storms

## Timeouts

### Why timeouts?
- Prevent hung steps from blocking the workflow
- Bound latency for user-facing operations

### Configuring timeouts

```go
// Ensure the step completes within 5 seconds
step, _ := orchwf.NewStepBuilder("db_query", "Query Database", queryExecutor).
    WithTimeout(5 * time.Second).
    Build()
```

### Tips
- Set timeouts slightly above expected p95 latency
- Combine with retries for robust behavior
- For parallel fan-out, use per-step timeouts to isolate slow branches

## Putting It Together

```go
// Example: robust notification sender
notify, _ := orchwf.NewStepBuilder("notify", "Send Notification", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    // call external service using ctx for deadline/cancel
    // ...
    return map[string]interface{}{"sent": true}, nil
}).
    WithTimeout(3 * time.Second).
    WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
        WithMaxAttempts(4).
        WithInitialInterval(800 * time.Millisecond).
        WithMultiplier(1.8).
        WithMaxInterval(8 * time.Second).
        WithRetryableErrors("timeout", "429", "5xx").
        Build()).
    Build()
```

## Diagnostics and Tuning Checklist
- Log attempt count and error cause per retry
- Track per-step duration and timeout rate
- Alert on sustained high retry rates
- Ensure operations are idempotent if they can be retried

## Next Steps
In the next lesson, we'll cover async and parallel execution patterns for higher throughput.


