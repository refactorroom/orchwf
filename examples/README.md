# OrchWF Examples

This directory contains comprehensive examples demonstrating various features and use cases of the OrchWF workflow orchestration library.

## Examples Overview

### 1. Simple (`simple/`)
**Basic workflow execution with sequential steps**
- Demonstrates basic workflow creation and execution
- Shows step dependencies and chaining
- Includes retry policies and timeouts
- Perfect for getting started with OrchWF

```bash
cd simple
go run main.go
```

### 2. Parallel (`parallel/`)
**Concurrent step execution**
- Shows how to run independent steps in parallel
- Demonstrates data aggregation from parallel steps
- Illustrates performance benefits of parallel execution
- Useful for I/O-bound operations

```bash
cd parallel
go run main.go
```

### 3. Error Handling (`error_handling/`)
**Robust error handling and retry mechanisms**
- Demonstrates retry policies with exponential backoff
- Shows optional vs required steps
- Includes error recovery scenarios
- Perfect for understanding fault tolerance

```bash
cd error_handling
go run main.go
```

### 4. Async (`async/`)
**Asynchronous workflow execution**
- Shows how to run multiple workflows concurrently
- Demonstrates async workflow management
- Includes workflow status monitoring
- Useful for high-throughput scenarios

```bash
cd async
go run main.go
```

### 5. Conditional (`conditional/`)
**Conditional workflow execution with branching**
- Demonstrates business logic branching
- Shows different execution paths based on input
- Includes user type-based processing
- Perfect for complex business workflows

```bash
cd conditional
go run main.go
```

### 6. Database (`database/`)
**Database persistence with PostgreSQL**
- Shows database state manager usage
- Demonstrates workflow persistence
- Includes database schema setup
- Perfect for production scenarios

**Prerequisites:**
- PostgreSQL database running
- Update connection string in `main.go`

```bash
cd database
go run main.go
```

### 7. Compensation (`compensation/`)
**Transaction-like behavior with compensation**
- Demonstrates compensation/rollback patterns
- Shows how to handle partial failures
- Includes travel booking scenario
- Perfect for distributed transactions

```bash
cd compensation
go run main.go
```

### 8. Webhook (`webhook/`)
**HTTP integration and webhook processing**
- Shows webhook integration with external services
- Demonstrates HTTP client usage
- Includes retry policies for network calls
- Perfect for microservices integration

```bash
cd webhook
go run main.go
```

### 9. Batch Processing (`batch_processing/`)
**High-volume batch processing**
- Demonstrates processing large datasets
- Shows concurrency control
- Includes performance monitoring
- Perfect for ETL and data processing

```bash
cd batch_processing
go run main.go
```

## Common Patterns Demonstrated

### Workflow Patterns
- **Sequential Processing**: Simple, parallel, conditional examples
- **Parallel Processing**: Parallel, async, batch processing examples
- **Error Handling**: Error handling, compensation examples
- **Integration**: Webhook, database examples

### State Management
- **In-Memory**: All examples use in-memory state manager
- **Database**: Database example shows persistent state
- **Custom State Managers**: Can be implemented for other backends

### Retry Policies
- **Exponential Backoff**: Most examples include retry policies
- **Custom Intervals**: Configurable retry intervals
- **Error-Specific Retries**: Retry only specific error types

### Timeouts and Limits
- **Step Timeouts**: Individual step timeout configuration
- **Workflow Timeouts**: Overall workflow timeout limits
- **Concurrency Limits**: Controlled parallel execution

## Running All Examples

To run all examples and see the full capabilities of OrchWF:

```bash
# Run each example individually
for dir in */; do
  echo "Running $dir"
  cd "$dir"
  go run main.go
  cd ..
  echo "---"
done
```

## Prerequisites

### Required
- Go 1.19 or later
- OrchWF library (see main README for installation)

### Optional
- PostgreSQL (for database example)
- Internet connection (for webhook example)

## Example Customization

Each example is designed to be easily customizable:

1. **Modify Input Data**: Change the test data in each example
2. **Adjust Timeouts**: Modify timeout values for your environment
3. **Add Steps**: Extend workflows with additional steps
4. **Change Retry Policies**: Adjust retry behavior for your needs
5. **Integrate with Your Services**: Replace mock services with real ones

## Best Practices Demonstrated

1. **Error Handling**: Always handle errors gracefully
2. **Retry Policies**: Use appropriate retry strategies
3. **Timeouts**: Set reasonable timeouts for all operations
4. **Logging**: Include meaningful log messages
5. **Monitoring**: Track workflow execution metrics
6. **Testing**: Include various test scenarios
7. **Documentation**: Document workflow behavior

## Contributing

When adding new examples:

1. Create a new directory with a descriptive name
2. Include a comprehensive `main.go` file
3. Add clear comments explaining the example
4. Update this README with the new example
5. Test the example thoroughly
6. Follow the existing code style and patterns

## Support

For questions about these examples or OrchWF in general:
- Check the main documentation
- Review the source code
- Open an issue on GitHub
- Join the community discussions
