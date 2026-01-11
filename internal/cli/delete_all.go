package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/middlendian/llima-box/internal/log"
	"github.com/middlendian/llima-box/pkg/env"
	"github.com/middlendian/llima-box/pkg/vm"
	"github.com/spf13/cobra"
)

// NewDeleteAllCommand creates the delete-all command.
func NewDeleteAllCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete-all",
		Short: "Delete all environments",
		Long: `Delete all isolated environments in the VM.

This removes all environment user accounts, home directories, and namespaces.
Any processes running in the environments will be terminated.

By default, prompts for confirmation before deletion. Use --force to skip.

WARNING: This cannot be undone!

Examples:
  # Delete all environments (with confirmation)
  llima-box delete-all

  # Delete all environments without confirmation
  llima-box delete-all --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteAll(cmd, args, force)
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Delete without confirmation")

	return cmd
}

func runDeleteAll(_ *cobra.Command, _ []string, force bool) error {
	// Check if VM exists
	vmManager := vm.NewManager("llima-box")

	exists, err := vmManager.Exists()
	if err != nil {
		return fmt.Errorf("failed to check VM existence: %w", err)
	}

	if !exists {
		log.Info("No VM exists. Nothing to delete.")
		return nil
	}

	// Check if VM is running
	running, err := vmManager.IsRunning()
	if err != nil {
		return fmt.Errorf("failed to check VM status: %w", err)
	}

	if !running {
		return fmt.Errorf("VM is not running (cannot delete environments)")
	}

	// List environments
	ctx := context.Background()
	envManager := env.NewManager(vmManager)
	defer func() { _ = envManager.Close() }()

	environments, err := envManager.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	if len(environments) == 0 {
		log.Info("No environments to delete.")
		return nil
	}

	// Show environments
	log.Info("Found %d environment(s):", len(environments))
	for _, e := range environments {
		projectPath := e.ProjectPath
		if projectPath == "" {
			projectPath = "(unknown)"
		}
		log.Plain("  - %s [%s]", e.Name, projectPath)
	}
	log.Plain("")

	// Confirm deletion
	if !force {
		log.Warning("Delete ALL %d environment(s)?", len(environments))
		log.Plain("This will terminate all processes and remove all data. Continue? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			log.Info("Cancelled")
			return nil
		}
	}

	// Delete all environments
	log.Info("Deleting environments...")
	successCount := 0
	failCount := 0

	for _, e := range environments {
		log.Plain("  Deleting %s... ", e.Name)
		if err := envManager.Delete(ctx, e.Name); err != nil {
			log.Error("FAILED: %v", err)
			failCount++
		} else {
			log.Success("OK")
			successCount++
		}
	}

	log.Plain("\nDeleted %d of %d environment(s)", successCount, len(environments))
	if failCount > 0 {
		log.Warning("%d failed", failCount)
		return fmt.Errorf("failed to delete %d environment(s)", failCount)
	}

	return nil
}
