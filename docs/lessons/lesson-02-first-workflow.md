# Lesson 2: Creating Your First Workflow

## Overview

In this lesson, we'll create a more complex workflow with multiple steps and learn about step dependencies.

## Example: Data Processing Workflow

Let's create a workflow that:
1. Validates input data
2. Processes the data
3. Saves the results
4. Sends a notification

## Step-by-Step Implementation

### 1. Define the Steps

```go
// Step 1: Validate input data
validateStep, err := orchwf.NewStepBuilder("validate", "Validate Data", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    data, ok := input["data"].(string)
    if !ok || data == "" {
        return nil, fmt.Errorf("invalid data")
    }
    
    fmt.Printf("Validating data: %s\n", data)
    return map[string]interface{}{
        "validated_data": data,
    }, nil
}).WithDescription("Validate input data").
    WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
        WithMaxAttempts(3).
        WithInitialInterval(1 * time.Second).
        Build()).
    Build()

// Step 2: Process data
processStep, err := orchwf.NewStepBuilder("process", "Process Data", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    validatedData := input["validated_data"].(string)
    
    fmt.Printf("Processing data: %s\n", validatedData)
    processedData := strings.ToUpper(validatedData)
    
    return map[string]interface{}{
        "processed_data": processedData,
    }, nil
}).WithDescription("Process validated data").
    WithDependencies("validate").
    WithTimeout(5 * time.Second).
    Build()

// Step 3: Save results
saveStep, err := orchwf.NewStepBuilder("save", "Save Results", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    processedData := input["processed_data"].(string)
    
    fmt.Printf("Saving data: %s\n", processedData)
    // Simulate database save
    time.Sleep(100 * time.Millisecond)
    
    return map[string]interface{}{
        "saved": true,
        "data":  processedData,
    }, nil
}).WithDescription("Save processed data").
    WithDependencies("process").
    Build()

// Step 4: Send notification (optional)
notifyStep, err := orchwf.NewStepBuilder("notify", "Send Notification", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    fmt.Println("Sending notification...")
    // Simulate notification sending
    time.Sleep(200 * time.Millisecond)
    
    return map[string]interface{}{
        "notification_sent": true,
    }, nil
}).WithDescription("Send notification about completed processing").
    WithDependencies("save").
    WithRequired(false). // This step is optional
    WithAsync(true).     // Run asynchronously
    Build()
```

### 2. Create the Workflow

```go
workflow, err := orchwf.NewWorkflowBuilder("data_processing", "Data Processing Workflow").
    WithDescription("A workflow that validates, processes, and saves data").
    WithVersion("1.0.0").
    AddStep(validateStep).
    AddStep(processStep).
    AddStep(saveStep).
    AddStep(notifyStep).
    Build()
```

### 3. Execute the Workflow

```go
// Create orchestrator
stateManager := orchwf.NewInMemoryStateManager()
orchestrator := orchwf.NewOrchestrator(stateManager)

// Register workflow
orchestrator.RegisterWorkflow(workflow)

// Execute workflow
result, err := orchestrator.StartWorkflow(context.Background(), "data_processing",
    map[string]interface{}{
        "data": "hello world",
    },
    map[string]interface{}{
        "trace_id": "trace_123",
    })

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Workflow completed: %+v\n", result)
```

## Key Concepts Explained

### Step Dependencies

The `WithDependencies()` method defines which steps must complete before this step can run:

```go
.WithDependencies("validate")  // This step depends on "validate" step
.WithDependencies("step1", "step2")  // This step depends on both "step1" and "step2"
```

### Retry Policies

Configure how many times a step should retry on failure:

```go
.WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
    WithMaxAttempts(3).                    // Retry up to 3 times
    WithInitialInterval(1 * time.Second).  // Wait 1 second before first retry
    WithMultiplier(2.0).                   // Double the wait time each retry
    WithMaxInterval(30 * time.Second).     // Maximum wait time
    Build())
```

### Timeouts

Set a maximum execution time for a step:

```go
.WithTimeout(5 * time.Second)
```

### Optional Steps

Steps that don't stop the workflow if they fail:

```go
.WithRequired(false)
```

### Async Steps

Steps that run asynchronously (in parallel with other ready steps):

```go
.WithAsync(true)
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"
    
    "github.com/akkaraponph/orchwf"
)

func main() {
    // Create state manager and orchestrator
    stateManager := orchwf.NewInMemoryStateManager()
    orchestrator := orchwf.NewOrchestrator(stateManager)
    
    // Define steps (as shown above)
    // ... step definitions ...
    
    // Create and register workflow
    workflow, _ := orchwf.NewWorkflowBuilder("data_processing", "Data Processing Workflow").
        WithDescription("A workflow that validates, processes, and saves data").
        AddStep(validateStep).
        AddStep(processStep).
        AddStep(saveStep).
        AddStep(notifyStep).
        Build()
    
    orchestrator.RegisterWorkflow(workflow)
    
    // Execute workflow
    result, err := orchestrator.StartWorkflow(context.Background(), "data_processing",
        map[string]interface{}{"data": "hello world"}, nil)
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Workflow completed: %+v\n", result)
}
```

## Next Steps

In the next lesson, we'll learn about advanced step chaining and parallel execution patterns.
