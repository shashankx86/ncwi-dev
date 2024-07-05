package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"nuc/utils"
)

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	Expiration  int64  `json:"expiration"`
}

func Authenticate(username, password, apiUrl string) (*AuthResponse, error) {
	authData := map[string]string{
		"username": username,
		"password": password,
	}
	authJSON, err := json.Marshal(authData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", apiUrl+"/api/authenticate", strings.NewReader(string(authJSON)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed: %s", resp.Status)
	}

	var authResponse AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	if err != nil {
		return nil, err
	}

	return &authResponse, nil
}

func SaveAuthResponse(authResponse *AuthResponse) error {
	tokenData, err := json.Marshal(authResponse)
	if err != nil {
		return err
	}

	tokenFilePath, err := utils.GetTokenFilePath()
	if err != nil {
		return err
	}

	return utils.SaveToFile(tokenFilePath, tokenData)
}
