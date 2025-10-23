# Priority Queue Feature

ORCHWF now supports priority-based step execution, allowing you to control the order in which steps are executed based on their priority levels. This feature is particularly useful for workflows where certain steps need to be executed before others, regardless of their position in the workflow definition.

## Overview

The priority queue feature allows you to assign priority levels to workflow steps. Steps with higher priority values are executed before steps with lower priority values, even if they appear later in the workflow definition.

### Key Features

- **Priority-based execution**: Steps are sorted by priority before execution
- **Dependency respect**: Priority only affects execution order among steps that have their dependencies satisfied
- **Database persistence**: Priority values are stored in the database for persistence
- **Backward compatibility**: Existing workflows continue to work (default priority is 0)

## How It Works

1. **Step Definition**: Each step can have a priority value assigned
2. **Execution Order**: When multiple steps are ready to execute, they are sorted by priority (highest first)
3. **Dependency Resolution**: Priority only affects steps that have all their dependencies satisfied
4. **Database Storage**: Priority values are persisted in the database

## Usage

### Basic Priority Assignment

```go
// Create a high-priority step
highPriorityStep, err := orchwf.NewStepBuilder("critical_step", "Critical Processing", executor).
    WithPriority(10).  // High priority
    Build()

// Create a low-priority step
lowPriorityStep, err := orchwf.NewStepBuilder("background_step", "Background Task", executor).
    WithPriority(1).    // Low priority
    Build()

// Create a normal priority step (default)
normalStep, err := orchwf.NewStepBuilder("normal_step", "Normal Processing", executor).
    // No WithPriority() call - defaults to 0
    Build()
```

### Priority Levels

- **High Priority**: 10 and above (executed first)
- **Normal Priority**: 0 (default)
- **Low Priority**: Negative values (executed last)

### Complete Example

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
    // Create state manager
    stateManager := orchwf.NewInMemoryStateManager()
    orchestrator := orchwf.NewOrchestrator(stateManager)
    
    // Define high-priority step (executes first)
    criticalStep, err := orchwf.NewStepBuilder("critical", "Critical Data Processing", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("🔥 CRITICAL: Processing critical data...")
        time.Sleep(500 * time.Millisecond)
        return map[string]interface{}{"critical_result": "processed"}, nil
    }).
        WithPriority(10).
        WithDescription("High-priority critical processing").
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Define normal-priority step
    normalStep, err := orchwf.NewStepBuilder("normal", "Normal Processing", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("📋 NORMAL: Processing normal data...")
        time.Sleep(300 * time.Millisecond)
        return map[string]interface{}{"normal_result": "processed"}, nil
    }).
        WithPriority(0).  // Default priority
        WithDescription("Normal processing").
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Define low-priority step (executes last)
    backgroundStep, err := orchwf.NewStepBuilder("background", "Background Task", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        fmt.Println("🔄 BACKGROUND: Running background task...")
        time.Sleep(200 * time.Millisecond)
        return map[string]interface{}{"background_result": "completed"}, nil
    }).
        WithPriority(-5).  // Low priority
        WithDescription("Low-priority background task").
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Build workflow
    workflow, err := orchwf.NewWorkflowBuilder("priority_workflow", "Priority-based Workflow").
        AddStep(criticalStep).
        AddStep(normalStep).
        AddStep(backgroundStep).
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Register and execute
    orchestrator.RegisterWorkflow(workflow)
    
    result, err := orchestrator.StartWorkflow(context.Background(), "priority_workflow", 
        map[string]interface{}{"data": "test"}, nil)
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Workflow completed: %+v\n", result)
}
```

### Expected Output

```
🔥 CRITICAL: Processing critical data...
📋 NORMAL: Processing normal data...
🔄 BACKGROUND: Running background task...
Workflow completed: &{Success:true ...}
```

## Advanced Usage

### Priority with Dependencies

Priority works in conjunction with dependencies. Steps with higher priority will execute first, but only among steps that have their dependencies satisfied.

```go
// Step A has no dependencies, priority 5
stepA, _ := orchwf.NewStepBuilder("step_a", "Step A", executorA).
    WithPriority(5).
    Build()

// Step B depends on Step A, priority 10 (higher than A)
stepB, _ := orchwf.NewStepBuilder("step_b", "Step B", executorB).
    WithDependencies("step_a").
    WithPriority(10).
    Build()

