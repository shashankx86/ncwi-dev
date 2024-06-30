package main

import (
    "bytes"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "bufio"
    "strings"
    "path/filepath"

    "github.com/spf13/cobra"
    "nuc/components"
)

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

// promptCredentials prompts the user for username, password, and encryption key if not provided as arguments
func promptCredentials() (string, string, string) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Username: ")
    username, _ := reader.ReadString('\n')
    username = strings.TrimSpace(username)

    fmt.Print("Password: ")
    password, _ := reader.ReadString('\n')
    password = strings.TrimSpace(password)

    fmt.Print("Encryption Key: ")
    encKey, _ := reader.ReadString('\n')
    encKey = strings.TrimSpace(encKey)

    return username, password, encKey
}

// login performs the login request and saves the tokens
func login(username, password, encKey string) {
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

    if err := saveTokens(tokenResponse, []byte(encKey)); err != nil {
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
        Use:   "auth [username] [password] [encryption key]",
        Short: "Authenticate to the napi",
        Long:  "Authenticate to the API using a username and password. You can provide the credentials and encryption key as arguments or input them interactively.",
        Run: func(cmd *cobra.Command, args []string) {
            var username, password, encKey string

            if len(args) == 3 {
                username = args[0]
                password = args[1]
                encKey = args[2]
            } else {
                username, password, encKey = promptCredentials()
            }

            login(username, password, encKey)
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
