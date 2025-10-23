package orchwf

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestPriorityQueueBasic tests basic priority queue functionality
func TestPriorityQueueBasic(t *testing.T) {
	stateManager := NewInMemoryStateManager()
	orchestrator := NewOrchestrator(stateManager)

	// Track execution order
	executionOrder := make([]string, 0)
	executionMutex := sync.Mutex{}

	// Create steps with different priorities
	highPriorityStep, err := NewStepBuilder("high", "High Priority Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "high")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "high"}, nil
	}).
		WithPriority(10).
		Build()

	if err != nil {
		t.Fatalf("Failed to create high priority step: %v", err)
	}

	normalPriorityStep, err := NewStepBuilder("normal", "Normal Priority Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "normal")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "normal"}, nil
	}).
		WithPriority(0).
		Build()

	if err != nil {
		t.Fatalf("Failed to create normal priority step: %v", err)
	}

	lowPriorityStep, err := NewStepBuilder("low", "Low Priority Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "low")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "low"}, nil
	}).
		WithPriority(-10).
		Build()

	if err != nil {
		t.Fatalf("Failed to create low priority step: %v", err)
	}

	// Build workflow
	workflow, err := NewWorkflowBuilder("priority_test", "Priority Test Workflow").
		AddStep(lowPriorityStep). // Add in reverse order to test priority
		AddStep(normalPriorityStep).
		AddStep(highPriorityStep).
		Build()

	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Register and execute
	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "priority_test",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Workflow failed: %v", result.Error)
	}

	// Verify execution order (high priority should execute first)
	expectedOrder := []string{"high", "normal", "low"}
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d steps to execute, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Expected step %d to be '%s', got '%s'", i, expected, executionOrder[i])
		}
	}
}

// TestPriorityWithDependencies tests priority with step dependencies
func TestPriorityWithDependencies(t *testing.T) {
	stateManager := NewInMemoryStateManager()
	orchestrator := NewOrchestrator(stateManager)

	executionOrder := make([]string, 0)
	executionMutex := sync.Mutex{}

	// Step A: No dependencies, priority 5
	stepA, err := NewStepBuilder("step_a", "Step A", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "step_a")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "a"}, nil
	}).
		WithPriority(5).
		Build()

	if err != nil {
		t.Fatalf("Failed to create step A: %v", err)
	}

	// Step B: No dependencies, priority 1 (lower than A)
	stepB, err := NewStepBuilder("step_b", "Step B", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "step_b")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "b"}, nil
	}).
		WithPriority(1).
		Build()

	if err != nil {
		t.Fatalf("Failed to create step B: %v", err)
	}

	// Step C: Depends on A, priority 10 (highest, but waits for A)
	stepC, err := NewStepBuilder("step_c", "Step C", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "step_c")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "c"}, nil
	}).
		WithDependencies("step_a").
		WithPriority(10).
		Build()

	if err != nil {
		t.Fatalf("Failed to create step C: %v", err)
	}

	// Build workflow
	workflow, err := NewWorkflowBuilder("dependency_priority_test", "Dependency Priority Test").
		AddStep(stepA).
		AddStep(stepB).
		AddStep(stepC).
		Build()

	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Register and execute
	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "dependency_priority_test",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Workflow failed: %v", result.Error)
	}

	// Verify execution order: A and B execute first (A has higher priority), then C
	expectedOrder := []string{"step_a", "step_b", "step_c"}
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d steps to execute, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Expected step %d to be '%s', got '%s'", i, expected, executionOrder[i])
		}
	}
}

