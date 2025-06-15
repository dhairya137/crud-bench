package databases

import (
	"fmt"

	"github.com/surrealdb/go-crud-bench/internal/benchmark"
	"github.com/surrealdb/go-crud-bench/internal/databases/mysql"
	"github.com/surrealdb/go-crud-bench/internal/databases/postgres"
)

// NewAdapter creates a new database adapter based on the database type
func NewAdapter(dbType, endpoint, image string, privileged bool) (benchmark.Adapter, error) {
	switch dbType {
	case "mysql":
		return mysql.NewAdapter(endpoint, image, privileged), nil
	case "postgres":
		return postgres.NewAdapter(endpoint, image, privileged), nil
	// Add more database types here as they are implemented
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
} 