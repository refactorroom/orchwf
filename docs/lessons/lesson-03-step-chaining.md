# Lesson 3: Advanced Step Chaining and Parallel Execution

## Overview

In this lesson, we'll explore advanced workflow patterns including:
- Complex step dependencies
- Parallel execution
- Mixed sync/async workflows
- Error handling strategies

## Complex Dependency Patterns

### Fan-Out Pattern

Multiple steps depend on a single step:

```go
// Step 1: Fetch data
fetchStep, _ := orchwf.NewStepBuilder("fetch", "Fetch Data", fetchExecutor).Build()

// Step 2: Process data (depends on fetch)
processStep, _ := orchwf.NewStepBuilder("process", "Process Data", processExecutor).
    WithDependencies("fetch").
    Build()

// Step 3: Validate data (depends on fetch)
validateStep, _ := orchwf.NewStepBuilder("validate", "Validate Data", validateExecutor).
    WithDependencies("fetch").
    Build()

// Step 4: Save data (depends on both process and validate)
saveStep, _ := orchwf.NewStepBuilder("save", "Save Data", saveExecutor).
    WithDependencies("process", "validate").
    Build()
```

### Fan-In Pattern

Multiple steps feed into a single step:

```go
// Steps 1-3: Independent operations
step1, _ := orchwf.NewStepBuilder("step1", "Operation 1", executor1).Build()
step2, _ := orchwf.NewStepBuilder("step2", "Operation 2", executor2).Build()
step3, _ := orchwf.NewStepBuilder("step3", "Operation 3", executor3).Build()

// Step 4: Combines results from all previous steps
combineStep, _ := orchwf.NewStepBuilder("combine", "Combine Results", combineExecutor).
    WithDependencies("step1", "step2", "step3").
    Build()
```

## Parallel Execution

### Async Steps

Steps marked as async run in parallel when their dependencies are met:

```go
// These steps will run in parallel after "fetch" completes
processStep, _ := orchwf.NewStepBuilder("process", "Process Data", processExecutor).
    WithDependencies("fetch").
    WithAsync(true).
    Build()

validateStep, _ := orchwf.NewStepBuilder("validate", "Validate Data", validateExecutor).
    WithDependencies("fetch").
    WithAsync(true).
    Build()

notifyStep, _ := orchwf.NewStepBuilder("notify", "Send Notification", notifyExecutor).
    WithDependencies("fetch").
    WithAsync(true).
    Build()
```

### Mixed Sync/Async Workflow

```go
// Sync step - runs first
fetchStep, _ := orchwf.NewStepBuilder("fetch", "Fetch Data", fetchExecutor).Build()

// Async steps - run in parallel after fetch
processStep, _ := orchwf.NewStepBuilder("process", "Process Data", processExecutor).
    WithDependencies("fetch").
    WithAsync(true).
    Build()

validateStep, _ := orchwf.NewStepBuilder("validate", "Validate Data", validateExecutor).
    WithDependencies("fetch").
    WithAsync(true).
    Build()

// Sync step - waits for both async steps
saveStep, _ := orchwf.NewStepBuilder("save", "Save Data", saveExecutor).
    WithDependencies("process", "validate").
    Build()
```

## Error Handling Strategies

### Required vs Optional Steps

```go
// Required step - workflow fails if this fails
criticalStep, _ := orchwf.NewStepBuilder("critical", "Critical Operation", criticalExecutor).
    WithRequired(true).
    Build()

// Optional step - workflow continues if this fails
optionalStep, _ := orchwf.NewStepBuilder("optional", "Optional Operation", optionalExecutor).
    WithRequired(false).
    Build()
```

### Retry Policies

```go
// Aggressive retry for network operations
networkStep, _ := orchwf.NewStepBuilder("network", "Network Call", networkExecutor).
    WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
        WithMaxAttempts(5).
        WithInitialInterval(1 * time.Second).
        WithMultiplier(2.0).
        WithMaxInterval(60 * time.Second).
        WithRetryableErrors("timeout", "connection_refused").
        Build()).
    Build()

// Conservative retry for database operations
dbStep, _ := orchwf.NewStepBuilder("db", "Database Operation", dbExecutor).
    WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
        WithMaxAttempts(3).
        WithInitialInterval(500 * time.Millisecond).
        WithMultiplier(1.5).
        Build()).
    Build()
```

## Complete Example: E-commerce Order Processing

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/akkaraponph/orchwf"
)

