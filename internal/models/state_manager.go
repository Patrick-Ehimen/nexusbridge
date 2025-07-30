package models

import (
	"context"
	"fmt"

	"nexus-bridge/pkg/types"

	"github.com/jmoiron/sqlx"
)

// StateManager implements the StateManager interface using database repositories
type StateManager struct {
	transferRepo  *TransferRepository
	signatureRepo *SignatureRepository
	tokenRepo     *SupportedTokenRepository
}

// NewStateManager creates a new state manager with database repositories
func NewStateManager(db *sqlx.DB) *StateManager {
	return &StateManager{
		transferRepo:  NewTransferRepository(db),
		signatureRepo: NewSignatureRepository(db),
		tokenRepo:     NewSupportedTokenRepository(db),
	}
}

// RecordTransfer records a new transfer in the database
func (sm *StateManager) RecordTransfer(ctx context.Context, transfer types.Transfer) error {
	return sm.transferRepo.Create(ctx, &transfer)
}

// GetTransferStatus returns the current status of a transfer
func (sm *StateManager) GetTransferStatus(ctx context.Context, transferID string) (*types.TransferStatus, error) {
	transfer, err := sm.transferRepo.GetByID(ctx, transferID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer: %w", err)
	}
	
	return &transfer.Status, nil
}

// UpdateTransferStatus updates the status of a transfer
func (sm *StateManager) UpdateTransferStatus(ctx context.Context, transferID string, status types.TransferStatus) error {
	return sm.transferRepo.UpdateStatus(ctx, transferID, status)
}

// MarkTransferComplete marks a transfer as completed
func (sm *StateManager) MarkTransferComplete(ctx context.Context, transferID string, destinationTxHash string) error {
	// Update the destination transaction hash
	if err := sm.transferRepo.UpdateDestinationTxHash(ctx, transferID, destinationTxHash); err != nil {
		return fmt.Errorf("failed to update destination tx hash: %w", err)
	}
	
	// Update the status to completed
	return sm.transferRepo.UpdateStatus(ctx, transferID, types.StatusCompleted)
}

// IsTransferProcessed checks if a transfer has already been processed
func (sm *StateManager) IsTransferProcessed(ctx context.Context, transferID string) (bool, error) {
	transfer, err := sm.transferRepo.GetByID(ctx, transferID)
	if err != nil {
		// If transfer doesn't exist, it hasn't been processed
		return false, nil
	}
	
	// Consider transfer processed if it's completed or failed
	return transfer.Status == types.StatusCompleted || transfer.Status == types.StatusFailed, nil
}

// GetTransfersInBlockRange returns transfers affected by a block range (for reorg handling)
func (sm *StateManager) GetTransfersInBlockRange(ctx context.Context, chainID types.ChainID, fromBlock, toBlock uint64) ([]types.Transfer, error) {
	return sm.transferRepo.GetByBlockRange(ctx, chainID, fromBlock, toBlock)
}

// RecordSignature records a signature for a transfer
func (sm *StateManager) RecordSignature(ctx context.Context, transferID string, signature types.Signature) error {
	return sm.signatureRepo.Create(ctx, transferID, &signature)
}

// GetSignatures returns all signatures for a transfer
func (sm *StateManager) GetSignatures(ctx context.Context, transferID string) ([]types.Signature, error) {
	return sm.signatureRepo.GetByTransferID(ctx, transferID)
}

// MarkTransferForReview marks a transfer for manual review
func (sm *StateManager) MarkTransferForReview(ctx context.Context, transferID string, reason string) error {
	return sm.transferRepo.MarkForReview(ctx, transferID, reason)
}

// GetTransfer returns a transfer by ID
func (sm *StateManager) GetTransfer(ctx context.Context, transferID string) (*types.Transfer, error) {
	return sm.transferRepo.GetByID(ctx, transferID)
}

// GetTransfersByStatus returns transfers with a specific status
func (sm *StateManager) GetTransfersByStatus(ctx context.Context, status types.TransferStatus) ([]types.Transfer, error) {
	return sm.transferRepo.GetByStatus(ctx, status)
}

// UpdateConfirmations updates the confirmation count for a transfer
func (sm *StateManager) UpdateConfirmations(ctx context.Context, transferID string, confirmations uint64) error {
	return sm.transferRepo.UpdateConfirmations(ctx, transferID, confirmations)
}

// IsTokenSupported checks if a token is supported on a specific chain
func (sm *StateManager) IsTokenSupported(ctx context.Context, chainID types.ChainID, tokenAddress string) (bool, error) {
	return sm.tokenRepo.IsSupported(ctx, chainID, tokenAddress)
}

// GetSupportedTokens returns all supported tokens for a chain
func (sm *StateManager) GetSupportedTokens(ctx context.Context, chainID types.ChainID) ([]types.SupportedToken, error) {
	return sm.tokenRepo.GetByChain(ctx, chainID)
}

// HasRelayerSigned checks if a relayer has already signed a transfer
func (sm *StateManager) HasRelayerSigned(ctx context.Context, transferID, relayerAddress string) (bool, error) {
	return sm.signatureRepo.HasSignature(ctx, transferID, relayerAddress)
}

// GetSignatureCount returns the number of signatures for a transfer
func (sm *StateManager) GetSignatureCount(ctx context.Context, transferID string) (int, error) {
	return sm.signatureRepo.CountByTransferID(ctx, transferID)
}