package cmd

import (
	"github.com/spf13/cobra"
	"nuc/utils"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open a reverse shell through the WebSocket server",
	Run: func(cmd *cobra.Command, args []string) {
		utils.ConnectToWebSocket()
	},
}
