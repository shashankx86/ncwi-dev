package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open a reverse shell through the WebSocket server",
	Run: func(cmd *cobra.Command, args []string) {
		u := url.URL{Scheme: "wss", Host: "nws.theaddicts.hackclub.app", Path: "/ws"}

		// Connect to the WebSocket server
		c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			if resp != nil {
				body, _ := ioutil.ReadAll(resp.Body)
				log.Fatalf("HTTP Response Error: %s\nResponse Body: %s", err, body)
			} else {
				log.Fatalf("Dial Error: %s", err)
			}
		}
		defer c.Close()

		go handleMessages(c)

		// Handle Ctrl+C signal
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, os.Interrupt)
		go handleInterrupt(c, stopChan)

		handleInput(c)
	},
}

func handleMessages(c *websocket.Conn) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Fatalf("Read Error: %s", err)
		}
		fmt.Print(string(message))
	}
}

func handleInterrupt(c *websocket.Conn, stopChan chan os.Signal) {
	<-stopChan
	fmt.Println("\nReceived interrupt signal. Closing connection...")
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Fatalf("Close Error: %s", err)
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
			log.Fatalf("Input Error: %s", err)
		}
		err = c.WriteMessage(websocket.TextMessage, bytes.TrimSpace([]byte(command)))
		if err != nil {
			log.Fatalf("Write Error: %s", err)
		}
	}
}
