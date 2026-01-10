// Package main provides the llima-box CLI tool for managing secure, isolated
// environments for LLM agents using Lima VMs.
package main

import (
	"fmt"
	"os"

	"github.com/middlendian/llima-box/internal/cli"
	"github.com/middlendian/llima-box/pkg/env"
	"github.com/middlendian/llima-box/pkg/vm"
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

var testVMCmd = &cobra.Command{
	Use:    "test-vm",
	Short:  "Test VM management (proof of concept)",
	Run:    runTestVM,
	Hidden: true, // Hide from main help
}

var testNamingCmd = &cobra.Command{
	Use:    "test-naming [path]",
	Short:  "Test environment naming (generate name from path)",
	Args:   cobra.MaximumNArgs(1),
	Run:    runTestNaming,
	Hidden: true, // Hide from main help
}

func runTestVM(_ *cobra.Command, _ []string) {
	fmt.Println("llima-box VM Management Test")
	fmt.Println("=============================")

	manager := vm.NewManager("llima-box")

	// Check if VM exists
	fmt.Print("\nChecking if VM exists... ")
	exists, err := manager.Exists()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%t\n", exists)

	if exists {
		// Check if running
		fmt.Print("Checking if VM is running... ")
		running, err := manager.IsRunning()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%t\n", running)

		// Get instance details
		fmt.Print("Getting instance details... ")
		inst, err := manager.GetInstance()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("OK\n")
		fmt.Printf("  Name: %s\n", inst.Name)
		fmt.Printf("  Status: %s\n", inst.Status)
		fmt.Printf("  Dir: %s\n", inst.Dir)
	} else {
		fmt.Println("\nVM does not exist. Use 'llima-box shell' to create it.")
	}
}

func runTestNaming(_ *cobra.Command, args []string) {
	var path string
	if len(args) == 0 {
		// Use current directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		path = cwd
	} else {
		path = args[0]
	}

	fmt.Println("llima-box Environment Naming Test")
	fmt.Println("==================================")
	fmt.Printf("\nProject Path: %s\n", path)

	name, err := env.GenerateName(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Environment Name: %s\n", name)

	// Validate
	if env.IsValidName(name) {
		fmt.Println("✓ Valid Linux username")
	} else {
		fmt.Println("✗ Invalid Linux username")
	}

	fmt.Printf("\nLength: %d characters (max 32)\n", len(name))
	fmt.Printf("Format: <basename>-<hash>\n")
	fmt.Printf("\nThis name will be used for:\n")
	fmt.Printf("  - Linux user account: %s\n", name)
	fmt.Printf("  - Home directory: /home/%s\n", name)
	fmt.Printf("  - Namespace file: /home/%s/namespace.mnt\n", name)
}

func init() {
	// Add production commands
	rootCmd.AddCommand(cli.NewShellCommand())
	rootCmd.AddCommand(cli.NewListCommand())
	rootCmd.AddCommand(cli.NewDeleteCommand())
	rootCmd.AddCommand(cli.NewDeleteAllCommand())

	// Add test commands (hidden)
	rootCmd.AddCommand(testVMCmd)
	rootCmd.AddCommand(testNamingCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
