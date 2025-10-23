# Priority Queue Testing Guide

This document provides comprehensive testing guidelines for the ORCHWF priority queue feature, including unit tests, integration tests, and performance tests.

## Test Coverage

The priority queue feature includes comprehensive test coverage across multiple test files:

### 1. Basic Functionality Tests (`priority_queue_test.go`)

- **TestPriorityQueueBasic**: Tests basic priority-based execution order
- **TestPriorityWithDependencies**: Tests priority with step dependencies
- **TestPriorityWithAsyncSteps**: Tests priority with async steps
- **TestPriorityDefaultValue**: Tests default priority value (0)
- **TestPriorityNegativeValues**: Tests negative priority values
- **TestPriorityWithMixedSyncAsync**: Tests priority with mixed sync/async steps

### 2. Database Persistence Tests (`priority_queue_db_test.go`)

- **TestPriorityDatabasePersistence**: Tests priority values are stored in database
- **TestPriorityDatabaseRetrieval**: Tests priority values are correctly retrieved
- **TestPriorityDatabaseIndex**: Tests priority-based queries use indexes efficiently
- **TestPriorityDatabaseMigration**: Tests database migration to add priority column

### 3. Edge Cases and Error Scenarios (`priority_queue_edge_cases_test.go`)

- **TestPriorityEdgeCases**: Tests extreme priority values
- **TestPriorityWithIdenticalPriorities**: Tests steps with identical priorities
- **TestPriorityWithLargeNumberOfSteps**: Tests priority with many steps (50+)
- **TestPriorityWithConcurrentExecution**: Tests priority with high concurrency
- **TestPriorityWithStepFailure**: Tests priority behavior when steps fail
- **TestPriorityWithTimeout**: Tests priority behavior with step timeouts
- **TestPriorityWithRetry**: Tests priority behavior with retry logic
- **TestPriorityWithComplexDependencies**: Tests priority with complex dependency chains

## Running the Tests

### Run All Priority Queue Tests

```bash
# Run all priority queue tests
go test -v -run "TestPriority"

# Run specific test categories
go test -v -run "TestPriorityQueue"
go test -v -run "TestPriorityDatabase"
go test -v -run "TestPriorityEdge"
```

### Run Individual Tests

```bash
# Run basic functionality tests
go test -v -run "TestPriorityQueueBasic"

# Run database tests
go test -v -run "TestPriorityDatabasePersistence"

# Run edge case tests
go test -v -run "TestPriorityWithLargeNumberOfSteps"
```

### Run Tests with Coverage

```bash
# Run tests with coverage report
go test -v -cover -run "TestPriority" ./...

# Generate detailed coverage report
go test -v -coverprofile=coverage.out -run "TestPriority" ./...
go tool cover -html=coverage.out -o coverage.html
```

## Test Categories

### 1. Unit Tests

**Purpose**: Test individual components in isolation

**Coverage**:
- Step builder with priority
- Priority field in StepDefinition
- Priority sorting in orchestrator
- Default priority values

**Example**:
```go
func TestPriorityStepBuilder(t *testing.T) {
    step, err := NewStepBuilder("test", "Test", executor).
        WithPriority(15).
        Build()
    
    if step.Priority != 15 {
        t.Errorf("Expected priority 15, got %d", step.Priority)
    }
}
```

### 2. Integration Tests

**Purpose**: Test priority queue with real workflow execution

**Coverage**:
- Priority-based execution order
- Priority with dependencies
- Priority with async steps
- Priority with database persistence

**Example**:
```go
func TestPriorityQueueBasic(t *testing.T) {
    // Create steps with different priorities
    highStep := createStep("high", 10)
    lowStep := createStep("low", -5)
    
    // Execute workflow and verify order
    result := executeWorkflow(highStep, lowStep)
    
    // Verify high priority executes first
    assertExecutionOrder(result, []string{"high", "low"})
}
```

### 3. Performance Tests

**Purpose**: Test priority queue performance with large workloads

**Coverage**:
- Large number of steps (50+)
- High concurrency (20+ async workers)
- Complex dependency chains
- Database query performance

**Example**:
```go
func TestPriorityWithLargeNumberOfSteps(t *testing.T) {
    // Create 50 steps with varying priorities
    steps := createManySteps(50)
    
    // Execute and measure performance
    start := time.Now()
    result := executeWorkflow(steps...)
    duration := time.Since(start)
    
    // Verify performance within acceptable limits
    if duration > 5*time.Second {
        t.Errorf("Execution took too long: %v", duration)
    }
}
```

### 4. Edge Case Tests

**Purpose**: Test boundary conditions and error scenarios

**Coverage**:
- Extreme priority values (±999999)
- Identical priorities
- Step failures with priority
- Timeout scenarios with priority
- Retry logic with priority

**Example**:
```go
func TestPriorityEdgeCases(t *testing.T) {
    // Test with extreme values
    veryHigh := createStep("very_high", 999999)
    veryLow := createStep("very_low", -999999)
    
    // Should still work correctly
    result := executeWorkflow(veryHigh, veryLow)
    assertSuccess(result)
}
```

