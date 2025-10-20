package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Up          string
	Down        string
}

// Migrator handles database migrations
type Migrator struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: getDefaultMigrations(),
	}
}

// NewMigratorWithMigrations creates a new migrator with custom migrations
func NewMigratorWithMigrations(db *sql.DB, migrations []Migration) *Migrator {
	return &Migrator{
		db:         db,
		migrations: migrations,
	}
}

// Migrate runs all pending migrations
func (m *Migrator) Migrate(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	// Apply pending migrations
	for _, migration := range m.migrations {
		if !applied[migration.Version] {
			if err := m.applyMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %v", migration.Version, err)
			}
			fmt.Printf("Applied migration: %s - %s\n", migration.Version, migration.Description)
		}
	}

	return nil
}

// Rollback rolls back the last migration
func (m *Migrator) Rollback(ctx context.Context) error {
	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	// Find the last applied migration
	var lastMigration *Migration
	for i := len(m.migrations) - 1; i >= 0; i-- {
		if applied[m.migrations[i].Version] {
			lastMigration = &m.migrations[i]
			break
		}
	}

	if lastMigration == nil {
		return fmt.Errorf("no migrations to rollback")
	}

	// Rollback the migration
	if err := m.rollbackMigration(ctx, *lastMigration); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %v", lastMigration.Version, err)
	}

	fmt.Printf("Rolled back migration: %s - %s\n", lastMigration.Version, lastMigration.Description)
	return nil
}

