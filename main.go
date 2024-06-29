package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "bufio"
    "strings"

    "github.com/spf13/cobra"
    "nuc/components"
)

var (
    configFilePath = filepath.Join(os.Getenv("HOME"), ".cli_tokens.json")
)

// TokenResponse represents the structure of the response from the login endpoint
type TokenResponse struct {
    Message           string `json:"message"`
    AccessToken       string `json:"access_token"`
    RefreshToken      string `json:"refresh_token"`
    RefreshExpiration int64  `json:"refresh_expiration"`
}

// saveTokens saves the tokens to a local file
func saveTokens(tokens TokenResponse) error {
    data, err := json.Marshal(tokens)
    if err != nil {
        return err
    }

    log.Printf("Saving tokens: %s", data)
    err = ioutil.WriteFile(configFilePath, data, 0600)
    if err != nil {
        return err
    }

    return nil
}


// promptCredentials prompts the user for username and password if not provided as arguments
func promptCredentials() (string, string) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Username: ")
    username, _ := reader.ReadString('\n')
    username = strings.TrimSpace(username)

    fmt.Print("Password: ")
    password, _ := reader.ReadString('\n')
    password = strings.TrimSpace(password)

    return username, password
}

// login performs the login request and saves the tokens
func login(username, password string) {
    url := "https://napi.theaddicts.hackclub.app/login"
    credentials := map[string]string{
        "username": username,
        "password": password,
    }

    jsonData, err := json.Marshal(credentials)
    if err != nil {
        log.Fatalf("Error encoding credentials: %v", err)
    }

    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Fatalf("Error making request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Fatalf("Login failed: %v", resp.Status)
    }

    var tokenResponse TokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
        log.Fatalf("Error decoding response: %v", err)
    }

    if err := saveTokens(tokenResponse); err != nil {
        log.Fatalf("Error saving tokens: %v", err)
    }

    fmt.Println("Login successful")
}

func main() {
    var rootCmd = &cobra.Command{
        Use:   "nuc",
        Short: "Nest User Control CLI",
        Long:  "A CLI tool control your Nest server provided by HACK CLUB <3",
    }

    var authCmd = &cobra.Command{
        Use:   "auth [username] [password]",
        Short: "Authenticate to the napi",
        Long:  "Authenticate to the API using a username and password. You can provide the credentials as arguments or input them interactively.",
        Run: func(cmd *cobra.Command, args []string) {
            var username, password string

            if len(args) == 2 {
                username = args[0]
                password = args[1]
            } else {
                username, password = promptCredentials()
            }

            login(username, password)
        },
    }

    var versionCmd = &cobra.Command{
        Use:   "napi-ver",
        Short: "Show the version of the API",
        Long:  "Show the version of the API.",
        Run: func(cmd *cobra.Command, args []string) {
            versionResponse, err := components.GetVersion()
            if err != nil {
                log.Fatalf("Error fetching version: %v", err)
            }
            fmt.Printf("API Version: %s\nUser: %s\n", versionResponse.Version, versionResponse.User)
        },
    }

    rootCmd.AddCommand(authCmd)
    rootCmd.AddCommand(versionCmd)

    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
