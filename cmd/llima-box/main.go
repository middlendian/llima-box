package main

import (
	"context"
	"fmt"
	"os"

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
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

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

		// Test EnsureRunning
		fmt.Println("\nTesting EnsureRunning()...")
		if err := manager.EnsureRunning(ctx); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("VM is now running!")
	},
}

func init() {
	rootCmd.AddCommand(testVMCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
