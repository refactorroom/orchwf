package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/refactorroom/orchwf"
	"github.com/refactorroom/orchwf/migrate"
)

func main() {
	// Database connection string - adjust for your environment
	// For this example, we'll use a simple in-memory SQLite database
	// In production, you would use PostgreSQL, MySQL, or another database
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=password dbname=orchwf sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Create database tables using the migration package
	fmt.Println("Setting up database tables...")
	if err := migrate.QuickSetup(db); err != nil {
		log.Fatal("Failed to setup database tables:", err)
	}
	fmt.Println("✓ Database tables created successfully!")

	// Create database state manager
	stateManager := orchwf.NewDBStateManager(db)

	// Create orchestrator
	orchestrator := orchwf.NewOrchestrator(stateManager)

	// Define workflow steps
	step1, err := orchwf.NewStepBuilder("create_order", "Create Order", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 1: Creating order in database...")
		time.Sleep(1 * time.Second)

		orderID := fmt.Sprintf("order_%d", time.Now().Unix())
		return map[string]interface{}{
			"order_id":    orderID,
			"customer_id": input["customer_id"],
			"amount":      input["amount"],
			"status":      "created",
			"created_at":  time.Now().Unix(),
		}, nil
	}).WithDescription("Create order record in database").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(3).
			WithInitialInterval(1 * time.Second).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	step2, err := orchwf.NewStepBuilder("validate_payment", "Validate Payment", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 2: Validating payment...")
		time.Sleep(2 * time.Second)

		amount := input["amount"].(float64)

		// Simulate payment validation
		if amount > 1000 {
			return nil, fmt.Errorf("payment validation failed: amount too high")
		}

		return map[string]interface{}{
			"payment_valid": true,
			"payment_id":    fmt.Sprintf("pay_%s", input["order_id"]),
			"processed_at":  time.Now().Unix(),
		}, nil
	}).WithDescription("Validate payment for the order").
		WithDependencies("create_order").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(2).
			WithInitialInterval(2 * time.Second).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	step3, err := orchwf.NewStepBuilder("update_inventory", "Update Inventory", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 3: Updating inventory...")
		time.Sleep(1 * time.Second)

		return map[string]interface{}{
			"inventory_updated": true,
			"items_reserved":    []string{"item_001", "item_002"},
			"reserved_at":       time.Now().Unix(),
		}, nil
	}).WithDescription("Update inventory for ordered items").
		WithDependencies("validate_payment").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	step4, err := orchwf.NewStepBuilder("send_confirmation", "Send Confirmation", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 4: Sending confirmation email...")
		time.Sleep(1 * time.Second)

		orderData := input["create_order"].(map[string]interface{})
		return map[string]interface{}{
			"email_sent": true,
			"email_id":   fmt.Sprintf("email_%s", orderData["order_id"]),
			"sent_at":    time.Now().Unix(),
			"recipient":  input["customer_email"],
		}, nil
	}).WithDescription("Send order confirmation email").
		WithDependencies("update_inventory").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Build workflow
	workflow, err := orchwf.NewWorkflowBuilder("order_processing", "Order Processing Workflow").
		WithDescription("Process customer orders with database persistence").
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

	// Test with multiple orders
	orders := []map[string]interface{}{
		{
			"customer_id":    123,
			"customer_email": "john@example.com",
			"amount":         99.99,
		},
		{
			"customer_id":    456,
			"customer_email": "jane@example.com",
			"amount":         1500.00, // This will fail payment validation
		},
		{
			"customer_id":    789,
			"customer_email": "bob@example.com",
			"amount":         299.99,
		},
	}

	for i, order := range orders {
		fmt.Printf("\n=== Processing Order %d ===\n", i+1)

		result, err := orchestrator.StartWorkflow(context.Background(), "order_processing",
			order,
			map[string]interface{}{
				"trace_id": fmt.Sprintf("order_trace_%d", i+1),
				"source":   "web",
			})

		if err != nil {
			fmt.Printf("Order processing failed: %v\n", err)
		} else {
			fmt.Printf("Order processed successfully!\n")
			fmt.Printf("Duration: %v\n", result.Duration)
			fmt.Printf("Order ID: %s\n", result.Output["create_order"].(map[string]interface{})["order_id"])
		}
	}

	// Demonstrate workflow status persistence
	fmt.Println("\n=== Checking Workflow Statuses ===")

	// Get all workflow instances from database using ListWorkflows
	workflows, total, err := stateManager.ListWorkflows(context.Background(), map[string]interface{}{}, 100, 0)
	if err != nil {
		log.Printf("Failed to get workflows: %v", err)
	} else {
		fmt.Printf("Found %d total workflows\n", total)

		completedCount := 0
		failedCount := 0
		for _, wf := range workflows {
			if wf.Status == orchwf.WorkflowStatusCompleted {
				completedCount++
				fmt.Printf("  ✓ Workflow %s: %s (Duration: %v)\n",
					wf.ID, wf.WorkflowID, wf.CompletedAt.Sub(wf.StartedAt))
			} else if wf.Status == orchwf.WorkflowStatusFailed {
				failedCount++
				errMsg := "unknown error"
				if wf.Error != nil {
					errMsg = *wf.Error
				}
				fmt.Printf("  ✗ Workflow %s: %s (Error: %s)\n",
					wf.ID, wf.WorkflowID, errMsg)
			}
		}

		fmt.Printf("\nSummary: %d completed, %d failed\n", completedCount, failedCount)
	}
}