// TestPriorityWithAsyncSteps tests priority with async steps
func TestPriorityWithAsyncSteps(t *testing.T) {
	stateManager := NewInMemoryStateManager()
	orchestrator := NewOrchestrator(stateManager)

	executionOrder := make([]string, 0)
	executionMutex := sync.Mutex{}

	// High priority async step
	asyncHigh, err := NewStepBuilder("async_high", "High Priority Async", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "async_high")
		executionMutex.Unlock()
		time.Sleep(50 * time.Millisecond)
		return map[string]interface{}{"result": "async_high"}, nil
	}).
		WithAsync(true).
		WithPriority(10).
		Build()

	if err != nil {
		t.Fatalf("Failed to create high priority async step: %v", err)
	}

	// Low priority async step
	asyncLow, err := NewStepBuilder("async_low", "Low Priority Async", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "async_low")
		executionMutex.Unlock()
		time.Sleep(50 * time.Millisecond)
		return map[string]interface{}{"result": "async_low"}, nil
	}).
		WithAsync(true).
		WithPriority(1).
		Build()

	if err != nil {
		t.Fatalf("Failed to create low priority async step: %v", err)
	}

	// Build workflow
	workflow, err := NewWorkflowBuilder("async_priority_test", "Async Priority Test").
		AddStep(asyncHigh).
		AddStep(asyncLow).
		Build()

	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Register and execute
	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "async_priority_test",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Workflow failed: %v", result.Error)
	}

	// Both async steps should execute, but high priority should be scheduled first
	// Note: Due to concurrency, exact order may vary, but we should have both steps
	if len(executionOrder) != 2 {
		t.Fatalf("Expected 2 async steps to execute, got %d", len(executionOrder))
	}

	// Verify both steps executed
	hasHigh := false
	hasLow := false
	for _, step := range executionOrder {
		if step == "async_high" {
			hasHigh = true
		}
		if step == "async_low" {
			hasLow = true
		}
	}

	if !hasHigh {
		t.Error("High priority async step did not execute")
	}
	if !hasLow {
		t.Error("Low priority async step did not execute")
	}
}

// TestPriorityDefaultValue tests that steps without priority have default value 0
func TestPriorityDefaultValue(t *testing.T) {
	stateManager := NewInMemoryStateManager()
	orchestrator := NewOrchestrator(stateManager)

	executionOrder := make([]string, 0)
	executionMutex := sync.Mutex{}

	// Step with explicit priority 0
	explicitZero, err := NewStepBuilder("explicit_zero", "Explicit Zero Priority", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "explicit_zero")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "explicit_zero"}, nil
	}).
		WithPriority(0).
		Build()

	if err != nil {
		t.Fatalf("Failed to create explicit zero priority step: %v", err)
	}

	// Step without priority (should default to 0)
	noPriority, err := NewStepBuilder("no_priority", "No Priority Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "no_priority")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "no_priority"}, nil
	}).
		Build()

	if err != nil {
		t.Fatalf("Failed to create no priority step: %v", err)
	}

	// High priority step to ensure ordering works
	highPriority, err := NewStepBuilder("high_priority", "High Priority Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "high_priority")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "high_priority"}, nil
	}).
		WithPriority(10).
		Build()

	if err != nil {
		t.Fatalf("Failed to create high priority step: %v", err)
	}

	// Build workflow
	workflow, err := NewWorkflowBuilder("default_priority_test", "Default Priority Test").
		AddStep(explicitZero).
		AddStep(noPriority).
		AddStep(highPriority).
		Build()

	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Register and execute
	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "default_priority_test",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Workflow failed: %v", result.Error)
	}

	// High priority should execute first, then the two zero priority steps
	// (order between zero priority steps may vary due to definition order)
	if len(executionOrder) != 3 {
		t.Fatalf("Expected 3 steps to execute, got %d", len(executionOrder))
	}

	// First step should be high priority
	if executionOrder[0] != "high_priority" {
		t.Errorf("Expected first step to be 'high_priority', got '%s'", executionOrder[0])
	}

	// Verify both zero priority steps executed
	hasExplicitZero := false
	hasNoPriority := false
	for i := 1; i < len(executionOrder); i++ {
		if executionOrder[i] == "explicit_zero" {
			hasExplicitZero = true
		}
		if executionOrder[i] == "no_priority" {
			hasNoPriority = true
		}
	}

	if !hasExplicitZero {
		t.Error("Explicit zero priority step did not execute")
	}
	if !hasNoPriority {
		t.Error("No priority step did not execute")
	}
}

