package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"nuc/config"
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
		apiUrl := strings.TrimSuffix(args[0], "/")
		err := config.SaveConfig(apiUrl)
		utils.HandleErr(err, "Error saving config")
		fmt.Println("Configuration saved successfully.")
	},
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate and obtain access tokens",
	Run: func(cmd *cobra.Command, args []string) {
		configData, err := config.LoadConfig()
		utils.HandleErr(err, "Error loading config\nUse the 'configure set-url <api-url>' command to set the API URL")

		if !utils.IsAPIServerOnline(configData.APIUrl) {
			fmt.Println("API server is offline")
			return
		}

		username, err := utils.PromptInput("Enter username: ", false)
		utils.HandleErr(err, "Error reading username")

		password, err := utils.PromptInput("Enter password: ", true)
		utils.HandleErr(err, "Error reading password")

		err = config.Authenticate(configData.APIUrl, username, password)
		utils.HandleErr(err, "Error during authentication")
	},
}

func init() {
	configureCmd.AddCommand(setURLCmd, authCmd)
}
