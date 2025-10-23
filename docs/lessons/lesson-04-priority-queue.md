# Lesson 4: Priority Queue Feature

Welcome to Lesson 4 of the ORCHWF tutorial series! In this lesson, you'll learn how to use the priority queue feature to control the execution order of workflow steps.

## What You'll Learn

- Understanding priority-based execution
- Setting step priorities
- Combining priority with dependencies
- Best practices for priority usage
- Real-world examples

## Prerequisites

- Completed Lesson 1: Introduction
- Completed Lesson 2: First Workflow
- Completed Lesson 3: Step Chaining
- Basic understanding of workflow orchestration

## Introduction to Priority Queues

In previous lessons, we learned that steps execute based on their dependencies. However, sometimes you need more control over execution order. The priority queue feature allows you to specify which steps should execute first, even when multiple steps are ready to run.

### Why Use Priority Queues?

1. **Performance Optimization**: Execute critical steps first
2. **Resource Management**: Prioritize steps that free up resources
3. **User Experience**: Show results to users as quickly as possible
4. **Business Logic**: Ensure important operations happen before less critical ones

## Basic Priority Usage

### Setting Step Priority

```go
// High priority step (executes first)
highPriorityStep, err := orchwf.NewStepBuilder("critical", "Critical Processing", executor).
    WithPriority(10).  // Higher number = higher priority
    Build()

// Normal priority step (default)
normalStep, err := orchwf.NewStepBuilder("normal", "Normal Processing", executor).
    // No WithPriority() call - defaults to 0
    Build()

// Low priority step (executes last)
lowPriorityStep, err := orchwf.NewStepBuilder("background", "Background Task", executor).
    WithPriority(-5).  // Negative number = low priority
    Build()
```

### Priority Levels

- **Critical**: 10-20 (system-critical operations)
- **High**: 5-9 (important business logic)
- **Normal**: 0-4 (standard processing)
- **Low**: -10 to -1 (background tasks)
- **Very Low**: -20 and below (maintenance)

## Hands-On Example

Let's create a workflow that demonstrates priority-based execution:

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
    // Create state manager and orchestrator
    stateManager := orchwf.NewInMemoryStateManager()
    orchestrator := orchwf.NewOrchestrator(stateManager)
    
    // Define steps with different priorities
    steps := createPrioritySteps()
    
    // Build workflow
    workflow, err := orchwf.NewWorkflowBuilder("priority_demo", "Priority Demo").
        AddStep(steps["critical"]).
        AddStep(steps["normal"]).
        AddStep(steps["background"]).
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Register and execute
    orchestrator.RegisterWorkflow(workflow)
    
    fmt.Println("Executing workflow with priority-based ordering...")
    result, err := orchestrator.StartWorkflow(context.Background(), "priority_demo", 
        map[string]interface{}{"data": "test"}, nil)
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Workflow completed: %+v\n", result)
}

func createPrioritySteps() map[string]*orchwf.StepDefinition {
    steps := make(map[string]*orchwf.StepDefinition)
    
    // Critical step (priority 20)
    criticalStep, _ := orchwf.NewStepBuilder("critical", "Critical Check", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("🔥 CRITICAL: System check...")
        time.Sleep(100 * time.Millisecond)
        return map[string]interface{}{"status": "healthy"}, nil
    }).
        WithPriority(20).
        Build()
    steps["critical"] = criticalStep
    
    // Normal step (priority 0)
    normalStep, _ := orchwf.NewStepBuilder("normal", "Normal Processing", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("📋 NORMAL: Processing...")
        time.Sleep(150 * time.Millisecond)
        return map[string]interface{}{"result": "processed"}, nil
    }).
        WithPriority(0).
        Build()
    steps["normal"] = normalStep
    
    // Background step (priority -10)
    backgroundStep, _ := orchwf.NewStepBuilder("background", "Background Task", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("🔄 BACKGROUND: Cleanup...")
        time.Sleep(200 * time.Millisecond)
        return map[string]interface{}{"cleanup": "done"}, nil
    }).
        WithPriority(-10).
        Build()
    steps["background"] = backgroundStep
    
    return steps
}
```

### Expected Output

```
🔥 CRITICAL: System check...
📋 NORMAL: Processing...
🔄 BACKGROUND: Cleanup...
Workflow completed: &{Success:true ...}
```

Notice how the critical step executes first, even though it was added first to the workflow. The priority system overrides the definition order.

## Priority with Dependencies

Priority works alongside dependencies. Steps with higher priority execute first, but only among steps that have their dependencies satisfied.

### Example: Priority + Dependencies

```go
// Step A: No dependencies, priority 5
stepA, _ := orchwf.NewStepBuilder("step_a", "Step A", executorA).
    WithPriority(5).
    Build()

// Step B: No dependencies, priority 1 (lower than A)
stepB, _ := orchwf.NewStepBuilder("step_b", "Step B", executorB).
    WithPriority(1).
    Build()

// Step C: Depends on A, priority 10 (highest, but waits for A)
stepC, _ := orchwf.NewStepBuilder("step_c", "Step C", executorC).
    WithDependencies("step_a").
    WithPriority(10).
    Build()
