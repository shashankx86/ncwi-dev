package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"nuc/utils"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open a reverse shell through the WebSocket server",
	Run: func(cmd *cobra.Command, args []string) {
		u := url.URL{Scheme: "wss", Host: "nws.theaddicts.hackclub.app", Path: "/ws"}

		c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			if resp != nil {
				body, _ := ioutil.ReadAll(resp.Body)
				log.Fatalf("Error connecting to WebSocket server: %v, Response: %s", err, string(body))
			} else {
				log.Fatalf("Error connecting to WebSocket server: %v", err)
			}
		}
		defer c.Close()

		done := make(chan struct{})
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)

		go func() {
			defer close(done)
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}
				fmt.Print(string(message))
			}
		}()

		go func() {
			reader := bufio.NewReader(os.Stdin)
			for {
				command, err := reader.ReadString('\n')
				if err != nil {
					log.Println("Error reading command:", err)
					return
				}

				err = c.WriteMessage(websocket.TextMessage, []byte(command))
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
		}()

		for {
			select {
			case <-done:
				return
			case <-interrupt:
				log.Println("interrupt")
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("Error closing connection:", err)
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				return
			}
		}
	},
}
