package components

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Service represents a system service
type Service struct {
	Unit        string `json:"UNIT"`
	Load        string `json:"LOAD"`
	Active      string `json:"ACTIVE"`
	Sub         string `json:"SUB"`
	Description string `json:"DESCRIPTION"`
}

// ServicesResponse represents the response structure from the API
type ServicesResponse struct {
	Services []Service `json:"services"`
}

// FetchServices fetches the services from the API and returns them
func FetchServices(apiUrl string, accessToken string) ([]Service, error) {
	fullUrl := fmt.Sprintf("%s/io/system/services", apiUrl)
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var servicesResponse ServicesResponse
	err = json.Unmarshal(body, &servicesResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return servicesResponse.Services, nil
}

// PrintServices prints the fetched services
func PrintServices(services []Service) {
	fmt.Println("Services:")
	for _, service := range services {
		fmt.Printf("Unit: %s, Load: %s, Active: %s, Sub: %s, Description: %s\n",
			service.Unit, service.Load, service.Active, service.Sub, service.Description)
	}
}
