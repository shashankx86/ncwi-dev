package cmd

import (
	"github.com/spf13/cobra"
	"nuc/utils"
)

var socketCmd = &cobra.Command{
	Use:   "socket",
	Short: "Connect to a WebSocket server",
	Run: func(cmd *cobra.Command, args []string) {
		utils.ConnectToWebSocket()
	},
}

func init() {
	rootCmd.AddCommand(socketCmd)
}
