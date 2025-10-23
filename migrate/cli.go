package migrate

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// CLI represents the command-line interface for migrations
type CLI struct {
	dbURL string
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{}
}

// Run runs the CLI with the given arguments
func (c *CLI) Run(args []string) error {
	// Parse command line flags
	var (
		dbURL = flag.String("db", "", "Database connection URL (e.g., postgres://user:pass@localhost/dbname?sslmode=disable)")
		help  = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	if *help {
		c.showHelp()
		return nil
	}

	if *dbURL == "" {
		// Try to get from environment variable
		*dbURL = os.Getenv("ORCHWF_DB_URL")
		if *dbURL == "" {
			return fmt.Errorf("database URL is required. Use -db flag or set ORCHWF_DB_URL environment variable")
		}
	}

	c.dbURL = *dbURL

	// Get the command
	if len(args) < 1 {
		return fmt.Errorf("command is required. Use 'orchwf-migrate help' for usage")
	}

	command := args[0]

	// Connect to database
	db, err := sql.Open("postgres", c.dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Create migrator
	migrator := NewMigrator(db)

	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch command {
	case "up", "migrate":
		return migrator.Migrate(ctx)
	case "down", "rollback":
		return migrator.Rollback(ctx)
	case "status":
		return migrator.Status(ctx)
	case "help":
		c.showHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s. Use 'orchwf-migrate help' for usage", command)
	}
}

// showHelp displays the help message
func (c *CLI) showHelp() {
	fmt.Println("OrchWF Migration Tool")
	fmt.Println("====================")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  orchwf-migrate [command] [flags]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  up, migrate    Apply all pending migrations")
	fmt.Println("  down, rollback Rollback the last migration")
	fmt.Println("  status         Show migration status")
	fmt.Println("  help           Show this help message")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -db string     Database connection URL")
	fmt.Println("  -help          Show help")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  ORCHWF_DB_URL  Database connection URL (alternative to -db flag)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Migrate using database URL")
	fmt.Println("  orchwf-migrate up -db 'postgres://user:pass@localhost/dbname?sslmode=disable'")
	fmt.Println("")
	fmt.Println("  # Migrate using environment variable")
	fmt.Println("  export ORCHWF_DB_URL='postgres://user:pass@localhost/dbname?sslmode=disable'")
	fmt.Println("  orchwf-migrate up")
	fmt.Println("")
	fmt.Println("  # Check migration status")
	fmt.Println("  orchwf-migrate status -db 'postgres://user:pass@localhost/dbname?sslmode=disable'")
	fmt.Println("")
	fmt.Println("  # Rollback last migration")
	fmt.Println("  orchwf-migrate down -db 'postgres://user:pass@localhost/dbname?sslmode=disable'")
}

// RunCLI is a convenience function to run the CLI
func RunCLI() {
	cli := NewCLI()
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