## Test Data Patterns

### 1. Priority Test Data

```go
// Standard priority levels for testing
const (
    PriorityCritical = 20
    PriorityHigh     = 10
    PriorityNormal   = 0
    PriorityLow       = -10
    PriorityBackground = -20
)

// Test step creation helper
func createTestStep(name string, priority int) *StepDefinition {
    return NewStepBuilder(name, name, func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        return map[string]interface{}{"result": name}, nil
    }).WithPriority(priority).Build()
}
```

### 2. Execution Order Verification

```go
// Helper to verify execution order
func assertExecutionOrder(t *testing.T, actual []string, expected []string) {
    if len(actual) != len(expected) {
        t.Fatalf("Expected %d steps, got %d", len(expected), len(actual))
    }
    
    for i, expectedStep := range expected {
        if actual[i] != expectedStep {
            t.Errorf("Step %d: expected '%s', got '%s'", i, expectedStep, actual[i])
        }
    }
}
```

### 3. Database Test Setup

```go
// Helper to create test database
func createTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    
    // Create tables
    if err := createTestTables(db); err != nil {
        t.Fatalf("Failed to create tables: %v", err)
    }
    
    return db
}
```

## Performance Benchmarks

### Benchmark Priority Sorting

```go
func BenchmarkPrioritySorting(b *testing.B) {
    steps := createManySteps(1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        sort.Slice(steps, func(i, j int) bool {
            return steps[i].Priority > steps[j].Priority
        })
    }
}
```

### Benchmark Priority Execution

```go
func BenchmarkPriorityExecution(b *testing.B) {
    stateManager := NewInMemoryStateManager()
    orchestrator := NewOrchestrator(stateManager)
    
    workflow := createPriorityWorkflow(100)
    orchestrator.RegisterWorkflow(workflow)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        orchestrator.StartWorkflow(context.Background(), "benchmark_workflow", 
            map[string]interface{}{"data": "test"}, nil)
    }
}
```

## Test Assertions

### 1. Priority Value Assertions

```go
// Assert step has correct priority
func assertPriority(t *testing.T, step *StepDefinition, expected int) {
    if step.Priority != expected {
        t.Errorf("Expected priority %d, got %d", expected, step.Priority)
    }
}
```

### 2. Execution Order Assertions

```go
// Assert execution order matches expected
func assertExecutionOrder(t *testing.T, actual []string, expected []string) {
    if !reflect.DeepEqual(actual, expected) {
        t.Errorf("Expected execution order %v, got %v", expected, actual)
    }
}
```

### 3. Database Assertions

```go
// Assert priority is stored in database
func assertPriorityInDB(t *testing.T, db *sql.DB, stepID string, expectedPriority int) {
    var actualPriority int
    err := db.QueryRow("SELECT priority FROM orchwf_step_instances WHERE step_id = ?", stepID).
        Scan(&actualPriority)
    
    if err != nil {
        t.Fatalf("Failed to query priority: %v", err)
    }
    
    if actualPriority != expectedPriority {
        t.Errorf("Expected priority %d in DB, got %d", expectedPriority, actualPriority)
    }
}
```

## Continuous Integration

### GitHub Actions Configuration

```yaml
name: Priority Queue Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Run Priority Queue Tests
      run: |
        go test -v -run "TestPriority" ./...
        go test -v -cover -run "TestPriority" ./...
    
    - name: Run Benchmarks
      run: go test -bench=BenchmarkPriority ./...
```

### Test Coverage Requirements

- **Unit Tests**: 100% coverage for priority-related code
- **Integration Tests**: 90% coverage for workflow execution
- **Edge Cases**: 100% coverage for boundary conditions
- **Performance**: All benchmarks must complete within time limits

## Troubleshooting Tests

### Common Test Issues

1. **Race Conditions**: Use mutexes for shared state in concurrent tests
2. **Timing Issues**: Use appropriate sleep durations for async tests
3. **Database Locks**: Use in-memory databases for isolated tests
4. **Memory Leaks**: Clean up resources in test teardown

### Debug Test Failures

```bash
# Run tests with verbose output
go test -v -run "TestPriorityQueueBasic"

# Run tests with race detection
go test -race -run "TestPriority"

# Run tests with memory profiling
go test -memprofile=mem.prof -run "TestPriority"
go tool pprof mem.prof
```

## Test Maintenance

### Adding New Tests

1. **Follow naming convention**: `TestPriority[FeatureName]`
2. **Use descriptive test names**: Clearly indicate what is being tested
3. **Include setup and teardown**: Clean up resources properly
4. **Add documentation**: Explain complex test scenarios

### Updating Existing Tests

1. **Maintain backward compatibility**: Don't break existing test logic
2. **Update assertions**: Ensure they match current behavior
3. **Review performance**: Update benchmarks if needed
4. **Document changes**: Update this guide when adding new test patterns

## Conclusion

The priority queue testing suite provides comprehensive coverage of all functionality, edge cases, and performance scenarios. Regular execution of these tests ensures the priority queue feature remains reliable and performant across all use cases.
