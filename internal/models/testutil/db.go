package testutil

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *sqlx.DB {
	// Use environment variable or default to test database
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/nexus_bridge_test?sslmode=disable"
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		t.Skipf("Skipping database tests: %v", err)
	}

	// Clean up any existing test data
	cleanupTestData(t, db)

	return db
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(t *testing.T, db *sqlx.DB) {
	cleanupTestData(t, db)
	db.Close()
}

// cleanupTestData removes all test data from tables
func cleanupTestData(t *testing.T, db *sqlx.DB) {
	tables := []string{
		"signatures",
		"transfers", 
		"supported_tokens",
		"relayer_config",
		"audit_log",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: failed to clean table %s: %v", table, err)
		}
	}
}

// CreateTestTransfer creates a test transfer for use in tests
func CreateTestTransfer() map[string]interface{} {
	return map[string]interface{}{
		"id":                "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		"source_chain":      1,
		"destination_chain": 137,
		"token":            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		"amount":           "1000000000000000000",
		"sender":           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		"recipient":        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
		"status":           "pending",
		"confirmations":    0,
	}
}

// CreateTestSignature creates a test signature for use in tests
func CreateTestSignature() map[string]interface{} {
	return map[string]interface{}{
		"relayer_address": "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		"signature":       []byte("test_signature_data"),
	}
}

// CreateTestSupportedToken creates a test supported token for use in tests
func CreateTestSupportedToken() map[string]interface{} {
	return map[string]interface{}{
		"chain_id":      1,
		"token_address": "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		"name":          "Test Token",
		"symbol":        "TEST",
		"decimals":      18,
		"is_native":     false,
		"enabled":       true,
	}
}