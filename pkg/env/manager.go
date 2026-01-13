package env

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/middlendian/llima-box/pkg/ssh"
	"github.com/middlendian/llima-box/pkg/vm"
)

// Environment represents an isolated environment for a project
type Environment struct {
	// Name is the environment name (e.g., "my-project-a1b2")
	Name string

	// ProjectPath is the absolute path to the project directory
	ProjectPath string
}

// Manager handles environment lifecycle operations
type Manager struct {
	vmManager    *vm.Manager
	sshClient    *ssh.Client
	instanceName string
}

// NewManager creates a new environment manager
func NewManager(vmManager *vm.Manager) *Manager {
	return &Manager{
		vmManager:    vmManager,
		instanceName: vmManager.GetInstanceName(),
	}
}

// ensureSSH ensures SSH client is connected
func (m *Manager) ensureSSH(ctx context.Context) error {
	if m.sshClient != nil && m.sshClient.IsConnected() {
		return nil
	}

	// Ensure VM is running
	if err := m.vmManager.EnsureRunning(ctx); err != nil {
		return fmt.Errorf("failed to ensure VM is running: %w", err)
	}

	// Create SSH client
	client, err := ssh.NewClient(m.instanceName)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	// Connect with retries
	retryConfig := ssh.RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 2 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}
	if err := client.ConnectWithRetry(retryConfig); err != nil {
		return fmt.Errorf("failed to connect SSH: %w", err)
	}

	m.sshClient = client
	return nil
}

// Create creates a new environment or returns existing one
func (m *Manager) Create(ctx context.Context, projectPath string) (*Environment, error) {
	// Get absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Generate environment name
	envName, err := GenerateName(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate environment name: %w", err)
	}

	// Ensure SSH connection
	if err := m.ensureSSH(ctx); err != nil {
		return nil, err
	}

	// Check if environment already exists
	exists, err := m.Exists(ctx, envName)
	if err != nil {
		return nil, err
	}

	env := &Environment{
		Name:        envName,
		ProjectPath: absPath,
	}

	if exists {
		// Environment already exists, return it
		return env, nil
	}

	// Create user account
	if err := m.createUser(ctx, envName); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create namespace
	if err := m.createNamespace(ctx, env); err != nil {
		// Try to clean up user on failure
		_ = m.deleteUser(ctx, envName)
		return nil, fmt.Errorf("failed to create namespace: %w", err)
	}

	return env, nil
}

// Exists checks if an environment exists
func (m *Manager) Exists(ctx context.Context, envName string) (bool, error) {
	if err := m.ensureSSH(ctx); err != nil {
		return false, err
	}

	// Check if user account exists
	cmd := fmt.Sprintf("id %[1]s && [ -e /envs/%[1]s/namespace.pid ] && kill -0 $(cat /envs/%[1]s/namespace.pid)", envName)
	_, err := m.sshClient.ExecContext(ctx, cmd)
	return err == nil, nil
}

