package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"nuc/utils"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure the CLI tool",
}

var setURLCmd = &cobra.Command{
	Use:   "set-url <api-url>",
	Short: "Set the API URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiUrl := strings.TrimSuffix(args[0], "/") // Remove trailing slash if present

		err := utils.SaveConfig(apiUrl)
		utils.HandleErr(err, "Error saving config")

		fmt.Println("Configuration saved successfully.")
	},
}

func init() {
	configureCmd.AddCommand(setURLCmd)
}