func main() {
    stateManager := orchwf.NewInMemoryStateManager()
    orchestrator := orchwf.NewOrchestrator(stateManager)
    
    // Step 1: Validate order
    validateOrder, _ := orchwf.NewStepBuilder("validate_order", "Validate Order", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        orderID := input["order_id"].(string)
        fmt.Printf("Validating order: %s\n", orderID)
        time.Sleep(100 * time.Millisecond)
        return map[string]interface{}{
            "order_id": orderID,
            "valid":    true,
        }, nil
    }).WithDescription("Validate order data").
        WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
            WithMaxAttempts(3).
            Build()).
        Build()
    
    // Step 2: Check inventory (async)
    checkInventory, _ := orchwf.NewStepBuilder("check_inventory", "Check Inventory", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        orderID := input["order_id"].(string)
        fmt.Printf("Checking inventory for order: %s\n", orderID)
        time.Sleep(200 * time.Millisecond)
        return map[string]interface{}{
            "in_stock": true,
        }, nil
    }).WithDescription("Check product availability").
        WithDependencies("validate_order").
        WithAsync(true).
        WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
            WithMaxAttempts(5).
            WithInitialInterval(1 * time.Second).
            Build()).
        Build()
    
    // Step 3: Process payment (async)
    processPayment, _ := orchwf.NewStepBuilder("process_payment", "Process Payment", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        orderID := input["order_id"].(string)
        fmt.Printf("Processing payment for order: %s\n", orderID)
        time.Sleep(300 * time.Millisecond)
        return map[string]interface{}{
            "payment_id": "pay_123",
            "status":     "success",
        }, nil
    }).WithDescription("Process payment").
        WithDependencies("validate_order").
        WithAsync(true).
        WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
            WithMaxAttempts(3).
            WithInitialInterval(2 * time.Second).
            Build()).
        Build()
    
    // Step 4: Reserve inventory (async)
    reserveInventory, _ := orchwf.NewStepBuilder("reserve_inventory", "Reserve Inventory", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        orderID := input["order_id"].(string)
        fmt.Printf("Reserving inventory for order: %s\n", orderID)
        time.Sleep(150 * time.Millisecond)
        return map[string]interface{}{
            "reserved": true,
        }, nil
    }).WithDescription("Reserve products").
        WithDependencies("check_inventory").
        WithAsync(true).
        Build()
    
    // Step 5: Create shipment (depends on payment and inventory)
    createShipment, _ := orchwf.NewStepBuilder("create_shipment", "Create Shipment", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        orderID := input["order_id"].(string)
        paymentID := input["payment_id"].(string)
        fmt.Printf("Creating shipment for order: %s, payment: %s\n", orderID, paymentID)
        time.Sleep(200 * time.Millisecond)
        return map[string]interface{}{
            "shipment_id": "ship_123",
            "tracking":    "TRK123456",
        }, nil
    }).WithDescription("Create shipment").
        WithDependencies("process_payment", "reserve_inventory").
        Build()
    
    // Step 6: Send confirmation email (optional, async)
    sendEmail, _ := orchwf.NewStepBuilder("send_email", "Send Confirmation Email", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        orderID := input["order_id"].(string)
        fmt.Printf("Sending confirmation email for order: %s\n", orderID)
        time.Sleep(100 * time.Millisecond)
        return map[string]interface{}{
            "email_sent": true,
        }, nil
    }).WithDescription("Send confirmation email").
        WithDependencies("create_shipment").
        WithRequired(false).
        WithAsync(true).
        Build()
    
    // Create workflow
    workflow, _ := orchwf.NewWorkflowBuilder("order_processing", "Order Processing Workflow").
        WithDescription("Complete e-commerce order processing workflow").
        WithVersion("1.0.0").
        AddStep(validateOrder).
        AddStep(checkInventory).
        AddStep(processPayment).
        AddStep(reserveInventory).
        AddStep(createShipment).
        AddStep(sendEmail).
        Build()
    
    // Register and execute
    orchestrator.RegisterWorkflow(workflow)
    
    result, err := orchestrator.StartWorkflow(context.Background(), "order_processing",
        map[string]interface{}{
            "order_id": "order_123",
        },
        map[string]interface{}{
            "trace_id":    "trace_123",
            "business_id": "business_456",
        })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Order processing completed: %+v\n", result)
}
```

## Key Takeaways

1. **Dependencies**: Use `WithDependencies()` to control step execution order
2. **Parallel Execution**: Use `WithAsync(true)` for steps that can run in parallel
3. **Error Handling**: Use `WithRequired(false)` for optional steps
4. **Retry Policies**: Configure retries based on the type of operation
5. **Mixed Patterns**: Combine sync and async steps for optimal performance

## Next Steps

In the next lesson, we'll learn about database persistence and production deployment considerations.
