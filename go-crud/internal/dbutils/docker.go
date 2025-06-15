package dbutils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/surrealdb/go-crud-bench/internal/docker"
)

// EnsureDockerImage checks if the specified Docker image is available locally
// and pulls it if necessary. Returns true if the image was pulled.
func EnsureDockerImage(imageName string) (bool, error) {
	// Check if image exists locally
	cmd := exec.Command("docker", "image", "inspect", imageName)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err == nil {
		// Image exists, nothing to do
		return false, nil
	}

	// Image doesn't exist, pull it
	fmt.Printf("Image not found locally, pulling %s...\n", imageName)
	pullCmd := exec.Command("docker", "pull", imageName)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	
	if err := pullCmd.Run(); err != nil {
		return false, fmt.Errorf("failed to pull Docker image %s: %w", imageName, err)
	}
	
	return true, nil
}

// CreateContainerWithRetry creates and starts a Docker container with automatic image pulling
// if needed. It handles retries if the image is not available.
func CreateContainerWithRetry(
	ctx context.Context, 
	containerName string,
	imageName string,
	ports map[string]string,
	privileged bool,
	env []string) (*docker.Container, error) {
	
	// First, ensure the image is available
	if _, err := EnsureDockerImage(imageName); err != nil {
		return nil, err
	}

	// Create container
	container, err := docker.NewContainer(containerName, imageName, ports, privileged, env)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container with retry if needed
	if err := container.Start(ctx); err != nil {
		// If container start fails, try manual image pull and retry
		if strings.Contains(err.Error(), "No such image") {
			fmt.Printf("Container start failed, trying to pull image %s manually...\n", imageName)
			
			// Manual pull as a fallback
			pullCmd := exec.Command("docker", "pull", imageName)
			pullCmd.Stdout = os.Stdout
			pullCmd.Stderr = os.Stderr
			if err := pullCmd.Run(); err != nil {
				return nil, fmt.Errorf("manual docker pull failed: %w", err)
			}
			
			// Try to create and start container again
			container, err = docker.NewContainer(containerName, imageName, ports, privileged, env)
			if err != nil {
				return nil, fmt.Errorf("failed to create container after image pull: %w", err)
			}
			
			if err := container.Start(ctx); err != nil {
				return nil, fmt.Errorf("failed to start container after image pull: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to start container: %w", err)
		}
	}

	return container, nil
} 