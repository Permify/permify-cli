package config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCoreConfigPersistence(t *testing.T) {
	tmpFile := "test_config.yaml"
	defer os.Remove(tmpFile)

	expected := CoreConfig{
		PermifyURL:  "https://localhost:8888",
		Tenant:      "test-tenant",
		Token:       "test-token-123",
		CertPath:    "/path/to/cert",
		CertKeyPath: "/path/to/key",
	}

	// Test Writing
	configs := make(map[string]CoreConfig)
	configs["default"] = expected
	data, _ := yaml.Marshal(configs)
	_ = os.WriteFile(tmpFile, data, 0644)

	// Test Reading
	var actualConfigs map[string]CoreConfig
	readData, _ := os.ReadFile(tmpFile)
	yaml.Unmarshal(readData, &actualConfigs)

	actual := actualConfigs["default"]

	if actual.Token != expected.Token {
		t.Errorf("Expected Token %s, got %s", expected.Token, actual.Token)
	}
	if actual.CertPath != expected.CertPath {
		t.Errorf("Expected CertPath %s, got %s", expected.CertPath, actual.CertPath)
	}
}
