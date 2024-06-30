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

// getSavePath returns the path to save the encrypted token file
func getSavePath() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err == nil {
        dataDir := filepath.Join(homeDir, ".nuc", "data")
        err = os.MkdirAll(dataDir, 0700)
        if err == nil {
            return filepath.Join(dataDir, "cli_tokens.json"), nil
        }
    }

    // If home directory is not accessible, use the binary directory
    exePath, err := os.Executable()
    if err != nil {
        return "", err
    }

    exeDir := filepath.Dir(exePath)
    dataDir := filepath.Join(exeDir, "data")
    err = os.MkdirAll(dataDir, 0700)
    if err != nil {
        return "", err
    }

    return filepath.Join(dataDir, "cli_tokens.json"), nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err == nil {
        dataDir := filepath.Join(homeDir, ".nuc", "data")
        err = os.MkdirAll(dataDir, 0700)
        if err == nil {
            return filepath.Join(dataDir, "config.json"), nil
        }
    }

    // If home directory is not accessible, use the binary directory
    exePath, err := os.Executable()
    if err != nil {
        return "", err
    }

    exeDir := filepath.Dir(exePath)
    dataDir := filepath.Join(exeDir, "data")
    err = os.MkdirAll(dataDir, 0700)
    if err != nil {
        return "", err
    }

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
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

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
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return nil, fmt.Errorf("ciphertext too short")
    }

    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    return gcm.Open(nil, nonce, ciphertext, nil)
}

// saveTokens saves the tokens to a local file, encrypted
func saveTokens(tokens TokenResponse, passphrase []byte) error {
    data, err := json.Marshal(tokens)
    if err != nil {
        return fmt.Errorf("error marshalling tokens: %v", err)
    }

    encryptedData, err := encrypt(data, passphrase)
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

    err = json.Unmarshal(data, &tokens)
    if err != nil {
        return tokens, fmt.Errorf("error unmarshalling tokens: %v", err)
    }

    return tokens, nil
}

// saveConfig saves the API URL to the configuration file
func saveConfig(apiUrl string) error {
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
            fmt.Print("•")
        }
    }

    return string(input), nil
}

// promptCredentials prompts the user for username, password, and encryption key if not provided as arguments
func promptCredentials() (string, string, string) {
    username, _ := promptInput("Username: ", false)
    password, _ := promptInput("Password: ", true)
    encKey, _ := promptInput("Encryption Key: ", true)

    return strings.TrimSpace(username), strings.TrimSpace(password), strings.TrimSpace(encKey)
}

// login performs the login request and saves the tokens
func login(username, password, encKey string, apiUrl string) {
    url := apiUrl + "/login"
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

    if err := saveTokens(tokenResponse, []byte(encKey)); err != nil {
        log.Fatalf("Error saving tokens: %v", err)
    }

    fmt.Println("Login successful")
}

func main() {
    var rootCmd = &cobra.Command{
        Use:   "nuc",
        Short: "Nest User Control CLI",
        Long:  "A CLI tool to control your Nest server provided by HACK CLUB <3",
    }

    // The configure command sets the API URL and saves it to the configuration file.
    var configureCmd = &cobra.Command{
        Use:   "configure",
        Short: "Configure the CLI tool",
        Long:  "Configure the CLI tool by setting the API URL.",
        Run: func(cmd *cobra.Command, args []string) {
            apiUrl, _ := cmd.Flags().GetString("api-url")
            if apiUrl == "" {
                fmt.Print("URL: ")
                reader := bufio.NewReader(os.Stdin)
                input, err := reader.ReadString('\n')
                if err != nil {
                    log.Fatalf("Error reading input: %v", err)
                }
                apiUrl = strings.TrimSpace(input)
            }

            if err := saveConfig(apiUrl); err != nil {
                log.Fatalf("Error saving config: %v", err)
            }

            fmt.Println("Configuration saved successfully.")
        },
    }

    configureCmd.Flags().String("api-url", "", "The API URL to configure")

    // The auth command authenticates the user to the API using provided or interactive credentials.
    var authCmd = &cobra.Command{
        Use:   "auth [username] [password] [encryption key]",
        Short: "Authenticate to the API",
        Long:  "Authenticate to the API using a username and password. You can provide the credentials and encryption key as arguments or input them interactively.",
        Run: func(cmd *cobra.Command, args []string) {
            var username, password, encKey string

            // Load the API URL from the configuration file
            config, err := loadConfig()
            if err != nil {
                log.Fatalf("Error loading config: %v", err)
            }

            // Check if the API URL is configured
            if config.APIUrl == "" {
                log.Fatalf("API URL is not configured. Please configure it using the 'configure --api-url' command.")
            }

            if len(args) == 3 {
                username = args[0]
                password = args[1]
                encKey = args[2]
            } else {
                username, password, encKey = promptCredentials()
            }

            login(username, password, encKey, config.APIUrl)
        },
    }

    // The version command shows the version of the API using the configured API URL.
    var versionCmd = &cobra.Command{
        Use:   "napi-ver",
        Short: "Show the version of the API",
        Long:  "Show the version of the API.",
        Run: func(cmd *cobra.Command, args []string) {
            config, err := loadConfig()
            if err != nil {
                log.Fatalf("Error loading config: %v", err)
            }

            if config.APIUrl == "" {
                log.Fatalf("API URL is not configured. Please configure it using the 'configure --api-url' command.")
            }

            // versionResponse, err := components.GetVersion(config.APIUrl)
            versionResponse, err := components.GetVersion()
            if err != nil {
                log.Fatalf("Error fetching version: %v", err)
            }
            fmt.Printf("API Version: %s\nUser: %s\n", versionResponse.Version, versionResponse.User)
        },
    }

    rootCmd.AddCommand(configureCmd)
    rootCmd.AddCommand(authCmd)
    rootCmd.AddCommand(versionCmd)

    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