```

**Execution Order**: A → B → C
- A and B execute first (A has higher priority)
- C waits for A to complete, then executes

## Real-World Example: E-commerce Order Processing

Let's create a realistic example for an e-commerce order processing workflow:

```go
func createEcommerceWorkflow() *orchwf.WorkflowDefinition {
    // Critical: Payment processing (highest priority)
    paymentStep, _ := orchwf.NewStepBuilder("payment", "Process Payment", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("💳 Processing payment...")
        time.Sleep(200 * time.Millisecond)
        return map[string]interface{}{"payment_id": "pay_123", "status": "success"}, nil
    }).
        WithPriority(20).
        WithRequired(true).
        Build()
    
    // High: Inventory check
    inventoryStep, _ := orchwf.NewStepBuilder("inventory", "Check Inventory", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("📦 Checking inventory...")
        time.Sleep(150 * time.Millisecond)
        return map[string]interface{}{"in_stock": true, "quantity": 5}, nil
    }).
        WithPriority(10).
        WithDependencies("payment").
        Build()
    
    // Normal: Order confirmation
    confirmationStep, _ := orchwf.NewStepBuilder("confirmation", "Send Confirmation", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("📧 Sending confirmation email...")
        time.Sleep(100 * time.Millisecond)
        return map[string]interface{}{"email_sent": true}, nil
    }).
        WithPriority(5).
        WithDependencies("inventory").
        Build()
    
    // Low: Analytics tracking
    analyticsStep, _ := orchwf.NewStepBuilder("analytics", "Track Analytics", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("📊 Tracking analytics...")
        time.Sleep(50 * time.Millisecond)
        return map[string]interface{}{"tracked": true}, nil
    }).
        WithPriority(-5).
        WithDependencies("confirmation").
        WithRequired(false).  // Don't fail workflow if analytics fails
        Build()
    
    // Build workflow
    workflow, _ := orchwf.NewWorkflowBuilder("ecommerce_order", "E-commerce Order Processing").
        AddStep(paymentStep).
        AddStep(inventoryStep).
        AddStep(confirmationStep).
        AddStep(analyticsStep).
        Build()
    
    return workflow
}
```

## Async Steps with Priority

Priority also works with async steps. Among ready async steps, higher priority steps are scheduled first.

```go
// High-priority async step
asyncHigh, _ := orchwf.NewStepBuilder("async_high", "High Priority Async", executor).
    WithAsync(true).
    WithPriority(10).
    Build()

// Low-priority async step
asyncLow, _ := orchwf.NewStepBuilder("async_low", "Low Priority Async", executor).
    WithAsync(true).
    WithPriority(1).
    Build()
```

## Best Practices

### 1. Use Consistent Priority Ranges

```go
// Good: Clear priority ranges
const (
    PriorityCritical = 20
    PriorityHigh     = 10
    PriorityNormal   = 0
    PriorityLow      = -10
    PriorityBackground = -20
)

// Usage
step.WithPriority(PriorityCritical)
```

### 2. Document Priority Levels

```go
// Document why a step has high priority
step, _ := orchwf.NewStepBuilder("payment", "Process Payment", executor).
    WithPriority(20).  // Critical: Must complete before other operations
    WithDescription("Processes customer payment - highest priority for business continuity").
    Build()
```

### 3. Use Priority for Performance

```go
// Prioritize steps that free up resources
cleanupStep, _ := orchwf.NewStepBuilder("cleanup", "Cleanup Resources", cleanupExecutor).
    WithPriority(15).  // High priority to free up memory/connections
    Build()
```

### 4. Combine with Required Flag

```go
// Critical step that must succeed
criticalStep, _ := orchwf.NewStepBuilder("critical", "Critical Operation", executor).
    WithPriority(20).
    WithRequired(true).  // Workflow fails if this step fails
    Build()

// Background step that can fail
backgroundStep, _ := orchwf.NewStepBuilder("background", "Background Task", executor).
    WithPriority(-10).
    WithRequired(false).  // Workflow continues even if this fails
    Build()
```

## Common Patterns

### 1. Critical Path Optimization

```go
// Ensure critical business logic executes first
paymentStep.WithPriority(20)
inventoryStep.WithPriority(15)
shippingStep.WithPriority(10)
```

### 2. Resource Management

```go
// Prioritize steps that free up resources
cleanupStep.WithPriority(15)
cacheStep.WithPriority(10)
loggingStep.WithPriority(-5)
```

### 3. User Experience

```go
// Show results to users quickly
displayStep.WithPriority(20)
notificationStep.WithPriority(15)
analyticsStep.WithPriority(-10)
```

## Troubleshooting

### Issue: Steps not executing in expected order

**Cause**: Dependencies not satisfied
**Solution**: Check that all dependencies are completed before priority takes effect

### Issue: Priority not working

**Cause**: Using old version without priority support
**Solution**: Update to latest ORCHWF version

### Issue: Database errors

**Cause**: Missing priority column
**Solution**: Run database migration to add priority column

## Exercise

Create a workflow for a content management system with the following requirements:

1. **Critical**: Validate user permissions (priority 20)
2. **High**: Process content (priority 10)
3. **Normal**: Update metadata (priority 0)
4. **Low**: Generate thumbnails (priority -5)
5. **Background**: Update search index (priority -10)

Make sure the workflow respects dependencies and priorities.

## Next Steps

- Learn about retry policies in Lesson 5
- Explore advanced workflow patterns
- Study performance optimization techniques

## Summary

The priority queue feature gives you fine-grained control over step execution order. Use it to:

- Optimize workflow performance
- Ensure critical operations execute first
- Improve user experience
- Manage resources efficiently

Remember: Priority only affects steps that have their dependencies satisfied. Plan your workflow dependencies carefully to get the most benefit from priority-based execution.
