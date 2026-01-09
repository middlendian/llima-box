package main

import (
	"fmt"
	"os"

	"github.com/middlendian/llima-box/pkg/env"
	"github.com/middlendian/llima-box/pkg/vm"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "llima-box",
	Short: "Secure multi-agent environment manager using Lima VMs",
	Long: `llima-box creates isolated environments for LLM agents using Lima VMs
and Linux mount namespaces. Each agent gets complete filesystem isolation
while sharing VM resources.`,
}

var testVMCmd = &cobra.Command{
	Use:   "test-vm",
	Short: "Test VM management (proof of concept)",
	Run:   runTestVM,
}

var testNamingCmd = &cobra.Command{
	Use:   "test-naming [path]",
	Short: "Test environment naming (generate name from path)",
	Args:  cobra.MaximumNArgs(1),
	Run:   runTestNaming,
}

func runTestVM(cmd *cobra.Command, args []string) {
	fmt.Println("llima-box VM Management Test")
	fmt.Println("=============================")

	manager := vm.NewManager("agents")

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

func runTestNaming(cmd *cobra.Command, args []string) {
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
	rootCmd.AddCommand(testVMCmd)
	rootCmd.AddCommand(testNamingCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
