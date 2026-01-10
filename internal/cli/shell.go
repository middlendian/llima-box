// Package cli implements the CLI commands for llima-box.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/middlendian/llima-box/pkg/env"
	"github.com/middlendian/llima-box/pkg/vm"
	"github.com/spf13/cobra"
)

// NewShellCommand creates the shell command.
func NewShellCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell [path] [-- command]",
		Short: "Enter an isolated environment shell",
		Long: `Enter an isolated environment shell for the specified project path.

The shell command creates a new isolated environment (if it doesn't exist) and
starts an interactive shell within that environment. Each environment has its
own filesystem view and user account.

Examples:
  # Enter shell for current directory
  llima-box shell

  # Enter shell for specific path
  llima-box shell /path/to/project

  # Run specific command in environment
  llima-box shell /path/to/project -- git status

  # Run command with arguments
  llima-box shell -- python script.py --arg value`,
		RunE: runShell,
	}

	return cmd
}

func runShell(cmd *cobra.Command, args []string) error {
	// Parse arguments
	projectPath, command, err := parseShellArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Ensure VM is running
	ctx := context.Background()
	fmt.Println("Ensuring VM is running...")
	vmManager := vm.NewManager("llima-box")

	exists, err := vmManager.Exists()
	if err != nil {
		return fmt.Errorf("failed to check VM existence: %w", err)
	}

	if !exists {
		fmt.Println("Creating VM (this may take a few minutes)...")
		if err := vmManager.Create(ctx); err != nil {
			return fmt.Errorf("failed to create VM: %w", err)
		}
	}

	if err := vmManager.EnsureRunning(ctx); err != nil {
		return fmt.Errorf("failed to start VM: %w", err)
	}

	// Create or get environment
	fmt.Printf("Setting up environment for %s...\n", projectPath)
	envManager := env.NewManager(vmManager)
	defer envManager.Close()

	environment, err := envManager.Create(ctx, projectPath)
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}

	fmt.Printf("Environment: %s\n", environment.Name)

	// Enter namespace and execute command
	if err := envManager.EnterNamespace(ctx, environment, command); err != nil {
		return fmt.Errorf("failed to enter namespace: %w", err)
	}

	return nil
}

// parseShellArgs parses the shell command arguments.
// Returns: (projectPath, command, error)
func parseShellArgs(args []string) (string, []string, error) {
	var projectPath string
	var command []string

	// Find "--" separator
	dashIndex := -1
	for i, arg := range args {
		if arg == "--" {
			dashIndex = i
			break
		}
	}

	// Parse based on "--" position
	if dashIndex == -1 {
		// No "--" separator
		if len(args) == 0 {
			// No args: use current directory, default shell
			cwd, err := os.Getwd()
			if err != nil {
				return "", nil, fmt.Errorf("failed to get current directory: %w", err)
			}
			projectPath = cwd
			command = []string{"bash"}
		} else {
			// One arg: treat as path, use default shell
			projectPath = args[0]
			command = []string{"bash"}
		}
	} else {
		// Has "--" separator
		if dashIndex == 0 {
			// "-- command": use current directory
			cwd, err := os.Getwd()
			if err != nil {
				return "", nil, fmt.Errorf("failed to get current directory: %w", err)
			}
			projectPath = cwd
			command = args[dashIndex+1:]
		} else {
			// "path -- command"
			projectPath = args[0]
			command = args[dashIndex+1:]
		}

		if len(command) == 0 {
			return "", nil, fmt.Errorf("no command specified after '--'")
		}
	}

	// Make path absolute
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Validate path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("path does not exist: %s", absPath)
		}
		return "", nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return "", nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	return absPath, command, nil
}

// formatCommand formats a command for display.
func formatCommand(command []string) string {
	if len(command) == 0 {
		return ""
	}
	return strings.Join(command, " ")
}