// TestPriorityNegativeValues tests negative priority values
func TestPriorityNegativeValues(t *testing.T) {
	stateManager := NewInMemoryStateManager()
	orchestrator := NewOrchestrator(stateManager)

	executionOrder := make([]string, 0)
	executionMutex := sync.Mutex{}

	// Very low priority step
	veryLow, err := NewStepBuilder("very_low", "Very Low Priority", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "very_low")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "very_low"}, nil
	}).
		WithPriority(-20).
		Build()

	if err != nil {
		t.Fatalf("Failed to create very low priority step: %v", err)
	}

	// Low priority step
	low, err := NewStepBuilder("low", "Low Priority", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "low")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "low"}, nil
	}).
		WithPriority(-5).
		Build()

	if err != nil {
		t.Fatalf("Failed to create low priority step: %v", err)
	}

	// High priority step
	high, err := NewStepBuilder("high", "High Priority", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "high")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "high"}, nil
	}).
		WithPriority(10).
		Build()

	if err != nil {
		t.Fatalf("Failed to create high priority step: %v", err)
	}

	// Build workflow
	workflow, err := NewWorkflowBuilder("negative_priority_test", "Negative Priority Test").
		AddStep(veryLow).
		AddStep(low).
		AddStep(high).
		Build()

	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Register and execute
	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "negative_priority_test",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Workflow failed: %v", result.Error)
	}

	// Verify execution order: high -> low -> very_low
	expectedOrder := []string{"high", "low", "very_low"}
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d steps to execute, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Expected step %d to be '%s', got '%s'", i, expected, executionOrder[i])
		}
	}
}

// TestPriorityWithMixedSyncAsync tests priority with mixed sync and async steps
func TestPriorityWithMixedSyncAsync(t *testing.T) {
	stateManager := NewInMemoryStateManager()
	orchestrator := NewOrchestrator(stateManager)

	executionOrder := make([]string, 0)
	executionMutex := sync.Mutex{}

	// High priority sync step
	syncHigh, err := NewStepBuilder("sync_high", "High Priority Sync", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "sync_high")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "sync_high"}, nil
	}).
		WithPriority(10).
		WithAsync(false).
		Build()

	if err != nil {
		t.Fatalf("Failed to create high priority sync step: %v", err)
	}

	// Low priority sync step
	syncLow, err := NewStepBuilder("sync_low", "Low Priority Sync", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "sync_low")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "sync_low"}, nil
	}).
		WithPriority(1).
		WithAsync(false).
		Build()

	if err != nil {
		t.Fatalf("Failed to create low priority sync step: %v", err)
	}

	// High priority async step
	asyncHigh, err := NewStepBuilder("async_high", "High Priority Async", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "async_high")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "async_high"}, nil
	}).
		WithPriority(10).
		WithAsync(true).
		Build()

	if err != nil {
		t.Fatalf("Failed to create high priority async step: %v", err)
	}

	// Low priority async step
	asyncLow, err := NewStepBuilder("async_low", "Low Priority Async", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "async_low")
		executionMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "async_low"}, nil
	}).
		WithPriority(1).
		WithAsync(true).
		Build()

	if err != nil {
		t.Fatalf("Failed to create low priority async step: %v", err)
	}

	// Build workflow
	workflow, err := NewWorkflowBuilder("mixed_priority_test", "Mixed Priority Test").
		AddStep(syncHigh).
		AddStep(syncLow).
		AddStep(asyncHigh).
		AddStep(asyncLow).
		Build()

	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Register and execute
	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "mixed_priority_test",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Workflow failed: %v", result.Error)
	}

	// Verify all steps executed
	if len(executionOrder) != 4 {
		t.Fatalf("Expected 4 steps to execute, got %d", len(executionOrder))
	}

	// Verify all step types executed
	hasSyncHigh := false
	hasSyncLow := false
	hasAsyncHigh := false
	hasAsyncLow := false

	for _, step := range executionOrder {
		switch step {
		case "sync_high":
			hasSyncHigh = true
		case "sync_low":
			hasSyncLow = true
		case "async_high":
			hasAsyncHigh = true
		case "async_low":
			hasAsyncLow = true
		}
	}

	if !hasSyncHigh {
		t.Error("High priority sync step did not execute")
	}
	if !hasSyncLow {
		t.Error("Low priority sync step did not execute")
	}
	if !hasAsyncHigh {
		t.Error("High priority async step did not execute")
	}
	if !hasAsyncLow {
		t.Error("Low priority async step did not execute")
	}
}

