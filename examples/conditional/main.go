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

	// Initial step to determine user type
	step1, err := orchwf.NewStepBuilder("determine_user_type", "Determine User Type", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 1: Analyzing user data...")
		time.Sleep(1 * time.Second)

		userID := input["user_id"].(int)
		userType := "premium"

		// Simulate business logic to determine user type
		if userID%3 == 0 {
			userType = "basic"
		} else if userID%5 == 0 {
			userType = "enterprise"
		}

		fmt.Printf("User type determined: %s\n", userType)
		return map[string]interface{}{
			"user_type": userType,
			"user_id":   userID,
		}, nil
	}).WithDescription("Determine user type based on business rules").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Premium user specific step
	step2, err := orchwf.NewStepBuilder("premium_features", "Premium Features", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 2: Enabling premium features...")
		time.Sleep(1 * time.Second)
		return map[string]interface{}{
			"premium_enabled": true,
			"features":        []string{"advanced_analytics", "priority_support", "custom_themes"},
		}, nil
	}).WithDescription("Enable premium features for premium users").
		WithDependencies("determine_user_type").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Enterprise user specific step
	step3, err := orchwf.NewStepBuilder("enterprise_setup", "Enterprise Setup", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 3: Setting up enterprise features...")
		time.Sleep(2 * time.Second)
		return map[string]interface{}{
			"enterprise_enabled": true,
			"features":           []string{"sso", "audit_logs", "custom_integrations", "dedicated_support"},
			"admin_contact":      "admin@company.com",
		}, nil
	}).WithDescription("Setup enterprise features for enterprise users").
		WithDependencies("determine_user_type").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Basic user step
	step4, err := orchwf.NewStepBuilder("basic_setup", "Basic Setup", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 4: Setting up basic features...")
		time.Sleep(500 * time.Millisecond)
		return map[string]interface{}{
			"basic_enabled": true,
			"features":      []string{"standard_analytics", "email_support"},
		}, nil
	}).WithDescription("Setup basic features for basic users").
		WithDependencies("determine_user_type").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Conditional step that runs different logic based on user type
	step5, err := orchwf.NewStepBuilder("conditional_processing", "Conditional Processing", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 5: Processing based on user type...")
		time.Sleep(1 * time.Second)

		// Get user type from previous step
		userTypeData := input["determine_user_type"].(map[string]interface{})
		userType := userTypeData["user_type"].(string)

		var processingResult map[string]interface{}

		switch userType {
		case "premium":
			premiumData := input["premium_features"].(map[string]interface{})
			processingResult = map[string]interface{}{
				"processing_type": "premium",
				"features":        premiumData["features"],
				"quota":           "unlimited",
				"support_level":   "priority",
			}
		case "enterprise":
			enterpriseData := input["enterprise_setup"].(map[string]interface{})
			processingResult = map[string]interface{}{
				"processing_type": "enterprise",
				"features":        enterpriseData["features"],
				"quota":           "unlimited",
				"support_level":   "dedicated",
				"admin_contact":   enterpriseData["admin_contact"],
			}
		case "basic":
			basicData := input["basic_setup"].(map[string]interface{})
			processingResult = map[string]interface{}{
				"processing_type": "basic",
				"features":        basicData["features"],
				"quota":           "1000_requests_per_month",
				"support_level":   "email",
			}
		default:
			return nil, fmt.Errorf("unknown user type: %s", userType)
		}

		return processingResult, nil
	}).WithDescription("Process user based on their type").
		WithDependencies("determine_user_type").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Final step that creates user profile
	step6, err := orchwf.NewStepBuilder("create_profile", "Create User Profile", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 6: Creating user profile...")
		time.Sleep(1 * time.Second)

		userTypeData := input["determine_user_type"].(map[string]interface{})
		processingData := input["conditional_processing"].(map[string]interface{})

		profile := map[string]interface{}{
			"user_id":         userTypeData["user_id"],
			"user_type":       userTypeData["user_type"],
			"profile_created": true,
			"created_at":      time.Now().Unix(),
			"processing_info": processingData,
		}

		// Add type-specific data if available
		if userTypeData["user_type"] == "premium" {
			if premiumData, exists := input["premium_features"]; exists {
				profile["premium_data"] = premiumData
			}
		} else if userTypeData["user_type"] == "enterprise" {
			if enterpriseData, exists := input["enterprise_setup"]; exists {
				profile["enterprise_data"] = enterpriseData
			}
		} else if userTypeData["user_type"] == "basic" {
			if basicData, exists := input["basic_setup"]; exists {
				profile["basic_data"] = basicData
			}
		}

		return profile, nil
	}).WithDescription("Create final user profile").
		WithDependencies("conditional_processing").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Build workflow
	workflow, err := orchwf.NewWorkflowBuilder("conditional_user_setup", "Conditional User Setup").
		WithDescription("Demonstrates conditional workflow execution based on user type").
		WithVersion("1.0.0").
		AddStep(step1).
		AddStep(step2).
		AddStep(step3).
		AddStep(step4).
		AddStep(step5).
		AddStep(step6).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Register workflow
	if err := orchestrator.RegisterWorkflow(workflow); err != nil {
		log.Fatal(err)
	}

	// Test with different user types
	testUsers := []int{1, 3, 5, 7, 9, 12, 15, 18, 21, 25}

	for _, userID := range testUsers {
		fmt.Printf("\n=== Testing with User ID: %d ===\n", userID)

		result, err := orchestrator.StartWorkflow(context.Background(), "conditional_user_setup",
			map[string]interface{}{
				"user_id": userID,
			},
			map[string]interface{}{
				"trace_id": fmt.Sprintf("conditional_%d", userID),
			})

		if err != nil {
			log.Printf("Workflow failed for user %d: %v", userID, err)
			continue
		}

		fmt.Printf("Workflow completed successfully!\n")
		fmt.Printf("Duration: %v\n", result.Duration)

		// Display the created profile
		if profile, ok := result.Output["create_profile"].(map[string]interface{}); ok {
			fmt.Printf("User Profile:\n")
			fmt.Printf("  Type: %s\n", profile["user_type"])
			fmt.Printf("  Features: %v\n", profile["processing_info"].(map[string]interface{})["features"])
			fmt.Printf("  Quota: %s\n", profile["processing_info"].(map[string]interface{})["quota"])
			fmt.Printf("  Support: %s\n", profile["processing_info"].(map[string]interface{})["support_level"])
		}
	}
}
