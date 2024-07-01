package main

import (
    "bytes"
    "encoding/gob"
    "encoding/json"
    "bufio"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"

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
    Message           string `json:"message"`
    AccessToken       string `json:"access_token"`
    RefreshToken      string `json:"refresh_token"`
    RefreshExpiration int64  `json:"refresh_expiration"`
}

// VERSION represents the CLI version
const VERSION = "0.0.1"

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
}

// main initializes the CLI commands and their flags
func main() {
    var rootCmd = &cobra.Command{Use: "nuc"}

    var configureCmd = &cobra.Command{
        Use:   "configure",
        Short: "Configure the API URL",
        Run: func(cmd *cobra.Command, args []string) {
            apiUrl, _ := cmd.Flags().GetString("api-url")
            if apiUrl == "" {
                apiUrl = "api.example.com"
                fmt.Println("No API URL provided, using example:", apiUrl)
            }
            err := saveConfig(apiUrl)
            handleErr(err, "Error saving config")
            fmt.Println("Configuration saved successfully")
        },
    }
    configureCmd.Flags().String("api-url", "", "API URL")
    rootCmd.AddCommand(configureCmd)

    var authCmd = &cobra.Command{
        Use:   "auth",
        Short: "Authenticate and get tokens",
        Run: func(cmd *cobra.Command, args []string) {
            username, err := promptInput("Enter username: ", false)
            handleErr(err, "Error reading username")
            password, err := promptInput("Enter password: ", true)
            handleErr(err, "Error reading password")

            config, err := loadConfig()
            if err != nil {
                fmt.Println("Configuration not found. Please run 'configure --api-url' first.")
                return
            }

            loginURL := fmt.Sprintf("%s/auth/login", config.APIUrl)
            payload := map[string]string{
                "username": username,
                "password": password,
            }
            jsonPayload, err := json.Marshal(payload)
            handleErr(err, "Error marshalling JSON")

            resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(jsonPayload))
            handleErr(err, "Error making POST request")
            defer resp.Body.Close()

            if resp.StatusCode != http.StatusOK {
                fmt.Println("Error: Unable to authenticate, please check your credentials")
                return
            }

            body, err := ioutil.ReadAll(resp.Body)
            handleErr(err, "Error reading response body")

            var tokenResponse TokenResponse
            err = json.Unmarshal(body, &tokenResponse)
            handleErr(err, "Error unmarshalling response")

            err = saveTokens(tokenResponse)
            handleErr(err, "Error saving tokens")

            fmt.Println("Authentication successful!")
        },
    }
    rootCmd.AddCommand(authCmd)

    var napiVerCmd = &cobra.Command{
        Use:   "napi-ver",
        Short: "Show the API version",
        Run: func(cmd *cobra.Command, args []string) {
            config, err := loadConfig()
            if err != nil {
                fmt.Println("Configuration not found. Please run 'configure --api-url' first.")
                return
            }

            tokens, err := loadTokens()
            if err != nil {
                fmt.Println("Error loading tokens. Please run 'auth' command to authenticate.")
                return
            }

            versionURL := fmt.Sprintf("%s/version", config.APIUrl)
            req, err := http.NewRequest("GET", versionURL, nil)
            handleErr(err, "Error creating request")
            req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))

            client := &http.Client{}
            resp, err := client.Do(req)
            handleErr(err, "Error making GET request")
            defer resp.Body.Close()

            if resp.StatusCode != http.StatusOK {
                fmt.Println("Error: Unable to get API version")
                return
            }

            body, err := ioutil.ReadAll(resp.Body)
            handleErr(err, "Error reading response body")

            fmt.Printf("API Version: %s\n", body)
        },
    }
    rootCmd.AddCommand(napiVerCmd)

    var listCmd = &cobra.Command{
        Use:   "list",
        Short: "List services",
        Run: func(cmd *cobra.Command, args []string) {
            config, err := loadConfig()
            if err != nil {
                fmt.Println("Configuration not found. Please run 'configure --api-url' first.")
                return
            }

            tokens, err := loadTokens()
            if err != nil {
                fmt.Println("Error loading tokens. Please run 'auth' command to authenticate.")
                return
            }

            servicesResponse, err := components.FetchServices(config.APIUrl, tokens.AccessToken)
            if err != nil {
                fmt.Printf("Error fetching services: %v\n", err)
                return
            }

            fmt.Println("Services:")
            for _, service := range servicesResponse.Services {
                fmt.Printf("Unit: %s, Load: %s, Active: %s, Sub: %s, Description: %s\n", service.Unit, service.Load, service.Active, service.Sub, service.Description)
            }
        },
    }
    rootCmd.AddCommand(listCmd)

    var artCmd = &cobra.Command{
        Use:   "art",
        Short: "Show ASCII art",
        Run: func(cmd *cobra.Command, args []string) {
            printASCIIArt()
        },
    }
    rootCmd.AddCommand(artCmd)

    rootCmd.Version = VERSION
    rootCmd.Execute()
}
