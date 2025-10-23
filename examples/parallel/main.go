package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/refactorroom/orchwf"
)

func main() {
	// Create in-memory state manager
	stateManager := orchwf.NewInMemoryStateManager()

	// Create orchestrator
	orchestrator := orchwf.NewOrchestrator(stateManager)

	// Define parallel steps that can run concurrently
	step1, err := orchwf.NewStepBuilder("fetch_user_data", "Fetch User Data", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 1: Fetching user data from API...")
		time.Sleep(2 * time.Second) // Simulate API call
		return map[string]interface{}{
			"user_id":   123,
			"user_name": "john_doe",
			"email":     "john@example.com",
		}, nil
	}).WithDescription("Fetch user information from external API").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(3).
			WithInitialInterval(1 * time.Second).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	step2, err := orchwf.NewStepBuilder("fetch_user_preferences", "Fetch User Preferences", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 2: Fetching user preferences...")
		time.Sleep(1 * time.Second) // Simulate database query
		return map[string]interface{}{
			"theme":         "dark",
			"language":      "en",
			"notifications": true,
		}, nil
	}).WithDescription("Fetch user preferences from database").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(2).
			WithInitialInterval(500 * time.Millisecond).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	step3, err := orchwf.NewStepBuilder("fetch_user_orders", "Fetch User Orders", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 3: Fetching user order history...")
		time.Sleep(1500 * time.Millisecond) // Simulate complex query
		return map[string]interface{}{
			"orders": []map[string]interface{}{
				{"id": 1, "amount": 99.99, "status": "completed"},
				{"id": 2, "amount": 149.99, "status": "shipped"},
			},
		}, nil
	}).WithDescription("Fetch user order history").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(2).
			WithInitialInterval(1 * time.Second).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Aggregation step that depends on all parallel steps
	step4, err := orchwf.NewStepBuilder("aggregate_data", "Aggregate User Data", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 4: Aggregating all user data...")
		time.Sleep(500 * time.Millisecond)

		// Access data from all previous steps
		userData := input["fetch_user_data"].(map[string]interface{})
		preferences := input["fetch_user_preferences"].(map[string]interface{})
		orders := input["fetch_user_orders"].(map[string]interface{})

		return map[string]interface{}{
			"user_profile": map[string]interface{}{
				"user_data":     userData,
				"preferences":   preferences,
				"order_history": orders,
			},
			"summary": fmt.Sprintf("User %s has %d orders",
				userData["user_name"],
				len(orders["orders"].([]map[string]interface{}))),
		}, nil
	}).WithDescription("Aggregate data from all parallel steps").
		WithDependencies("fetch_user_data", "fetch_user_preferences", "fetch_user_orders").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Build workflow with parallel execution
	workflow, err := orchwf.NewWorkflowBuilder("parallel_user_data", "Parallel User Data Fetching").
		WithDescription("Demonstrates parallel execution of independent steps").
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

	// Execute workflow
	fmt.Println("Starting parallel workflow...")
	start := time.Now()

	result, err := orchestrator.StartWorkflow(context.Background(), "parallel_user_data",
		map[string]interface{}{
			"user_id": 123,
		},
		map[string]interface{}{
			"trace_id": "parallel_trace_123",
		})

	if err != nil {
		log.Fatal(err)
	}

	duration := time.Since(start)
	fmt.Printf("\nWorkflow completed successfully!\n")
	fmt.Printf("Total execution time: %v\n", duration)
	fmt.Printf("Result: %+v\n", result.Output)

	// Note: In a real scenario, steps 1, 2, and 3 would run in parallel
	// The total time should be approximately max(step1_time, step2_time, step3_time) + step4_time
	// rather than sum of all step times
}
