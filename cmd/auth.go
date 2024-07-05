package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"nuc/utils"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate and obtain access tokens",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := utils.LoadConfig()
		utils.HandleErr(err, "Error loading config\nUse the 'configure set-url <api-url>' command to set the API URL")

		// Check if the API server is online
		if !utils.IsAPIServerOnline(config.APIUrl) {
			fmt.Println("API server is offline")
			return
		}

		username, err := utils.PromptInput("Enter username: ", false)
		utils.HandleErr(err, "Error reading username")

		password, err := utils.PromptInput("Enter password: ", true)
		utils.HandleErr(err, "Error reading password")

		tokens, err := utils.Authenticate(username, password, config.APIUrl)
		utils.HandleErr(err, "Error during authentication")

		tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix() // Set token expiration to one month

		err = utils.SaveTokens(tokens)
		utils.HandleErr(err, "Error saving tokens")

		fmt.Println(tokens.Message)
	},
}
