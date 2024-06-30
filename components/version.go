package components

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
)

// VersionResponse represents the structure of the version response from the API
type VersionResponse struct {
    Version string `json:"version"`
    User    string `json:"user"`
}

// LoadTokens loads the tokens from the local file
func LoadTokens() (map[string]interface{}, error) {  // Change map type to interface{} to accommodate different value types
    configFilePath := filepath.Join(os.Getenv("HOME"), ".cli_tokens.json")
    log.Printf("Loading tokens from %s", configFilePath)
    data, err := ioutil.ReadFile(configFilePath)
    if err != nil {
        log.Printf("Error reading token file: %v", err)
        return nil, err
    }

    var tokens map[string]interface{}  // Use interface{} to handle mixed types
    if err := json.Unmarshal(data, &tokens); err != nil {
        log.Printf("Error unmarshalling token file: %v", err)
        return nil, err
    }

    log.Printf("Tokens loaded: %v", tokens)
    return tokens, nil
}

// GetVersion fetches the version from the API if the user is authenticated
func GetVersion() (*VersionResponse, error) {
    tokens, err := LoadTokens()
    if err != nil {
        return nil, fmt.Errorf("please log in using the auth command")
    }

    accessToken, ok := tokens["access_token"].(string)  // Assert to string type
    if !ok {
        return nil, fmt.Errorf("please log in using the auth command")
    }

    req, err := http.NewRequest("GET", "https://napi.theaddicts.hackclub.app/version", nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer "+accessToken)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to fetch version: %v", resp.Status)
    }

    var versionResponse VersionResponse
    if err := json.NewDecoder(resp.Body).Decode(&versionResponse); err != nil {
        return nil, err
    }

    return &versionResponse, nil
}
