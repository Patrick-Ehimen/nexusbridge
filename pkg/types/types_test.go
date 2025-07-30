package types

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"
)

func TestBigInt_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		bigInt   *BigInt
		expected string
	}{
		{
			name:     "positive number",
			bigInt:   NewBigInt(big.NewInt(1000000000000000000)),
			expected: `"1000000000000000000"`,
		},
		{
			name:     "zero",
			bigInt:   NewBigInt(big.NewInt(0)),
			expected: `"0"`,
		},
		{
			name:     "nil big.Int",
			bigInt:   NewBigInt(nil),
			expected: `"0"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.bigInt)
			if err != nil {
				t.Fatalf("Failed to marshal BigInt: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestBigInt_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		expected  *big.Int
		expectErr bool
	}{
		{
			name:     "positive number",
			jsonData: `"1000000000000000000"`,
			expected: big.NewInt(1000000000000000000),
		},
		{
			name:     "zero",
			jsonData: `"0"`,
			expected: big.NewInt(0),
		},
		{
			name:      "invalid number",
			jsonData:  `"invalid"`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bigInt BigInt
			err := json.Unmarshal([]byte(tt.jsonData), &bigInt)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to unmarshal BigInt: %v", err)
			}

			if bigInt.Cmp(tt.expected) != 0 {
				t.Errorf("Expected %s, got %s", tt.expected.String(), bigInt.String())
			}
		})
	}
}

func TestChainID_String(t *testing.T) {
	tests := []struct {
		chainID  ChainID
		expected string
	}{
		{ChainEthereum, "ethereum"},
		{ChainPolygon, "polygon"},
		{ChainCosmos, "cosmos"},
		{ChainID(999), "unknown-999"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.chainID.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTransfer_Validate(t *testing.T) {
	validTransfer := &Transfer{
		ID:               "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		SourceChain:      ChainEthereum,
		DestinationChain: ChainPolygon,
		Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Amount:           NewBigInt(big.NewInt(1000000000000000000)),
		Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
		Status:           StatusPending,
	}

	// Test valid transfer
	err := validTransfer.Validate()
	if err != nil {
		t.Errorf("Valid transfer should not return error: %v", err)
	}

	// Test invalid transfers
	tests := []struct {
		name     string
		modify   func(*Transfer)
		expectErr bool
	}{
		{
			name: "empty ID",
			modify: func(t *Transfer) {
				t.ID = ""
			},
			expectErr: true,
		},
		{
			name: "same source and destination",
			modify: func(t *Transfer) {
				t.DestinationChain = t.SourceChain
			},
			expectErr: true,
		},
		{
			name: "zero amount",
			modify: func(t *Transfer) {
				t.Amount = NewBigInt(big.NewInt(0))
			},
			expectErr: true,
		},
		{
			name: "negative amount",
			modify: func(t *Transfer) {
				t.Amount = NewBigInt(big.NewInt(-1))
			},
			expectErr: true,
		},
		{
			name: "empty token",
			modify: func(t *Transfer) {
				t.Token = ""
			},
			expectErr: true,
		},
		{
			name: "empty sender",
			modify: func(t *Transfer) {
				t.Sender = ""
			},
			expectErr: true,
		},
		{
			name: "empty recipient",
			modify: func(t *Transfer) {
				t.Recipient = ""
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the valid transfer
			transfer := *validTransfer
			transfer.Amount = NewBigInt(new(big.Int).Set(validTransfer.Amount.Int))
			
			// Apply modification
			tt.modify(&transfer)
			
			err := transfer.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestChainConfig_Validate(t *testing.T) {
	validConfig := &ChainConfig{
		ChainID:               ChainEthereum,
		Name:                  "Ethereum",
		Type:                  ChainTypeEthereum,
		RPC:                   "https://mainnet.infura.io/v3/test",
		BridgeContract:        "0x1234567890123456789012345678901234567890",
		RequiredConfirmations: 12,
		BlockTime:             12 * time.Second,
		GasLimit:              21000,
		Enabled:               true,
	}

	// Test valid config
	err := validConfig.Validate()
	if err != nil {
		t.Errorf("Valid config should not return error: %v", err)
	}

	// Test invalid configs
	tests := []struct {
		name     string
		modify   func(*ChainConfig)
		expectErr bool
	}{
		{
			name: "empty name",
			modify: func(c *ChainConfig) {
				c.Name = ""
			},
			expectErr: true,
		},
		{
			name: "empty RPC",
			modify: func(c *ChainConfig) {
				c.RPC = ""
			},
			expectErr: true,
		},
		{
			name: "empty bridge contract",
			modify: func(c *ChainConfig) {
				c.BridgeContract = ""
			},
			expectErr: true,
		},
		{
			name: "zero confirmations",
			modify: func(c *ChainConfig) {
				c.RequiredConfirmations = 0
			},
			expectErr: true,
		},
		{
			name: "zero block time",
			modify: func(c *ChainConfig) {
				c.BlockTime = 0
			},
			expectErr: true,
		},
		{
			name: "negative block time",
			modify: func(c *ChainConfig) {
				c.BlockTime = -1 * time.Second
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the valid config
			config := *validConfig
			
			// Apply modification
			tt.modify(&config)
			
			err := config.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestEvent_Serialization(t *testing.T) {
	event := Event{
		ID:          "event-123",
		Type:        EventTypeLock,
		ChainID:     ChainEthereum,
		TxHash:      "0xabcdef1234567890",
		BlockNumber: 12345,
		TransferID:  "transfer-456",
		Transfer: Transfer{
			ID:               "transfer-456",
			SourceChain:      ChainEthereum,
			DestinationChain: ChainPolygon,
			Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
			Amount:           NewBigInt(big.NewInt(1000000000000000000)),
			Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
			Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
			Status:           StatusPending,
		},
		Timestamp: time.Now(),
	}

	// Test JSON serialization
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Test JSON deserialization
	var unmarshaled Event
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	// Verify key fields
	if unmarshaled.ID != event.ID {
		t.Errorf("Expected ID %s, got %s", event.ID, unmarshaled.ID)
	}
	if unmarshaled.Type != event.Type {
		t.Errorf("Expected Type %s, got %s", event.Type, unmarshaled.Type)
	}
	if unmarshaled.Transfer.Amount.Cmp(event.Transfer.Amount.Int) != 0 {
		t.Errorf("Expected Amount %s, got %s", event.Transfer.Amount.String(), unmarshaled.Transfer.Amount.String())
	}
}