package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/akkaraponph/orchwf"
)

func main() {
	// Create in-memory state manager
	stateManager := orchwf.NewInMemoryStateManager()

	// Create orchestrator with custom async workers
	orchestrator := orchwf.NewOrchestratorWithAsyncWorkers(stateManager, 5)

	// Define a long-running workflow
	step1, err := orchwf.NewStepBuilder("data_processing", "Data Processing", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Printf("Step 1: Processing data for workflow %s...\n", input["workflow_id"])
		time.Sleep(3 * time.Second) // Simulate long processing
		return map[string]interface{}{
			"processed_items": 1000,
			"processing_time": "3s",
		}, nil
	}).WithDescription("Process large dataset").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(2).
			WithInitialInterval(2 * time.Second).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	step2, err := orchwf.NewStepBuilder("data_validation", "Data Validation", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Printf("Step 2: Validating data for workflow %s...\n", input["workflow_id"])
		time.Sleep(2 * time.Second)
		return map[string]interface{}{
			"valid_items":   950,
			"invalid_items": 50,
		}, nil
	}).WithDescription("Validate processed data").
		WithDependencies("data_processing").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	step3, err := orchwf.NewStepBuilder("data_export", "Data Export", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Printf("Step 3: Exporting data for workflow %s...\n", input["workflow_id"])
		time.Sleep(2 * time.Second)
		return map[string]interface{}{
			"exported_file": fmt.Sprintf("export_%s.csv", input["workflow_id"]),
			"file_size":     "2.5MB",
		}, nil
	}).WithDescription("Export validated data").
		WithDependencies("data_validation").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Build workflow
	workflow, err := orchwf.NewWorkflowBuilder("async_data_processing", "Async Data Processing").
		WithDescription("Demonstrates asynchronous workflow execution").
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

	// Start multiple workflows asynchronously
	fmt.Println("Starting multiple workflows asynchronously...")

	var wg sync.WaitGroup
	workflowCount := 5

	// Channel to collect results
	results := make(chan *orchwf.WorkflowResult, workflowCount)
	errors := make(chan error, workflowCount)

	for i := 1; i <= workflowCount; i++ {
		wg.Add(1)
		go func(workflowID int) {
			defer wg.Done()

			workflowIDStr := fmt.Sprintf("workflow_%d", workflowID)
			fmt.Printf("Starting workflow %s...\n", workflowIDStr)

			// Start workflow asynchronously
			workflowInstanceID, err := orchestrator.StartWorkflowAsync(context.Background(), "async_data_processing",
				map[string]interface{}{
					"workflow_id": workflowIDStr,
					"batch_id":    "batch_001",
				},
				map[string]interface{}{
					"trace_id": fmt.Sprintf("async_trace_%d", workflowID),
				})

			if err != nil {
				errors <- fmt.Errorf("workflow %s failed to start: %v", workflowIDStr, err)
				return
			}

			// Wait for completion by polling status
			for {
				status, err := orchestrator.GetWorkflowStatus(context.Background(), workflowInstanceID)
				if err != nil {
					errors <- fmt.Errorf("workflow %s status check failed: %v", workflowIDStr, err)
					return
				}

				if status.Status == orchwf.WorkflowStatusCompleted {
					// Get the completed workflow result
					workflow, err := stateManager.GetWorkflow(context.Background(), workflowInstanceID)
					if err != nil {
						errors <- fmt.Errorf("workflow %s result retrieval failed: %v", workflowIDStr, err)
						return
					}

					result := &orchwf.WorkflowResult{
						Success:      true,
						WorkflowInst: workflow,
						Output:       workflow.Output,
						Duration:     workflow.CompletedAt.Sub(workflow.StartedAt),
					}
					results <- result
					break
				} else if status.Status == orchwf.WorkflowStatusFailed {
					errors <- fmt.Errorf("workflow %s failed", workflowIDStr)
					return
				}

				// Wait before next check
				time.Sleep(100 * time.Millisecond)
			}
			fmt.Printf("Workflow %s completed successfully!\n", workflowIDStr)
		}(i)
	}

	// Wait for all workflows to complete
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collect and display results
	successCount := 0
	errorCount := 0

	fmt.Println("\nWaiting for workflows to complete...")

	// Process results
	for {
		select {
		case result, ok := <-results:
			if !ok {
				// Results channel closed
				goto done
			}
			successCount++
			fmt.Printf("✓ Workflow completed in %v\n", result.Duration)

		case err, ok := <-errors:
			if !ok {
				// Errors channel closed
				goto done
			}
			errorCount++
			fmt.Printf("✗ Workflow error: %v\n", err)
		}
	}

done:
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Successful workflows: %d\n", successCount)
	fmt.Printf("Failed workflows: %d\n", errorCount)
	fmt.Printf("Total workflows: %d\n", workflowCount)

	// Demonstrate workflow status checking
	fmt.Println("\n=== Checking workflow statuses ===")
	for i := 1; i <= workflowCount; i++ {
		workflowID := fmt.Sprintf("workflow_%d", i)
		status, err := orchestrator.GetWorkflowStatus(context.Background(), workflowID)
		if err != nil {
			fmt.Printf("Workflow %s: Status unknown (%v)\n", workflowID, err)
		} else {
			fmt.Printf("Workflow %s: %s\n", workflowID, status.Status)
		}
	}
}
