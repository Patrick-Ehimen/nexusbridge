package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	Chains   map[string]ChainConfig
	API      APIConfig
	Relayer  RelayerConfig
	Logging  LoggingConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// ChainConfig holds blockchain-specific configuration
type ChainConfig struct {
	ChainID               uint64
	Name                  string
	Type                  string // ethereum, cosmos
	RPCURL                string
	WSSURL                string
	BridgeContract        string
	RequiredConfirmations uint64
	BlockTime             time.Duration
	GasLimit              uint64
	Enabled               bool
}

// APIConfig holds API server configuration
type APIConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// RelayerConfig holds relayer-specific configuration
type RelayerConfig struct {
	Port              string
	PrivateKey        string
	SignatureThreshold uint64
	RelayerCount      uint64
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://nexus:nexus_password@localhost:5432/nexus_bridge?sslmode=disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", "5m"),
		},
		Chains: map[string]ChainConfig{
			"ethereum": {
				ChainID:               1,
				Name:                  "Ethereum Mainnet",
				Type:                  "ethereum",
				RPCURL:                getEnv("ETHEREUM_RPC_URL", "http://localhost:8545"),
				WSSURL:                getEnv("ETHEREUM_WSS_URL", "ws://localhost:8545"),
				BridgeContract:        getEnv("ETHEREUM_BRIDGE_CONTRACT", ""),
				RequiredConfirmations: uint64(getEnvAsInt("ETHEREUM_CONFIRMATIONS", 12)),
				BlockTime:             getEnvAsDuration("ETHEREUM_BLOCK_TIME", "12s"),
				GasLimit:              uint64(getEnvAsInt("ETHEREUM_GAS_LIMIT", 21000)),
				Enabled:               getEnvAsBool("ETHEREUM_ENABLED", true),
			},
			"polygon": {
				ChainID:               137,
				Name:                  "Polygon",
				Type:                  "ethereum",
				RPCURL:                getEnv("POLYGON_RPC_URL", "http://localhost:8546"),
				WSSURL:                getEnv("POLYGON_WSS_URL", "ws://localhost:8546"),
				BridgeContract:        getEnv("POLYGON_BRIDGE_CONTRACT", ""),
				RequiredConfirmations: uint64(getEnvAsInt("POLYGON_CONFIRMATIONS", 20)),
				BlockTime:             getEnvAsDuration("POLYGON_BLOCK_TIME", "2s"),
				GasLimit:              uint64(getEnvAsInt("POLYGON_GAS_LIMIT", 21000)),
				Enabled:               getEnvAsBool("POLYGON_ENABLED", true),
			},
			"hardhat": {
				ChainID:               31337,
				Name:                  "Hardhat Local",
				Type:                  "ethereum",
				RPCURL:                getEnv("HARDHAT_RPC_URL", "http://localhost:8545"),
				WSSURL:                getEnv("HARDHAT_WSS_URL", "ws://localhost:8545"),
				BridgeContract:        getEnv("HARDHAT_BRIDGE_CONTRACT", ""),
				RequiredConfirmations: uint64(getEnvAsInt("HARDHAT_CONFIRMATIONS", 1)),
				BlockTime:             getEnvAsDuration("HARDHAT_BLOCK_TIME", "1s"),
				GasLimit:              uint64(getEnvAsInt("HARDHAT_GAS_LIMIT", 21000)),
				Enabled:               getEnvAsBool("HARDHAT_ENABLED", true),
			},
		},
		API: APIConfig{
			Port:         getEnv("API_PORT", "8080"),
			ReadTimeout:  getEnvAsDuration("API_READ_TIMEOUT", "10s"),
			WriteTimeout: getEnvAsDuration("API_WRITE_TIMEOUT", "10s"),
			IdleTimeout:  getEnvAsDuration("API_IDLE_TIMEOUT", "60s"),
		},
		Relayer: RelayerConfig{
			Port:               getEnv("RELAYER_PORT", "8081"),
			PrivateKey:         getEnv("RELAYER_PRIVATE_KEY", ""),
			SignatureThreshold: uint64(getEnvAsInt("SIGNATURE_THRESHOLD", 2)),
			RelayerCount:       uint64(getEnvAsInt("RELAYER_COUNT", 3)),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return time.Minute
}