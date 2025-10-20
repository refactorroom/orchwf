# OrchWF Migration Package

The `migrate` package provides easy-to-use database migration utilities for OrchWF. It handles the creation and management of all required database tables, indexes, and triggers.

## Features

- **Easy Setup**: One-line database setup with `QuickSetup()`
- **Migration Management**: Full migration lifecycle with version tracking
- **Rollback Support**: Safe rollback of migrations
- **CLI Tool**: Command-line interface for migration management
- **Custom Migrations**: Support for custom migration scripts
- **Multiple Databases**: PostgreSQL support (easily extensible)

## Quick Start

### Method 1: Quick Setup (Recommended for new projects)

```go
package main

import (
    "database/sql"
    "github.com/refactorroom/orchwf/migrate"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // One-line setup!
    if err := migrate.QuickSetup(db); err != nil {
        log.Fatal(err)
    }
    
    // Your database is ready for OrchWF!
}
```

### Method 2: Using Migrator

```go
package main

import (
    "context"
    "database/sql"
    "github.com/refactorroom/orchwf/migrate"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    ctx := context.Background()
    migrator := migrate.NewMigrator(db)
    
    // Apply all pending migrations
    if err := migrator.Migrate(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Check migration status
    if err := migrator.Status(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## CLI Tool

The package includes a command-line tool for migration management:

### Installation

```bash
go install github.com/refactorroom/orchwf/migrate
```

### Usage

```bash
# Apply all pending migrations
orchwf-migrate up -db "postgres://user:pass@localhost/dbname?sslmode=disable"

# Check migration status
orchwf-migrate status -db "postgres://user:pass@localhost/dbname?sslmode=disable"

# Rollback last migration
orchwf-migrate down -db "postgres://user:pass@localhost/dbname?sslmode=disable"

# Show help
orchwf-migrate help
```

### Environment Variables

You can use the `ORCHWF_DB_URL` environment variable instead of the `-db` flag:

```bash
export ORCHWF_DB_URL="postgres://user:pass@localhost/dbname?sslmode=disable"
orchwf-migrate up
```

## API Reference

### Functions

#### `QuickSetup(db *sql.DB) error`
Quickly sets up all OrchWF tables. This is the simplest way to get started.

#### `QuickSetupWithContext(ctx context.Context, db *sql.DB) error`
Same as `QuickSetup` but with context support for timeouts and cancellation.

### Types

#### `Migration`
```go
type Migration struct {
    Version     string  // Unique version identifier
    Description string  // Human-readable description
    Up          string  // SQL to apply the migration
    Down        string  // SQL to rollback the migration
}
```

#### `Migrator`
```go
type Migrator struct {
    db         *sql.DB
    migrations []Migration
}
```

### Methods

#### `NewMigrator(db *sql.DB) *Migrator`
Creates a new migrator with default OrchWF migrations.

#### `NewMigratorWithMigrations(db *sql.DB, migrations []Migration) *Migrator`
Creates a new migrator with custom migrations.

#### `Migrate(ctx context.Context) error`
Applies all pending migrations.

#### `Rollback(ctx context.Context) error`
Rolls back the last applied migration.

#### `Status(ctx context.Context) error`
Shows the status of all migrations.

## Database Schema

The migration creates the following tables:

### `orchwf_workflow_instances`
Stores workflow execution instances.

**Columns:**
- `id` (VARCHAR(36)) - Primary key
- `workflow_id` (VARCHAR(255)) - Workflow definition ID
- `status` (VARCHAR(50)) - Current status
- `input` (JSONB) - Workflow input data
- `output` (JSONB) - Workflow output data
- `context` (JSONB) - Workflow context
- `current_step_id` (VARCHAR(255)) - Currently executing step
- `started_at` (TIMESTAMP) - When workflow started
- `completed_at` (TIMESTAMP) - When workflow completed
- `error` (TEXT) - Error message if failed
- `retry_count` (INT) - Number of retries
- `last_retry_at` (TIMESTAMP) - Last retry timestamp
- `metadata` (JSONB) - Additional metadata
- `trace_id` (VARCHAR(255)) - Tracing ID
- `correlation_id` (VARCHAR(255)) - Correlation ID
- `business_id` (VARCHAR(255)) - Business identifier
- `created_at` (TIMESTAMP) - Record creation time
- `updated_at` (TIMESTAMP) - Record update time

### `orchwf_step_instances`
Stores individual step execution instances.

**Columns:**
- `id` (VARCHAR(36)) - Primary key
- `step_id` (VARCHAR(255)) - Step definition ID
- `workflow_inst_id` (VARCHAR(36)) - Parent workflow instance ID
- `status` (VARCHAR(50)) - Current status
- `input` (JSONB) - Step input data
- `output` (JSONB) - Step output data
- `started_at` (TIMESTAMP) - When step started
- `completed_at` (TIMESTAMP) - When step completed
- `error` (TEXT) - Error message if failed
- `retry_count` (INT) - Number of retries
- `last_retry_at` (TIMESTAMP) - Last retry timestamp
- `duration_ms` (BIGINT) - Step duration in milliseconds
- `execution_order` (INT) - Step execution order
- `created_at` (TIMESTAMP) - Record creation time
- `updated_at` (TIMESTAMP) - Record update time

### `orchwf_workflow_events`
Stores workflow and step events for auditing and monitoring.

**Columns:**
- `id` (VARCHAR(36)) - Primary key
- `workflow_inst_id` (VARCHAR(36)) - Workflow instance ID
- `step_inst_id` (VARCHAR(36)) - Step instance ID (optional)
- `event_type` (VARCHAR(100)) - Event type
- `event_data` (JSONB) - Event data
- `timestamp` (TIMESTAMP) - Event timestamp
- `created_at` (TIMESTAMP) - Record creation time

### Indexes

The migration creates optimized indexes for:
- Workflow lookups by ID, status, trace ID, correlation ID, business ID
- Step lookups by step ID, workflow instance ID, status
- Event lookups by workflow instance ID, step instance ID, event type
- Time-based queries with descending order

### Triggers

Automatic `updated_at` timestamp updates for:
- `orchwf_workflow_instances`
- `orchwf_step_instances`

## Custom Migrations

You can create custom migrations for additional tables or modifications:

```go
customMigrations := []migrate.Migration{
    {
        Version:     "001",
        Description: "Create OrchWF tables",
        Up:          getOrchWFTablesSQL(),
        Down:        getOrchWFTablesRollbackSQL(),
    },
    {
        Version:     "002",
        Description: "Add custom business tables",
        Up:          getCustomTablesSQL(),
        Down:        getCustomTablesRollbackSQL(),
    },
}

migrator := migrate.NewMigratorWithMigrations(db, customMigrations)
err := migrator.Migrate(ctx)
```

## Error Handling

The migration package provides detailed error messages for common issues:

- Database connection failures
- SQL syntax errors
- Migration conflicts
- Rollback failures
- Timeout errors

## Best Practices

1. **Always backup your database** before running migrations in production
2. **Test migrations** in a development environment first
3. **Use transactions** for complex migrations (handled automatically)
4. **Version your migrations** with meaningful version numbers
5. **Write rollback scripts** for all migrations
6. **Monitor migration status** regularly

## Examples

See the `examples/migration_example/` directory for comprehensive examples showing:
- Basic migration setup
- Custom migrations
- CLI tool usage
- Error handling
- Integration with OrchWF workflows

## Support

For issues and questions:
- Check the examples directory
- Review the source code
- Open an issue on GitHub
- Join the community discussions