// List returns all environments
func (m *Manager) List(ctx context.Context) ([]*Environment, error) {
	if err := m.ensureSSH(ctx); err != nil {
		return nil, err
	}

	cmd := "find /envs/. -type d -maxdepth 1 -mindepth 1 | xargs basename"
	output, err := m.sshClient.ExecContext(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var envs []*Environment

	for _, line := range lines {
		if line == "" {
			continue
		}

		env := &Environment{
			Name:        line,
			ProjectPath: "", // Unknown without metadata
		}

		envs = append(envs, env)
	}

	return envs, nil
}

// Delete deletes an environment
func (m *Manager) Delete(ctx context.Context, envName string) error {
	if err := m.ensureSSH(ctx); err != nil {
		return err
	}

	// Check if environment exists
	exists, err := m.Exists(ctx, envName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("environment %s does not exist", envName)
	}

	// Kill processes in the namespace
	if err := m.killNamespaceProcesses(ctx, envName); err != nil {
		// Log but continue - processes might already be dead
		fmt.Printf("Warning: failed to kill namespace processes: %v\n", err)
	}

	// Delete user account (includes home directory)
	if err := m.deleteUser(ctx, envName); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// EnterNamespace enters an environment's namespace and executes a command
func (m *Manager) EnterNamespace(ctx context.Context, env *Environment, cmd []string) error {
	if err := m.ensureSSH(ctx); err != nil {
		return err
	}

	pidFile := fmt.Sprintf("/envs/%s/namespace.pid", env.Name)

	// Build the nsenter command to enter the namespace and run as the environment user
	// The working directory is /workspace (where the project is bind-mounted)
	var sshCmd string
	if len(cmd) == 0 {
		// Interactive shell - start bash in /workspace without login shell
		// This ensures we start in the project directory, not the home directory
		sshCmd = fmt.Sprintf(
			"sudo nsenter --target=$(sudo cat %s) --mount su %s --command 'cd /workspace && exec bash'",
			pidFile,
			env.Name,
		)
	} else {
		// Specific command - run it in /workspace
		command := strings.Join(cmd, " ")
		sshCmd = fmt.Sprintf(
			"sudo nsenter --target=$(sudo cat %s) --mount su %s --command 'cd /workspace && %s'",
			pidFile,
			env.Name,
			command,
		)
	}

	// Execute interactively
	return m.sshClient.ExecInteractive(sshCmd)
}

// createUser creates a Linux user account for the environment
func (m *Manager) createUser(ctx context.Context, username string) error {
	// Create user with home directory
	cmd := fmt.Sprintf("sudo useradd -m -s /bin/bash %s", username)

	fmt.Fprintf(os.Stderr, "\033[90mDEBUG\033[0m: Creating user: %s\n", cmd)

	output, err := m.sshClient.ExecContext(ctx, cmd)
	if err != nil {
		if output != "" {
			fmt.Fprintf(os.Stderr, "\033[90mDEBUG\033[0m: User creation output: %s\n", output)
		}
		return fmt.Errorf("failed to create user account: %w", err)
	}

	return nil
}

// deleteUser deletes a Linux user account
func (m *Manager) deleteUser(ctx context.Context, username string) error {
	// Delete user and home directory
	cmd := fmt.Sprintf("sudo userdel -r %s", username)
	_, err := m.sshClient.ExecContext(ctx, cmd)
	return err
}

// createNamespace creates a persistent namespace for the environment
func (m *Manager) createNamespace(ctx context.Context, env *Environment) error {
	pidFile := fmt.Sprintf("/envs/%s/namespace.pid", env.Name)

	// Create the /envs directory
	mkdirCmd := fmt.Sprintf("sudo mkdir -p /envs/%s", env.Name)
	if _, err := m.sshClient.ExecContext(ctx, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create namespace directory: %w", err)
	}

	// Build the unshare command with proper namespace setup
	// This creates mount and PID namespaces, runs a sleep process to keep them alive
	unshareCmd := fmt.Sprintf(`sudo unshare --mount --pid --fork --propagation private bash -c 'sleep infinity' >/dev/null 2>&1 & echo $! | sudo tee %s >/dev/null`, pidFile)

	fmt.Fprintf(os.Stderr, "\033[90mDEBUG\033[0m: Creating namespace: %s\n", unshareCmd)

	// Execute the unshare command
	if _, err := m.sshClient.ExecContext(ctx, unshareCmd); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// Wait a moment for the namespace to stabilize
	time.Sleep(500 * time.Millisecond)

	// Verify namespace PID file exists
	fmt.Fprintf(os.Stderr, "\033[90mDEBUG\033[0m: Verifying namespace PID file at: %s\n", pidFile)

	// Read the PID file
	catCmd := fmt.Sprintf("sudo cat %s 2>&1", pidFile)
	catOutput, catErr := m.sshClient.ExecContext(ctx, catCmd)
	if catErr != nil {
		fmt.Fprintf(os.Stderr, "\033[90mDEBUG\033[0m: Failed to read PID file: %v\nOutput: %s\n", catErr, catOutput)
		return fmt.Errorf("namespace PID file not created: %s (error: %w, output: %s)", pidFile, catErr, catOutput)
	}

	pid := strings.TrimSpace(catOutput)
	fmt.Fprintf(os.Stderr, "\033[90mDEBUG\033[0m: Namespace PID: %s\n", pid)

	// Verify the namespace process is still running
	checkProcCmd := fmt.Sprintf("sudo kill -0 %s 2>&1", pid)
	checkOutput, checkErr := m.sshClient.ExecContext(ctx, checkProcCmd)
	if checkErr != nil {
		fmt.Fprintf(os.Stderr, "\033[90mDEBUG\033[0m: Namespace process check failed: %v\nOutput: %s\n", checkErr, checkOutput)
		return fmt.Errorf("namespace process (PID %s) is not running: %w", pid, checkErr)
	}

	fmt.Fprintf(os.Stderr, "\033[90mDEBUG\033[0m: Namespace ready (PID %s is running)\n", pid)

	// Set up filesystem isolation inside the namespace
	// This bind-mounts only the project directory to /workspace
	if err := m.setupNamespaceFilesystem(ctx, env, pid); err != nil {
		return fmt.Errorf("failed to setup namespace filesystem: %w", err)
	}

	return nil
}

// setupNamespaceFilesystem sets up the isolated filesystem view inside a namespace
func (m *Manager) setupNamespaceFilesystem(ctx context.Context, env *Environment, namespacePID string) error {
	// Enter the namespace and set up bind mount for the project directory
	// Note: Each command needs sudo since nsenter runs bash as the lima user
	setupCmd := fmt.Sprintf(`sudo nsenter --mount --target=%s -- bash -c 'sudo mkdir -p /workspace && sudo mount --bind %s /workspace && sudo chown -R %s:%s /workspace'`,
		namespacePID, env.ProjectPath, env.Name, env.Name)

	fmt.Fprintf(os.Stderr, "\033[90mDEBUG\033[0m: Setting up namespace filesystem: %s\n", setupCmd)

	output, err := m.sshClient.ExecContext(ctx, setupCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[90mDEBUG\033[0m: Filesystem setup failed: %v\nOutput: %s\n", err, output)
		return fmt.Errorf("failed to setup bind mount: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\033[90mDEBUG\033[0m: Namespace filesystem isolation ready\n")
	return nil
}

// killNamespaceProcesses kills all processes running in the namespace
func (m *Manager) killNamespaceProcesses(ctx context.Context, username string) error {
	// Kill all processes owned by the user
	cmd := fmt.Sprintf("sudo pkill -u %s || true", username)
	_, err := m.sshClient.ExecContext(ctx, cmd)
	return err
}

// GetProjectPath attempts to recover the project path for an environment
// This is a best-effort operation since we don't store metadata yet
func (m *Manager) GetProjectPath(ctx context.Context, envName string) (string, error) {
	if err := m.ensureSSH(ctx); err != nil {
		return "", err
	}

	// Read the namespace PID
	pidFile := fmt.Sprintf("/envs/%s/namespace.pid", envName)
	pidOutput, err := m.sshClient.ExecContext(ctx, fmt.Sprintf("cat %s", pidFile))
	if err != nil {
		return "", fmt.Errorf("failed to read namespace PID: %w", err)
	}
	pid := strings.TrimSpace(pidOutput)

	// Try to find the project path from the namespace mounts
	// This is a heuristic - look for bind mounts in /proc/mounts
	cmd := fmt.Sprintf("sudo nsenter --mount=/proc/%s/ns/mnt findmnt -n -o TARGET | grep -E '^/Users|^/home' | grep -v '^/home/%s$' | head -n1 || echo ''",
		pid, envName)

	output, err := m.sshClient.ExecContext(ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("failed to get project path: %w", err)
	}

	projectPath := strings.TrimSpace(output)
	if projectPath == "" {
		return "", fmt.Errorf("could not determine project path for %s", envName)
	}

	return projectPath, nil
}

// DeleteAll deletes all environments
func (m *Manager) DeleteAll(ctx context.Context) error {
	envs, err := m.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	var errors []string
	for _, env := range envs {
		if err := m.Delete(ctx, env.Name); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", env.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete some environments: %s", strings.Join(errors, "; "))
	}

	return nil
}

// IsValidEnvironmentName checks if a name is a valid environment name
// This helps filter out system users when listing environments
func IsValidEnvironmentName(name string) bool {
	// Should match our generated name pattern: <base>-<hash>
	// where hash is 4 hex characters
	pattern := regexp.MustCompile(`^[a-z][a-z0-9-]*-[0-9a-f]{4}$`)
	return pattern.MatchString(name)
}

// Close closes the SSH connection
func (m *Manager) Close() error {
	if m.sshClient != nil {
		return m.sshClient.Close()
	}
	return nil
}
