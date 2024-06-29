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

    "github.com/spf13/cobra"
)

var (
    username string
    password string
    configFilePath = filepath.Join(os.Getenv("HOME"), ".cli_tokens.json")
)

// TokenResponse represents the structure of the response from the login endpoint
type TokenResponse struct {
    Message          string `json:"message"`
    AccessToken      string `json:"access_token"`
    RefreshToken     string `json:"refresh_token"`
    RefreshExpiration int64  `json:"refresh_expiration"`
}

// saveTokens saves the tokens to a local file
func saveTokens(tokens TokenResponse) error {
    data, err := json.Marshal(tokens)
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(configFilePath, data, 0600)
    if err != nil {
        return err
    }

    return nil
}

// login performs the login request and saves the tokens
func login(cmd *cobra.Command, args []string) {
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
    var rootCmd = &cobra.Command{Use: "cli"}

    var loginCmd = &cobra.Command{
        Use:   "login",
        Short: "Log in to the API",
        Run:   login,
    }

    loginCmd.Flags().StringVarP(&username, "username", "u", "", "Username")
    loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password")
    loginCmd.MarkFlagRequired("username")
    loginCmd.MarkFlagRequired("password")

    rootCmd.AddCommand(loginCmd)
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
