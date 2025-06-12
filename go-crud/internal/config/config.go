package config

import (
	"encoding/json"
	"fmt"
)

// Config represents the main configuration for the benchmark
type Config struct {
	Name       string
	Database   string
	Image      string
	Privileged bool
	Endpoint   string
	Blocking   int
	Workers    int
	Clients    int
	Threads    int
	Samples    int
	Random     bool
	KeyType    string
	Value      string
	ShowSample bool
	PID        int
	Scans      []ScanConfig
}

// ScanConfig represents a scan operation configuration
type ScanConfig struct {
	Name       string `json:"name"`
	Samples    int    `json:"samples"`
	Projection string `json:"projection"` // ID, FULL, COUNT
	Start      int    `json:"start,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Expect     int    `json:"expect,omitempty"`
}

// ValidKeyTypes contains all supported key types
var ValidKeyTypes = []string{"integer", "string26", "string90", "string250", "string506", "uuid"}

// ValidDatabases contains all supported database types
var ValidDatabases = []string{
	"dry", "map", "arangodb", "dragonfly", "fjall", "keydb", "lmdb", 
	"mongodb", "mysql", "neo4j", "postgres", "redb", "redis", "rocksdb", 
	"scylladb", "sqlite", "surrealkv", "surrealdb", "surrealdb-memory", 
	"surrealdb-rocksdb", "surrealdb-surrealkv",
}

// ParseScans parses the JSON string into a slice of ScanConfig
func ParseScans(scansJSON string) ([]ScanConfig, error) {
	var scans []ScanConfig
	err := json.Unmarshal([]byte(scansJSON), &scans)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scans JSON: %w", err)
	}
	return scans, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Database == "" {
		return fmt.Errorf("database is required")
	}

	if c.Samples <= 0 {
		return fmt.Errorf("samples must be greater than 0")
	}

	// Validate key type
	validKey := false
	for _, k := range ValidKeyTypes {
		if c.KeyType == k {
			validKey = true
			break
		}
	}
	if !validKey {
		return fmt.Errorf("invalid key type: %s", c.KeyType)
	}

	// Validate database
	validDB := false
	for _, db := range ValidDatabases {
		if c.Database == db {
			validDB = true
			break
		}
	}
	if !validDB {
		return fmt.Errorf("invalid database: %s", c.Database)
	}

	return nil
} 