package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/eiannone/keyboard"
	"github.com/spf13/cobra"
	"nuc/components"
)

// Config represents the structure of the configuration file
type Config struct {
	APIUrl string `json:"api_url"`
}

// TokenResponse represents the structure of the response from the login endpoint
type TokenResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	Expiration  int64  `json:"expiration"`
}

// VERSION represents the CLI version
const VERSION = "0.0.2"

// handleErr is a reusable error handling function that provides appropriate messages to the user
func handleErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

// getSavePath returns the path to save the token file
func getSavePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	handleErr(err, "Error getting home directory")

	dataDir := filepath.Join(homeDir, ".nuc", "data")
	err = os.MkdirAll(dataDir, 0700)
	handleErr(err, "Error creating data directory")

	return filepath.Join(dataDir, "data.bin"), nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	handleErr(err, "Error getting home directory")

	dataDir := filepath.Join(homeDir, ".nuc", "data")
	err = os.MkdirAll(dataDir, 0700)
	handleErr(err, "Error creating data directory")

	return filepath.Join(dataDir, "config.json"), nil
}

// saveTokens saves the tokens to a local file
func saveTokens(tokens TokenResponse) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tokens)
	if err != nil {
		return fmt.Errorf("error encoding tokens: %v", err)
	}

	savePath, err := getSavePath()
	if err != nil {
		return fmt.Errorf("error getting save path: %v", err)
	}

	err = ioutil.WriteFile(savePath, buffer.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("error writing tokens to file: %v", err)
	}

	return nil
}

// loadTokens loads the tokens from the local file
func loadTokens() (TokenResponse, error) {
	var tokens TokenResponse

	savePath, err := getSavePath()
	if err != nil {
		return tokens, fmt.Errorf("error getting save path: %v", err)
	}

	data, err := ioutil.ReadFile(savePath)
	if err != nil {
		return tokens, fmt.Errorf("error reading tokens from file: %v", err)
	}

	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	err = decoder.Decode(&tokens)
	if err != nil {
		return tokens, fmt.Errorf("error decoding tokens: %v", err)
	}

	return tokens, nil
}

// saveConfig saves the API URL to the configuration file
func saveConfig(apiUrl string) error {
	apiUrl = strings.TrimSuffix(apiUrl, "/") // Remove trailing slash if present
	config := Config{APIUrl: apiUrl}
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %v", err)
	}

	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("error getting config path: %v", err)
	}

	err = ioutil.WriteFile(configPath, data, 0600)
	if err != nil {
		return fmt.Errorf("error writing config to file: %v", err)
	}

	return nil
}

// loadConfig loads the API URL from the configuration file
func loadConfig() (Config, error) {
	var config Config

	configPath, err := getConfigPath()
	if err != nil {
		return config, fmt.Errorf("error getting config path: %v", err)
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("error reading config from file: %v", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshalling config: %v", err)
	}

	return config, nil
}

// promptInput prompts the user for input and masks the input with dots
func promptInput(prompt string, maskInput bool) (string, error) {
	fmt.Print(prompt)
	if !maskInput {
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(input), nil
	}

	err := keyboard.Open()
	if err != nil {
		return "", err
	}
	defer keyboard.Close()

	var input []rune
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			return "", err
		}
		if key == keyboard.KeyEnter {
			fmt.Println()
			break
		}
		if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
			if len(input) > 0 {
				input = input[:len(input)-1]
				fmt.Print("\b \b")
			}
		} else {
			input = append(input, char)
			fmt.Print("*")
		}
	}

	return strings.TrimSpace(string(input)), nil
}

// printASCIIArt prints a dummy ASCII art
func printASCIIArt() {
	fmt.Println("                                                  ")
	fmt.Println("b.             8 8 8888      88     ,o888888o.    ")
	fmt.Println("888o.          8 8 8888      88    8888     `88.  ")
	fmt.Println("Y88888o.       8 8 8888      88 ,8 8888       `8. ")
	fmt.Println(".`Y888888o.    8 8 8888      88 88 8888           ")
	fmt.Println("8o. `Y888888o. 8 8 8888      88 88 8888           ")
	fmt.Println("8`Y8o. `Y88888o8 8 8888      88 88 8888           ")
	fmt.Println("8   `Y8o. `Y8888 8 8888      88 88 8888           ")
	fmt.Println("8      `Y8o. `Y8 ` 8888     ,8P `8 8888       .8' ")
	fmt.Println("8         `Y8o.`   8888   ,d8P     8888     ,88'  ")
	fmt.Println("8            `Yo    `Y88888P'       `8888888P'    ")
	fmt.Println("                                    by shashankx86")
	fmt.Println("                                                  ")
}

