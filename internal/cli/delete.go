package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/middlendian/llima-box/pkg/env"
	"github.com/middlendian/llima-box/pkg/vm"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the delete command.
func NewDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [path]",
		Short: "Delete an environment",
		Long: `Delete an isolated environment for the specified project path.

This removes the environment's user account, home directory, and namespace.
Any processes running in the environment will be terminated.

By default, prompts for confirmation before deletion. Use --force to skip.

Examples:
  # Delete environment for current directory
  llima-box delete

  # Delete environment for specific path
  llima-box delete /path/to/project

  # Delete without confirmation
  llima-box delete --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, args, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Delete without confirmation")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string, force bool) error {
	// Parse path
	projectPath, err := parseDeletePath(args)
	if err != nil {
		return err
	}

	// Generate environment name
	envName, err := env.GenerateName(projectPath)
	if err != nil {
		return fmt.Errorf("failed to generate environment name: %w", err)
	}

	// Check if VM exists
	vmManager := vm.NewManager("llima-box")

	exists, err := vmManager.Exists()
	if err != nil {
		return fmt.Errorf("failed to check VM existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("VM does not exist (no environments to delete)")
	}

	// Check if VM is running
	running, err := vmManager.IsRunning()
	if err != nil {
		return fmt.Errorf("failed to check VM status: %w", err)
	}

	if !running {
		return fmt.Errorf("VM is not running (cannot delete environments)")
	}

	// Check if environment exists
	ctx := context.Background()
	envManager := env.NewManager(vmManager)
	defer envManager.Close()

	envExists, err := envManager.Exists(ctx, envName)
	if err != nil {
		return fmt.Errorf("failed to check environment existence: %w", err)
	}

	if !envExists {
		return fmt.Errorf("environment %s does not exist", envName)
	}

	// Confirm deletion
	if !force {
		fmt.Printf("Delete environment '%s' for project '%s'?\n", envName, projectPath)
		fmt.Print("This will terminate all processes and remove all data. Continue? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Delete environment
	fmt.Printf("Deleting environment %s...\n", envName)
	if err := envManager.Delete(ctx, envName); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	fmt.Println("Environment deleted successfully.")

	return nil
}

// parseDeletePath parses the delete command path argument.
func parseDeletePath(args []string) (string, error) {
	var projectPath string

	if len(args) == 0 {
		// Use current directory
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		projectPath = cwd
	} else {
		projectPath = args[0]
	}

	// Make path absolute
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return absPath, nil
}
