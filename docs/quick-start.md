# Quick Start Guide

Get up and running with ORCHWF in just a few minutes!

## Installation

```bash
go get github.com/refactorroom/orchwf
```

## Basic Example

Here's a simple workflow that processes data:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/refactorroom/orchwf"
)

func main() {
    // 1. Create state manager
    stateManager := orchwf.NewInMemoryStateManager()
    
    // 2. Create orchestrator
    orchestrator := orchwf.NewOrchestrator(stateManager)
    
    // 3. Define a step
    step, err := orchwf.NewStepBuilder("process", "Process Data", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        data := input["data"].(string)
        fmt.Printf("Processing: %s\n", data)
        return map[string]interface{}{
            "result": "processed_" + data,
        }, nil
    }).Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. Create workflow
    workflow, err := orchwf.NewWorkflowBuilder("my_workflow", "My Workflow").
        AddStep(step).
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    // 5. Register and execute
    orchestrator.RegisterWorkflow(workflow)
    
    result, err := orchestrator.StartWorkflow(context.Background(), "my_workflow",
        map[string]interface{}{"data": "hello"}, nil)
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %+v\n", result)
}
```

## With Database

For production use, use database persistence:

```go
package main

import (
    "database/sql"
    "log"
    
    _ "github.com/lib/pq"
    "github.com/refactorroom/orchwf"
)

func main() {
    // Connect to database
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/db?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create database state manager
    stateManager := orchwf.NewDBStateManager(db)
    
    // Create orchestrator
    orchestrator := orchwf.NewOrchestrator(stateManager)
    
    // Run migrations (see migrations/001_create_orchwf_tables.sql)
    // ... run migrations ...
    
    // Use orchestrator...
}
```

## Next Steps

- Read [Lesson 1: Introduction](lessons/lesson-01-introduction.md) for a detailed walkthrough
- Check out [examples/](examples/) for more complex workflows
- See [API Reference](api-reference.md) for complete documentation