// Step C has no dependencies, priority 1 (lower than A)
stepC, _ := orchwf.NewStepBuilder("step_c", "Step C", executorC).
    WithPriority(1).
    Build()
```

**Execution Order**: A → C → B (A and C execute first based on priority, then B executes after A completes)

### Priority with Async Steps

Priority also works with async steps. Among ready async steps, higher priority steps will be scheduled first.

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

## Database Schema

The priority feature adds a `priority` column to the `orchwf_step_instances` table:

```sql
ALTER TABLE orchwf_step_instances 
ADD COLUMN priority INT DEFAULT 0;

CREATE INDEX idx_orchwf_step_instances_priority 
ON orchwf_step_instances(workflow_inst_id, priority DESC);
```

## Migration

If you're upgrading from a previous version, run the database migration to add the priority column:

```sql
-- Add priority column to existing step instances
ALTER TABLE orchwf_step_instances 
ADD COLUMN priority INT DEFAULT 0;

-- Create index for efficient priority-based queries
CREATE INDEX idx_orchwf_step_instances_priority 
ON orchwf_step_instances(workflow_inst_id, priority DESC);
```

## Best Practices

### Priority Level Guidelines

- **Critical Steps**: Priority 10-20 (system-critical operations)
- **High Priority**: Priority 5-9 (important business logic)
- **Normal Priority**: Priority 0-4 (standard processing)
- **Low Priority**: Priority -10 to -1 (background tasks, cleanup)
- **Very Low Priority**: Priority -20 and below (maintenance tasks)

### Use Cases

1. **Critical Path Optimization**: Ensure critical business logic executes first
2. **Resource Management**: Prioritize steps that free up resources
3. **User Experience**: Execute user-facing operations before background tasks
4. **Error Recovery**: Prioritize recovery steps over normal processing
5. **Performance**: Execute expensive operations last

### Example: E-commerce Order Processing

```go
// Critical: Payment processing (highest priority)
paymentStep, _ := orchwf.NewStepBuilder("payment", "Process Payment", paymentExecutor).
    WithPriority(20).
    WithRequired(true).
    Build()

// High: Inventory check
inventoryStep, _ := orchwf.NewStepBuilder("inventory", "Check Inventory", inventoryExecutor).
    WithPriority(10).
    WithDependencies("payment").
    Build()

// Normal: Order confirmation
confirmationStep, _ := orchwf.NewStepBuilder("confirmation", "Send Confirmation", confirmationExecutor).
    WithPriority(5).
    WithDependencies("inventory").
    Build()

// Low: Analytics tracking
analyticsStep, _ := orchwf.NewStepBuilder("analytics", "Track Analytics", analyticsExecutor).
    WithPriority(-5).
    WithDependencies("confirmation").
    WithRequired(false).  // Don't fail workflow if analytics fails
    Build()
```

## Performance Considerations

- **Index Usage**: The priority index ensures efficient sorting
- **Memory Usage**: Priority sorting adds minimal overhead
- **Database Queries**: Priority-based queries are optimized with indexes
- **Concurrency**: Priority works seamlessly with async execution

## Troubleshooting

### Common Issues

1. **Steps not executing in expected order**: Check that dependencies are satisfied
2. **Priority not working**: Ensure you're using the latest version with priority support
3. **Database errors**: Run the migration to add the priority column

### Debugging

Enable debug logging to see step execution order:

```go
// The orchestrator will log step execution order including priority values
// Check your logs for messages like:
// "Executing step 'step_name' with priority 10"
```

## API Reference

### StepBuilder Methods

```go
// Set step priority
func (b *StepBuilder) WithPriority(priority int) *StepBuilder
```

### StepDefinition Fields

```go
type StepDefinition struct {
    // ... other fields ...
    Priority int  // Higher number = higher priority (default: 0)
}
```

### Database Model

```go
type ORCHStepInstance struct {
    // ... other fields ...
    Priority int  // Priority value stored in database
}
```

## Conclusion

The priority queue feature provides fine-grained control over step execution order while maintaining the flexibility and power of ORCHWF's dependency system. Use it to optimize workflow performance, ensure critical operations execute first, and create more efficient business processes.