// isAPIServerOnline checks if the API server is online
func isAPIServerOnline(apiUrl string) bool {
	resp, err := http.Get(apiUrl + "/ping")
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

var cmdShell = &cobra.Command{
	Use:   "shell",
	Short: "Open a reverse shell through the WebSocket server",
	Run: func(cmd *cobra.Command, args []string) {
		u := url.URL{Scheme: "wss", Host: "nws.theaddicts.hackclub.app", Path: "/ws"}

		// Connect to the WebSocket server
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

		// Signal handling to close the WebSocket connection gracefully
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)

		// Reader goroutine to read from the WebSocket and print to stdout
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

		// Writer goroutine to read from stdin and send to the WebSocket
		go func() {
			reader := bufio.NewReader(os.Stdin)
			for {
				command, err := reader.ReadString('\n')
				if err != nil {
					log.Println("read stdin:", err)
					break
				}

				if strings.TrimSpace(command) == "exit" {
					c.WriteMessage(websocket.TextMessage, []byte(command))
					break
				}

				err = c.WriteMessage(websocket.TextMessage, []byte(command))
				if err != nil {
					log.Println("write:", err)
					break
				}
			}
		}()

		for {
			select {
			case <-done:
				return
			case <-interrupt:
				log.Println("interrupt")
				// Gracefully close the WebSocket connection
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
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


func main() {
	// Get the name of the executable
	exeName := filepath.Base(os.Args[0])

	// Create the root command with the executable name
	var rootCmd = &cobra.Command{
		Use:   exeName,
		Short: "CLI tool for API interaction",
		Run: func(cmd *cobra.Command, args []string) {
			// Display the help message if no arguments are provided
			cmd.Help()
		},
	}

	var cmdConfigure = &cobra.Command{
		Use:   "configure",
		Short: "Configure the CLI tool",
	}

	var cmdSetURL = &cobra.Command{
		Use:   "set-url <api-url>",
		Short: "Set the API URL",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			apiUrl := strings.TrimSuffix(args[0], "/") // Remove trailing slash if present

			err := saveConfig(apiUrl)
			handleErr(err, "Error saving config")

			fmt.Println("Configuration saved successfully.")
		},
	}

	var cmdAuth = &cobra.Command{
		Use:   "auth",
		Short: "Authenticate and obtain access tokens",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := loadConfig()
			handleErr(err, "Error loading config\nUse the 'configure set-url <api-url>' command to set the API URL")
	
			// Check if the API server is online
			if !isAPIServerOnline(config.APIUrl) {
				fmt.Println("API server is offline")
				return
			}
	
			username, err := promptInput("Enter username: ", false)
			handleErr(err, "Error reading username")
	
			password, err := promptInput("Enter password: ", true)
			handleErr(err, "Error reading password")
	
			requestBody, err := json.Marshal(map[string]string{
				"username": username,
				"password": password,
			})
			handleErr(err, "Error marshalling request body")
	
			resp, err := http.Post(config.APIUrl+"/login", "application/json", bytes.NewBuffer(requestBody))
			handleErr(err, "Error sending login request")
			defer resp.Body.Close()
	
			var tokenResponse TokenResponse
			err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
			handleErr(err, "Error decoding login response")
	
			tokenResponse.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix() // Set token expiration to one month
	
			err = saveTokens(tokenResponse)
			handleErr(err, "Error saving tokens")
	
			fmt.Println(tokenResponse.Message)
		},
	}
	
	var cmdVersion = &cobra.Command{
		Use:   "api-version",
			Short: "Get the API version",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := loadConfig()
			handleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")
	
			// Check if the API server is online
			if (!isAPIServerOnline(config.APIUrl)) {
				fmt.Println("API server is offline")
				return
			}
	
			tokens, err := loadTokens()
			handleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")
	
			// Check token expiration
			if tokens.Expiration < time.Now().Unix() {
				log.Fatal("Token has expired. Please authenticate again.")
			}
	
			// Reset token expiration
			tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
			err = saveTokens(tokens)
			handleErr(err, "Error updating token expiration")
	
			versionResponse, err := components.GetVersion(tokens.AccessToken, config.APIUrl)
			handleErr(err, "Error getting version")
	
			fmt.Printf("API Version: %s\nUser: %s\n", versionResponse.Version, versionResponse.User)
		},
	}
	
	var cmdList = &cobra.Command{
		Use:   "list",
		Short: "List all services",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := loadConfig()
			handleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")
	
			// Check if the API server is online
			if (!isAPIServerOnline(config.APIUrl)) {
				fmt.Println("API server is offline")
				return
			}
	
			tokens, err := loadTokens()
			handleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")
	
			// Check token expiration
			if tokens.Expiration < time.Now().Unix() {
				log.Fatal("Token has expired. Please authenticate again.")
			}
	
			// Reset token expiration
			tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
			err = saveTokens(tokens)
			handleErr(err, "Error updating token expiration")
	
			// Fetch services and handle the returned error
			services, _, err := components.FetchServices(config.APIUrl, tokens.AccessToken)
			handleErr(err, "Error fetching services")
	
			// Print fetched services
			components.PrintServices(services)
		},
	}
	
	var cmdSocketList = &cobra.Command{
		Use:   "list",
		Short: "List all sockets",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := loadConfig()
			handleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")
	
			// Check if the API server is online
			if (!isAPIServerOnline(config.APIUrl)) {
				fmt.Println("API server is offline")
				return
			}
	
			tokens, err := loadTokens()
			handleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")
	
			// Check token expiration
			if tokens.Expiration < time.Now().Unix() {
				log.Fatal("Token has expired. Please authenticate again.")
			}
	
			// Reset token expiration
			tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
			err = saveTokens(tokens)
			handleErr(err, "Error updating token expiration")
	
			// Fetch sockets and handle the returned error
			_, sockets, err := components.FetchServices(config.APIUrl, tokens.AccessToken)
			handleErr(err, "Error fetching services")
	
			// Print fetched sockets
			components.PrintSockets(sockets)
		},
	}	

	// Add subcommands for stopping and starting services
	var cmdStopService = &cobra.Command{
		Use:   "stop [service]",
		Short: "Stop a specific service",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service := args[0]
			config, err := loadConfig()
			handleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")

			// Check if the API server is online
			if !isAPIServerOnline(config.APIUrl) {
				fmt.Println("API server is offline")
				return
			}

			tokens, err := loadTokens()
			handleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")

			// Check token expiration
			if tokens.Expiration < time.Now().Unix() {
				log.Fatal("Token has expired. Please authenticate again.")
			}

			// Reset token expiration
			tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
			err = saveTokens(tokens)
			handleErr(err, "Error updating token expiration")

			resp, err := http.PostForm(fmt.Sprintf("%s/system/services/stop", config.APIUrl), url.Values{"target": {service}})
			handleErr(err, "Error sending stop service request")
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Printf("Service %s stopped successfully.\n", service)
			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				log.Fatalf("Error stopping service: %s", string(body))
			}
		},
	}

	var cmdStartService = &cobra.Command{
		Use:   "start [service]",
		Short: "Start a specific service",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service := args[0]
			config, err := loadConfig()
			handleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")

			// Check if the API server is online
			if !isAPIServerOnline(config.APIUrl) {
				fmt.Println("API server is offline")
				return
			}

			tokens, err := loadTokens()
			handleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")

			// Check token expiration
			if tokens.Expiration < time.Now().Unix() {
				log.Fatal("Token has expired. Please authenticate again.")
			}

			// Reset token expiration
			tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
			err = saveTokens(tokens)
			handleErr(err, "Error updating token expiration")

			resp, err := http.PostForm(fmt.Sprintf("%s/system/services/start", config.APIUrl), url.Values{"target": {service}})
			handleErr(err, "Error sending start service request")
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Printf("Service %s started successfully.\n", service)
			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				log.Fatalf("Error starting service: %s", string(body))
			}
		},
	}

	// Add subcommands for docker commands
	var cmdDockerPS = &cobra.Command{
		Use:   "ps",
		Short: "List all currently running Docker containers",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := loadConfig()
			handleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")

			// Check if the API server is online
			if !isAPIServerOnline(config.APIUrl) {
				fmt.Println("API server is offline")
				return
			}

			tokens, err := loadTokens()
			handleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")

			// Check token expiration
			if tokens.Expiration < time.Now().Unix() {
				log.Fatal("Token has expired. Please authenticate again.")
			}

			// Reset token expiration
			tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
			err = saveTokens(tokens)
			handleErr(err, "Error updating token expiration")

			client := &http.Client{}
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/io/docker/running", config.APIUrl), nil)
			handleErr(err, "Error creating request")

			// Add authorization header with access token
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))

			resp, err := client.Do(req)
			handleErr(err, "Error sending list Docker containers request")
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var result struct {
					Containers []struct {
						ContainerID string `json:"CONTAINER_ID"`
						Image       string `json:"IMAGE"`
						Command     string `json:"COMMAND"`
						Created     string `json:"CREATED"`
						Status      string `json:"STATUS"`
						Ports       string `json:"PORTS"`
						Names       string `json:"NAMES"`
					} `json:"containers"`
				}

				err = json.NewDecoder(resp.Body).Decode(&result)
				handleErr(err, "Error decoding Docker containers response")

				for _, container := range result.Containers {
					fmt.Printf("ID: %s\nImage: %s\nCommand: %s\nCreated: %s\nStatus: %s\nPorts: %s\nNames: %s\n\n",
						container.ContainerID, container.Image, container.Command, container.Created, container.Status, container.Ports, container.Names)
				}
			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				log.Fatalf("Error listing Docker containers: %s", string(body))
			}
		},
	}

	// Define the start command for Docker
	var cmdDockerStart = &cobra.Command{
		Use:   "start [container]",
		Short: "Start a specific Docker container",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			container := args[0]
			config, err := loadConfig()
			handleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")

			// Check if the API server is online
			if (!isAPIServerOnline(config.APIUrl)) {
				fmt.Println("API server is offline")
				return
			}

			tokens, err := loadTokens()
			handleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")

			// Check token expiration
			if (tokens.Expiration < time.Now().Unix()) {
				log.Fatal("Token has expired. Please authenticate again.")
			}

			// Reset token expiration
			tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
			err = saveTokens(tokens)
			handleErr(err, "Error updating token expiration")

			resp, err := http.PostForm(fmt.Sprintf("%s/io/docker/start", config.APIUrl), url.Values{"target": {container}})
			handleErr(err, "Error sending start container request")
			defer resp.Body.Close()

			if (resp.StatusCode == http.StatusOK) {
				fmt.Printf("Container %s started successfully.\n", container)
			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				log.Fatalf("Error starting container: %s", string(body))
			}
		},
	}

	// Define the stop command for Docker
	var cmdDockerStop = &cobra.Command{
		Use:   "stop [container]",
		Short: "Stop a specific Docker container",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			container := args[0]
			config, err := loadConfig()
			handleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")

			// Check if the API server is online
			if (!isAPIServerOnline(config.APIUrl)) {
				fmt.Println("API server is offline")
				return
			}

			tokens, err := loadTokens()
			handleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")

			// Check token expiration
			if (tokens.Expiration < time.Now().Unix()) {
				log.Fatal("Token has expired. Please authenticate again.")
			}

			// Reset token expiration
			tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
			err = saveTokens(tokens)
			handleErr(err, "Error updating token expiration")

			resp, err := http.PostForm(fmt.Sprintf("%s/io/docker/stop", config.APIUrl), url.Values{"target": {container}})
			handleErr(err, "Error sending stop container request")
			defer resp.Body.Close()

			if (resp.StatusCode == http.StatusOK) {
				fmt.Printf("Container %s stopped successfully.\n", container)
			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				log.Fatalf("Error stopping container: %s", string(body))
			}
		},
	}

	// Define the rm command for Docker
	var cmdDockerRm = &cobra.Command{
		Use:   "rm [image]",
		Short: "Remove a specific Docker image",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			image := args[0]
			config, err := loadConfig()
			handleErr(err, "Error loading config\nUse the 'configure set-url' command to set the API URL")

			// Check if the API server is online
			if !isAPIServerOnline(config.APIUrl) {
				fmt.Println("API server is offline")
				return
			}

			tokens, err := loadTokens()
			handleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")

			// Check token expiration
			if tokens.Expiration < time.Now().Unix() {
				log.Fatal("Token has expired. Please authenticate again.")
			}

			// Reset token expiration
			tokens.Expiration = time.Now().Add(30 * 24 * time.Hour).Unix()
			err = saveTokens(tokens)
			handleErr(err, "Error updating token expiration")

			// Check if force flag is set
			force, _ := cmd.Flags().GetBool("force")

			client := &http.Client{}
			req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/docker/image/rm?targetid=%s&toforce=%t", config.APIUrl, image, force), nil)
			handleErr(err, "Error creating request")

			// Add authorization header with access token
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))

			resp, err := client.Do(req)
			handleErr(err, "Error sending delete image request")
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var result struct {
					Message string `json:"message"`
				}
				err = json.NewDecoder(resp.Body).Decode(&result)
				handleErr(err, "Error decoding delete image response")

				fmt.Println(result.Message)
			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				log.Fatalf("Error removing Docker image: %s", string(body))
			}
		},
	}

	var cmdSystem = &cobra.Command{
		Use:   "system",
		Short: "System commands",
	}

	var cmdServices = &cobra.Command{
		Use:   "services",
		Short: "Manage services",
	}

	var cmdSocket = &cobra.Command{
		Use:   "sockets",
		Short: "Socket commands",
	}

	var cmdDocker = &cobra.Command{
		Use:   "docker",
		Short: "Docker commands",
	}

	// Add commands to root and configure command
	cmdDockerRm.Flags().Bool("force", false, "Force remove the image")

	rootCmd.AddCommand(cmdConfigure, cmdVersion, cmdSystem, cmdShell, cmdDocker)
	cmdConfigure.AddCommand(cmdSetURL, cmdAuth)
	cmdSystem.AddCommand(cmdServices)
	cmdServices.AddCommand(cmdList, cmdStopService, cmdStartService)
	cmdSystem.AddCommand(cmdSocket)
	cmdSocket.AddCommand(cmdSocketList)
	cmdDocker.AddCommand(cmdDockerPS, cmdDockerStart, cmdDockerStop, cmdDockerRm)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
