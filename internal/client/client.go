package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	permify "github.com/Permify/permify-go/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	credentialsFileName = "credentials"
	permifyDirName      = ".permify"
)

var logger *zap.Logger

type Credentials struct {
	Endpoint    string `json:"endpoint"`
	Token       string `json:"token"`
	CertPath    string `json:"certPath"`
	CertKeyPath string `json:"certKeyPath"`
}

func init() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = customTimeEncoder
	logger, _ = config.Build()
	defer logger.Sync()
	logger.Info("Client initialized")
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func New(endpoint string) (*permify.Client, error) {
	creds, err := retrieveCredentials()
	if err != nil {
		if errors.Is(err, errCredentialsNotFound) {
			// Credentials not found, get from user?
			if err := getCredentialsFromUser(); err != nil {
				return nil, fmt.Errorf("failed to create credentials: %v", err)
			}
			return New(endpoint)
		}
		logger.Error("Failed to retrieve credentials from file", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve credentials: %v", err)
	}

	if err := validateCredentials(creds); err != nil {
		logger.Warn("Invalid credentials", zap.Any("invalidCredentials", creds), zap.Error(err))
		return nil, fmt.Errorf("invalid credentials: %v", err)
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: creds.Endpoint,
			// Other necessary configurations
		},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("Failed to create Permify client", zap.Error(err))
		return nil, fmt.Errorf("failed to create Permify client: %v", err)
	}

	logger.Info("Permify client created successfully")

	return client, nil
}

var errCredentialsNotFound = errors.New("credentials file not found")

func retrieveCredentials() (*Credentials, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get user home directory", zap.Error(err))
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	credentialsPath := filepath.Join(homeDir, permifyDirName, credentialsFileName)

	fileContent, err := os.ReadFile(credentialsPath)
	if os.IsNotExist(err) {
		return nil, errCredentialsNotFound
	} else if err != nil {
		logger.Error("Failed to read credentials file", zap.String("file", credentialsPath), zap.Error(err))
		return nil, fmt.Errorf("failed to read credentials file '%s': %v", credentialsPath, err)
	}

	var creds Credentials
	if err := json.Unmarshal(fileContent, &creds); err != nil {
		logger.Error("Failed to unmarshal credentials", zap.String("file", credentialsPath), zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal credentials: %v", err)
	}

	return &creds, nil
}

func getCredentialsFromUser() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Credentials file not found. Do you want to create one? (y/n): ")
	answer, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %v", err)
	}

	answer = strings.TrimSpace(answer)
	if answer != "y" && answer != "yes" {
		return errors.New("user chose not to create credentials")
	}

	creds := Credentials{}
	fmt.Print("Enter Permify endpoint: ")
	creds.Endpoint, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %v", err)
	}
	creds.Endpoint = strings.TrimSpace(creds.Endpoint)

	fmt.Print("Enter Permify token: ")
	creds.Token, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %v", err)
	}
	creds.Token = strings.TrimSpace(creds.Token)

	fmt.Print("Enter path to the certificate file: ")
	creds.CertPath, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %v", err)
	}
	creds.CertPath = strings.TrimSpace(creds.CertPath)

	fmt.Print("Enter path to the private key file: ")
	creds.CertKeyPath, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %v", err)
	}
	creds.CertKeyPath = strings.TrimSpace(creds.CertKeyPath)
	// Create .permify directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	permifyDirPath := filepath.Join(homeDir, permifyDirName)
	if _, err := os.Stat(permifyDirPath); os.IsNotExist(err) {
		err := os.Mkdir(permifyDirPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create .permify directory: %v", err)
		}
	}

	credentialsPath := filepath.Join(permifyDirPath, credentialsFileName)
	file, err := os.Create(credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to create credentials file: %v", err)
	}
	defer file.Close()

	credsJSON, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials to JSON: %v", err)
	}

	_, err = file.Write(credsJSON)
	if err != nil {
		return fmt.Errorf("failed to write credentials to file: %v", err)
	}

	fmt.Printf("Credentials saved to %s\n", credentialsPath)
	return nil
}

func validateCredentials(creds *Credentials) error {
	if creds.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if creds.Token == "" {
		return errors.New("token is required")
	}
	if creds.CertPath == "" {
		return errors.New("certPath is required")
	}
	if creds.CertKeyPath == "" {
		return errors.New("certKeyPath is required")
	}

	return nil
}
