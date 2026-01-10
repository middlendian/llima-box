// Package vm provides Lima VM lifecycle management functionality.
package vm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lima-vm/lima/pkg/instance"
	"github.com/lima-vm/lima/pkg/store"
)

const (
	// DefaultInstanceName is the name of the Lima VM instance
	DefaultInstanceName = "llima-box"
)

// Manager handles Lima VM lifecycle operations
type Manager struct {
	instanceName string
}

// NewManager creates a new VM manager
func NewManager(instanceName string) *Manager {
	if instanceName == "" {
		instanceName = DefaultInstanceName
	}
	return &Manager{
		instanceName: instanceName,
	}
}

// Exists checks if the VM instance exists
func (m *Manager) Exists() (bool, error) {
	instances, err := store.Instances()
	if err != nil {
		return false, fmt.Errorf("failed to list instances: %w", err)
	}

	for _, name := range instances {
		if name == m.instanceName {
			return true, nil
		}
	}
	return false, nil
}

// IsRunning checks if the VM is currently running
func (m *Manager) IsRunning() (bool, error) {
	inst, err := store.Inspect(m.instanceName)
	if err != nil {
		return false, fmt.Errorf("failed to inspect instance: %w", err)
	}

	return inst.Status == store.StatusRunning, nil
}

// GetInstance returns the instance details
func (m *Manager) GetInstance() (*store.Instance, error) {
	return store.Inspect(m.instanceName)
}

// Create creates a new Lima VM instance with the embedded configuration
func (m *Manager) Create(ctx context.Context) error {
	exists, err := m.Exists()
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("instance %s already exists", m.instanceName)
	}

	// Get configuration YAML
	configYAML, err := GetEmbeddedConfig()
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	// Create instance
	_, err = instance.Create(ctx, m.instanceName, []byte(configYAML), false)
	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}

	return nil
}

// Start starts the Lima VM instance
func (m *Manager) Start(ctx context.Context) error {
	inst, err := m.GetInstance()
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// Check if already running
	if inst.Status == store.StatusRunning {
		return nil
	}

	// Start the instance
	// Empty string for limactl means use current executable
	// false means don't run in foreground
	err = instance.Start(ctx, inst, "", false)
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	return nil
}

// Stop stops the Lima VM instance gracefully
func (m *Manager) Stop(ctx context.Context) error {
	inst, err := m.GetInstance()
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	err = instance.StopGracefully(ctx, inst, false)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	return nil
}

// Delete deletes the Lima VM instance
func (m *Manager) Delete(ctx context.Context, force bool) error {
	inst, err := m.GetInstance()
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	err = instance.Delete(ctx, inst, force)
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	return nil
}

// EnsureRunning ensures the VM is running, starting it if necessary
func (m *Manager) EnsureRunning(ctx context.Context) error {
	exists, err := m.Exists()
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("Creating Lima VM instance '%s'...\n", m.instanceName)
		if err := m.Create(ctx); err != nil {
			return err
		}
	}

	running, err := m.IsRunning()
	if err != nil {
		return err
	}

	if !running {
		fmt.Printf("Starting Lima VM instance '%s'...\n", m.instanceName)
		if err := m.Start(ctx); err != nil {
			return err
		}
		fmt.Println("VM started successfully")
	}

	return nil
}

// GetConfigPath returns the path to the Lima configuration file
func (m *Manager) GetConfigPath() (string, error) {
	limaDir := store.Directory()
	configPath := filepath.Join(limaDir, m.instanceName, "lima.yaml")
	return configPath, nil
}

// WriteDefaultConfig writes the default configuration to disk
func (m *Manager) WriteDefaultConfig() error {
	limaDir := store.Directory()
	instanceDir := filepath.Join(limaDir, m.instanceName)

	// #nosec G301 -- Standard permissions for Lima instance directories
	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		return fmt.Errorf("failed to create instance directory: %w", err)
	}

	configPath := filepath.Join(instanceDir, "lima.yaml")
	configYAML, err := GetEmbeddedConfig()
	if err != nil {
		return err
	}

	// #nosec G306 -- Standard permissions for config files
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetInstanceName returns the instance name
func (m *Manager) GetInstanceName() string {
	return m.instanceName
}
