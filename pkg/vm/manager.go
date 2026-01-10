// Package vm provides Lima VM lifecycle management functionality.
package vm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// DefaultInstanceName is the name of the Lima VM instance
	DefaultInstanceName = "llima-box"
)

// Instance represents a Lima VM instance
type Instance struct {
	Name         string          `json:"name"`
	Status       string          `json:"status"`
	Dir          string          `json:"dir"`
	Arch         string          `json:"arch"`
	CPUs         int             `json:"cpus"`
	Memory       int64           `json:"memory"`
	Disk         int64           `json:"disk"`
	SSHLocalPort int             `json:"sshLocalPort"`
	HostAgentPID int             `json:"hostAgentPID"`
	DriverPID    int             `json:"driverPID"`
	Config       *InstanceConfig `json:"config,omitempty"`
}

// InstanceConfig represents Lima instance configuration
type InstanceConfig struct {
	User *UserConfig `json:"user,omitempty"`
}

// UserConfig represents Lima user configuration
type UserConfig struct {
	Name *string `json:"name,omitempty"`
}

// Manager handles Lima VM lifecycle operations
type Manager struct {
	instanceName string
	limactl      string
}

// NewManager creates a new VM manager
func NewManager(instanceName string) *Manager {
	if instanceName == "" {
		instanceName = DefaultInstanceName
	}
	return &Manager{
		instanceName: instanceName,
		limactl:      "limactl",
	}
}

// findLimactl finds the limactl binary in PATH
func (m *Manager) findLimactl() error {
	_, err := exec.LookPath(m.limactl)
	if err != nil {
		return fmt.Errorf("limactl not found in PATH. Please install Lima: https://lima-vm.io/docs/installation/")
	}
	return nil
}

// execLimactl executes a limactl command
func (m *Manager) execLimactl(ctx context.Context, args ...string) ([]byte, error) {
	if err := m.findLimactl(); err != nil {
		return nil, err
	}

	// #nosec G204 -- args are controlled internally and validated
	cmd := exec.CommandContext(ctx, m.limactl, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("limactl %s failed: %w\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// Exists checks if the VM instance exists
func (m *Manager) Exists() (bool, error) {
	instances, err := m.listInstances()
	if err != nil {
		return false, fmt.Errorf("failed to list instances: %w", err)
	}

	for _, inst := range instances {
		if inst.Name == m.instanceName {
			return true, nil
		}
	}
	return false, nil
}

// listInstances lists all Lima instances
func (m *Manager) listInstances() ([]Instance, error) {
	output, err := m.execLimactl(context.Background(), "list", "--json")
	if err != nil {
		return nil, err
	}

	// Try to unmarshal as array first
	var instances []Instance
	if err := json.Unmarshal(output, &instances); err != nil {
		// If that fails, try as a single object
		var instance Instance
		if err2 := json.Unmarshal(output, &instance); err2 != nil {
			return nil, fmt.Errorf("failed to parse limactl list output as array or object: array error: %w, object error: %v", err, err2)
		}
		instances = []Instance{instance}
	}

	return instances, nil
}

// IsRunning checks if the VM is currently running
func (m *Manager) IsRunning() (bool, error) {
	inst, err := m.GetInstance()
	if err != nil {
		return false, fmt.Errorf("failed to inspect instance: %w", err)
	}

	return inst.Status == "Running", nil
}

// GetInstance returns the instance details
func (m *Manager) GetInstance() (*Instance, error) {
	instances, err := m.listInstances()
	if err != nil {
		return nil, err
	}

	for _, inst := range instances {
		if inst.Name == m.instanceName {
			return &inst, nil
		}
	}

	return nil, fmt.Errorf("instance %s not found", m.instanceName)
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

	// Write config to temporary file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, fmt.Sprintf("llima-box-%s.yaml", m.instanceName))
	// #nosec G306 -- Temporary config file
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		return fmt.Errorf("failed to write temporary config: %w", err)
	}
	defer func() {
		_ = os.Remove(configPath) // Best effort cleanup
	}()

	// Create instance with limactl
	_, err = m.execLimactl(ctx, "create", "--name="+m.instanceName, configPath)
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
	if inst.Status == "Running" {
		return nil
	}

	// Start the instance
	_, err = m.execLimactl(ctx, "start", m.instanceName)
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	return nil
}

// Stop stops the Lima VM instance gracefully
func (m *Manager) Stop(ctx context.Context) error {
	_, err := m.GetInstance()
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	_, err = m.execLimactl(ctx, "stop", m.instanceName)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	return nil
}

// Delete deletes the Lima VM instance
func (m *Manager) Delete(ctx context.Context, force bool) error {
	_, err := m.GetInstance()
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	args := []string{"delete", m.instanceName}
	if force {
		args = append(args, "--force")
	}

	_, err = m.execLimactl(ctx, args...)
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
	inst, err := m.GetInstance()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(inst.Dir, "lima.yaml")
	return configPath, nil
}

// GetLimaHome returns the Lima home directory
func (m *Manager) GetLimaHome() (string, error) {
	// Check LIMA_HOME environment variable
	if limaHome := os.Getenv("LIMA_HOME"); limaHome != "" {
		return limaHome, nil
	}

	// Default to ~/.lima
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".lima"), nil
}

// WriteDefaultConfig writes the default configuration to disk
func (m *Manager) WriteDefaultConfig() error {
	limaHome, err := m.GetLimaHome()
	if err != nil {
		return err
	}

	instanceDir := filepath.Join(limaHome, m.instanceName)

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
