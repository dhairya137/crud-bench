# Adding a New Database Adapter to Go-CRUD-Bench

This document provides detailed instructions on how to add support for a new database to the Go-CRUD-Bench benchmarking tool. The tool is designed to benchmark CRUD (Create, Read, Update, Delete) operations and scan queries against various databases.

## Table of Contents

1. [Overview](#overview)
2. [Step-by-Step Guide](#step-by-step-guide)
3. [Adapter Interface Requirements](#adapter-interface-requirements)
4. [Docker Integration](#docker-integration)
5. [Error Handling and Logging](#error-handling-and-logging)
6. [Testing Your Adapter](#testing-your-adapter)
7. [Example Implementation: MySQL](#example-implementation-mysql)
8. [Using Docker Helper Utilities](#using-docker-helper-utilities)

## Overview

The Go-CRUD-Bench tool uses a plugin-like architecture where each database is supported through an adapter that implements the `benchmark.Adapter` interface. This allows the core benchmarking logic to remain database-agnostic while supporting many different databases.

## Step-by-Step Guide

### 1. Create a New Package for Your Database

Create a new directory under `internal/databases/` with the name of your database:

```bash
mkdir -p internal/databases/yourdatabase
```

### 2. Create the Adapter Implementation File

Create a new file named after your database inside this directory (e.g., `yourdatabase.go`).

### 3. Implement the Adapter Interface

Your adapter must implement the `benchmark.Adapter` interface, which consists of the following methods:

```go
type Adapter interface {
    Initialize(ctx context.Context) error
    Cleanup(ctx context.Context) error
    Create(ctx context.Context, key string, value map[string]interface{}) error
    Read(ctx context.Context, key string) (map[string]interface{}, error)
    Update(ctx context.Context, key string, value map[string]interface{}) error
    Delete(ctx context.Context, key string) error
    Scan(ctx context.Context, scanConfig config.ScanConfig) (int, error)
    Name() string
}
```

### 4. Register Your Adapter in the Factory

Update `internal/databases/factory.go` to include your new adapter:

```go
func NewAdapter(dbType, endpoint, image string, privileged bool) (benchmark.Adapter, error) {
    switch dbType {
    case "mysql":
        return mysql.NewAdapter(endpoint, image, privileged), nil
    case "postgres":
        return postgres.NewAdapter(endpoint, image, privileged), nil
    case "yourdatabase":
        return yourdatabase.NewAdapter(endpoint, image, privileged), nil
    // Add more database types here
    default:
        return nil, fmt.Errorf("unsupported database type: %s", dbType)
    }
}
```

### 5. Add Your Database to Valid Databases List

Update `internal/config/config.go` to include your database in the `ValidDatabases` slice:

```go
var ValidDatabases = []string{
    "dry", "map", "arangodb", "dragonfly", "fjall", "keydb", "lmdb",
    "mongodb", "mysql", "neo4j", "postgres", "redb", "redis", "rocksdb",
    "scylladb", "sqlite", "surrealkv", "surrealdb", "yourdatabase",
    // Add your database here
}
```

## Adapter Interface Requirements

Let's look at each method your adapter must implement:

### Initialize

```go
Initialize(ctx context.Context) error
```

This method should:

- Set up the database connection
- Start a Docker container if needed (when no endpoint is provided)
- Create any necessary tables/collections/schemas
- Ensure the database is ready for benchmarking

### Cleanup

```go
Cleanup(ctx context.Context) error
```

This method should:

- Close the database connection
- Stop and remove any Docker containers that were started
- Clean up any other resources

### Create

```go
Create(ctx context.Context, key string, value map[string]interface{}) error
```

This method should:

- Insert a new record with the given key and value
- Handle any database-specific serialization (e.g., converting to JSON)

### Read

```go
Read(ctx context.Context, key string) (map[string]interface{}, error)
```

This method should:

- Retrieve a record with the given key
- Deserialize the record to a map[string]interface{}
- Return an error if the record doesn't exist

### Update

```go
Update(ctx context.Context, key string, value map[string]interface{}) error
```

This method should:

- Update an existing record with the given key and value
- Handle any database-specific serialization

### Delete

```go
Delete(ctx context.Context, key string) error
```

This method should:

- Remove a record with the given key

### Scan

```go
Scan(ctx context.Context, scanConfig config.ScanConfig) (int, error)
```

This method should:

- Perform a scan operation based on the provided configuration
- Support different projection types: "ID", "FULL", "COUNT"
- Support LIMIT and OFFSET if specified
- Return the count of records found

### Name

```go
Name() string
```

This method should:

- Return the name of the database (used for reporting and file naming)

## Docker Integration

For databases that should be run in Docker containers during benchmarks, follow these steps:

### 1. Define Constants

```go
const (
    // Default Docker image
    defaultImage = "yourdatabase:latest"

    // Default port
    defaultPort = "1234"

    // Default credentials
    defaultUser = "user"
    defaultPassword = "password"
    defaultDatabase = "bench"

    // Table/collection name
    tableName = "bench_table"

    // Container name prefix
    containerNamePrefix = "crud-bench-yourdatabase"
)
```

### 2. Create a startContainer Method

Instead of implementing your own container start logic, use the helper utilities from `internal/dbutils` to ensure proper image pulling and container creation:

```go
import (
    "github.com/surrealdb/go-crud-bench/internal/dbutils"
    "github.com/surrealdb/go-crud-bench/internal/docker"
)

// startContainer starts a Docker container for your database
func (a *Adapter) startContainer(ctx context.Context) (*docker.Container, error) {
    // Generate unique container name with timestamp
    containerName := fmt.Sprintf("%s-%d", containerNamePrefix, time.Now().Unix())

    // Configure container
    ports := map[string]string{
        "1234/tcp": defaultPort,
    }

    env := []string{
        fmt.Sprintf("DB_USER=%s", defaultUser),
        fmt.Sprintf("DB_PASSWORD=%s", defaultPassword),
        fmt.Sprintf("DB_DATABASE=%s", defaultDatabase),
    }

    fmt.Printf("Starting YourDatabase container '%s' with image '%s'...\n", containerName, a.image)

    // Create and start container with the common utility
    container, err := dbutils.CreateContainerWithRetry(ctx, containerName, a.image, ports, a.privileged, env)
    if err != nil {
        return nil, fmt.Errorf("failed to start YourDatabase container: %w", err)
    }

    fmt.Printf("YourDatabase container started, waiting for it to be ready...\n")

    // Implement health check logic for your database
    // ...

    return container, nil
}
```

## Error Handling and Logging

- Use `fmt.Errorf` with error wrapping (`%w`) for proper error context.
- Silence excessive logs during container startup when appropriate.
- Use user-friendly messages for progress updates.

Example for silencing driver logs:

```go
// setupLogSilencer disables noisy driver logs during container startup
func setupLogSilencer() {
    // Create a silent logger that discards all output
    silentLogger := log.New(io.Discard, "", 0)
    // Set the database driver to use our silent logger
    driverpackage.SetLogger(silentLogger)
}
```

## Testing Your Adapter

To test your adapter, run the benchmarking tool with your database:

```bash
./bin/crud-bench -d yourdatabase -s 10 -c 1 -t 1
```

## Example Implementation: MySQL

Below is a simplified example of the MySQL adapter implementation to use as a reference:

### Structure and Constants

```go
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

    mysqldriver "github.com/go-sql-driver/mysql"
    _ "github.com/go-sql-driver/mysql"
    "github.com/surrealdb/go-crud-bench/internal/config"
    "github.com/surrealdb/go-crud-bench/internal/dbutils"
    "github.com/surrealdb/go-crud-bench/internal/docker"
)

const (
    defaultImage = "mysql:8.0"
    defaultPort = "3306"
    defaultUser = "root"
    defaultPassword = "mysql"
    defaultDatabase = "bench"
    tableName = "bench_table"
    containerNamePrefix = "crud-bench-mysql"
)

// setupLogSilencer disables noisy MySQL driver logs during container startup
func setupLogSilencer() {
    silentLogger := log.New(io.Discard, "", 0)
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
```

### Constructor

```go
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
```

### Using Docker Helper Utilities

The adapter uses the common Docker helper utilities to start containers and ensure images are available:

```go
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

    // Create and start container with the common utility
    container, err := dbutils.CreateContainerWithRetry(ctx, containerName, a.image, ports, a.privileged, env)
    if err != nil {
        return nil, fmt.Errorf("failed to start MySQL container: %w", err)
    }

    // Continue with health check logic
    // ...
}
```

## Using Docker Helper Utilities

The application provides helper utilities in the `internal/dbutils` package to simplify Docker container management and image pulling. These utilities ensure that your database adapter will work correctly even when the required Docker images are not already present on the user's system.

### Import the dbutils Package

```go
import "github.com/surrealdb/go-crud-bench/internal/dbutils"
```

### Available Utilities

#### EnsureDockerImage

This function checks if the specified Docker image exists locally and pulls it if necessary:

```go
pulled, err := dbutils.EnsureDockerImage("yourdatabase:latest")
if err != nil {
    return fmt.Errorf("failed to ensure Docker image: %w", err)
}
if pulled {
    fmt.Println("Image was pulled successfully")
}
```

#### CreateContainerWithRetry

This function handles the entire process of creating and starting a Docker container, including automatic image pulling if needed:

```go
container, err := dbutils.CreateContainerWithRetry(
    ctx,
    containerName,
    imageName,
    ports,
    privileged,
    env,
)
if err != nil {
    return nil, fmt.Errorf("failed to create and start container: %w", err)
}
```

### Benefits of Using These Utilities

1. **Better User Experience**: Users won't encounter errors just because they don't have the required Docker images.
2. **Automatic Image Pulling**: The system handles image pulls transparently with proper console feedback.
3. **Error Resilience**: Even if the image pull fails initially, the system will attempt a manual pull as a fallback.
4. **Consistent Implementation**: All database adapters will have the same reliable container creation behavior.

By using these utilities in your adapter implementation, you ensure that your database adapter will be robust against missing Docker images and provide a consistent user experience.

---

By following this guide, you should be able to successfully implement a new database adapter for the Go-CRUD-Bench tool. The adapter will be automatically used when you specify your database with the `-d` flag when running the benchmarking tool.

If you encounter any issues or have questions, please refer to the existing implementations or file an issue on the project repository.
