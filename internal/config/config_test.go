package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Test with default values
	config := LoadConfig()

	if config.Database.URL == "" {
		t.Error("Database URL should not be empty")
	}

	if len(config.Chains) == 0 {
		t.Error("Chains configuration should not be empty")
	}

	if config.API.Port == "" {
		t.Error("API port should not be empty")
	}

	// Test that ethereum chain is configured
	if _, exists := config.Chains["ethereum"]; !exists {
		t.Error("Ethereum chain should be configured")
	}

	// Test that polygon chain is configured
	if _, exists := config.Chains["polygon"]; !exists {
		t.Error("Polygon chain should be configured")
	}
}

func TestGetEnvFunctions(t *testing.T) {
	// Test getEnv
	os.Setenv("TEST_STRING", "test_value")
	defer os.Unsetenv("TEST_STRING")

	if getEnv("TEST_STRING", "default") != "test_value" {
		t.Error("getEnv should return environment value")
	}

	if getEnv("NON_EXISTENT", "default") != "default" {
		t.Error("getEnv should return default value for non-existent key")
	}

	// Test getEnvAsInt
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	if getEnvAsInt("TEST_INT", 0) != 42 {
		t.Error("getEnvAsInt should return parsed integer")
	}

	if getEnvAsInt("NON_EXISTENT_INT", 10) != 10 {
		t.Error("getEnvAsInt should return default value for non-existent key")
	}

	// Test getEnvAsBool
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	if !getEnvAsBool("TEST_BOOL", false) {
		t.Error("getEnvAsBool should return parsed boolean")
	}

	if getEnvAsBool("NON_EXISTENT_BOOL", false) {
		t.Error("getEnvAsBool should return default value for non-existent key")
	}

	// Test getEnvAsDuration
	os.Setenv("TEST_DURATION", "30s")
	defer os.Unsetenv("TEST_DURATION")

	if getEnvAsDuration("TEST_DURATION", "10s") != 30*time.Second {
		t.Error("getEnvAsDuration should return parsed duration")
	}

	if getEnvAsDuration("NON_EXISTENT_DURATION", "5s") != 5*time.Second {
		t.Error("getEnvAsDuration should return default value for non-existent key")
	}
}