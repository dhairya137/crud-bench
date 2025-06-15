package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/surrealdb/go-crud-bench/internal/config"
	"github.com/surrealdb/go-crud-bench/internal/docker"
)

// Default MySQL Docker image
const (
	defaultImage = "mysql:8.0"
	
	// Default MySQL port
	defaultPort = "3306"
	
	// Default MySQL credentials
	defaultUser     = "root"
	defaultPassword = "mysql"
	defaultDatabase = "bench"
	
	// Table name
	tableName = "bench_table"
	
	// Container name prefix
	containerNamePrefix = "crud-bench-mysql"
)

// setupLogSilencer disables noisy MySQL driver logs during container startup
func setupLogSilencer() {
	// Create a silent logger that discards all output
	silentLogger := log.New(io.Discard, "", 0)
	// Set the MySQL driver to use our silent logger
	mysqldriver.SetLogger(silentLogger)
}

// Adapter implements the benchmark.Adapter interface for MySQL
type Adapter struct {
	db         *sql.DB
	container  *docker.Container
	endpoint   string
	image      string
	privileged bool
	containerID string
}

// NewAdapter creates a new MySQL adapter
func NewAdapter(endpoint, image string, privileged bool) *Adapter {
	// Silence MySQL driver logs during container startup
	setupLogSilencer()
	
	if image == "" {
		image = defaultImage
	}
	
	return &Adapter{
		endpoint:   endpoint,
		image:      image,
		privileged: privileged,
	}
}

// Initialize sets up the MySQL database
func (a *Adapter) Initialize(ctx context.Context) error {
	var dsn string
	
	// If no endpoint is provided, start a Docker container
	if a.endpoint == "" {
		container, err := a.startContainer(ctx)
		if err != nil {
			return fmt.Errorf("failed to start MySQL container: %w", err)
		}
		
		a.container = container
		a.containerID = container.ID
		dsn = fmt.Sprintf("%s:%s@tcp(127.0.0.1:%s)/", defaultUser, defaultPassword, defaultPort)
	} else {
		// Use provided endpoint
		dsn = a.endpoint
	}
	
	// Connect to MySQL server
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	
	// Set connection pool parameters
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Hour)
	
	// Test connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}
	
	a.db = db
	
	// Create database if it doesn't exist
	if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", defaultDatabase)); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	
	// Use the database
	if _, err := db.ExecContext(ctx, fmt.Sprintf("USE %s", defaultDatabase)); err != nil {
		return fmt.Errorf("failed to use database: %w", err)
	}
	
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
			return fmt.Errorf("failed to close MySQL connection: %w", err)
		}
	}
	
	// Stop and remove container if it was started
	if a.container != nil {
		fmt.Printf("Cleaning up MySQL container %s...\n", a.containerID)
		if err := a.container.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop MySQL container: %w", err)
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
	placeholders := []string{"?"}
	values := []interface{}{key}
	
	// Check for specific fields we know about
	if textVal, ok := value["text"].(string); ok {
		columns = append(columns, "text_val")
		placeholders = append(placeholders, "?")
		values = append(values, textVal)
	}
	
	if intVal, ok := value["integer"].(float64); ok {
		columns = append(columns, "integer_val")
		placeholders = append(placeholders, "?")
		values = append(values, int(intVal))
	}
	
	// Add JSON data column
	columns = append(columns, "data")
	placeholders = append(placeholders, "?")
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
	query := fmt.Sprintf("SELECT data FROM %s WHERE id = ?", tableName)
	
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
	
	// Check for specific fields we know about
	if textVal, ok := value["text"].(string); ok {
		setClauses = append(setClauses, "text_val = ?")
		values = append(values, textVal)
	}
	
	if intVal, ok := value["integer"].(float64); ok {
		setClauses = append(setClauses, "integer_val = ?")
		values = append(values, int(intVal))
	}
	
	// Add JSON data column
	setClauses = append(setClauses, "data = ?")
	values = append(values, string(jsonData))
	
	// Add key for WHERE clause
	values = append(values, key)
	
	// Prepare SQL statement
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ?",
		tableName,
		strings.Join(setClauses, ", "),
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
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	
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
	return "mysql"
}

// createTable creates the benchmark table
func (a *Adapter) createTable(ctx context.Context) error {
	// Create table with id and data columns
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(255) PRIMARY KEY,
			text_val VARCHAR(255),
			integer_val INT,
			data JSON
		)
	`, tableName)
	
	_, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	
	return nil
}

// startContainer starts a MySQL Docker container
func (a *Adapter) startContainer(ctx context.Context) (*docker.Container, error) {
	// Generate unique container name with timestamp
	containerName := fmt.Sprintf("%s-%d", containerNamePrefix, time.Now().Unix())
	
	// Configure container
	ports := map[string]string{
		"3306/tcp": defaultPort,
	}
	
	env := []string{
		fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", defaultPassword),
		fmt.Sprintf("MYSQL_DATABASE=%s", defaultDatabase),
	}
	
	fmt.Printf("Starting MySQL container '%s' with image '%s'...\n", containerName, a.image)
	
	// Create container
	container, err := docker.NewContainer(containerName, a.image, ports, a.privileged, env)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	
	// Start container
	if err := container.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}
	
	fmt.Printf("MySQL container started, waiting for it to be ready...\n")
	
	printedStartup := false
	attemptCount := 0
	// Wait for MySQL to be ready with increased timeout (90 seconds)
	checkFunc := func(ctx context.Context) error {
		if !printedStartup {
			fmt.Println("MySQL container is starting up...")
			printedStartup = true
		} else {
			attemptCount++
			if attemptCount % 5 == 0 {
				// Print status update every 5 attempts
				fmt.Println("Still waiting for MySQL to be ready...")
			}
		}
		
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:%s)/", defaultUser, defaultPassword, defaultPort))
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
		
		// Create database if it doesn't exist
		_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", defaultDatabase))
		if err != nil {
			// Not printing error message, just returning it
			return err
		}
		
		// Select the database
		_, err = db.ExecContext(ctx, fmt.Sprintf("USE %s", defaultDatabase))
		if err != nil {
			// Not printing error message, just returning it
			return err
		}
		
		// Try to create a simple test table to verify MySQL is really ready
		_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS health_check (id INT)")
		if err != nil {
			// Not printing error message, just returning it
			return err
		}
		
		fmt.Printf("MySQL is ready!\n")
		return nil
	}
	
	if err := container.WaitForHealthy(ctx, 90*time.Second, checkFunc); err != nil {
		// Clean up container if health check fails
		_ = container.Stop(ctx)
		return nil, fmt.Errorf("MySQL health check failed: %w", err)
	}
	
	return container, nil
} 