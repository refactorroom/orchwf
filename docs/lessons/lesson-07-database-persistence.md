# Lesson 7: Database Persistence

## Overview

Persist workflow state to a database for durability, observability, and recovery.

## State Managers

- In-memory (default): simple, fast, non-durable
- Database (`DBStateManager`): durable, production-ready

## Quick Setup with Migration Package

```go
db, _ := sql.Open("postgres", "postgres://user:pass@localhost/db?sslmode=disable")
if err := migrate.QuickSetup(db); err != nil { panic(err) }

state := orchwf.NewDBStateManager(db)
orch := orchwf.NewOrchestrator(state)
```

## What Gets Persisted

- Workflow instances: status, input, output, context, timing
- Step instances: status, input, output, retries, timing
- Events: audit trail for workflow and step lifecycle

## Listing and Filtering Workflows

```go
filters := map[string]interface{}{"status": orchwf.WorkflowStatusCompleted}
workflows, total, err := state.ListWorkflows(ctx, filters, 50, 0)
_ = workflows; _ = total; _ = err
```

## Operations and Transactions

Use `WithTransaction` for atomic multi-write operations inside a step when needed.

## Monitoring

- Query `orchwf_workflow_instances` for duration and error rates
- Use `orchwf_workflow_events` for detailed timelines

## Best Practices

- Run migrations on deploy
- Index by status, created_at, business IDs for fast lookups
- Store trace/correlation IDs for cross-service observability

## Next Steps
Next, we’ll integrate workflows with external systems via webhooks.


