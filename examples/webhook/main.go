package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/refactorroom/orchwf"
)

// WebhookClient simulates external webhook calls
type WebhookClient struct {
	baseURL string
	client  *http.Client
}

func NewWebhookClient(baseURL string) *WebhookClient {
	return &WebhookClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *WebhookClient) SendWebhook(ctx context.Context, endpoint string, payload map[string]interface{}) error {
	url := w.baseURL + endpoint

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "orchwf-webhook-client/1.0")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func main() {
	// Create in-memory state manager
	stateManager := orchwf.NewInMemoryStateManager()

	// Create orchestrator
	orchestrator := orchwf.NewOrchestrator(stateManager)

	// Create webhook client (using a mock service for demo)
	webhookClient := NewWebhookClient("https://httpbin.org") // httpbin.org is a testing service

	// Step 1: Send order notification webhook
	step1, err := orchwf.NewStepBuilder("send_order_webhook", "Send Order Webhook", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 1: Sending order notification webhook...")

		payload := map[string]interface{}{
			"event_type":  "order.created",
			"order_id":    input["order_id"],
			"customer_id": input["customer_id"],
			"amount":      input["amount"],
			"timestamp":   time.Now().Unix(),
		}

		// Send webhook to order service
		err := webhookClient.SendWebhook(ctx, "/post", payload)
		if err != nil {
			return nil, fmt.Errorf("order webhook failed: %v", err)
		}

		fmt.Printf("Order webhook sent successfully for order %s\n", input["order_id"])
		return map[string]interface{}{
			"webhook_sent": true,
			"webhook_type": "order.created",
			"sent_at":      time.Now().Unix(),
		}, nil
	}).WithDescription("Send order creation webhook to external service").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(3).
			WithInitialInterval(2*time.Second).
			WithMaxInterval(10*time.Second).
			WithMultiplier(2.0).
			WithRetryableErrors("webhook request failed", "timeout").
			Build()).
		WithTimeout(30 * time.Second).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Step 2: Send inventory update webhook
	step2, err := orchwf.NewStepBuilder("send_inventory_webhook", "Send Inventory Webhook", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 2: Sending inventory update webhook...")

		payload := map[string]interface{}{
			"event_type": "inventory.updated",
			"order_id":   input["order_id"],
			"items":      input["items"],
			"action":     "reserve",
			"timestamp":  time.Now().Unix(),
		}

		// Send webhook to inventory service
		err := webhookClient.SendWebhook(ctx, "/post", payload)
		if err != nil {
			return nil, fmt.Errorf("inventory webhook failed: %v", err)
		}

		fmt.Printf("Inventory webhook sent successfully for order %s\n", input["order_id"])
		return map[string]interface{}{
			"webhook_sent": true,
			"webhook_type": "inventory.updated",
			"sent_at":      time.Now().Unix(),
		}, nil
	}).WithDescription("Send inventory update webhook").
		WithDependencies("send_order_webhook").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(3).
			WithInitialInterval(1 * time.Second).
			WithMaxInterval(5 * time.Second).
			WithMultiplier(1.5).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Step 3: Send payment webhook
	step3, err := orchwf.NewStepBuilder("send_payment_webhook", "Send Payment Webhook", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 3: Sending payment webhook...")

		payload := map[string]interface{}{
			"event_type": "payment.processed",
			"order_id":   input["order_id"],
			"amount":     input["amount"],
			"currency":   "USD",
			"status":     "completed",
			"timestamp":  time.Now().Unix(),
		}

		// Send webhook to payment service
		err := webhookClient.SendWebhook(ctx, "/post", payload)
		if err != nil {
			return nil, fmt.Errorf("payment webhook failed: %v", err)
		}

		fmt.Printf("Payment webhook sent successfully for order %s\n", input["order_id"])
		return map[string]interface{}{
			"webhook_sent": true,
			"webhook_type": "payment.processed",
			"sent_at":      time.Now().Unix(),
		}, nil
	}).WithDescription("Send payment confirmation webhook").
		WithDependencies("send_inventory_webhook").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(3).
			WithInitialInterval(1 * time.Second).
			WithMaxInterval(5 * time.Second).
			WithMultiplier(1.5).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Step 4: Send notification webhook
	step4, err := orchwf.NewStepBuilder("send_notification_webhook", "Send Notification Webhook", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 4: Sending customer notification webhook...")

		payload := map[string]interface{}{
			"event_type":  "notification.order_confirmed",
			"order_id":    input["order_id"],
			"customer_id": input["customer_id"],
			"email":       input["customer_email"],
			"message":     "Your order has been confirmed and is being processed",
			"timestamp":   time.Now().Unix(),
		}

		// Send webhook to notification service
		err := webhookClient.SendWebhook(ctx, "/post", payload)
		if err != nil {
			return nil, fmt.Errorf("notification webhook failed: %v", err)
		}

		fmt.Printf("Notification webhook sent successfully for order %s\n", input["order_id"])
		return map[string]interface{}{
			"webhook_sent": true,
			"webhook_type": "notification.order_confirmed",
			"sent_at":      time.Now().Unix(),
		}, nil
	}).WithDescription("Send customer notification webhook").
		WithDependencies("send_payment_webhook").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(2).
			WithInitialInterval(1 * time.Second).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Build workflow
	workflow, err := orchwf.NewWorkflowBuilder("webhook_integration", "Webhook Integration Workflow").
		WithDescription("Demonstrates webhook integration with external services").
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

	// Test with sample orders
	orders := []map[string]interface{}{
		{
			"order_id":       "order_001",
			"customer_id":    123,
			"customer_email": "john@example.com",
			"amount":         99.99,
			"items":          []string{"item_001", "item_002"},
		},
		{
			"order_id":       "order_002",
			"customer_id":    456,
			"customer_email": "jane@example.com",
			"amount":         149.99,
			"items":          []string{"item_003", "item_004", "item_005"},
		},
		{
			"order_id":       "order_003",
			"customer_id":    789,
			"customer_email": "bob@example.com",
			"amount":         299.99,
			"items":          []string{"item_006"},
		},
	}

	for i, order := range orders {
		fmt.Printf("\n=== Processing Order %d: %s ===\n", i+1, order["order_id"])

		result, err := orchestrator.StartWorkflow(context.Background(), "webhook_integration",
			order,
			map[string]interface{}{
				"trace_id": fmt.Sprintf("webhook_trace_%d", i+1),
				"source":   "api",
			})

		if err != nil {
			fmt.Printf("Webhook workflow failed: %v\n", err)
		} else {
			fmt.Printf("Webhook workflow completed successfully!\n")
			fmt.Printf("Duration: %v\n", result.Duration)

			// Count successful webhooks
			webhookCount := 0
			for stepName, stepOutput := range result.Output {
				if stepData, ok := stepOutput.(map[string]interface{}); ok {
					if sent, exists := stepData["webhook_sent"]; exists && sent.(bool) {
						webhookCount++
						fmt.Printf("  ✓ %s: %s\n", stepName, stepData["webhook_type"])
					}
				}
			}
			fmt.Printf("Total webhooks sent: %d\n", webhookCount)
		}

		// Add delay between orders to avoid rate limiting
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n=== Webhook Integration Summary ===")
	fmt.Println("This example demonstrates:")
	fmt.Println("- Sending webhooks to external services")
	fmt.Println("- Retry policies for webhook failures")
	fmt.Println("- Sequential webhook execution")
	fmt.Println("- Error handling and timeout management")
	fmt.Println("- Integration with real HTTP services (httpbin.org)")
}
