package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"nuc/components"
	"nuc/utils"
)

var versionCmd = &cobra.Command{
	Use:   "api-version",
	Short: "Get the API version",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := utils.LoadConfig()
		utils.HandleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")

		// Check if the API server is online
		if !utils.IsAPIServerOnline(config.APIUrl) {
			fmt.Println("API server is offline")
			return
		}

		tokens, err := utils.LoadTokens()
		utils.HandleErr(err, "Error loading tokens\nPlease authenticate using the 'auth' command")

		// Check token expiration
		if tokens.Expiration < time.Now().Unix() {
			log.Fatal("Token has expired. Please authenticate again.")
		}

		// Reset token expiration
		tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
		err = utils.SaveTokens(tokens)
		utils.HandleErr(err, "Error updating token expiration")

		versionResponse, err := components.GetVersion(tokens.AccessToken, config.APIUrl)
		utils.HandleErr(err, "Error getting version")

		fmt.Printf("API Version: %s\nUser: %s\n", versionResponse.Version, versionResponse.User)
	},
}
