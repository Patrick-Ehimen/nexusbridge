package models

import (
	"context"
	"math/big"
	"testing"

	"nexus-bridge/internal/models/testutil"
	"nexus-bridge/pkg/types"
)

func TestSignatureRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	// First create a transfer
	transferRepo := NewTransferRepository(db)
	transfer := createTestTransfer()
	err := transferRepo.Create(context.Background(), transfer)
	if err != nil {
		t.Fatalf("Failed to create transfer: %v", err)
	}

	// Now test signature creation
	sigRepo := NewSignatureRepository(db)
	signature := &types.Signature{
		RelayerAddress: "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Signature:      []byte("test_signature_data"),
	}

	err = sigRepo.Create(context.Background(), transfer.ID, signature)
	if err != nil {
		t.Fatalf("Failed to create signature: %v", err)
	}

	// Verify the signature was created
	signatures, err := sigRepo.GetByTransferID(context.Background(), transfer.ID)
	if err != nil {
		t.Fatalf("Failed to get signatures: %v", err)
	}

	if len(signatures) != 1 {
		t.Errorf("Expected 1 signature, got %d", len(signatures))
	}

	if signatures[0].RelayerAddress != signature.RelayerAddress {
		t.Errorf("Expected relayer address %s, got %s", signature.RelayerAddress, signatures[0].RelayerAddress)
	}
}

func TestSignatureRepository_CreateInvalidSignature(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	sigRepo := NewSignatureRepository(db)

	tests := []struct {
		name      string
		signature *types.Signature
		expectErr bool
	}{
		{
			name: "empty relayer address",
			signature: &types.Signature{
				RelayerAddress: "",
				Signature:      []byte("test_signature_data"),
			},
			expectErr: true,
		},
		{
			name: "empty signature data",
			signature: &types.Signature{
				RelayerAddress: "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
				Signature:      []byte{},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sigRepo.Create(context.Background(), "test-transfer-id", tt.signature)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestSignatureRepository_GetByTransferIDAndRelayer(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	// Create transfer and signature
	transferRepo := NewTransferRepository(db)
	transfer := createTestTransfer()
	err := transferRepo.Create(context.Background(), transfer)
	if err != nil {
		t.Fatalf("Failed to create transfer: %v", err)
	}

	sigRepo := NewSignatureRepository(db)
	signature := &types.Signature{
		RelayerAddress: "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Signature:      []byte("test_signature_data"),
	}

	err = sigRepo.Create(context.Background(), transfer.ID, signature)
	if err != nil {
		t.Fatalf("Failed to create signature: %v", err)
	}

	// Test retrieval
	retrieved, err := sigRepo.GetByTransferIDAndRelayer(context.Background(), transfer.ID, signature.RelayerAddress)
	if err != nil {
		t.Fatalf("Failed to get signature: %v", err)
	}

	if retrieved.RelayerAddress != signature.RelayerAddress {
		t.Errorf("Expected relayer address %s, got %s", signature.RelayerAddress, retrieved.RelayerAddress)
	}

	if string(retrieved.Signature) != string(signature.Signature) {
		t.Errorf("Expected signature %s, got %s", string(signature.Signature), string(retrieved.Signature))
	}
}

func TestSignatureRepository_CountByTransferID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	// Create transfer
	transferRepo := NewTransferRepository(db)
	transfer := createTestTransfer()
	err := transferRepo.Create(context.Background(), transfer)
	if err != nil {
		t.Fatalf("Failed to create transfer: %v", err)
	}

	sigRepo := NewSignatureRepository(db)

	// Initially should have 0 signatures
	count, err := sigRepo.CountByTransferID(context.Background(), transfer.ID)
	if err != nil {
		t.Fatalf("Failed to count signatures: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 signatures, got %d", count)
	}

	// Add signatures from different relayers
	signatures := []*types.Signature{
		{
			RelayerAddress: "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
			Signature:      []byte("signature1"),
		},
		{
			RelayerAddress: "0x8ba1f109551bD432803012645Hac136c22C4C4C",
			Signature:      []byte("signature2"),
		},
	}

	for _, sig := range signatures {
		err = sigRepo.Create(context.Background(), transfer.ID, sig)
		if err != nil {
			t.Fatalf("Failed to create signature: %v", err)
		}
	}

	// Should now have 2 signatures
	count, err = sigRepo.CountByTransferID(context.Background(), transfer.ID)
	if err != nil {
		t.Fatalf("Failed to count signatures: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 signatures, got %d", count)
	}
}

func TestSignatureRepository_HasSignature(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	// Create transfer
	transferRepo := NewTransferRepository(db)
	transfer := createTestTransfer()
	err := transferRepo.Create(context.Background(), transfer)
	if err != nil {
		t.Fatalf("Failed to create transfer: %v", err)
	}

	sigRepo := NewSignatureRepository(db)
	relayerAddress := "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C"

	// Initially should not have signature
	hasSignature, err := sigRepo.HasSignature(context.Background(), transfer.ID, relayerAddress)
	if err != nil {
		t.Fatalf("Failed to check signature existence: %v", err)
	}
	if hasSignature {
		t.Error("Expected no signature initially")
	}

	// Add signature
	signature := &types.Signature{
		RelayerAddress: relayerAddress,
		Signature:      []byte("test_signature_data"),
	}

	err = sigRepo.Create(context.Background(), transfer.ID, signature)
	if err != nil {
		t.Fatalf("Failed to create signature: %v", err)
	}

	// Should now have signature
	hasSignature, err = sigRepo.HasSignature(context.Background(), transfer.ID, relayerAddress)
	if err != nil {
		t.Fatalf("Failed to check signature existence: %v", err)
	}
	if !hasSignature {
		t.Error("Expected signature to exist")
	}
}

// Helper function to create a test transfer
func createTestTransfer() *types.Transfer {
	return &types.Transfer{
		ID:               "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		SourceChain:      types.ChainEthereum,
		DestinationChain: types.ChainPolygon,
		Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
		Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
		Status:           types.StatusPending,
	}
}