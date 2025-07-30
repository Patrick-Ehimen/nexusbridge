package models

import (
	"context"
	"testing"

	"nexus-bridge/internal/models/testutil"
	"nexus-bridge/pkg/types"
)

func TestSupportedTokenRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewSupportedTokenRepository(db)

	token := &types.SupportedToken{
		ChainID:      types.ChainEthereum,
		TokenAddress: "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Name:         "Test Token",
		Symbol:       "TEST",
		Decimals:     18,
		IsNative:     false,
		Enabled:      true,
	}

	err := repo.Create(context.Background(), token)
	if err != nil {
		t.Fatalf("Failed to create supported token: %v", err)
	}

	if token.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	// Verify the token was created
	retrieved, err := repo.GetByChainAndAddress(context.Background(), token.ChainID, token.TokenAddress)
	if err != nil {
		t.Fatalf("Failed to retrieve token: %v", err)
	}

	if retrieved.Name != token.Name {
		t.Errorf("Expected name %s, got %s", token.Name, retrieved.Name)
	}
	if retrieved.Symbol != token.Symbol {
		t.Errorf("Expected symbol %s, got %s", token.Symbol, retrieved.Symbol)
	}
	if retrieved.Decimals != token.Decimals {
		t.Errorf("Expected decimals %d, got %d", token.Decimals, retrieved.Decimals)
	}
}

