package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/surrealdb/go-crud-bench/internal/config"
	"github.com/surrealdb/go-crud-bench/internal/docker"
)

const (
	// Default PostgreSQL Docker image
	defaultImage = "postgres:15"

	// Default PostgreSQL port
	defaultPort = "5432"

	// Default PostgreSQL credentials
	defaultUser     = "postgres"
	defaultPassword = "postgres"
	defaultDatabase = "bench"

	// Table name
	tableName = "bench_table"

	// Container name prefix
	containerNamePrefix = "crud-bench-postgres"
)

// Adapter implements the benchmark.Adapter interface for PostgreSQL
type Adapter struct {
	db          *sql.DB
	container   *docker.Container
	endpoint    string
	image       string
	privileged  bool
	containerID string
}

// NewAdapter creates a new PostgreSQL adapter
func NewAdapter(endpoint, image string, privileged bool) *Adapter {
	if image == "" {
		image = defaultImage
	}

	return &Adapter{
		endpoint:   endpoint,
		image:      image,
		privileged: privileged,
	}
}

// Initialize sets up the PostgreSQL database
func (a *Adapter) Initialize(ctx context.Context) error {
	var dsn string

	// If no endpoint is provided, start a Docker container
	if a.endpoint == "" {
		container, err := a.startContainer(ctx)
		if err != nil {
			return fmt.Errorf("failed to start PostgreSQL container: %w", err)
		}

		a.container = container
		a.containerID = container.ID
		dsn = fmt.Sprintf("host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable",
			defaultPort, defaultUser, defaultPassword, defaultDatabase)
	} else {
		// Use provided endpoint
		dsn = a.endpoint
	}

	// Connect to PostgreSQL server
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	a.db = db

	// Create table
	if err := a.createTable(ctx); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// Cleanup performs cleanup operations
func (a *Adapter) Cleanup(ctx context.Context) error {
	// Close database connection
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
		}
	}

	// Stop and remove container if it was started
	if a.container != nil {
		fmt.Printf("Cleaning up PostgreSQL container %s...\n", a.containerID)
		if err := a.container.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop PostgreSQL container: %w", err)
		}
	}

	return nil
}

// Create inserts a new record
func (a *Adapter) Create(ctx context.Context, key string, value map[string]interface{}) error {
	// Convert value to JSON
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value to JSON: %w", err)
	}

	// Extract first-level fields for columns
	columns := []string{"id"}
	placeholders := []string{"$1"}
	values := []interface{}{key}
	paramCount := 1

	// Check for specific fields we know about
	if textVal, ok := value["text"].(string); ok {
		paramCount++
		columns = append(columns, "text_val")
		placeholders = append(placeholders, fmt.Sprintf("$%d", paramCount))
		values = append(values, textVal)
	}

	if intVal, ok := value["integer"].(float64); ok {
		paramCount++
		columns = append(columns, "integer_val")
		placeholders = append(placeholders, fmt.Sprintf("$%d", paramCount))
		values = append(values, int(intVal))
	}

	// Add JSON data column
	paramCount++
	columns = append(columns, "data")
	placeholders = append(placeholders, fmt.Sprintf("$%d", paramCount))
	values = append(values, string(jsonData))

	// Prepare SQL statement
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// Execute query
	_, err = a.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}

	return nil
}

// Read retrieves a record
func (a *Adapter) Read(ctx context.Context, key string) (map[string]interface{}, error) {
	// Prepare SQL statement
	query := fmt.Sprintf("SELECT data FROM %s WHERE id = $1", tableName)

	// Execute query
	var jsonData string
	err := a.db.QueryRowContext(ctx, query, key).Scan(&jsonData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("record not found: %s", key)
		}
		return nil, fmt.Errorf("failed to read record: %w", err)
	}

	// Parse JSON data
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}

	return result, nil
}

// Update updates a record
func (a *Adapter) Update(ctx context.Context, key string, value map[string]interface{}) error {
	// Convert value to JSON
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value to JSON: %w", err)
	}

	// Extract first-level fields for columns
	setClauses := []string{}
	values := []interface{}{}
	paramCount := 0

	// Check for specific fields we know about
	if textVal, ok := value["text"].(string); ok {
		paramCount++
		setClauses = append(setClauses, fmt.Sprintf("text_val = $%d", paramCount))
		values = append(values, textVal)
	}

	if intVal, ok := value["integer"].(float64); ok {
		paramCount++
		setClauses = append(setClauses, fmt.Sprintf("integer_val = $%d", paramCount))
		values = append(values, int(intVal))
	}

	// Add JSON data column
	paramCount++
	setClauses = append(setClauses, fmt.Sprintf("data = $%d", paramCount))
	values = append(values, string(jsonData))

	// Add key for WHERE clause
	paramCount++
	values = append(values, key)

	// Prepare SQL statement
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d",
		tableName,
		strings.Join(setClauses, ", "),
		paramCount,
	)

	// Execute query
	_, err = a.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	return nil
}

