package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Container represents a Docker container
type Container struct {
	ID         string
	Name       string
	Image      string
	Ports      map[string]string
	Privileged bool
	Env        []string
	Client     *client.Client
}

// NewContainer creates a new Docker container configuration
func NewContainer(name, image string, ports map[string]string, privileged bool, env []string) (*Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Container{
		Name:       name,
		Image:      image,
		Ports:      ports,
		Privileged: privileged,
		Env:        env,
		Client:     cli,
	}, nil
}

// Start starts the Docker container
func (c *Container) Start(ctx context.Context) error {
	// Pull the image
	_, err := c.Client.ImagePull(ctx, c.Image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull Docker image %s: %w", c.Image, err)
	}

	// Prepare port bindings
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	for containerPort, hostPort := range c.Ports {
		port := nat.Port(containerPort)
		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	// Create container
	resp, err := c.Client.ContainerCreate(
		ctx,
		&container.Config{
			Image:        c.Image,
			ExposedPorts: exposedPorts,
			Env:          c.Env,
		},
		&container.HostConfig{
			PortBindings: portBindings,
			Privileged:   c.Privileged,
		},
		&network.NetworkingConfig{},
		nil,
		c.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	c.ID = resp.ID

	// Start container
	if err := c.Client.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

// Stop stops and removes the Docker container
func (c *Container) Stop(ctx context.Context) error {
	if c.ID == "" {
		return nil
	}

	// Stop container
	timeout := 10 * time.Second
	if err := c.Client.ContainerStop(ctx, c.ID, &timeout); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	// Remove container
	if err := c.Client.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

// WaitForHealthy waits for the container to be healthy
func (c *Container) WaitForHealthy(ctx context.Context, timeout time.Duration, checkFunc func(ctx context.Context) error) error {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		// Check if container is running
		inspect, err := c.Client.ContainerInspect(ctx, c.ID)
		if err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}
		
		if !inspect.State.Running {
			return fmt.Errorf("container is not running")
		}
		
		// Run custom health check
		if checkFunc != nil {
			if err := checkFunc(ctx); err == nil {
				return nil
			}
		}
		
		time.Sleep(500 * time.Millisecond)
	}
	
	return fmt.Errorf("container health check timed out after %v", timeout)
} 