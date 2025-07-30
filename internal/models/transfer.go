package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"nexus-bridge/pkg/types"

	"github.com/jmoiron/sqlx"
)

// TransferRepository handles database operations for transfers
type TransferRepository struct {
	db *sqlx.DB
}

// NewTransferRepository creates a new transfer repository
func NewTransferRepository(db *sqlx.DB) *TransferRepository {
	return &TransferRepository{db: db}
}

// Create inserts a new transfer into the database
func (r *TransferRepository) Create(ctx context.Context, transfer *types.Transfer) error {
	if err := transfer.Validate(); err != nil {
		return fmt.Errorf("transfer validation failed: %w", err)
	}

	query := `
		INSERT INTO transfers (
			id, source_chain, destination_chain, token, amount, sender, recipient,
			status, source_tx_hash, destination_tx_hash, block_number, confirmations,
			fee, created_at, updated_at
		) VALUES (
			:id, :source_chain, :destination_chain, :token, :amount, :sender, :recipient,
			:status, :source_tx_hash, :destination_tx_hash, :block_number, :confirmations,
			:fee, :created_at, :updated_at
		)`

	transfer.CreatedAt = time.Now()
	transfer.UpdatedAt = time.Now()

	_, err := r.db.NamedExecContext(ctx, query, transfer)
	if err != nil {
		return fmt.Errorf("failed to create transfer: %w", err)
	}

	return nil
}

// GetByID retrieves a transfer by its ID
func (r *TransferRepository) GetByID(ctx context.Context, id string) (*types.Transfer, error) {
	var transfer types.Transfer
	query := `
		SELECT id, source_chain, destination_chain, token, amount, sender, recipient,
			   status, source_tx_hash, destination_tx_hash, block_number, confirmations,
			   fee, created_at, updated_at
		FROM transfers
		WHERE id = $1`

	err := r.db.GetContext(ctx, &transfer, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transfer not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get transfer: %w", err)
	}

	return &transfer, nil
}

// UpdateStatus updates the status of a transfer
func (r *TransferRepository) UpdateStatus(ctx context.Context, id string, status types.TransferStatus) error {
	query := `
		UPDATE transfers 
		SET status = $1, updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update transfer status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transfer not found: %s", id)
	}

	return nil
}

// UpdateDestinationTxHash updates the destination transaction hash
func (r *TransferRepository) UpdateDestinationTxHash(ctx context.Context, id string, txHash string) error {
	query := `
		UPDATE transfers 
		SET destination_tx_hash = $1, updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, txHash, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update destination tx hash: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transfer not found: %s", id)
	}

	return nil
}

// UpdateConfirmations updates the confirmation count for a transfer
func (r *TransferRepository) UpdateConfirmations(ctx context.Context, id string, confirmations uint64) error {
	query := `
		UPDATE transfers 
		SET confirmations = $1, updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, confirmations, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update confirmations: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transfer not found: %s", id)
	}

	return nil
}

// GetByBlockRange retrieves transfers within a block range for a specific chain
func (r *TransferRepository) GetByBlockRange(ctx context.Context, chainID types.ChainID, fromBlock, toBlock uint64) ([]types.Transfer, error) {
	var transfers []types.Transfer
	query := `
		SELECT id, source_chain, destination_chain, token, amount, sender, recipient,
			   status, source_tx_hash, destination_tx_hash, block_number, confirmations,
			   fee, created_at, updated_at
		FROM transfers
		WHERE source_chain = $1 AND block_number >= $2 AND block_number <= $3
		ORDER BY block_number ASC`

	err := r.db.SelectContext(ctx, &transfers, query, chainID, fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfers by block range: %w", err)
	}

	return transfers, nil
}

// List retrieves transfers with pagination
func (r *TransferRepository) List(ctx context.Context, limit, offset int) ([]types.Transfer, error) {
	var transfers []types.Transfer
	query := `
		SELECT id, source_chain, destination_chain, token, amount, sender, recipient,
			   status, source_tx_hash, destination_tx_hash, block_number, confirmations,
			   fee, created_at, updated_at
		FROM transfers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	err := r.db.SelectContext(ctx, &transfers, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list transfers: %w", err)
	}

	return transfers, nil
}

// GetByStatus retrieves transfers by status
func (r *TransferRepository) GetByStatus(ctx context.Context, status types.TransferStatus) ([]types.Transfer, error) {
	var transfers []types.Transfer
	query := `
		SELECT id, source_chain, destination_chain, token, amount, sender, recipient,
			   status, source_tx_hash, destination_tx_hash, block_number, confirmations,
			   fee, created_at, updated_at
		FROM transfers
		WHERE status = $1
		ORDER BY created_at ASC`

	err := r.db.SelectContext(ctx, &transfers, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfers by status: %w", err)
	}

	return transfers, nil
}

// MarkForReview marks a transfer for manual review
func (r *TransferRepository) MarkForReview(ctx context.Context, id string, reason string) error {
	query := `
		UPDATE transfers 
		SET status = $1, updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, types.StatusUnderReview, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to mark transfer for review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transfer not found: %s", id)
	}

	// TODO: Log the reason for review in a separate audit table
	return nil
}