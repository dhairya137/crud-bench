package benchmark

import (
	"context"
	"time"

	"github.com/surrealdb/go-crud-bench/internal/config"
)

// Operation represents a benchmark operation type
type Operation string

const (
	// OperationCreate represents a create operation
	OperationCreate Operation = "CREATE"
	// OperationRead represents a read operation
	OperationRead Operation = "READ"
	// OperationUpdate represents an update operation
	OperationUpdate Operation = "UPDATE"
	// OperationDelete represents a delete operation
	OperationDelete Operation = "DELETE"
	// OperationScan represents a scan operation
	OperationScan Operation = "SCAN"
)

// Result represents the result of a benchmark operation
type Result struct {
	Operation Operation
	Name      string
	Duration  time.Duration
	Error     error
	Count     int
}

// Adapter defines the interface that all database adapters must implement
type Adapter interface {
	// Initialize sets up the database connection and creates necessary tables/collections
	Initialize(ctx context.Context) error
	
	// Cleanup performs any necessary cleanup operations
	Cleanup(ctx context.Context) error
	
	// Create inserts a new record with the given key and value
	Create(ctx context.Context, key string, value map[string]interface{}) error
	
	// Read retrieves a record with the given key
	Read(ctx context.Context, key string) (map[string]interface{}, error)
	
	// Update updates a record with the given key and value
	Update(ctx context.Context, key string, value map[string]interface{}) error
	
	// Delete removes a record with the given key
	Delete(ctx context.Context, key string) error
	
	// Scan performs a scan operation based on the given configuration
	Scan(ctx context.Context, scanConfig config.ScanConfig) (int, error)
	
	// Name returns the name of the database adapter
	Name() string
}

// Runner is responsible for running benchmark operations
type Runner struct {
	Adapter  Adapter
	Config   *config.Config
	Results  []Result
}

// NewRunner creates a new benchmark runner
func NewRunner(adapter Adapter, cfg *config.Config) *Runner {
	return &Runner{
		Adapter: adapter,
		Config:  cfg,
		Results: []Result{},
	}
}

// Run executes the benchmark
func (r *Runner) Run(ctx context.Context) ([]Result, error) {
	// Initialize the database
	if err := r.Adapter.Initialize(ctx); err != nil {
		return nil, err
	}
	
	// Ensure cleanup happens
	defer func() {
		_ = r.Adapter.Cleanup(ctx)
	}()
	
	// Run the benchmark operations
	if err := r.runCreate(ctx); err != nil {
		return r.Results, err
	}
	
	if err := r.runRead(ctx); err != nil {
		return r.Results, err
	}
	
	if err := r.runUpdate(ctx); err != nil {
		return r.Results, err
	}
	
	if err := r.runScans(ctx); err != nil {
		return r.Results, err
	}
	
	if err := r.runDelete(ctx); err != nil {
		return r.Results, err
	}
	
	return r.Results, nil
} 