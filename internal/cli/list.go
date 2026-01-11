package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/middlendian/llima-box/internal/log"
	"github.com/middlendian/llima-box/pkg/env"
	"github.com/middlendian/llima-box/pkg/vm"
	"github.com/spf13/cobra"
)

// NewListCommand creates the list command.
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all environments",
		Long: `List all isolated environments running in the VM.

Shows the environment name and associated project path (if available).
Environments are created automatically when you run 'llima-box shell'.

Example:
  llima-box list`,
		RunE:         runList,
		SilenceUsage: true,
	}

	return cmd
}

func runList(_ *cobra.Command, _ []string) error {
	// Check if VM exists
	vmManager := vm.NewManager("llima-box")

	exists, err := vmManager.Exists()
	if err != nil {
		return fmt.Errorf("failed to check VM existence: %w", err)
	}

	if !exists {
		log.Info("No VM created yet. Use 'llima-box shell' to create one.")
		return nil
	}

	// Check if VM is running
	running, err := vmManager.IsRunning()
	if err != nil {
		return fmt.Errorf("failed to check VM status: %w", err)
	}

	if !running {
		log.Info("VM is not running. Use 'llima-box shell' to start it.")
		return nil
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
		log.Info("No environments found. Use 'llima-box shell' to create one.")
		return nil
	}

	// Print table to stdout (so it can be captured/redirected)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ENVIRONMENT\tPROJECT PATH")
	_, _ = fmt.Fprintln(w, "-----------\t------------")

	for _, e := range environments {
		projectPath := e.ProjectPath
		if projectPath == "" {
			projectPath = "(unknown)"
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\n", e.Name, projectPath)
	}

	_ = w.Flush()

	// Summary to stderr
	log.Plain("\nTotal: %d environment(s)", len(environments))

	return nil
}