// Status shows the status of all migrations
func (m *Migrator) Status(ctx context.Context) error {
	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	fmt.Println("Migration Status:")
	fmt.Println("=================")
	for _, migration := range m.migrations {
		status := "PENDING"
		if applied[migration.Version] {
			status = "APPLIED"
		}
		fmt.Printf("%s - %s: %s\n", migration.Version, migration.Description, status)
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS orchwf_migrations (
		version VARCHAR(255) PRIMARY KEY,
		description TEXT,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getAppliedMigrations returns a map of applied migration versions
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	query := `SELECT version FROM orchwf_migrations ORDER BY applied_at`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute the migration
	if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
		return err
	}

	// Record the migration
	query := `INSERT INTO orchwf_migrations (version, description) VALUES ($1, $2)`
	if _, err := tx.ExecContext(ctx, query, migration.Version, migration.Description); err != nil {
		return err
	}

	return tx.Commit()
}

// rollbackMigration rolls back a single migration
func (m *Migrator) rollbackMigration(ctx context.Context, migration Migration) error {
	if migration.Down == "" {
		return fmt.Errorf("no rollback script for migration %s", migration.Version)
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute the rollback
	if _, err := tx.ExecContext(ctx, migration.Down); err != nil {
		return err
	}

	// Remove the migration record
	query := `DELETE FROM orchwf_migrations WHERE version = $1`
	if _, err := tx.ExecContext(ctx, query, migration.Version); err != nil {
		return err
	}

	return tx.Commit()
}

// getDefaultMigrations returns the default OrchWF migrations
func getDefaultMigrations() []Migration {
	return []Migration{
		{
			Version:     "001",
			Description: "Create OrchWF tables",
			Up:          getOrchWFTablesSQL(),
			Down:        getOrchWFTablesRollbackSQL(),
		},
	}
}

// getOrchWFTablesSQL returns the SQL for creating OrchWF tables
func getOrchWFTablesSQL() string {
	return `-- Create ORCHWF (Workflow Orchestration) tables

-- Workflow instances table
CREATE TABLE IF NOT EXISTS orchwf_workflow_instances (
    id VARCHAR(36) PRIMARY KEY,
    workflow_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    input JSONB DEFAULT '{}',
    output JSONB DEFAULT '{}',
    context JSONB DEFAULT '{}',
    current_step_id VARCHAR(255),
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    error TEXT,
    retry_count INT DEFAULT 0,
    last_retry_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    trace_id VARCHAR(255),
    correlation_id VARCHAR(255),
    business_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for workflow instances
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_workflow_id ON orchwf_workflow_instances(workflow_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_status ON orchwf_workflow_instances(status);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_trace_id ON orchwf_workflow_instances(trace_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_correlation_id ON orchwf_workflow_instances(correlation_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_business_id ON orchwf_workflow_instances(business_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_created_at ON orchwf_workflow_instances(created_at DESC);

-- Step instances table
CREATE TABLE IF NOT EXISTS orchwf_step_instances (
    id VARCHAR(36) PRIMARY KEY,
    step_id VARCHAR(255) NOT NULL,
    workflow_inst_id VARCHAR(36) NOT NULL,
    status VARCHAR(50) NOT NULL,
    input JSONB DEFAULT '{}',
    output JSONB DEFAULT '{}',
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error TEXT,
    retry_count INT DEFAULT 0,
    last_retry_at TIMESTAMP,
    duration_ms BIGINT DEFAULT 0,
    execution_order INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (workflow_inst_id) REFERENCES orchwf_workflow_instances(id) ON DELETE CASCADE
);

-- Create indexes for step instances
CREATE INDEX IF NOT EXISTS idx_orchwf_step_instances_step_id ON orchwf_step_instances(step_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_step_instances_workflow_inst_id ON orchwf_step_instances(workflow_inst_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_step_instances_status ON orchwf_step_instances(status);
CREATE INDEX IF NOT EXISTS idx_orchwf_step_instances_execution_order ON orchwf_step_instances(workflow_inst_id, execution_order);

-- Workflow events table
CREATE TABLE IF NOT EXISTS orchwf_workflow_events (
    id VARCHAR(36) PRIMARY KEY,
    workflow_inst_id VARCHAR(36) NOT NULL,
    step_inst_id VARCHAR(36),
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB DEFAULT '{}',
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (workflow_inst_id) REFERENCES orchwf_workflow_instances(id) ON DELETE CASCADE,
    FOREIGN KEY (step_inst_id) REFERENCES orchwf_step_instances(id) ON DELETE CASCADE
);

-- Create indexes for workflow events
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_events_workflow_inst_id ON orchwf_workflow_events(workflow_inst_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_events_step_inst_id ON orchwf_workflow_events(step_inst_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_events_event_type ON orchwf_workflow_events(event_type);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_events_timestamp ON orchwf_workflow_events(timestamp DESC);

-- Create updated_at trigger function for workflow instances
CREATE OR REPLACE FUNCTION update_orchwf_workflow_instances_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for workflow instances
DROP TRIGGER IF EXISTS trigger_update_orchwf_workflow_instances_updated_at ON orchwf_workflow_instances;
CREATE TRIGGER trigger_update_orchwf_workflow_instances_updated_at
    BEFORE UPDATE ON orchwf_workflow_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_orchwf_workflow_instances_updated_at();

-- Create updated_at trigger function for step instances
CREATE OR REPLACE FUNCTION update_orchwf_step_instances_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for step instances
DROP TRIGGER IF EXISTS trigger_update_orchwf_step_instances_updated_at ON orchwf_step_instances;
CREATE TRIGGER trigger_update_orchwf_step_instances_updated_at
    BEFORE UPDATE ON orchwf_step_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_orchwf_step_instances_updated_at();`
}

// getOrchWFTablesRollbackSQL returns the SQL for rolling back OrchWF tables
func getOrchWFTablesRollbackSQL() string {
	return `-- Rollback OrchWF tables

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_orchwf_step_instances_updated_at ON orchwf_step_instances;
DROP TRIGGER IF EXISTS trigger_update_orchwf_workflow_instances_updated_at ON orchwf_workflow_instances;

-- Drop functions
DROP FUNCTION IF EXISTS update_orchwf_step_instances_updated_at();
DROP FUNCTION IF EXISTS update_orchwf_workflow_instances_updated_at();

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS orchwf_workflow_events;
DROP TABLE IF EXISTS orchwf_step_instances;
DROP TABLE IF EXISTS orchwf_workflow_instances;`
}

// LoadMigrationsFromFile loads migrations from a SQL file
func LoadMigrationsFromFile(filePath string) ([]Migration, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration file %s: %v", filePath, err)
	}

	// For now, return a single migration with the file content
	// In a more sophisticated implementation, you might parse multiple migrations from a single file
	return []Migration{
		{
			Version:     filepath.Base(filePath),
			Description: fmt.Sprintf("Migration from %s", filePath),
			Up:          string(content),
			Down:        "", // No rollback by default
		},
	}, nil
}

// QuickSetup is a convenience function that sets up OrchWF tables quickly
func QuickSetup(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	migrator := NewMigrator(db)
	return migrator.Migrate(ctx)
}

// QuickSetupWithContext is a convenience function that sets up OrchWF tables with context
func QuickSetupWithContext(ctx context.Context, db *sql.DB) error {
	migrator := NewMigrator(db)
	return migrator.Migrate(ctx)
}
