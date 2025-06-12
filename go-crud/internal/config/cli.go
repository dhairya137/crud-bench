package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// FromCommand parses the command line arguments into a Config struct
func FromCommand(cmd *cobra.Command) (*Config, error) {
	// Get all values from flags
	name, _ := cmd.Flags().GetString("name")
	database, _ := cmd.Flags().GetString("database")
	image, _ := cmd.Flags().GetString("image")
	privileged, _ := cmd.Flags().GetBool("privileged")
	endpoint, _ := cmd.Flags().GetString("endpoint")
	blocking, _ := cmd.Flags().GetInt("blocking")
	workers, _ := cmd.Flags().GetInt("workers")
	clients, _ := cmd.Flags().GetInt("clients")
	threads, _ := cmd.Flags().GetInt("threads")
	samples, _ := cmd.Flags().GetInt("samples")
	random, _ := cmd.Flags().GetBool("random")
	keyType, _ := cmd.Flags().GetString("key")
	value, _ := cmd.Flags().GetString("value")
	showSample, _ := cmd.Flags().GetBool("show-sample")
	pid, _ := cmd.Flags().GetInt("pid")
	scansJSON, _ := cmd.Flags().GetString("scans")

	// Parse scans from JSON
	scans, err := ParseScans(scansJSON)
	if err != nil {
		return nil, fmt.Errorf("invalid scans configuration: %w", err)
	}

	// Create config
	config := &Config{
		Name:       name,
		Database:   database,
		Image:      image,
		Privileged: privileged,
		Endpoint:   endpoint,
		Blocking:   blocking,
		Workers:    workers,
		Clients:    clients,
		Threads:    threads,
		Samples:    samples,
		Random:     random,
		KeyType:    keyType,
		Value:      value,
		ShowSample: showSample,
		PID:        pid,
		Scans:      scans,
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
} 