# OrchWF Migration Guide

This guide explains how to use the OrchWF migration package to easily set up and manage database tables for your workflow orchestration needs.

## Overview

The OrchWF migration package provides:
- **Easy database setup** with one-line functions
- **Migration management** with version tracking
- **Rollback support** for safe database changes
- **CLI tool** for command-line migration management
- **Custom migrations** for additional requirements

## Quick Start

### 1. Basic Setup (Recommended)

```go
package main

import (
    "database/sql"
    "log"
    
    "github.com/akkaraponph/orchwf/migrate"
    _ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
    // Connect to your database
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // One-line database setup!
    if err := migrate.QuickSetup(db); err != nil {
        log.Fatal(err)
    }
    
    // Your database is ready for OrchWF!
}
```

### 2. Using the CLI Tool

```bash
# Install the CLI tool
go install github.com/akkaraponph/orchwf/cmd/orchwf-migrate

# Apply migrations
orchwf-migrate up -db "postgres://user:pass@localhost/dbname?sslmode=disable"

# Check status
orchwf-migrate status -db "postgres://user:pass@localhost/dbname?sslmode=disable"

# Rollback if needed
orchwf-migrate down -db "postgres://user:pass@localhost/dbname?sslmode=disable"
```

## Migration Package Features

### Core Functions

| Function | Description | Usage |
|----------|-------------|-------|
| `QuickSetup(db)` | One-line database setup | New projects, development |
| `QuickSetupWithContext(ctx, db)` | Setup with context | Production with timeouts |
| `NewMigrator(db)` | Create migrator instance | Advanced usage |
| `NewMigratorWithMigrations(db, migrations)` | Custom migrations | Complex setups |

### Migration Management

| Method | Description | Usage |
|--------|-------------|-------|
| `Migrate(ctx)` | Apply pending migrations | Deploy changes |
| `Rollback(ctx)` | Rollback last migration | Undo changes |
| `Status(ctx)` | Show migration status | Check current state |

## Database Schema

The migration creates three main tables:

### 1. `orchwf_workflow_instances`
Stores workflow execution instances with full metadata.

**Key Features:**
- JSONB columns for flexible data storage
- Comprehensive indexing for performance
- Automatic timestamp management
- Support for tracing and correlation

### 2. `orchwf_step_instances`
Stores individual step execution details.

**Key Features:**
- Foreign key relationship to workflows
- Execution order tracking
- Duration measurement
- Retry count tracking

### 3. `orchwf_workflow_events`
Stores audit trail and monitoring events.

**Key Features:**
- Event-based architecture
- Flexible event data storage
- Time-series optimized indexes
- Optional step-level events

## Usage Examples

### Example 1: Simple Setup

```go
package main

import (
    "database/sql"
    "log"
    
    "github.com/akkaraponph/orchwf"
    "github.com/akkaraponph/orchwf/migrate"
    _ "github.com/lib/pq"
)

func main() {
    // Setup database
    db, _ := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
    migrate.QuickSetup(db)
    
    // Use with OrchWF
    stateManager := orchwf.NewDBStateManager(db)
    orchestrator := orchwf.NewOrchestrator(stateManager)
    
    // Your workflows are ready!
}
```

### Example 2: Custom Migrations

```go
package main

import (
    "context"
    "database/sql"
    "github.com/akkaraponph/orchwf/migrate"
    _ "github.com/lib/pq"
)

func main() {
    db, _ := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
    
    // Define custom migrations
    customMigrations := []migrate.Migration{
        {
            Version:     "001",
            Description: "Create OrchWF tables",
            Up:          getOrchWFTablesSQL(),
            Down:        getOrchWFTablesRollbackSQL(),
        },
        {
            Version:     "002",
            Description: "Add business-specific tables",
            Up:          getBusinessTablesSQL(),
            Down:        getBusinessTablesRollbackSQL(),
        },
    }
    
    // Apply custom migrations
    migrator := migrate.NewMigratorWithMigrations(db, customMigrations)
    migrator.Migrate(context.Background())
}
```

### Example 3: Production Setup

```go
package main

import (
    "context"
    "database/sql"
    "time"
    
    "github.com/akkaraponph/orchwf/migrate"
    _ "github.com/lib/pq"
)

func main() {
    db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    
    // Production setup with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := migrate.QuickSetupWithContext(ctx, db); err != nil {
        log.Fatal("Migration failed:", err)
    }
    
    // Check migration status
    migrator := migrate.NewMigrator(db)
    migrator.Status(ctx)
}
```

## CLI Tool Usage

### Installation

```bash
# Install from source
go install github.com/akkaraponph/orchwf/cmd/orchwf-migrate

# Or build locally
go build -o orchwf-migrate github.com/akkaraponph/orchwf/cmd/orchwf-migrate
```

### Commands

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

```bash
# Set database URL
export ORCHWF_DB_URL="postgres://user:pass@localhost/dbname?sslmode=disable"

# Use without -db flag
orchwf-migrate up
orchwf-migrate status
```

## Integration with Examples

The migration package is integrated into the database example:

```go
// Before (manual table creation)
if err := createTables(db); err != nil {
    log.Fatal("Failed to create tables:", err)
}

// After (using migration package)
if err := migrate.QuickSetup(db); err != nil {
    log.Fatal("Failed to setup database tables:", err)
}
```

## Best Practices

### 1. Development
- Use `QuickSetup()` for rapid prototyping
- Test migrations in development first
- Use version control for migration files

### 2. Production
- Use `QuickSetupWithContext()` with timeouts
- Always backup before migrations
- Monitor migration status
- Use the CLI tool for operations

### 3. Custom Migrations
- Write both UP and DOWN scripts
- Use meaningful version numbers
- Test rollback procedures
- Document migration purposes

## Troubleshooting

### Common Issues

1. **Connection Errors**
   ```
   Error: failed to connect to database
   Solution: Check database URL and credentials
   ```

2. **Migration Conflicts**
   ```
   Error: migration already applied
   Solution: Check migration status first
   ```

3. **Rollback Failures**
   ```
   Error: no rollback script available
   Solution: Ensure DOWN script is provided
   ```

### Debug Tips

1. **Check Migration Status**
   ```bash
   orchwf-migrate status -db "your-db-url"
   ```

2. **Verify Database Connection**
   ```go
   if err := db.Ping(); err != nil {
       log.Fatal("Database connection failed:", err)
   }
   ```

3. **Check Table Creation**
   ```sql
   \dt orchwf_*
   ```

## Migration Files

The package includes a comprehensive migration file at:
- `migrations/001_create_orchwf_tables.sql`

This file contains:
- Complete table definitions
- Optimized indexes
- Automatic triggers
- Foreign key constraints
- Rollback scripts

## Examples Directory

See the following examples for complete usage:
- `examples/database/` - Basic database usage
- `examples/migration_example/` - Migration package examples

## Support

For issues and questions:
- Check the examples directory
- Review the migration package source
- Open an issue on GitHub
- Join community discussions

The migration package makes it easy to get started with OrchWF while providing the flexibility needed for complex production environments.
