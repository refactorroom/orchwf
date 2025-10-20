package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/akkaraponph/orchwf"
)

func main() {
	// Create in-memory state manager
	stateManager := orchwf.NewInMemoryStateManager()

	// Create orchestrator
	orchestrator := orchwf.NewOrchestrator(stateManager)

	// Define steps
	step1, err := orchwf.NewStepBuilder("step1", "Process Data", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 1: Processing data...")
		time.Sleep(1 * time.Second)
		return map[string]interface{}{
			"processed_data": "step1_result",
		}, nil
	}).WithDescription("Process input data").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(3).
			WithInitialInterval(1 * time.Second).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	step2, err := orchwf.NewStepBuilder("step2", "Validate Data", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 2: Validating data...")
		time.Sleep(500 * time.Millisecond)
		return map[string]interface{}{
			"validation_result": "valid",
		}, nil
	}).WithDescription("Validate processed data").
		WithDependencies("step1").
		WithTimeout(5 * time.Second).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	step3, err := orchwf.NewStepBuilder("step3", "Save Data", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 3: Saving data...")
		time.Sleep(1 * time.Second)
		return map[string]interface{}{
			"saved": true,
		}, nil
	}).WithDescription("Save validated data").
		WithDependencies("step2").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Build workflow
	workflow, err := orchwf.NewWorkflowBuilder("simple_workflow", "Simple Data Processing Workflow").
		WithDescription("A simple workflow that processes, validates, and saves data").
		WithVersion("1.0.0").
		AddStep(step1).
		AddStep(step2).
		AddStep(step3).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Register workflow
	if err := orchestrator.RegisterWorkflow(workflow); err != nil {
		log.Fatal(err)
	}

	// Execute workflow synchronously
	fmt.Println("Starting workflow synchronously...")
	result, err := orchestrator.StartWorkflow(context.Background(), "simple_workflow",
		map[string]interface{}{
			"data": "test_data",
		},
		map[string]interface{}{
			"trace_id": "trace_123",
		})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Workflow completed successfully: %+v\n", result)
	fmt.Printf("Duration: %v\n", result.Duration)
}
