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

var listServicesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services",
	Run: func(cmd *cobra.Command, args []string) {
		configData, err := config.LoadConfig()
		utils.HandleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")

		// Check if the API server is online
		if !config.IsAPIServerOnline(configData.APIUrl) {
			fmt.Println("API server is offline")
			return
		}

		tokens, err := utils.LoadTokens()
		utils.HandleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")

		// Check token expiration
		if tokens.Expiration < time.Now().Unix() {
			log.Fatal("Token has expired. Please authenticate again.")
		}

		// Reset token expiration
		tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
		err = utils.SaveTokens(tokens)
		utils.HandleErr(err, "Error updating token expiration")

		// Fetch services and handle the returned error
		services, _, err := components.FetchServices(configData.APIUrl, tokens.AccessToken)
		utils.HandleErr(err, "Error fetching services")

		// Print fetched services
		components.PrintServices(services)
	},
}

var listSocketsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sockets",
	Run: func(cmd *cobra.Command, args []string) {
		configData, err := config.LoadConfig()
		utils.HandleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")

		// Check if the API server is online
		if !config.IsAPIServerOnline(configData.APIUrl) {
			fmt.Println("API server is offline")
			return
		}

		tokens, err := utils.LoadTokens()
		utils.HandleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")

		// Check token expiration
		if tokens.Expiration < time.Now().Unix() {
			log.Fatal("Token has expired. Please authenticate again.")
		}

		// Reset token expiration
		tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
		err = utils.SaveTokens(tokens)
		utils.HandleErr(err, "Error updating token expiration")

		// Fetch sockets and handle the returned error
		_, sockets, err := components.FetchServices(configData.APIUrl, tokens.AccessToken)
		utils.HandleErr(err, "Error fetching services")

		// Print fetched sockets
		components.PrintSockets(sockets)
	},
}

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage services",
}

var socketsCmd = &cobra.Command{
	Use:   "sockets",
	Short: "Socket commands",
}

func init() {
	servicesCmd.AddCommand(listServicesCmd)
	socketsCmd.AddCommand(listSocketsCmd)
	systemCmd.AddCommand(servicesCmd, socketsCmd)
}
