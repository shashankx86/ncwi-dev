package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"nuc/components"
	"nuc/config"
	"nuc/utils"
)

var socketsCmd = &cobra.Command{
	Use:   "sockets",
	Short: "Socket commands",
}

var listSocketsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sockets",
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

		_, sockets, err := components.FetchServices(configData.APIUrl, tokens.AccessToken)
		utils.HandleErr(err, "Error fetching services")

		components.PrintSockets(sockets)
	},
}

func init() {
	socketsCmd.AddCommand(listSocketsCmd)
}
