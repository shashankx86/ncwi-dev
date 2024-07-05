package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var SocketCmd = &cobra.Command{
	Use:   "socket",
	Short: "Connect to a WebSocket server",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatalf("Usage: nuc socket <websocket-url>")
		}
		websocketURL := args[0]
		connectToWebSocket(websocketURL)
	},
}

func connectToWebSocket(url string) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer c.Close()

	go handleMessages(c)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)
	go handleInterrupt(c, stopChan)

	handleInput(c)
}

func handleMessages(c *websocket.Conn) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Fatalf("Read Error: %v", err)
		}
		fmt.Print(string(message))
	}
}

func handleInterrupt(c *websocket.Conn, stopChan chan os.Signal) {
	<-stopChan
	fmt.Println("\nReceived interrupt signal. Closing connection...")
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Fatalf("Close Error: %v", err)
	}
	time.Sleep(time.Second)
	os.Exit(0)
}

func handleInput(c *websocket.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")
		command, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Input Error: %v", err)
		}
		err = c.WriteMessage(websocket.TextMessage, bytes.TrimSpace([]byte(command)))
		if err != nil {
			log.Fatalf("Write Error: %v", err)
		}
	}
}