func TestSupportedTokenRepository_CreateInvalidToken(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewSupportedTokenRepository(db)

	tests := []struct {
		name      string
		token     *types.SupportedToken
		expectErr bool
	}{
		{
			name: "zero chain ID",
			token: &types.SupportedToken{
				ChainID:      0,
				TokenAddress: "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
				Name:         "Test Token",
				Symbol:       "TEST",
				Decimals:     18,
			},
			expectErr: true,
		},
		{
			name: "empty token address",
			token: &types.SupportedToken{
				ChainID:      types.ChainEthereum,
				TokenAddress: "",
				Name:         "Test Token",
				Symbol:       "TEST",
				Decimals:     18,
			},
			expectErr: true,
		},
		{
			name: "empty name",
			token: &types.SupportedToken{
				ChainID:      types.ChainEthereum,
				TokenAddress: "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
				Name:         "",
				Symbol:       "TEST",
				Decimals:     18,
			},
			expectErr: true,
		},
		{
			name: "empty symbol",
			token: &types.SupportedToken{
				ChainID:      types.ChainEthereum,
				TokenAddress: "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
				Name:         "Test Token",
				Symbol:       "",
				Decimals:     18,
			},
			expectErr: true,
		},
		{
			name: "too many decimals",
			token: &types.SupportedToken{
				ChainID:      types.ChainEthereum,
				TokenAddress: "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
				Name:         "Test Token",
				Symbol:       "TEST",
				Decimals:     19,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(context.Background(), tt.token)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestSupportedTokenRepository_GetByChain(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewSupportedTokenRepository(db)

	// Create tokens for different chains
	tokens := []*types.SupportedToken{
		{
			ChainID:      types.ChainEthereum,
			TokenAddress: "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
			Name:         "Ethereum Token 1",
			Symbol:       "ETH1",
			Decimals:     18,
			Enabled:      true,
		},
		{
			ChainID:      types.ChainEthereum,
			TokenAddress: "0xB0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
			Name:         "Ethereum Token 2",
			Symbol:       "ETH2",
			Decimals:     6,
			Enabled:      true,
		},
		{
			ChainID:      types.ChainPolygon,
			TokenAddress: "0xC0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
			Name:         "Polygon Token",
			Symbol:       "POLY",
			Decimals:     18,
			Enabled:      true,
		},
		{
			ChainID:      types.ChainEthereum,
			TokenAddress: "0xD0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
			Name:         "Disabled Token",
			Symbol:       "DIS",
			Decimals:     18,
			Enabled:      false, // This should not be returned
		},
	}

	for _, token := range tokens {
		err := repo.Create(context.Background(), token)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}
	}

	// Get Ethereum tokens (should return 2 enabled tokens)
	ethTokens, err := repo.GetByChain(context.Background(), types.ChainEthereum)
	if err != nil {
		t.Fatalf("Failed to get Ethereum tokens: %v", err)
	}

	if len(ethTokens) != 2 {
		t.Errorf("Expected 2 Ethereum tokens, got %d", len(ethTokens))
	}

	// Get Polygon tokens (should return 1 token)
	polyTokens, err := repo.GetByChain(context.Background(), types.ChainPolygon)
	if err != nil {
		t.Fatalf("Failed to get Polygon tokens: %v", err)
	}

	if len(polyTokens) != 1 {
		t.Errorf("Expected 1 Polygon token, got %d", len(polyTokens))
	}
}

func TestSupportedTokenRepository_IsSupported(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewSupportedTokenRepository(db)

	token := &types.SupportedToken{
		ChainID:      types.ChainEthereum,
		TokenAddress: "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Name:         "Test Token",
		Symbol:       "TEST",
		Decimals:     18,
		Enabled:      true,
	}

	// Initially should not be supported
	supported, err := repo.IsSupported(context.Background(), token.ChainID, token.TokenAddress)
	if err != nil {
		t.Fatalf("Failed to check token support: %v", err)
	}
	if supported {
		t.Error("Token should not be supported initially")
	}

	// Create the token
	err = repo.Create(context.Background(), token)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Should now be supported
	supported, err = repo.IsSupported(context.Background(), token.ChainID, token.TokenAddress)
	if err != nil {
		t.Fatalf("Failed to check token support: %v", err)
	}
	if !supported {
		t.Error("Token should be supported after creation")
	}

	// Disable the token
	err = repo.UpdateEnabled(context.Background(), token.ID, false)
	if err != nil {
		t.Fatalf("Failed to disable token: %v", err)
	}

	// Should no longer be supported
	supported, err = repo.IsSupported(context.Background(), token.ChainID, token.TokenAddress)
	if err != nil {
		t.Fatalf("Failed to check token support: %v", err)
	}
	if supported {
		t.Error("Disabled token should not be supported")
	}
}

func TestSupportedTokenRepository_UpdateEnabled(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewSupportedTokenRepository(db)

	token := &types.SupportedToken{
		ChainID:      types.ChainEthereum,
		TokenAddress: "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Name:         "Test Token",
		Symbol:       "TEST",
		Decimals:     18,
		Enabled:      true,
	}

	err := repo.Create(context.Background(), token)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Disable the token
	err = repo.UpdateEnabled(context.Background(), token.ID, false)
	if err != nil {
		t.Fatalf("Failed to update enabled status: %v", err)
	}

	// Verify the token is disabled
	retrieved, err := repo.GetByChainAndAddress(context.Background(), token.ChainID, token.TokenAddress)
	if err != nil {
		t.Fatalf("Failed to retrieve token: %v", err)
	}

	if retrieved.Enabled {
		t.Error("Token should be disabled")
	}

	// Re-enable the token
	err = repo.UpdateEnabled(context.Background(), token.ID, true)
	if err != nil {
		t.Fatalf("Failed to update enabled status: %v", err)
	}

	// Verify the token is enabled
	retrieved, err = repo.GetByChainAndAddress(context.Background(), token.ChainID, token.TokenAddress)
	if err != nil {
		t.Fatalf("Failed to retrieve token: %v", err)
	}

	if !retrieved.Enabled {
		t.Error("Token should be enabled")
	}
}

func TestSupportedTokenRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewSupportedTokenRepository(db)

	token := &types.SupportedToken{
		ChainID:      types.ChainEthereum,
		TokenAddress: "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Name:         "Test Token",
		Symbol:       "TEST",
		Decimals:     18,
		Enabled:      true,
	}

	err := repo.Create(context.Background(), token)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Delete the token
	err = repo.Delete(context.Background(), token.ID)
	if err != nil {
		t.Fatalf("Failed to delete token: %v", err)
	}

	// Verify the token is deleted
	_, err = repo.GetByChainAndAddress(context.Background(), token.ChainID, token.TokenAddress)
	if err == nil {
		t.Error("Expected error when getting deleted token")
	}
}