// Delete removes a record
func (a *Adapter) Delete(ctx context.Context, key string) error {
	// Prepare SQL statement
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)

	// Execute query
	_, err := a.db.ExecContext(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}

// Scan performs a scan operation
func (a *Adapter) Scan(ctx context.Context, scanConfig config.ScanConfig) (int, error) {
	var query string
	var args []interface{}
	var count int

	// Build query based on projection type
	switch scanConfig.Projection {
	case "ID":
		query = fmt.Sprintf("SELECT id FROM %s", tableName)
	case "FULL":
		query = fmt.Sprintf("SELECT * FROM %s", tableName)
	case "COUNT":
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	default:
		return 0, fmt.Errorf("unsupported projection type: %s", scanConfig.Projection)
	}

	// Add LIMIT and OFFSET if specified
	if scanConfig.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", scanConfig.Limit)

		if scanConfig.Start > 0 {
			query += fmt.Sprintf(" OFFSET %d", scanConfig.Start)
		}
	}

	// Execute query
	if scanConfig.Projection == "COUNT" {
		err := a.db.QueryRowContext(ctx, query, args...).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("failed to execute count scan: %w", err)
		}
		return count, nil
	}

	// For ID and FULL projections, execute query and count rows
	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute scan: %w", err)
	}
	defer rows.Close()

	// Count rows
	for rows.Next() {
		count++
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("error while scanning rows: %w", err)
	}

	return count, nil
}

// Name returns the adapter name
func (a *Adapter) Name() string {
	return "postgres"
}

// createTable creates the benchmark table
func (a *Adapter) createTable(ctx context.Context) error {
	// Create table with id and data columns
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(255) PRIMARY KEY,
			text_val VARCHAR(255),
			integer_val INTEGER,
			data JSONB
		)
	`, tableName)

	_, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// startContainer starts a PostgreSQL Docker container
func (a *Adapter) startContainer(ctx context.Context) (*docker.Container, error) {
	// Generate unique container name with timestamp
	containerName := fmt.Sprintf("%s-%d", containerNamePrefix, time.Now().Unix())

	// Configure container
	ports := map[string]string{
		"5432/tcp": defaultPort,
	}

	env := []string{
		fmt.Sprintf("POSTGRES_USER=%s", defaultUser),
		fmt.Sprintf("POSTGRES_PASSWORD=%s", defaultPassword),
		fmt.Sprintf("POSTGRES_DB=%s", defaultDatabase),
	}

	fmt.Printf("Starting PostgreSQL container '%s' with image '%s'...\n", containerName, a.image)

	// Create container
	container, err := docker.NewContainer(containerName, a.image, ports, a.privileged, env)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := container.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("PostgreSQL container started, waiting for it to be ready...\n")

	printedStartup := false
	attemptCount := 0
	// Wait for PostgreSQL to be ready with increased timeout (90 seconds)
	checkFunc := func(ctx context.Context) error {
		if !printedStartup {
			fmt.Println("PostgreSQL container is starting up...")
			printedStartup = true
		} else {
			attemptCount++
			if attemptCount % 5 == 0 {
				// Print status update every 5 attempts
				fmt.Println("Still waiting for PostgreSQL to be ready...")
			}
		}

		db, err := sql.Open("postgres", fmt.Sprintf("host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable",
			defaultPort, defaultUser, defaultPassword, defaultDatabase))
		if err != nil {
			return err
		}
		defer db.Close()

		// Set a short timeout for the connection attempt
		db.SetConnMaxLifetime(5 * time.Second)

		// Try to ping the database
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		err = db.PingContext(ctx)
		if err != nil {
			// Not printing error message, just returning it
			return err
		}

		// Try to create a simple test table to verify PostgreSQL is really ready
		_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS health_check (id INT)")
		if err != nil {
			// Not printing error message, just returning it
			return err
		}

		fmt.Printf("PostgreSQL is ready!\n")
		return nil
	}

	if err := container.WaitForHealthy(ctx, 90*time.Second, checkFunc); err != nil {
		// Clean up container if health check fails
		_ = container.Stop(ctx)
		return nil, fmt.Errorf("PostgreSQL health check failed: %w", err)
	}

	return container, nil
} 