package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/akkaraponph/orchwf"
)

// BatchProcessor handles batch processing of workflows
type BatchProcessor struct {
	orchestrator *orchwf.Orchestrator
	concurrency  int
}

func NewBatchProcessor(orchestrator *orchwf.Orchestrator, concurrency int) *BatchProcessor {
	return &BatchProcessor{
		orchestrator: orchestrator,
		concurrency:  concurrency,
	}
}

// ProcessBatch processes a batch of items using workflows
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, workflowID string, items []map[string]interface{}) ([]*orchwf.WorkflowResult, []error) {
	results := make([]*orchwf.WorkflowResult, len(items))
	errors := make([]error, len(items))

	// Create semaphore to limit concurrency
	semaphore := make(chan struct{}, bp.concurrency)
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		go func(index int, data map[string]interface{}) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Execute workflow
			result, err := bp.orchestrator.StartWorkflow(ctx, workflowID, data, map[string]interface{}{
				"trace_id": fmt.Sprintf("batch_%d_%d", time.Now().Unix(), index),
				"batch_id": fmt.Sprintf("batch_%d", time.Now().Unix()),
			})

			results[index] = result
			errors[index] = err
		}(i, item)
	}

	wg.Wait()
	return results, errors
}

func main() {
	// Create in-memory state manager
	stateManager := orchwf.NewInMemoryStateManager()

	// Create orchestrator with more async workers for batch processing
	orchestrator := orchwf.NewOrchestratorWithAsyncWorkers(stateManager, 20)

	// Define a data processing workflow
	step1, err := orchwf.NewStepBuilder("validate_data", "Validate Data", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		itemID := input["item_id"].(string)
		fmt.Printf("Validating item %s...\n", itemID)
		time.Sleep(100 * time.Millisecond) // Simulate validation

		// Simulate validation logic
		value := input["value"].(float64)
		if value < 0 {
			return nil, fmt.Errorf("invalid value: %f", value)
		}

		return map[string]interface{}{
			"valid":        true,
			"item_id":      itemID,
			"validated_at": time.Now().Unix(),
		}, nil
	}).WithDescription("Validate input data").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(2).
			WithInitialInterval(500 * time.Millisecond).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	step2, err := orchwf.NewStepBuilder("process_data", "Process Data", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		itemID := input["item_id"].(string)
		fmt.Printf("Processing item %s...\n", itemID)
		time.Sleep(200 * time.Millisecond) // Simulate processing

		value := input["value"].(float64)
		processedValue := value * 1.1 // Apply 10% markup

		return map[string]interface{}{
			"processed":       true,
			"item_id":         itemID,
			"original_value":  value,
			"processed_value": processedValue,
			"processed_at":    time.Now().Unix(),
		}, nil
	}).WithDescription("Process validated data").
		WithDependencies("validate_data").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(2).
			WithInitialInterval(1 * time.Second).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	step3, err := orchwf.NewStepBuilder("save_result", "Save Result", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		itemID := input["item_id"].(string)
		fmt.Printf("Saving result for item %s...\n", itemID)
		time.Sleep(150 * time.Millisecond) // Simulate database save

		validationData := input["validate_data"].(map[string]interface{})
		processingData := input["process_data"].(map[string]interface{})

		return map[string]interface{}{
			"saved":           true,
			"item_id":         itemID,
			"final_value":     processingData["processed_value"],
			"saved_at":        time.Now().Unix(),
			"processing_time": time.Now().Unix() - validationData["validated_at"].(int64),
		}, nil
	}).WithDescription("Save processing result").
		WithDependencies("process_data").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Build workflow
	workflow, err := orchwf.NewWorkflowBuilder("data_processing", "Data Processing Workflow").
		WithDescription("Process individual data items in batch").
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

	// Create batch processor
	batchProcessor := NewBatchProcessor(orchestrator, 10) // Process 10 items concurrently

	// Generate test data
	fmt.Println("Generating test data...")
	var testData []map[string]interface{}
	for i := 1; i <= 50; i++ {
		testData = append(testData, map[string]interface{}{
			"item_id":  fmt.Sprintf("item_%03d", i),
			"value":    float64(i * 10),
			"category": fmt.Sprintf("category_%d", (i-1)%5+1),
		})
	}

	// Add some invalid data to test error handling
	testData = append(testData, map[string]interface{}{
		"item_id":  "item_invalid",
		"value":    -100.0, // This will fail validation
		"category": "invalid",
	})

	fmt.Printf("Generated %d test items\n", len(testData))

	// Process batch
	fmt.Println("\nStarting batch processing...")
	start := time.Now()

	results, errors := batchProcessor.ProcessBatch(context.Background(), "data_processing", testData)

	duration := time.Since(start)

	// Analyze results
	successCount := 0
	errorCount := 0
	totalProcessedValue := 0.0

	fmt.Printf("\n=== Batch Processing Results ===\n")
	fmt.Printf("Total items: %d\n", len(testData))
	fmt.Printf("Processing time: %v\n", duration)
	fmt.Printf("Throughput: %.2f items/second\n", float64(len(testData))/duration.Seconds())

	for i, result := range results {
		if errors[i] != nil {
			errorCount++
			fmt.Printf("✗ Item %s failed: %v\n", testData[i]["item_id"], errors[i])
		} else {
			successCount++
			if saveData, ok := result.Output["save_result"].(map[string]interface{}); ok {
				finalValue := saveData["final_value"].(float64)
				totalProcessedValue += finalValue
			}
		}
	}

	fmt.Printf("\nSuccessful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", errorCount)
	fmt.Printf("Success rate: %.2f%%\n", float64(successCount)/float64(len(testData))*100)
	fmt.Printf("Total processed value: %.2f\n", totalProcessedValue)

	// Demonstrate workflow status monitoring
	fmt.Println("\n=== Workflow Status Monitoring ===")

	// Get all workflow instances using ListWorkflows
	allWorkflows, total, err := stateManager.ListWorkflows(context.Background(), map[string]interface{}{}, 1000, 0)
	if err != nil {
		log.Printf("Failed to get workflow statuses: %v", err)
	} else {
		statusCounts := make(map[orchwf.WorkflowStatus]int)
		for _, wf := range allWorkflows {
			statusCounts[wf.Status]++
		}

		fmt.Printf("Workflow status distribution (showing %d of %d total):\n", len(allWorkflows), total)
		for status, count := range statusCounts {
			fmt.Printf("  %s: %d\n", status, count)
		}
	}

	// Demonstrate batch processing with different concurrency levels
	fmt.Println("\n=== Concurrency Comparison ===")

	concurrencyLevels := []int{1, 5, 10, 20}
	for _, concurrency := range concurrencyLevels {
		fmt.Printf("\nTesting with concurrency level: %d\n", concurrency)

		// Create new batch processor with different concurrency
		testProcessor := NewBatchProcessor(orchestrator, concurrency)

		// Use smaller dataset for testing
		testDataSmall := testData[:10]

		start := time.Now()
		_, _ = testProcessor.ProcessBatch(context.Background(), "data_processing", testDataSmall)
		duration := time.Since(start)

		fmt.Printf("  Processed %d items in %v (%.2f items/second)\n",
			len(testDataSmall), duration, float64(len(testDataSmall))/duration.Seconds())
	}
}
