package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/refactorroom/orchwf"
)

func main() {
	// Create in-memory state manager
	stateManager := orchwf.NewInMemoryStateManager()

	// Create orchestrator
	orchestrator := orchwf.NewOrchestrator(stateManager)

	// Step that sometimes fails (simulates network issues)
	step1, err := orchwf.NewStepBuilder("unreliable_api_call", "Unreliable API Call", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 1: Making unreliable API call...")
		time.Sleep(1 * time.Second)

		// Simulate 30% failure rate
		if rand.Float32() < 0.3 {
			return nil, fmt.Errorf("API call failed: network timeout")
		}

		return map[string]interface{}{
			"api_response": "success",
			"data":         "api_data_123",
		}, nil
	}).WithDescription("Make an API call that sometimes fails").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(5).
			WithInitialInterval(1*time.Second).
			WithMaxInterval(10*time.Second).
			WithMultiplier(2.0).
			WithRetryableErrors("network timeout", "connection refused").
			Build()).
		WithTimeout(30 * time.Second).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Step that fails after retries (demonstrates final failure)
	step2, err := orchwf.NewStepBuilder("always_fails", "Always Fails Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 2: This step always fails...")
		time.Sleep(500 * time.Millisecond)
		return nil, fmt.Errorf("permanent failure: service unavailable")
	}).WithDescription("A step that always fails to demonstrate retry limits").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(3).
			WithInitialInterval(500 * time.Millisecond).
			WithMaxInterval(2 * time.Second).
			WithMultiplier(1.5).
			Build()).
		WithDependencies("unreliable_api_call").
		WithRequired(false). // This step is not required, so workflow continues
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Step that succeeds after some failures
	step3, err := orchwf.NewStepBuilder("eventually_succeeds", "Eventually Succeeds", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 3: Attempting to process data...")
		time.Sleep(800 * time.Millisecond)

		// Simulate 50% failure rate initially, but succeeds after a few attempts
		if rand.Float32() < 0.5 {
			return nil, fmt.Errorf("temporary processing error")
		}

		return map[string]interface{}{
			"processed": true,
			"result":    "processing_complete",
		}, nil
	}).WithDescription("A step that eventually succeeds after retries").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(4).
			WithInitialInterval(1 * time.Second).
			WithMaxInterval(5 * time.Second).
			WithMultiplier(1.8).
			WithRetryableErrors("temporary processing error").
			Build()).
		WithDependencies("unreliable_api_call").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Final step that processes successful results
	step4, err := orchwf.NewStepBuilder("finalize", "Finalize Results", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 4: Finalizing results...")
		time.Sleep(500 * time.Millisecond)

		// Check what data we have available
		results := map[string]interface{}{
			"finalized": true,
			"timestamp": time.Now().Unix(),
		}

		// Add API data if available
		if apiData, exists := input["unreliable_api_call"]; exists {
			results["api_data"] = apiData
		}

		// Add processed data if available
		if processedData, exists := input["eventually_succeeds"]; exists {
			results["processed_data"] = processedData
		}

		// Note: always_fails step data won't be available since it failed
		results["failed_steps"] = []string{"always_fails"}

		return results, nil
	}).WithDescription("Finalize results from successful steps").
		WithDependencies("eventually_succeeds").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Build workflow
	workflow, err := orchwf.NewWorkflowBuilder("error_handling_demo", "Error Handling and Retry Demo").
		WithDescription("Demonstrates error handling, retries, and optional steps").
		WithVersion("1.0.0").
		AddStep(step1).
		AddStep(step2).
		AddStep(step3).
		AddStep(step4).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Register workflow
	if err := orchestrator.RegisterWorkflow(workflow); err != nil {
		log.Fatal(err)
	}

	// Execute workflow multiple times to see different outcomes
	for i := 1; i <= 3; i++ {
		fmt.Printf("\n=== Execution %d ===\n", i)

		result, err := orchestrator.StartWorkflow(context.Background(), "error_handling_demo",
			map[string]interface{}{
				"execution_id": i,
			},
			map[string]interface{}{
				"trace_id": fmt.Sprintf("error_demo_%d", i),
			})

		if err != nil {
			fmt.Printf("Workflow failed: %v\n", err)
		} else {
			fmt.Printf("Workflow completed successfully!\n")
			fmt.Printf("Result: %+v\n", result.Output)
			fmt.Printf("Duration: %v\n", result.Duration)
		}

		// Reset random seed for different outcomes
		rand.Seed(time.Now().UnixNano())
	}
}
