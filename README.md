# ORCHWF - Workflow Orchestration Package

ORCHWF is a powerful Go package for orchestrating complex workflows with support for both synchronous and asynchronous execution patterns. It provides flexible state management options (in-memory or database) and uses Go routines for async execution instead of external queue systems.

## Features

- **Dual Execution Modes**: Synchronous and asynchronous workflow execution
- **Flexible State Management**: In-memory or database persistence
- **Dependency Management**: Step dependencies with parallel execution support
- **Retry Logic**: Configurable retry policies with exponential backoff
- **Event System**: Comprehensive workflow and step event tracking
- **Transaction Support**: Database transaction support for consistency
- **Builder Pattern**: Fluent API for building workflows and steps
- **Pure Go**: Uses only Go standard library (no external dependencies except UUID)

## Installation

```bash
go get github.com/refactorroom/orchwf
```

## Quick Start

### 1. Basic Workflow

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/refactorroom/orchwf"
)

func main() {
    // Create in-memory state manager
    stateManager := orchwf.NewInMemoryStateManager()
    
    // Create orchestrator
    orchestrator := orchwf.NewOrchestrator(stateManager)
    
    // Define a simple step
    step1, _ := orchwf.NewStepBuilder("step1", "Process Data", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("Processing data...")
        return map[string]interface{}{
            "result": "processed",
        }, nil
    }).Build()
    
    // Build workflow
    workflow, _ := orchwf.NewWorkflowBuilder("simple_workflow", "Simple Workflow").
        AddStep(step1).
        Build()
    
    // Register workflow
    orchestrator.RegisterWorkflow(workflow)
    
    // Execute workflow
    result, err := orchestrator.StartWorkflow(context.Background(), "simple_workflow", 
        map[string]interface{}{"data": "test"}, nil)
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Workflow completed: %+v\n", result)
}
```

### 2. Database State Management

```go
package main

import (
    "database/sql"
    "log"
    
    _ "github.com/lib/pq" // PostgreSQL driver
    "github.com/refactorroom/orchwf"
)

func main() {
    // Connect to database
    db, err := sql.Open("postgres", "postgres://user:password@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create database state manager
    stateManager := orchwf.NewDBStateManager(db)
    
    // Create orchestrator
    orchestrator := orchwf.NewOrchestrator(stateManager)
    
    // Run migrations (you need to implement this)
    // runMigrations(db)
    
    // Use orchestrator...
}
```

## State Management Options

### In-Memory State Manager

Perfect for development, testing, or simple workflows:

```go
stateManager := orchwf.NewInMemoryStateManager()
```

**Pros:**
- Fast execution
- No database setup required
- Perfect for testing

**Cons:**
- Data lost on restart
- Not suitable for production

### Database State Manager

Production-ready persistence using standard `database/sql`:

```go
db, _ := sql.Open("postgres", connectionString)
stateManager := orchwf.NewDBStateManager(db)
```

**Pros:**
- Persistent storage
- Transaction support
- Production ready
- Scalable

**Cons:**
- Requires database setup
- Slightly slower than in-memory

## Execution Patterns

### Synchronous Execution

All steps run in sequence, waiting for each to complete:

```go
result, err := orchestrator.StartWorkflow(ctx, "workflow_id", input, metadata)
```

### Asynchronous Execution

Workflow starts immediately and runs in background:

```go
workflowID, err := orchestrator.StartWorkflowAsync(ctx, "workflow_id", input, metadata)
```

### Mixed Execution

Steps can be marked as async within a workflow:

```go
step1, _ := orchwf.NewStepBuilder("step1", "Sync Step", syncExecutor).
    WithAsync(false).
    Build()

step2, _ := orchwf.NewStepBuilder("step2", "Async Step", asyncExecutor).
    WithAsync(true).
    WithDependencies("step1").
    Build()
```

## Advanced Features

### Retry Policies

```go
retryPolicy := orchwf.NewRetryPolicyBuilder().
    WithMaxAttempts(3).
    WithInitialInterval(1 * time.Second).
    WithMultiplier(2.0).
    WithMaxInterval(30 * time.Second).
    WithRetryableErrors("network_error", "timeout").
    Build()

step, _ := orchwf.NewStepBuilder("step", "Name", executor).
    WithRetryPolicy(retryPolicy).
    Build()
```

### Step Dependencies

```go
step1, _ := orchwf.NewStepBuilder("step1", "First Step", executor1).Build()
step2, _ := orchwf.NewStepBuilder("step2", "Second Step", executor2).
    WithDependencies("step1").
    Build()
step3, _ := orchwf.NewStepBuilder("step3", "Third Step", executor3).
    WithDependencies("step1", "step2").
    Build()
```

### Timeouts

```go
step, _ := orchwf.NewStepBuilder("step", "Name", executor).
    WithTimeout(30 * time.Second).
    Build()
```

### Non-Required Steps

Steps that don't stop the workflow on failure:

```go
step, _ := orchwf.NewStepBuilder("step", "Optional Step", executor).
    WithRequired(false).
    Build()
```

## Database Setup

### PostgreSQL

Run the migration script:

```sql
-- See migrations/001_create_orchwf_tables.sql
```

### Other Databases

The package uses standard SQL, so it should work with any database that supports:
- JSON/JSONB columns
- Standard SQL syntax
- Transactions

## API Reference

### Orchestrator

- `NewOrchestrator(stateManager)` - Create new orchestrator
- `NewOrchestratorWithAsyncWorkers(stateManager, workers)` - Create with custom worker count
- `RegisterWorkflow(workflow)` - Register a workflow definition
- `StartWorkflow(ctx, id, input, metadata)` - Start workflow synchronously
- `StartWorkflowAsync(ctx, id, input, metadata)` - Start workflow asynchronously
- `ResumeWorkflow(ctx, instanceID)` - Resume a failed workflow
- `GetWorkflowStatus(ctx, instanceID)` - Get workflow status
- `ListWorkflows(ctx, filters, limit, offset)` - List workflows

### State Managers

- `NewInMemoryStateManager()` - Create in-memory state manager
- `NewDBStateManager(db)` - Create database state manager

### Builders

- `NewWorkflowBuilder(id, name)` - Create workflow builder
- `NewStepBuilder(id, name, executor)` - Create step builder
- `NewRetryPolicyBuilder()` - Create retry policy builder

## Examples

See the `examples/` directory for complete working examples.

## License

MIT License - see LICENSE file for details.
