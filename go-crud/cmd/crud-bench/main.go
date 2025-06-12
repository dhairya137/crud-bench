package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/surrealdb/go-crud-bench/internal/benchmark"
	"github.com/surrealdb/go-crud-bench/internal/config"
	"github.com/surrealdb/go-crud-bench/internal/databases"
	"github.com/surrealdb/go-crud-bench/internal/generators"
)

var (
	// CLI flags
	name       string
	database   string
	image      string
	privileged bool
	endpoint   string
	blocking   int
	workers    int
	clients    int
	threads    int
	samples    int
	random     bool
	keyType    string
	value      string
	showSample bool
	pid        int
	scans      string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "crud-bench",
		Short: "CRUD benchmarking tool for various databases",
		Long: `The crud-bench benchmarking tool is an open-source benchmarking tool for testing 
and comparing the performance of a number of different workloads on embedded, 
networked, and remote databases. It can be used to compare both SQL and NoSQL platforms.`,
		Run: runBenchmark,
	}

	// Define flags
	rootCmd.Flags().StringVarP(&name, "name", "n", "", "An optional name for the test, used as a suffix for the JSON result file name")
	rootCmd.Flags().StringVarP(&database, "database", "d", "", "The database to benchmark")
	rootCmd.MarkFlagRequired("database")
	rootCmd.Flags().StringVarP(&image, "image", "i", "", "Specify a custom Docker image")
	rootCmd.Flags().BoolVarP(&privileged, "privileged", "p", false, "Whether to run Docker in privileged mode")
	rootCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "Specify a custom endpoint to connect to")
	rootCmd.Flags().IntVarP(&blocking, "blocking", "b", 12, "Maximum number of blocking threads")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 12, "Number of async runtime workers")
	rootCmd.Flags().IntVarP(&clients, "clients", "c", 1, "Number of concurrent clients")
	rootCmd.Flags().IntVarP(&threads, "threads", "t", 1, "Number of concurrent threads per client")
	rootCmd.Flags().IntVarP(&samples, "samples", "s", 0, "Number of samples to be created, read, updated, and deleted")
	rootCmd.MarkFlagRequired("samples")
	rootCmd.Flags().BoolVarP(&random, "random", "r", false, "Generate the keys in a pseudo-randomized order")
	rootCmd.Flags().StringVarP(&keyType, "key", "k", "integer", "The type of the key")
	rootCmd.Flags().StringVarP(&value, "value", "v", "{\n\t\"text\": \"string:50\",\n\t\"integer\": \"int\"\n}", "Size of the text value")
	rootCmd.Flags().BoolVar(&showSample, "show-sample", false, "Print-out an example of a generated value")
	rootCmd.Flags().IntVar(&pid, "pid", 0, "Collect system information for a given pid")
	rootCmd.Flags().StringVarP(&scans, "scans", "a", "[\n\t{ \"name\": \"count_all\", \"samples\": 100, \"projection\": \"COUNT\" },\n\t{ \"name\": \"limit_id\", \"samples\": 100, \"projection\": \"ID\", \"limit\": 100, \"expect\": 100 }\n]", "An array of scan specifications")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runBenchmark(cmd *cobra.Command, args []string) {
	// Parse configuration
	cfg, err := config.FromCommand(cmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Show sample if requested
	if cfg.ShowSample {
		sampleJSON, err := generators.GenerateSample(cfg.Value)
		if err != nil {
			fmt.Printf("Error generating sample: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(sampleJSON)
		return
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		fmt.Println("\nReceived interrupt signal. Shutting down...")
		cancel()
	}()

	// Create database adapter
	adapter, err := databases.NewAdapter(cfg.Database, cfg.Endpoint, cfg.Image, cfg.Privileged)
	if err != nil {
		fmt.Printf("Error creating database adapter: %v\n", err)
		os.Exit(1)
	}

	// Create benchmark runner
	runner := benchmark.NewRunner(adapter, cfg)

	// Run benchmark
	fmt.Printf("Starting benchmark for %s with %d samples...\n", adapter.Name(), cfg.Samples)
	startTime := time.Now()
	
	results, err := runner.Run(ctx)
	if err != nil {
		fmt.Printf("Error running benchmark: %v\n", err)
		os.Exit(1)
	}
	
	duration := time.Since(startTime)

	// Print results
	fmt.Printf("\nBenchmark completed in %v\n\n", duration)
	
	// Print results table
	fmt.Printf("%-15s %-15s %-15s\n", "OPERATION", "DURATION", "COUNT")
	fmt.Printf("%-15s %-15s %-15s\n", "---------", "--------", "-----")
	
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("%-15s %-15s %-15s\n", result.Operation, result.Duration, fmt.Sprintf("ERROR: %v", result.Error))
		} else {
			fmt.Printf("%-15s %-15s %-15d\n", result.Operation, result.Duration, result.Count)
		}
	}
	
	// Save results to JSON file
	outputFilename := fmt.Sprintf("results-%s-%s.json", adapter.Name(), time.Now().Format("20060102-150405"))
	if cfg.Name != "" {
		outputFilename = fmt.Sprintf("results-%s-%s-%s.json", adapter.Name(), cfg.Name, time.Now().Format("20060102-150405"))
	}
	
	outputData := map[string]interface{}{
		"database":   adapter.Name(),
		"samples":    cfg.Samples,
		"clients":    cfg.Clients,
		"threads":    cfg.Threads,
		"duration":   duration.String(),
		"operations": results,
	}
	
	jsonData, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling results: %v\n", err)
	} else {
		if err := os.WriteFile(outputFilename, jsonData, 0644); err != nil {
			fmt.Printf("Error writing results file: %v\n", err)
		} else {
			fmt.Printf("\nResults saved to %s\n", outputFilename)
		}
	}
} 