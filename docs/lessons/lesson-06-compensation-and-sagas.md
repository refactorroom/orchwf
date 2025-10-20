# Lesson 6: Compensation and Sagas

## Overview

Use compensation to undo side-effects when multi-step workflows fail mid-flight.

## Compensation Basics

Each step can include a compensator that runs if the workflow needs to roll back effects of completed steps.

```go
reserveHotel, _ := orchwf.NewStepBuilder("reserve_hotel", "Reserve Hotel", reserveHotelExec).
    WithCompensator(func(ctx context.Context, input map[string]interface{}) error {
        // call provider to cancel reservation
        return nil
    }).
    Build()
```

## Saga Pattern

1. Execute steps that create side-effects (reservations, payments, writes)
2. On failure of a downstream step, run compensators for previously completed steps in reverse order

## Example

```go
bookFlight, _ := orchwf.NewStepBuilder("book_flight", "Book Flight", bookFlightExec).
    WithCompensator(cancelFlight).
    WithDependencies("reserve_hotel").
    Build()

reserveCar, _ := orchwf.NewStepBuilder("reserve_car", "Reserve Car", reserveCarExec).
    WithCompensator(cancelCar).
    WithDependencies("book_flight").
    Build()

// If reserve_car fails (e.g., age restriction), compensators for book_flight and reserve_hotel will run.
```

## Best Practices

- Make compensators idempotent (safe to call multiple times)
- Store external reference IDs required for compensation
- Compensate only completed steps; skipped/failed steps don't need compensation
- Log both forward actions and compensations for auditability

## When Not to Compensate

- Pure reads or ephemeral cache writes
- Steps with no external side-effects

## Next Steps
We'll now explore database persistence to make workflows durable across restarts.