// TestPriorityStepBuilder tests the WithPriority builder method
func TestPriorityStepBuilder(t *testing.T) {
	executor := func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "test"}, nil
	}

	// Test with positive priority
	step, err := NewStepBuilder("test_step", "Test Step", executor).
		WithPriority(15).
		Build()

	if err != nil {
		t.Fatalf("Failed to create step with priority: %v", err)
	}

	if step.Priority != 15 {
		t.Errorf("Expected priority 15, got %d", step.Priority)
	}

	// Test with negative priority
	step, err = NewStepBuilder("test_step_neg", "Test Step Negative", executor).
		WithPriority(-5).
		Build()

	if err != nil {
		t.Fatalf("Failed to create step with negative priority: %v", err)
	}

	if step.Priority != -5 {
		t.Errorf("Expected priority -5, got %d", step.Priority)
	}

	// Test with zero priority
	step, err = NewStepBuilder("test_step_zero", "Test Step Zero", executor).
		WithPriority(0).
		Build()

	if err != nil {
		t.Fatalf("Failed to create step with zero priority: %v", err)
	}

	if step.Priority != 0 {
		t.Errorf("Expected priority 0, got %d", step.Priority)
	}
}

// TestPriorityWithoutBuilder tests that steps without WithPriority have default priority
func TestPriorityWithoutBuilder(t *testing.T) {
	executor := func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "test"}, nil
	}

	// Create step without WithPriority
	step, err := NewStepBuilder("test_step", "Test Step", executor).
		Build()

	if err != nil {
		t.Fatalf("Failed to create step without priority: %v", err)
	}

	// Should have default priority of 0
	if step.Priority != 0 {
		t.Errorf("Expected default priority 0, got %d", step.Priority)
	}
}

// TestPriorityWithRetryPolicy tests priority with retry policy
func TestPriorityWithRetryPolicy(t *testing.T) {
	stateManager := NewInMemoryStateManager()
	orchestrator := NewOrchestrator(stateManager)

	executionCount := 0
	executionMutex := sync.Mutex{}

	// Create step with priority and retry policy
	step, err := NewStepBuilder("retry_step", "Retry Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionCount++
		executionMutex.Unlock()

		// Fail first two times, succeed on third
		if executionCount < 3 {
			return nil, fmt.Errorf("temporary failure")
		}
		return map[string]interface{}{"result": "success"}, nil
	}).
		WithPriority(10).
		WithRetryPolicy(NewRetryPolicyBuilder().
			WithMaxAttempts(3).
			WithInitialInterval(10 * time.Millisecond).
			Build()).
		Build()

	if err != nil {
		t.Fatalf("Failed to create step with retry policy: %v", err)
	}

	// Build workflow
	workflow, err := NewWorkflowBuilder("retry_priority_test", "Retry Priority Test").
		AddStep(step).
		Build()

	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Register and execute
	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "retry_priority_test",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Workflow failed: %v", result.Error)
	}

	// Should have executed 3 times (2 failures + 1 success)
	if executionCount != 3 {
		t.Errorf("Expected 3 executions (2 retries + 1 success), got %d", executionCount)
	}
}

