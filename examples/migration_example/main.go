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
	// Database connection string
	dbURL := "postgres://postgres:password@localhost:5432/orchwf?sslmode=disable"

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("=== OrchWF Migration Example ===")
	fmt.Println()

	// Method 1: Quick Setup (simplest way)
	fmt.Println("1. Quick Setup Method:")
	if err := migrate.QuickSetup(db); err != nil {
		log.Printf("Quick setup failed: %v", err)
	} else {
		fmt.Println("✓ Database tables created successfully using QuickSetup()")
	}
	fmt.Println()

	// Method 2: Using Migrator with Context
	fmt.Println("2. Using Migrator with Context:")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	migrator := migrate.NewMigrator(db)

	// Check migration status
	fmt.Println("Checking migration status...")
	if err := migrator.Status(ctx); err != nil {
		log.Printf("Failed to check status: %v", err)
	} else {
		fmt.Println("✓ Migration status checked successfully")
	}
	fmt.Println()

	// Method 3: Custom Migrations
	fmt.Println("3. Using Custom Migrations:")
	customMigrations := []migrate.Migration{
		{
			Version:     "001",
			Description: "Create OrchWF tables",
			Up:          getCustomMigrationSQL(),
			Down:        getCustomRollbackSQL(),
		},
		{
			Version:     "002",
			Description: "Add custom indexes",
			Up:          getCustomIndexesSQL(),
			Down:        getCustomIndexesRollbackSQL(),
		},
	}

	customMigrator := migrate.NewMigratorWithMigrations(db, customMigrations)
	if err := customMigrator.Migrate(ctx); err != nil {
		log.Printf("Custom migration failed: %v", err)
	} else {
		fmt.Println("✓ Custom migrations applied successfully")
	}
	fmt.Println()

	// Method 4: Load from File
	fmt.Println("4. Loading Migrations from File:")
	// This would load from a SQL file
	// migrations, err := migrate.LoadMigrationsFromFile("migrations/custom_migration.sql")
	// if err != nil {
	//     log.Printf("Failed to load migrations from file: %v", err)
	// } else {
	//     fileMigrator := migrate.NewMigratorWithMigrations(db, migrations)
	//     if err := fileMigrator.Migrate(ctx); err != nil {
	//         log.Printf("File migration failed: %v", err)
	//     } else {
	//         fmt.Println("✓ File migrations applied successfully")
	//     }
	// }
	fmt.Println("(File loading example commented out - would require actual SQL file)")
	fmt.Println()

	// Demonstrate using the migrated database
	fmt.Println("5. Using Migrated Database:")
	demonstrateDatabaseUsage(db)
}

// demonstrateDatabaseUsage shows how to use the migrated database
func demonstrateDatabaseUsage(db *sql.DB) {
	// Create database state manager
	stateManager := orchwf.NewDBStateManager(db)

	// Create orchestrator
	orchestrator := orchwf.NewOrchestrator(stateManager)

	// Define a simple workflow
	step1, err := orchwf.NewStepBuilder("test_step", "Test Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		fmt.Println("  Executing test step...")
		time.Sleep(100 * time.Millisecond)
		return map[string]interface{}{
			"result":    "success",
			"timestamp": time.Now().Unix(),
		}, nil
	}).WithDescription("A simple test step").
		Build()

	if err != nil {
		log.Printf("Failed to create step: %v", err)
		return
	}

	// Build workflow
	workflow, err := orchwf.NewWorkflowBuilder("migration_test", "Migration Test Workflow").
		WithDescription("Test workflow to verify database migration").
		WithVersion("1.0.0").
		AddStep(step1).
		Build()

	if err != nil {
		log.Printf("Failed to create workflow: %v", err)
		return
	}

	// Register workflow
	if err := orchestrator.RegisterWorkflow(workflow); err != nil {
		log.Printf("Failed to register workflow: %v", err)
		return
	}

	// Execute workflow
	fmt.Println("  Executing test workflow...")
	result, err := orchestrator.StartWorkflow(context.Background(), "migration_test",
		map[string]interface{}{
			"test_data": "migration_test",
		},
		map[string]interface{}{
			"trace_id": "migration_test_trace",
		})

	if err != nil {
		log.Printf("Workflow execution failed: %v", err)
		return
	}

	fmt.Printf("  ✓ Workflow completed successfully in %v\n", result.Duration)
	fmt.Println("  ✓ Database migration is working correctly!")
}

// getCustomMigrationSQL returns custom migration SQL
func getCustomMigrationSQL() string {
	return `-- Custom migration example
CREATE TABLE IF NOT EXISTS orchwf_custom_table (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`
}

// getCustomRollbackSQL returns custom rollback SQL
func getCustomRollbackSQL() string {
	return `-- Custom rollback example
DROP TABLE IF EXISTS orchwf_custom_table;`
}

// getCustomIndexesSQL returns custom indexes SQL
func getCustomIndexesSQL() string {
	return `-- Custom indexes example
CREATE INDEX IF NOT EXISTS idx_orchwf_custom_table_name ON orchwf_custom_table(name);
CREATE INDEX IF NOT EXISTS idx_orchwf_custom_table_created_at ON orchwf_custom_table(created_at DESC);`
}

// getCustomIndexesRollbackSQL returns custom indexes rollback SQL
func getCustomIndexesRollbackSQL() string {
	return `-- Custom indexes rollback example
DROP INDEX IF EXISTS idx_orchwf_custom_table_created_at;
DROP INDEX IF EXISTS idx_orchwf_custom_table_name;`
}
