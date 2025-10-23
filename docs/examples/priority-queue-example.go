package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/akkaraponph/orchwf"
)

// Priority Queue Example
// This example demonstrates how to use the priority queue feature
// to control the execution order of workflow steps.

func main() {
	fmt.Println("🚀 ORCHWF Priority Queue Example")
	fmt.Println("=================================")

	// Create state manager
	stateManager := orchwf.NewInMemoryStateManager()
	orchestrator := orchwf.NewOrchestrator(stateManager)

	// Define steps with different priorities
	steps := createPrioritySteps()

	// Build workflow
	workflow, err := orchwf.NewWorkflowBuilder("priority_example", "Priority Queue Example").
		AddStep(steps["critical"]).
		AddStep(steps["high"]).
		AddStep(steps["normal"]).
		AddStep(steps["low"]).
		AddStep(steps["background"]).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Register workflow
	orchestrator.RegisterWorkflow(workflow)

	// Execute workflow
	fmt.Println("\n📋 Executing workflow with priority-based ordering...")
	fmt.Println("Expected order: Critical → High → Normal → Low → Background")
	fmt.Println()

	result, err := orchestrator.StartWorkflow(context.Background(), "priority_example",
		map[string]interface{}{"data": "priority_test"}, nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n✅ Workflow completed successfully!\n")
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Output: %+v\n", result.Output)
}

func createPrioritySteps() map[string]*orchwf.StepDefinition {
	steps := make(map[string]*orchwf.StepDefinition)

	// Critical priority step (executes first)
	criticalStep, err := orchwf.NewStepBuilder("critical", "Critical System Check", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("🔥 [CRITICAL] Performing critical system check...")
		time.Sleep(100 * time.Millisecond)
		fmt.Println("✅ [CRITICAL] System check completed")
		return map[string]interface{}{
			"critical_status": "healthy",
			"timestamp":       time.Now().Unix(),
		}, nil
	}).
		WithPriority(20).
		WithDescription("Critical system validation that must run first").
		WithTimeout(5 * time.Second).
		Build()

	if err != nil {
		log.Fatal(err)
	}
	steps["critical"] = criticalStep

	// High priority step
	highStep, err := orchwf.NewStepBuilder("high", "High Priority Processing", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("⚡ [HIGH] Processing high-priority data...")
		time.Sleep(150 * time.Millisecond)
		fmt.Println("✅ [HIGH] High-priority processing completed")
		return map[string]interface{}{
			"high_result": "processed",
			"priority":    10,
		}, nil
	}).
		WithPriority(10).
		WithDescription("High-priority business logic").
		Build()

	if err != nil {
		log.Fatal(err)
	}
	steps["high"] = highStep

	// Normal priority step (default)
	normalStep, err := orchwf.NewStepBuilder("normal", "Normal Processing", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("📋 [NORMAL] Executing normal processing...")
		time.Sleep(200 * time.Millisecond)
		fmt.Println("✅ [NORMAL] Normal processing completed")
		return map[string]interface{}{
			"normal_result": "completed",
			"priority":      0,
		}, nil
	}).
		WithPriority(0).
		WithDescription("Standard processing step").
		Build()

	if err != nil {
		log.Fatal(err)
	}
	steps["normal"] = normalStep

	// Low priority step
	lowStep, err := orchwf.NewStepBuilder("low", "Low Priority Task", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("🐌 [LOW] Running low-priority task...")
		time.Sleep(300 * time.Millisecond)
		fmt.Println("✅ [LOW] Low-priority task completed")
		return map[string]interface{}{
			"low_result": "finished",
			"priority":   -5,
		}, nil
	}).
		WithPriority(-5).
		WithDescription("Low-priority maintenance task").
		Build()

	if err != nil {
		log.Fatal(err)
	}
	steps["low"] = lowStep

	// Background priority step (lowest)
	backgroundStep, err := orchwf.NewStepBuilder("background", "Background Cleanup", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("🔄 [BACKGROUND] Performing background cleanup...")
		time.Sleep(250 * time.Millisecond)
		fmt.Println("✅ [BACKGROUND] Background cleanup completed")
		return map[string]interface{}{
			"background_result": "cleaned",
			"priority":          -10,
		}, nil
	}).
		WithPriority(-10).
		WithDescription("Background maintenance and cleanup").
		WithRequired(false). // Don't fail workflow if background task fails
		Build()

	if err != nil {
		log.Fatal(err)
	}
	steps["background"] = backgroundStep

	return steps
}

