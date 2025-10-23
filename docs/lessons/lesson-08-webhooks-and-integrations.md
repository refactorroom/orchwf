# Lesson 8: Webhooks and Integrations

## Overview

Call external HTTP services safely from workflow steps with retries, timeouts, and error handling.

## HTTP Calls in Steps

```go
client := &http.Client{Timeout: 10 * time.Second}

sendWebhook, _ := orchwf.NewStepBuilder("send_webhook", "Send Webhook", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    payload := map[string]interface{}{"event": "order.created", "id": input["order_id"]}
    body, _ := json.Marshal(payload)

    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://httpbin.org/post", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := client.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        b, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("webhook failed: %d %s", resp.StatusCode, string(b))
    }
    return map[string]interface{}{"webhook_sent": true}, nil
}).
    WithTimeout(15 * time.Second).
    WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
        WithMaxAttempts(3).
        WithInitialInterval(2 * time.Second).
        WithMultiplier(2.0).
        WithMaxInterval(12 * time.Second).
        WithRetryableErrors("timeout", "connection_refused").
        Build()).
    Build()
```

## Idempotency

- Include idempotency keys in requests (e.g., `Idempotency-Key` header)
- Make downstream operations safe to retry

## Sequencing and Fan-Out

- Use dependencies to control order (e.g., send inventory after order)
- Fan-out to multiple services, then fan-in to aggregate results

## Security Considerations

- Validate destinations and TLS
- Sign payloads (HMAC) and verify signatures on receivers
- Limit outbound timeouts and retries to avoid amplification

## Observability

- Add trace IDs to headers
- Log response codes and latencies per attempt

## Next Steps
You now have end-to-end skills: robust retries/timeouts, parallelism, compensation, persistence, and integrations.


