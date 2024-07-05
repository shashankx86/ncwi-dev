package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"nuc/components"
	"nuc/config"
	"nuc/utils"
)

var versionCmd = &cobra.Command{
	Use:   "api-version",
	Short: "Get the API version",
	Run: func(cmd *cobra.Command, args []string) {
		configData, err := config.LoadConfig()
		utils.HandleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")

		if !utils.IsAPIServerOnline(configData.APIUrl) {
			fmt.Println("API server is offline")
			return
		}

		tokens, err := config.LoadTokens()
		utils.HandleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")

		if tokens.Expiration < time.Now().Unix() {
			log.Fatal("Token has expired. Please authenticate again.")
		}

		tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
		err = config.SaveTokens(tokens)
		utils.HandleErr(err, "Error updating token expiration")

		versionResponse, err := components.GetVersion(tokens.AccessToken, configData.APIUrl)
		utils.HandleErr(err, "Error getting version")

		fmt.Printf("API Version: %s\nUser: %s\n", versionResponse.Version, versionResponse.User)
	},
}
