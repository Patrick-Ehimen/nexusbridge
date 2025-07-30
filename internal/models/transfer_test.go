package models

import (
	"context"
	"math/big"
	"testing"

	"nexus-bridge/internal/models/testutil"
	"nexus-bridge/pkg/types"
)

func TestTransferRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewTransferRepository(db)

	transfer := &types.Transfer{
		ID:               "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		SourceChain:      types.ChainEthereum,
		DestinationChain: types.ChainPolygon,
		Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
		Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
		Status:           types.StatusPending,
		Confirmations:    0,
	}

	err := repo.Create(context.Background(), transfer)
	if err != nil {
		t.Fatalf("Failed to create transfer: %v", err)
	}

	// Verify the transfer was created
	retrieved, err := repo.GetByID(context.Background(), transfer.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve transfer: %v", err)
	}

	if retrieved.ID != transfer.ID {
		t.Errorf("Expected ID %s, got %s", transfer.ID, retrieved.ID)
	}
	if retrieved.SourceChain != transfer.SourceChain {
		t.Errorf("Expected SourceChain %d, got %d", transfer.SourceChain, retrieved.SourceChain)
	}
	if retrieved.Amount.Cmp(transfer.Amount.Int) != 0 {
		t.Errorf("Expected Amount %s, got %s", transfer.Amount.String(), retrieved.Amount.String())
	}
}

func TestTransferRepository_CreateInvalidTransfer(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewTransferRepository(db)

	// Test with empty ID
	transfer := &types.Transfer{
		SourceChain:      types.ChainEthereum,
		DestinationChain: types.ChainPolygon,
		Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
		Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
		Status:           types.StatusPending,
	}

	err := repo.Create(context.Background(), transfer)
	if err == nil {
		t.Error("Expected error for transfer with empty ID")
	}
}

func TestTransferRepository_UpdateStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewTransferRepository(db)

	// Create a transfer first
	transfer := &types.Transfer{
		ID:               "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		SourceChain:      types.ChainEthereum,
		DestinationChain: types.ChainPolygon,
		Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
		Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
		Status:           types.StatusPending,
	}

	err := repo.Create(context.Background(), transfer)
	if err != nil {
		t.Fatalf("Failed to create transfer: %v", err)
	}

	// Update status
	err = repo.UpdateStatus(context.Background(), transfer.ID, types.StatusConfirming)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Verify status was updated
	retrieved, err := repo.GetByID(context.Background(), transfer.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve transfer: %v", err)
	}

	if retrieved.Status != types.StatusConfirming {
		t.Errorf("Expected status %s, got %s", types.StatusConfirming, retrieved.Status)
	}
}

func TestTransferRepository_GetByBlockRange(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewTransferRepository(db)

	// Create transfers with different block numbers
	transfers := []*types.Transfer{
		{
			ID:               "0x1111111111111111111111111111111111111111111111111111111111111111",
			SourceChain:      types.ChainEthereum,
			DestinationChain: types.ChainPolygon,
			Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
			Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
			Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
			Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
			Status:           types.StatusPending,
			BlockNumber:      100,
		},
		{
			ID:               "0x2222222222222222222222222222222222222222222222222222222222222222",
			SourceChain:      types.ChainEthereum,
			DestinationChain: types.ChainPolygon,
			Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
			Amount:           types.NewBigInt(big.NewInt(2000000000000000000)),
			Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
			Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
			Status:           types.StatusPending,
			BlockNumber:      150,
		},
		{
			ID:               "0x3333333333333333333333333333333333333333333333333333333333333333",
			SourceChain:      types.ChainEthereum,
			DestinationChain: types.ChainPolygon,
			Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
			Amount:           types.NewBigInt(big.NewInt(3000000000000000000)),
			Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
			Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
			Status:           types.StatusPending,
			BlockNumber:      200,
		},
	}

	for _, transfer := range transfers {
		err := repo.Create(context.Background(), transfer)
		if err != nil {
			t.Fatalf("Failed to create transfer: %v", err)
		}
	}

	// Get transfers in block range 120-180
	retrieved, err := repo.GetByBlockRange(context.Background(), types.ChainEthereum, 120, 180)
	if err != nil {
		t.Fatalf("Failed to get transfers by block range: %v", err)
	}

	if len(retrieved) != 1 {
		t.Errorf("Expected 1 transfer, got %d", len(retrieved))
	}

	if len(retrieved) > 0 && retrieved[0].ID != transfers[1].ID {
		t.Errorf("Expected transfer ID %s, got %s", transfers[1].ID, retrieved[0].ID)
	}
}

func TestTransfer_Validate(t *testing.T) {
	tests := []struct {
		name      string
		transfer  *types.Transfer
		expectErr bool
	}{
		{
			name: "valid transfer",
			transfer: &types.Transfer{
				ID:               "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
				SourceChain:      types.ChainEthereum,
				DestinationChain: types.ChainPolygon,
				Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
				Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
				Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
				Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
				Status:           types.StatusPending,
			},
			expectErr: false,
		},
		{
			name: "empty ID",
			transfer: &types.Transfer{
				SourceChain:      types.ChainEthereum,
				DestinationChain: types.ChainPolygon,
				Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
				Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
				Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
				Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
			},
			expectErr: true,
		},
		{
			name: "same source and destination chain",
			transfer: &types.Transfer{
				ID:               "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
				SourceChain:      types.ChainEthereum,
				DestinationChain: types.ChainEthereum,
				Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
				Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
				Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
				Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
			},
			expectErr: true,
		},
		{
			name: "zero amount",
			transfer: &types.Transfer{
				ID:               "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
				SourceChain:      types.ChainEthereum,
				DestinationChain: types.ChainPolygon,
				Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
				Amount:           types.NewBigInt(big.NewInt(0)),
				Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
				Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transfer.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}