// Advanced Priority Example with Dependencies
func advancedPriorityExample() {
	fmt.Println("\n🔧 Advanced Priority Example with Dependencies")
	fmt.Println("=============================================")

	stateManager := orchwf.NewInMemoryStateManager()
	orchestrator := orchwf.NewOrchestrator(stateManager)

	// Step A: No dependencies, priority 5
	stepA, _ := orchwf.NewStepBuilder("step_a", "Step A", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("🅰️  [STEP A] Executing...")
		time.Sleep(100 * time.Millisecond)
		return map[string]interface{}{"result_a": "done"}, nil
	}).
		WithPriority(5).
		Build()

	// Step B: No dependencies, priority 1 (lower than A)
	stepB, _ := orchwf.NewStepBuilder("step_b", "Step B", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("🅱️  [STEP B] Executing...")
		time.Sleep(100 * time.Millisecond)
		return map[string]interface{}{"result_b": "done"}, nil
	}).
		WithPriority(1).
		Build()

	// Step C: Depends on A, priority 10 (highest, but waits for A)
	stepC, _ := orchwf.NewStepBuilder("step_c", "Step C", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("🅲  [STEP C] Executing after A...")
		time.Sleep(100 * time.Millisecond)
		return map[string]interface{}{"result_c": "done"}, nil
	}).
		WithDependencies("step_a").
		WithPriority(10).
		Build()

	// Build and execute
	workflow, _ := orchwf.NewWorkflowBuilder("advanced_priority", "Advanced Priority Example").
		AddStep(stepA).
		AddStep(stepB).
		AddStep(stepC).
		Build()

	orchestrator.RegisterWorkflow(workflow)

	fmt.Println("Expected execution order: A → B → C")
	fmt.Println("(A and B execute first based on priority, then C executes after A completes)")
	fmt.Println()

	result, err := orchestrator.StartWorkflow(context.Background(), "advanced_priority",
		map[string]interface{}{"data": "advanced_test"}, nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Advanced workflow completed: %+v\n", result.Output)
}

// Async Priority Example
func asyncPriorityExample() {
	fmt.Println("\n⚡ Async Priority Example")
	fmt.Println("=========================")

	stateManager := orchwf.NewInMemoryStateManager()
	orchestrator := orchwf.NewOrchestrator(stateManager)

	// High-priority async step
	asyncHigh, _ := orchwf.NewStepBuilder("async_high", "High Priority Async", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("🚀 [ASYNC HIGH] Starting high-priority async task...")
		time.Sleep(200 * time.Millisecond)
		fmt.Println("✅ [ASYNC HIGH] High-priority async completed")
		return map[string]interface{}{"async_high": "done"}, nil
	}).
		WithAsync(true).
		WithPriority(10).
		Build()

	// Low-priority async step
	asyncLow, _ := orchwf.NewStepBuilder("async_low", "Low Priority Async", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("🐌 [ASYNC LOW] Starting low-priority async task...")
		time.Sleep(200 * time.Millisecond)
		fmt.Println("✅ [ASYNC LOW] Low-priority async completed")
		return map[string]interface{}{"async_low": "done"}, nil
	}).
		WithAsync(true).
		WithPriority(1).
		Build()

	// Build and execute
	workflow, _ := orchwf.NewWorkflowBuilder("async_priority", "Async Priority Example").
		AddStep(asyncHigh).
		AddStep(asyncLow).
		Build()

	orchestrator.RegisterWorkflow(workflow)

	fmt.Println("Both async steps will run concurrently, but high-priority will be scheduled first")
	fmt.Println()

	result, err := orchestrator.StartWorkflow(context.Background(), "async_priority",
		map[string]interface{}{"data": "async_test"}, nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Async workflow completed: %+v\n", result.Output)
}
