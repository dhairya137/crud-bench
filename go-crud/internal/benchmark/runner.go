package benchmark

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/surrealdb/go-crud-bench/internal/generators"
)

// runCreate executes the create benchmark
func (r *Runner) runCreate(ctx context.Context) error {
	fmt.Printf("Running CREATE benchmark with %d samples...\n", r.Config.Samples)
	
	// Generate keys
	keys, err := generators.GenerateKeys(r.Config.KeyType, r.Config.Samples, r.Config.Random)
	if err != nil {
		return fmt.Errorf("failed to generate keys: %w", err)
	}
	
	// Generate sample value template
	valueTemplate, err := generators.ProcessTemplate(r.Config.Value)
	if err != nil {
		return fmt.Errorf("failed to process value template: %w", err)
	}
	
	// Start timer
	startTime := time.Now()
	
	// Create records
	var wg sync.WaitGroup
	errCh := make(chan error, r.Config.Clients*r.Config.Threads)
	
	// Process in batches based on client and thread count
	batchSize := r.Config.Samples / (r.Config.Clients * r.Config.Threads)
	if batchSize == 0 {
		batchSize = 1
	}
	
	for c := 0; c < r.Config.Clients; c++ {
		for t := 0; t < r.Config.Threads; t++ {
			wg.Add(1)
			
			go func(clientID, threadID int) {
				defer wg.Done()
				
				// Calculate start and end indices for this worker
				start := (clientID*r.Config.Threads + threadID) * batchSize
				end := start + batchSize
				
				if end > r.Config.Samples {
					end = r.Config.Samples
				}
				
				if start >= r.Config.Samples {
					return
				}
				
				// Process assigned keys
				for i := start; i < end; i++ {
					select {
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					default:
						// Generate a unique value for this record
						value := make(map[string]interface{})
						for k, v := range valueTemplate {
							value[k] = generators.ProcessValue(v)
						}
						
						if err := r.Adapter.Create(ctx, keys[i], value); err != nil {
							errCh <- fmt.Errorf("failed to create record %d: %w", i, err)
							return
						}
					}
				}
			}(c, t)
		}
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
	
	// Check for errors
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	
	// Record result
	duration := time.Since(startTime)
	r.Results = append(r.Results, Result{
		Operation: OperationCreate,
		Name:      "create_all",
		Duration:  duration,
		Count:     r.Config.Samples,
	})
	
	fmt.Printf("CREATE completed in %v\n", duration)
	return nil
}

// runRead executes the read benchmark
func (r *Runner) runRead(ctx context.Context) error {
	fmt.Printf("Running READ benchmark with %d samples...\n", r.Config.Samples)
	
	// Generate keys (same order as create)
	keys, err := generators.GenerateKeys(r.Config.KeyType, r.Config.Samples, r.Config.Random)
	if err != nil {
		return fmt.Errorf("failed to generate keys: %w", err)
	}
	
	// Start timer
	startTime := time.Now()
	
	// Read records
	var wg sync.WaitGroup
	errCh := make(chan error, r.Config.Clients*r.Config.Threads)
	
	// Process in batches based on client and thread count
	batchSize := r.Config.Samples / (r.Config.Clients * r.Config.Threads)
	if batchSize == 0 {
		batchSize = 1
	}
	
	for c := 0; c < r.Config.Clients; c++ {
		for t := 0; t < r.Config.Threads; t++ {
			wg.Add(1)
			
			go func(clientID, threadID int) {
				defer wg.Done()
				
				// Calculate start and end indices for this worker
				start := (clientID*r.Config.Threads + threadID) * batchSize
				end := start + batchSize
				
				if end > r.Config.Samples {
					end = r.Config.Samples
				}
				
				if start >= r.Config.Samples {
					return
				}
				
				// Process assigned keys
				for i := start; i < end; i++ {
					select {
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					default:
						if _, err := r.Adapter.Read(ctx, keys[i]); err != nil {
							errCh <- fmt.Errorf("failed to read record %d: %w", i, err)
							return
						}
					}
				}
			}(c, t)
		}
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
	
	// Check for errors
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	
	// Record result
	duration := time.Since(startTime)
	r.Results = append(r.Results, Result{
		Operation: OperationRead,
		Name:      "read_all",
		Duration:  duration,
		Count:     r.Config.Samples,
	})
	
	fmt.Printf("READ completed in %v\n", duration)
	return nil
}

// runUpdate executes the update benchmark
func (r *Runner) runUpdate(ctx context.Context) error {
	fmt.Printf("Running UPDATE benchmark with %d samples...\n", r.Config.Samples)
	
	// Generate keys (same order as create)
	keys, err := generators.GenerateKeys(r.Config.KeyType, r.Config.Samples, r.Config.Random)
	if err != nil {
		return fmt.Errorf("failed to generate keys: %w", err)
	}
	
	// Generate sample value template
	valueTemplate, err := generators.ProcessTemplate(r.Config.Value)
	if err != nil {
		return fmt.Errorf("failed to process value template: %w", err)
	}
	
	// Start timer
	startTime := time.Now()
	
	// Update records
	var wg sync.WaitGroup
	errCh := make(chan error, r.Config.Clients*r.Config.Threads)
	
	// Process in batches based on client and thread count
	batchSize := r.Config.Samples / (r.Config.Clients * r.Config.Threads)
	if batchSize == 0 {
		batchSize = 1
	}
	
	for c := 0; c < r.Config.Clients; c++ {
		for t := 0; t < r.Config.Threads; t++ {
			wg.Add(1)
			
			go func(clientID, threadID int) {
				defer wg.Done()
				
				// Calculate start and end indices for this worker
				start := (clientID*r.Config.Threads + threadID) * batchSize
				end := start + batchSize
				
				if end > r.Config.Samples {
					end = r.Config.Samples
				}
				
				if start >= r.Config.Samples {
					return
				}
				
				// Process assigned keys
				for i := start; i < end; i++ {
					select {
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					default:
						// Generate a unique value for this record
						value := make(map[string]interface{})
						for k, v := range valueTemplate {
							value[k] = generators.ProcessValue(v)
						}
						
						if err := r.Adapter.Update(ctx, keys[i], value); err != nil {
							errCh <- fmt.Errorf("failed to update record %d: %w", i, err)
							return
						}
					}
				}
			}(c, t)
		}
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
	
	// Check for errors
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	
	// Record result
	duration := time.Since(startTime)
	r.Results = append(r.Results, Result{
		Operation: OperationUpdate,
		Name:      "update_all",
		Duration:  duration,
		Count:     r.Config.Samples,
	})
	
	fmt.Printf("UPDATE completed in %v\n", duration)
	return nil
}

// runScans executes the scan benchmarks
func (r *Runner) runScans(ctx context.Context) error {
	fmt.Printf("Running SCAN benchmarks...\n")
	
	for _, scanConfig := range r.Config.Scans {
		fmt.Printf("Running scan '%s'...\n", scanConfig.Name)
		
		// Start timer
		startTime := time.Now()
		
		// Execute scan
		count, err := r.Adapter.Scan(ctx, scanConfig)
		if err != nil {
			return fmt.Errorf("failed to execute scan '%s': %w", scanConfig.Name, err)
		}
		
		// Verify count if expected
		if scanConfig.Expect > 0 && count != scanConfig.Expect {
			return fmt.Errorf("scan '%s' returned %d rows, expected %d", scanConfig.Name, count, scanConfig.Expect)
		}
		
		// Record result
		duration := time.Since(startTime)
		r.Results = append(r.Results, Result{
			Operation: OperationScan,
			Name:      scanConfig.Name,
			Duration:  duration,
			Count:     count,
		})
		
		fmt.Printf("Scan '%s' completed in %v with %d rows\n", scanConfig.Name, duration, count)
	}
	
	return nil
}

// runDelete executes the delete benchmark
func (r *Runner) runDelete(ctx context.Context) error {
	fmt.Printf("Running DELETE benchmark with %d samples...\n", r.Config.Samples)
	
	// Generate keys (same order as create)
	keys, err := generators.GenerateKeys(r.Config.KeyType, r.Config.Samples, r.Config.Random)
	if err != nil {
		return fmt.Errorf("failed to generate keys: %w", err)
	}
	
	// Start timer
	startTime := time.Now()
	
	// Delete records
	var wg sync.WaitGroup
	errCh := make(chan error, r.Config.Clients*r.Config.Threads)
	
	// Process in batches based on client and thread count
	batchSize := r.Config.Samples / (r.Config.Clients * r.Config.Threads)
	if batchSize == 0 {
		batchSize = 1
	}
	
	for c := 0; c < r.Config.Clients; c++ {
		for t := 0; t < r.Config.Threads; t++ {
			wg.Add(1)
			
			go func(clientID, threadID int) {
				defer wg.Done()
				
				// Calculate start and end indices for this worker
				start := (clientID*r.Config.Threads + threadID) * batchSize
				end := start + batchSize
				
				if end > r.Config.Samples {
					end = r.Config.Samples
				}
				
				if start >= r.Config.Samples {
					return
				}
				
				// Process assigned keys
				for i := start; i < end; i++ {
					select {
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					default:
						if err := r.Adapter.Delete(ctx, keys[i]); err != nil {
							errCh <- fmt.Errorf("failed to delete record %d: %w", i, err)
							return
						}
					}
				}
			}(c, t)
		}
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
	
	// Check for errors
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	
	// Record result
	duration := time.Since(startTime)
	r.Results = append(r.Results, Result{
		Operation: OperationDelete,
		Name:      "delete_all",
		Duration:  duration,
		Count:     r.Config.Samples,
	})
	
	fmt.Printf("DELETE completed in %v\n", duration)
	return nil
} 