package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func Execute() {
	// Get the name of the executable
	exeName := filepath.Base(os.Args[0])

	// Create the root command with the executable name
	var rootCmd = &cobra.Command{
		Use:   exeName,
		Short: "CLI tool for API interaction",
		Run: func(cmd *cobra.Command, args []string) {
			// Display the help message if no arguments are provided
			cmd.Help()
		},
	}

	// Add commands to root
	rootCmd.AddCommand(configureCmd, versionCmd, systemCmd, shellCmd)
	
	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
