package components

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
)

// Service represents the structure of a service in the response
type Service struct {
    Unit        string `json:"UNIT"`
    Load        string `json:"LOAD"`
    Active      string `json:"ACTIVE"`
    Sub         string `json:"SUB"`
    Description string `json:"DESCRIPTION"`
}

// ServicesResponse represents the structure of the response containing services
type ServicesResponse struct {
    Services []Service `json:"services"`
}

// FetchServices fetches the services data from the API and returns a ServicesResponse
func FetchServices(apiURL string, token string) (*ServicesResponse, error) {
    // Append the endpoint to the base API URL
    fullURL := fmt.Sprintf("%s/io/system", apiURL)
    
    // Create a new request
    req, err := http.NewRequest("GET", fullURL, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }

    // Set the Authorization header
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
    
    // Send the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error sending request: %v", err)
    }
    defer resp.Body.Close()
    
    // Read the response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading response body: %v", err)
    }
    
    // Parse the response body into a ServicesResponse
    var servicesResponse ServicesResponse
    err = json.Unmarshal(body, &servicesResponse)
    if err != nil {
        return nil, fmt.Errorf("error unmarshalling response: %v", err)
    }
    
    return &servicesResponse, nil
}
