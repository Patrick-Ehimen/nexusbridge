package models

import (
	"context"
	"fmt"
	"time"

	"nexus-bridge/pkg/types"

	"github.com/jmoiron/sqlx"
)

// SignatureRepository handles database operations for signatures
type SignatureRepository struct {
	db *sqlx.DB
}

// NewSignatureRepository creates a new signature repository
func NewSignatureRepository(db *sqlx.DB) *SignatureRepository {
	return &SignatureRepository{db: db}
}

// Create inserts a new signature into the database
func (r *SignatureRepository) Create(ctx context.Context, transferID string, signature *types.Signature) error {
	if signature.RelayerAddress == "" {
		return fmt.Errorf("relayer address is required")
	}
	if len(signature.Signature) == 0 {
		return fmt.Errorf("signature data is required")
	}

	query := `
		INSERT INTO signatures (transfer_id, relayer_address, signature, created_at)
		VALUES ($1, $2, $3, $4)`

	signature.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query, transferID, signature.RelayerAddress, signature.Signature, signature.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create signature: %w", err)
	}

	return nil
}

// GetByTransferID retrieves all signatures for a transfer
func (r *SignatureRepository) GetByTransferID(ctx context.Context, transferID string) ([]types.Signature, error) {
	var signatures []types.Signature
	query := `
		SELECT relayer_address, signature, created_at
		FROM signatures
		WHERE transfer_id = $1
		ORDER BY created_at ASC`

	err := r.db.SelectContext(ctx, &signatures, query, transferID)
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures: %w", err)
	}

	return signatures, nil
}

// GetByTransferIDAndRelayer retrieves a signature for a specific transfer and relayer
func (r *SignatureRepository) GetByTransferIDAndRelayer(ctx context.Context, transferID, relayerAddress string) (*types.Signature, error) {
	var signature types.Signature
	query := `
		SELECT relayer_address, signature, created_at
		FROM signatures
		WHERE transfer_id = $1 AND relayer_address = $2`

	err := r.db.GetContext(ctx, &signature, query, transferID, relayerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get signature: %w", err)
	}

	return &signature, nil
}

// CountByTransferID returns the number of signatures for a transfer
func (r *SignatureRepository) CountByTransferID(ctx context.Context, transferID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM signatures WHERE transfer_id = $1`

	err := r.db.GetContext(ctx, &count, query, transferID)
	if err != nil {
		return 0, fmt.Errorf("failed to count signatures: %w", err)
	}

	return count, nil
}

// HasSignature checks if a relayer has already signed a transfer
func (r *SignatureRepository) HasSignature(ctx context.Context, transferID, relayerAddress string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM signatures WHERE transfer_id = $1 AND relayer_address = $2`

	err := r.db.GetContext(ctx, &count, query, transferID, relayerAddress)
	if err != nil {
		return false, fmt.Errorf("failed to check signature existence: %w", err)
	}

	return count > 0, nil
}

// DeleteByTransferID deletes all signatures for a transfer (used in case of reorg)
func (r *SignatureRepository) DeleteByTransferID(ctx context.Context, transferID string) error {
	query := `DELETE FROM signatures WHERE transfer_id = $1`

	_, err := r.db.ExecContext(ctx, query, transferID)
	if err != nil {
		return fmt.Errorf("failed to delete signatures: %w", err)
	}

	return nil
}