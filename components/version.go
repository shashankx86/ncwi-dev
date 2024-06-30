package components

import (
    "encoding/json"
    "fmt"
    "net/http"
)

// VersionResponse represents the structure of the version response from the API
type VersionResponse struct {
    Version string `json:"version"`
    User    string `json:"user"`
}

// GetVersion fetches the version from the API if the user is authenticated
func GetVersion(token string, apiUrl string) (*VersionResponse, error) {
    req, err := http.NewRequest("GET", apiUrl+"/version", nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer "+token)

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
