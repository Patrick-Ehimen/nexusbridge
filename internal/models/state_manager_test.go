package models

import (
	"context"
	"math/big"
	"testing"

	"nexus-bridge/internal/models/testutil"
	"nexus-bridge/pkg/types"
)

func TestStateManager_RecordTransfer(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	sm := NewStateManager(db)

	transfer := types.Transfer{
		ID:               "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		SourceChain:      types.ChainEthereum,
		DestinationChain: types.ChainPolygon,
		Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
		Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
		Status:           types.StatusPending,
	}

	err := sm.RecordTransfer(context.Background(), transfer)
	if err != nil {
		t.Fatalf("Failed to record transfer: %v", err)
	}

	// Verify the transfer was recorded
	status, err := sm.GetTransferStatus(context.Background(), transfer.ID)
	if err != nil {
		t.Fatalf("Failed to get transfer status: %v", err)
	}

	if *status != types.StatusPending {
		t.Errorf("Expected status %s, got %s", types.StatusPending, *status)
	}
}

func TestStateManager_UpdateTransferStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	sm := NewStateManager(db)

	transfer := types.Transfer{
		ID:               "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		SourceChain:      types.ChainEthereum,
		DestinationChain: types.ChainPolygon,
		Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
		Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
		Status:           types.StatusPending,
	}

	err := sm.RecordTransfer(context.Background(), transfer)
	if err != nil {
		t.Fatalf("Failed to record transfer: %v", err)
	}

	// Update status
	err = sm.UpdateTransferStatus(context.Background(), transfer.ID, types.StatusConfirming)
	if err != nil {
		t.Fatalf("Failed to update transfer status: %v", err)
	}

	// Verify status was updated
	status, err := sm.GetTransferStatus(context.Background(), transfer.ID)
	if err != nil {
		t.Fatalf("Failed to get transfer status: %v", err)
	}

	if *status != types.StatusConfirming {
		t.Errorf("Expected status %s, got %s", types.StatusConfirming, *status)
	}
}

func TestStateManager_MarkTransferComplete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	sm := NewStateManager(db)

	transfer := types.Transfer{
		ID:               "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		SourceChain:      types.ChainEthereum,
		DestinationChain: types.ChainPolygon,
		Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
		Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
		Status:           types.StatusPending,
	}

	err := sm.RecordTransfer(context.Background(), transfer)
	if err != nil {
		t.Fatalf("Failed to record transfer: %v", err)
	}

	destinationTxHash := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	// Mark transfer complete
	err = sm.MarkTransferComplete(context.Background(), transfer.ID, destinationTxHash)
	if err != nil {
		t.Fatalf("Failed to mark transfer complete: %v", err)
	}

	// Verify status and destination tx hash
	retrievedTransfer, err := sm.GetTransfer(context.Background(), transfer.ID)
	if err != nil {
		t.Fatalf("Failed to get transfer: %v", err)
	}

	if retrievedTransfer.Status != types.StatusCompleted {
		t.Errorf("Expected status %s, got %s", types.StatusCompleted, retrievedTransfer.Status)
	}

	if retrievedTransfer.DestinationTxHash != destinationTxHash {
		t.Errorf("Expected destination tx hash %s, got %s", destinationTxHash, retrievedTransfer.DestinationTxHash)
	}
}

func TestStateManager_IsTransferProcessed(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	sm := NewStateManager(db)

	transferID := "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456"

	// Initially should not be processed (doesn't exist)
	processed, err := sm.IsTransferProcessed(context.Background(), transferID)
	if err != nil {
		t.Fatalf("Failed to check if transfer is processed: %v", err)
	}
	if processed {
		t.Error("Non-existent transfer should not be processed")
	}

	// Create a pending transfer
	transfer := types.Transfer{
		ID:               transferID,
		SourceChain:      types.ChainEthereum,
		DestinationChain: types.ChainPolygon,
		Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
		Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
		Status:           types.StatusPending,
	}

	err = sm.RecordTransfer(context.Background(), transfer)
	if err != nil {
		t.Fatalf("Failed to record transfer: %v", err)
	}

	// Pending transfer should not be considered processed
	processed, err = sm.IsTransferProcessed(context.Background(), transferID)
	if err != nil {
		t.Fatalf("Failed to check if transfer is processed: %v", err)
	}
	if processed {
		t.Error("Pending transfer should not be processed")
	}

	// Complete the transfer
	err = sm.UpdateTransferStatus(context.Background(), transferID, types.StatusCompleted)
	if err != nil {
		t.Fatalf("Failed to update transfer status: %v", err)
	}

	// Completed transfer should be considered processed
	processed, err = sm.IsTransferProcessed(context.Background(), transferID)
	if err != nil {
		t.Fatalf("Failed to check if transfer is processed: %v", err)
	}
	if !processed {
		t.Error("Completed transfer should be processed")
	}
}

func TestStateManager_RecordAndGetSignatures(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	sm := NewStateManager(db)

	// Create a transfer first
	transfer := types.Transfer{
		ID:               "0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456",
		SourceChain:      types.ChainEthereum,
		DestinationChain: types.ChainPolygon,
		Token:            "0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C",
		Amount:           types.NewBigInt(big.NewInt(1000000000000000000)),
		Sender:           "0x742d35Cc6634C0532925a3b8D4C9db96590C4C4C",
		Recipient:        "0x8ba1f109551bD432803012645Hac136c22C4C4C",
		Status:           types.StatusPending,
	}

	err := sm.RecordTransfer(context.Background(), transfer)
	if err != nil {
		t.Fatalf("Failed to record transfer: %v", err)
	}

	// Record signatures
	signatures := []types.Signature{
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
		err = sm.RecordSignature(context.Background(), transfer.ID, sig)
		if err != nil {
			t.Fatalf("Failed to record signature: %v", err)
		}
	}

	// Get signatures
	retrievedSigs, err := sm.GetSignatures(context.Background(), transfer.ID)
	if err != nil {
		t.Fatalf("Failed to get signatures: %v", err)
	}

	if len(retrievedSigs) != 2 {
		t.Errorf("Expected 2 signatures, got %d", len(retrievedSigs))
	}

	// Check signature count
	count, err := sm.GetSignatureCount(context.Background(), transfer.ID)
	if err != nil {
		t.Fatalf("Failed to get signature count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected signature count 2, got %d", count)
	}

	// Check if relayer has signed
	hasSigned, err := sm.HasRelayerSigned(context.Background(), transfer.ID, signatures[0].RelayerAddress)
	if err != nil {
		t.Fatalf("Failed to check if relayer has signed: %v", err)
	}

	if !hasSigned {
		t.Error("Relayer should have signed")
	}
}

func TestStateManager_GetTransfersInBlockRange(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	sm := NewStateManager(db)

	// Create transfers with different block numbers
	transfers := []types.Transfer{
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
	}

	for _, transfer := range transfers {
		err := sm.RecordTransfer(context.Background(), transfer)
		if err != nil {
			t.Fatalf("Failed to record transfer: %v", err)
		}
	}

	// Get transfers in block range 120-180
	retrieved, err := sm.GetTransfersInBlockRange(context.Background(), types.ChainEthereum, 120, 180)
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