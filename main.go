package main

import (
    "bytes"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "bufio"
    "encoding/gob"
    "fmt"
    "io"
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

// getSavePath returns the path to save the encrypted token file
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

// hashKey hashes the key using SHA-256 to ensure it is of the correct length
func hashKey(key []byte) []byte {
    hash := sha256.Sum256(key)
    return hash[:]
}

// encrypt encrypts data using AES-256
func encrypt(data []byte, passphrase []byte) (string, error) {
    block, err := aes.NewCipher(hashKey(passphrase))
    handleErr(err, "Error creating AES cipher")

    gcm, err := cipher.NewGCM(block)
    handleErr(err, "Error creating GCM")

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    ciphertext := gcm.Seal(nonce, nonce, data, nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts data using AES-256
func decrypt(data string, passphrase []byte) ([]byte, error) {
    ciphertext, _ := base64.StdEncoding.DecodeString(data)

    block, err := aes.NewCipher(hashKey(passphrase))
    handleErr(err, "Error creating AES cipher")

    gcm, err := cipher.NewGCM(block)
    handleErr(err, "Error creating GCM")

    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return nil, fmt.Errorf("ciphertext too short")
    }

    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    return gcm.Open(nil, nonce, ciphertext, nil)
}

// saveTokens saves the tokens to a local file, encrypted
func saveTokens(tokens TokenResponse, passphrase []byte) error {
    var buffer bytes.Buffer
    encoder := gob.NewEncoder(&buffer)
    err := encoder.Encode(tokens)
    if err != nil {
        return fmt.Errorf("error encoding tokens: %v", err)
    }

    encryptedData, err := encrypt(buffer.Bytes(), passphrase)
    if err != nil {
        return fmt.Errorf("error encrypting tokens: %v", err)
    }

    savePath, err := getSavePath()
    if err != nil {
        return fmt.Errorf("error getting save path: %v", err)
    }

    err = ioutil.WriteFile(savePath, []byte(encryptedData), 0600)
    if err != nil {
        return fmt.Errorf("error writing tokens to file: %v", err)
    }

    return nil
}

// loadTokens loads the tokens from the local file, decrypted
func loadTokens(passphrase []byte) (TokenResponse, error) {
    var tokens TokenResponse

    savePath, err := getSavePath()
    if err != nil {
        return tokens, fmt.Errorf("error getting save path: %v", err)
    }

    encryptedData, err := ioutil.ReadFile(savePath)
    if err != nil {
        return tokens, fmt.Errorf("error reading tokens from file: %v", err)
    }

    data, err := decrypt(string(encryptedData), passphrase)
    if err != nil {
        return tokens, fmt.Errorf("error decrypting tokens: %v", err)
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

func main() {
    // Get the name of the executable
    exeName := filepath.Base(os.Args[0])

    // Create the root command with the executable name
    var rootCmd = &cobra.Command{
        Use: exeName,
        Short: "CLI tool for API interaction",
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

            passphrase := []byte(password)
            err = saveTokens(tokenResponse, passphrase)
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

            passphrase, err := promptInput("Enter passphrase: ", true)
            handleErr(err, "Error reading passphrase")

            tokens, err := loadTokens([]byte(passphrase))
            handleErr(err, "Error loading tokens\nPlease authenticate using the 'configure auth' command")

            versionResponse, err := components.GetVersion(tokens.AccessToken, config.APIUrl)
            handleErr(err, "Error getting version")

            fmt.Printf("API Version: %s\nUser: %s\n", versionResponse.Version, versionResponse.User)
        },
    }

    // Add commands to root and configure command
    cmdConfigure.AddCommand(cmdSetURL)
    cmdConfigure.AddCommand(cmdAuth)
    rootCmd.AddCommand(cmdConfigure, cmdVersion)
    
    // Execute the root command
    if err := rootCmd.Execute(); err != nil {
        log.Fatalf("Error executing command: %v", err)
    }
}
