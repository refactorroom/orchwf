# Lesson 1: Introduction to ORCHWF

## What is ORCHWF?

ORCHWF is a Go package for orchestrating complex workflows. It allows you to:

- Define workflows as a series of steps
- Execute steps synchronously or asynchronously
- Handle step dependencies
- Retry failed steps
- Persist workflow state
- Track workflow events

## Key Concepts

### Workflow
A workflow is a collection of steps that are executed in a specific order based on dependencies.

### Step
A step is a single unit of work that can be executed. Each step has:
- An executor function
- Dependencies on other steps
- Retry policy
- Timeout settings
- Async/sync execution mode

### State Manager
The state manager handles persistence of workflow and step states. Two options:
- **In-Memory**: Fast, but data is lost on restart
- **Database**: Persistent, production-ready

## Basic Example

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
    step, err := orchwf.NewStepBuilder("hello", "Hello World", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("Hello, World!")
        return map[string]interface{}{
            "message": "Hello, World!",
        }, nil
    }).Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. Create workflow
    workflow, err := orchwf.NewWorkflowBuilder("hello_workflow", "Hello Workflow").
        AddStep(step).
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    // 5. Register workflow
    orchestrator.RegisterWorkflow(workflow)
    
    // 6. Execute workflow
    result, err := orchestrator.StartWorkflow(context.Background(), "hello_workflow", nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Workflow completed: %+v\n", result)
}
```

## Step Executor Function

The step executor is a function that:
- Takes a context and input map
- Returns an output map and error
- Performs the actual work

```go
func myStepExecutor(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    // Get input data
    data := input["data"].(string)
    
    // Do some work
    result := processData(data)
    
    // Return output
    return map[string]interface{}{
        "result": result,
    }, nil
}
```

## Next Steps

In the next lesson, we'll learn how to create more complex workflows with multiple steps and dependencies.