// TestPriorityWithTimeoutBasic tests priority with timeout
func TestPriorityWithTimeoutBasic(t *testing.T) {
	stateManager := NewInMemoryStateManager()
	orchestrator := NewOrchestrator(stateManager)

	// Create step with priority and timeout
	step, err := NewStepBuilder("timeout_step", "Timeout Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		// Simple step that completes quickly
		time.Sleep(10 * time.Millisecond)
		return map[string]interface{}{"result": "success"}, nil
	}).
		WithPriority(10).
		WithTimeout(100 * time.Millisecond). // Generous timeout
		Build()

	if err != nil {
		t.Fatalf("Failed to create step with timeout: %v", err)
	}

	// Build workflow
	workflow, err := NewWorkflowBuilder("timeout_priority_test", "Timeout Priority Test").
		AddStep(step).
		Build()

	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Register and execute
	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "timeout_priority_test",
		map[string]interface{}{"data": "test"}, nil)

	// Workflow should complete successfully
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Workflow failed: %v", result.Error)
	}

	// Verify the step completed successfully
	if len(result.WorkflowInst.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(result.WorkflowInst.Steps))
	}

	stepInst := result.WorkflowInst.Steps[0]
	if stepInst.Status != StepStatusCompleted {
		t.Errorf("Expected step to be completed, got status: %s", stepInst.Status)
	}
}

// TestPriorityWithRequired tests priority with required flag
func TestPriorityWithRequired(t *testing.T) {
	stateManager := NewInMemoryStateManager()
	orchestrator := NewOrchestrator(stateManager)

	// Create high priority required step that fails
	highPriorityFail, err := NewStepBuilder("high_fail", "High Priority Fail", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return nil, fmt.Errorf("high priority step failed")
	}).
		WithPriority(10).
		WithRequired(true).
		Build()

	if err != nil {
		t.Fatalf("Failed to create high priority failing step: %v", err)
	}

	// Create low priority step
	lowPriority, err := NewStepBuilder("low_success", "Low Priority Success", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "low_success"}, nil
	}).
		WithPriority(1).
		Build()

	if err != nil {
		t.Fatalf("Failed to create low priority step: %v", err)
	}

	// Build workflow
	workflow, err := NewWorkflowBuilder("required_priority_test", "Required Priority Test").
		AddStep(highPriorityFail).
		AddStep(lowPriority).
		Build()

	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Register and execute
	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "required_priority_test",
		map[string]interface{}{"data": "test"}, nil)

	// Should fail because high priority required step failed
	if err == nil {
		t.Fatal("Expected workflow to fail due to required step failure")
	}

	if result.Success {
		t.Fatal("Expected workflow to fail due to required step failure")
	}
}

// TestPriorityWithNonRequired tests priority with non-required steps
func TestPriorityWithNonRequired(t *testing.T) {
	stateManager := NewInMemoryStateManager()
	orchestrator := NewOrchestrator(stateManager)

	executionOrder := make([]string, 0)
	executionMutex := sync.Mutex{}

	// Create high priority non-required step that fails
	highPriorityFail, err := NewStepBuilder("high_fail", "High Priority Fail", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "high_fail")
		executionMutex.Unlock()
		return nil, fmt.Errorf("high priority step failed")
	}).
		WithPriority(10).
		WithRequired(false). // Don't fail workflow
		Build()

	if err != nil {
		t.Fatalf("Failed to create high priority failing step: %v", err)
	}

	// Create low priority step that succeeds
	lowPrioritySuccess, err := NewStepBuilder("low_success", "Low Priority Success", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		executionMutex.Lock()
		executionOrder = append(executionOrder, "low_success")
		executionMutex.Unlock()
		return map[string]interface{}{"result": "low_success"}, nil
	}).
		WithPriority(1).
		Build()

	if err != nil {
		t.Fatalf("Failed to create low priority step: %v", err)
	}

	// Build workflow
	workflow, err := NewWorkflowBuilder("non_required_priority_test", "Non-Required Priority Test").
		AddStep(highPriorityFail).
		AddStep(lowPrioritySuccess).
		Build()

	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Register and execute
	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "non_required_priority_test",
		map[string]interface{}{"data": "test"}, nil)

	// Should succeed because high priority step is not required
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Workflow failed: %v", result.Error)
	}

	// Both steps should have executed (high priority first, then low)
	expectedOrder := []string{"high_fail", "low_success"}
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d steps to execute, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Expected step %d to be '%s', got '%s'", i, expected, executionOrder[i])
		}
	}
}
