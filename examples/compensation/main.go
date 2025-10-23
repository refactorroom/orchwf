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

	// Step 1: Reserve hotel room
	step1, err := orchwf.NewStepBuilder("reserve_hotel", "Reserve Hotel Room", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 1: Reserving hotel room...")
		time.Sleep(1 * time.Second)

		hotelID := input["hotel_id"].(string)
		roomNumber := fmt.Sprintf("room_%s_%d", hotelID, time.Now().Unix()%1000)

		fmt.Printf("Hotel room reserved: %s\n", roomNumber)
		return map[string]interface{}{
			"room_number": roomNumber,
			"hotel_id":    hotelID,
			"reserved_at": time.Now().Unix(),
			"price":       150.00,
		}, nil
	}).WithDescription("Reserve a hotel room").
		WithCompensator(func(ctx context.Context, input map[string]interface{}) error {
			fmt.Println("Compensating: Cancelling hotel reservation...")
			time.Sleep(500 * time.Millisecond)

			// In a real scenario, this would call the hotel API to cancel
			roomNumber := input["room_number"].(string)
			fmt.Printf("Hotel reservation cancelled for room: %s\n", roomNumber)
			return nil
		}).
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(2).
			WithInitialInterval(1 * time.Second).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Step 2: Book flight
	step2, err := orchwf.NewStepBuilder("book_flight", "Book Flight", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 2: Booking flight...")
		time.Sleep(1500 * time.Millisecond)

		flightNumber := fmt.Sprintf("FL%d", time.Now().Unix()%10000)
		seatNumber := fmt.Sprintf("%dA", (time.Now().Unix()%30)+1)

		fmt.Printf("Flight booked: %s, Seat: %s\n", flightNumber, seatNumber)
		return map[string]interface{}{
			"flight_number": flightNumber,
			"seat_number":   seatNumber,
			"booked_at":     time.Now().Unix(),
			"price":         400.00,
		}, nil
	}).WithDescription("Book a flight").
		WithCompensator(func(ctx context.Context, input map[string]interface{}) error {
			fmt.Println("Compensating: Cancelling flight booking...")
			time.Sleep(800 * time.Millisecond)

			// In a real scenario, this would call the airline API to cancel
			flightNumber := input["flight_number"].(string)
			fmt.Printf("Flight booking cancelled: %s\n", flightNumber)
			return nil
		}).
		WithDependencies("reserve_hotel").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(2).
			WithInitialInterval(1 * time.Second).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Step 3: Reserve car rental (this will fail to demonstrate compensation)
	step3, err := orchwf.NewStepBuilder("reserve_car", "Reserve Car Rental", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 3: Reserving car rental...")
		time.Sleep(1 * time.Second)

		// Simulate failure condition
		userAge := input["user_age"].(int)
		if userAge < 25 {
			return nil, fmt.Errorf("car rental failed: user must be 25 or older")
		}

		carType := "economy"
		licensePlate := fmt.Sprintf("CAR%d", time.Now().Unix()%10000)

		fmt.Printf("Car reserved: %s (%s)\n", licensePlate, carType)
		return map[string]interface{}{
			"car_type":      carType,
			"license_plate": licensePlate,
			"reserved_at":   time.Now().Unix(),
			"price":         80.00,
		}, nil
	}).WithDescription("Reserve a car rental").
		WithCompensator(func(ctx context.Context, input map[string]interface{}) error {
			fmt.Println("Compensating: Cancelling car rental...")
			time.Sleep(500 * time.Millisecond)

			// In a real scenario, this would call the car rental API to cancel
			licensePlate := input["license_plate"].(string)
			fmt.Printf("Car rental cancelled: %s\n", licensePlate)
			return nil
		}).
		WithDependencies("book_flight").
		WithRetryPolicy(orchwf.NewRetryPolicyBuilder().
			WithMaxAttempts(1). // Don't retry for age restriction
			WithInitialInterval(1 * time.Second).
			Build()).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Step 4: Send confirmation (only runs if all previous steps succeed)
	step4, err := orchwf.NewStepBuilder("send_confirmation", "Send Travel Confirmation", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("Step 4: Sending travel confirmation...")
		time.Sleep(1 * time.Second)

		hotelData := input["reserve_hotel"].(map[string]interface{})
		flightData := input["book_flight"].(map[string]interface{})
		carData := input["reserve_car"].(map[string]interface{})

		totalPrice := hotelData["price"].(float64) + flightData["price"].(float64) + carData["price"].(float64)

		confirmation := map[string]interface{}{
			"confirmation_id": fmt.Sprintf("TRAVEL_%d", time.Now().Unix()),
			"hotel":           hotelData,
			"flight":          flightData,
			"car":             carData,
			"total_price":     totalPrice,
			"sent_at":         time.Now().Unix(),
		}

		fmt.Printf("Travel confirmation sent! Total: $%.2f\n", totalPrice)
		return confirmation, nil
	}).WithDescription("Send complete travel confirmation").
		WithDependencies("reserve_car").
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Build workflow
	workflow, err := orchwf.NewWorkflowBuilder("travel_booking", "Travel Booking with Compensation").
		WithDescription("Demonstrates compensation pattern for travel booking").
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

	// Test scenarios
	testCases := []map[string]interface{}{
		{
			"user_name": "John Doe",
			"user_age":  30,
			"hotel_id":  "hotel_001",
			"email":     "john@example.com",
		},
		{
			"user_name": "Jane Smith",
			"user_age":  22, // This will cause car rental to fail
			"hotel_id":  "hotel_002",
			"email":     "jane@example.com",
		},
		{
			"user_name": "Bob Johnson",
			"user_age":  28,
			"hotel_id":  "hotel_003",
			"email":     "bob@example.com",
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("\n=== Test Case %d: %s (Age: %d) ===\n",
			i+1, testCase["user_name"], testCase["user_age"])

		result, err := orchestrator.StartWorkflow(context.Background(), "travel_booking",
			testCase,
			map[string]interface{}{
				"trace_id": fmt.Sprintf("travel_%d", i+1),
			})

		if err != nil {
			fmt.Printf("Travel booking failed: %v\n", err)
			fmt.Println("Compensation steps should have been executed automatically.")
		} else {
			fmt.Printf("Travel booking completed successfully!\n")
			fmt.Printf("Duration: %v\n", result.Duration)

			if confirmation, ok := result.Output["send_confirmation"].(map[string]interface{}); ok {
				fmt.Printf("Confirmation ID: %s\n", confirmation["confirmation_id"])
				fmt.Printf("Total Price: $%.2f\n", confirmation["total_price"])
			}
		}
	}

	// Demonstrate manual compensation (if needed)
	fmt.Println("\n=== Manual Compensation Demo ===")
	fmt.Println("In a real scenario, you might need to manually trigger compensation")
	fmt.Println("for specific workflow instances that failed after partial completion.")

	// This would typically be done through the orchestrator API
	// orchestrator.CompensateWorkflow(context.Background(), "workflow_instance_id")
}
