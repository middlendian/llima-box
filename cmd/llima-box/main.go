// Package main provides the llima-box CLI tool for managing secure, isolated
// environments for LLM agents using Lima VMs.
package main

import (
	"fmt"
	"os"

	"github.com/middlendian/llima-box/internal/cli"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "llima-box",
	Short: "Secure multi-agent environment manager using Lima VMs",
	Long: `llima-box creates isolated environments for LLM agents using Lima VMs
and Linux mount namespaces. Each agent gets complete filesystem isolation
while sharing VM resources.

Commands:
  shell       Enter an isolated environment shell
  list        List all environments
  delete      Delete an environment
  delete-all  Delete all environments

Use "llima-box <command> --help" for more information about a command.`,
}

func init() {
	rootCmd.AddCommand(cli.NewShellCommand())
	rootCmd.AddCommand(cli.NewListCommand())
	rootCmd.AddCommand(cli.NewDeleteCommand())
	rootCmd.AddCommand(cli.NewDeleteAllCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
