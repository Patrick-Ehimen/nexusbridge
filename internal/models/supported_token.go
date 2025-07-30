package models

import (
	"context"
	"database/sql"
	"fmt"

	"nexus-bridge/pkg/types"

	"github.com/jmoiron/sqlx"
)

// SupportedTokenRepository handles database operations for supported tokens
type SupportedTokenRepository struct {
	db *sqlx.DB
}

// NewSupportedTokenRepository creates a new supported token repository
func NewSupportedTokenRepository(db *sqlx.DB) *SupportedTokenRepository {
	return &SupportedTokenRepository{db: db}
}

// Create inserts a new supported token into the database
func (r *SupportedTokenRepository) Create(ctx context.Context, token *types.SupportedToken) error {
	if err := r.validateToken(token); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	query := `
		INSERT INTO supported_tokens (chain_id, token_address, name, symbol, decimals, is_native, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := r.db.GetContext(ctx, &token.ID, query,
		token.ChainID, token.TokenAddress, token.Name, token.Symbol,
		token.Decimals, token.IsNative, token.Enabled)
	if err != nil {
		return fmt.Errorf("failed to create supported token: %w", err)
	}

	return nil
}

// GetByChainAndAddress retrieves a supported token by chain ID and token address
func (r *SupportedTokenRepository) GetByChainAndAddress(ctx context.Context, chainID types.ChainID, tokenAddress string) (*types.SupportedToken, error) {
	var token types.SupportedToken
	query := `
		SELECT id, chain_id, token_address, name, symbol, decimals, is_native, enabled
		FROM supported_tokens
		WHERE chain_id = $1 AND token_address = $2`

	err := r.db.GetContext(ctx, &token, query, chainID, tokenAddress)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("token not found: chain=%d, address=%s", chainID, tokenAddress)
		}
		return nil, fmt.Errorf("failed to get supported token: %w", err)
	}

	return &token, nil
}

// GetByChain retrieves all supported tokens for a specific chain
func (r *SupportedTokenRepository) GetByChain(ctx context.Context, chainID types.ChainID) ([]types.SupportedToken, error) {
	var tokens []types.SupportedToken
	query := `
		SELECT id, chain_id, token_address, name, symbol, decimals, is_native, enabled
		FROM supported_tokens
		WHERE chain_id = $1 AND enabled = true
		ORDER BY name ASC`

	err := r.db.SelectContext(ctx, &tokens, query, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supported tokens by chain: %w", err)
	}

	return tokens, nil
}

// GetAll retrieves all supported tokens
func (r *SupportedTokenRepository) GetAll(ctx context.Context) ([]types.SupportedToken, error) {
	var tokens []types.SupportedToken
	query := `
		SELECT id, chain_id, token_address, name, symbol, decimals, is_native, enabled
		FROM supported_tokens
		ORDER BY chain_id ASC, name ASC`

	err := r.db.SelectContext(ctx, &tokens, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all supported tokens: %w", err)
	}

	return tokens, nil
}

// UpdateEnabled updates the enabled status of a token
func (r *SupportedTokenRepository) UpdateEnabled(ctx context.Context, id int, enabled bool) error {
	query := `UPDATE supported_tokens SET enabled = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, enabled, id)
	if err != nil {
		return fmt.Errorf("failed to update token enabled status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found: %d", id)
	}

	return nil
}

// IsSupported checks if a token is supported on a specific chain
func (r *SupportedTokenRepository) IsSupported(ctx context.Context, chainID types.ChainID, tokenAddress string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM supported_tokens
		WHERE chain_id = $1 AND token_address = $2 AND enabled = true`

	err := r.db.GetContext(ctx, &count, query, chainID, tokenAddress)
	if err != nil {
		return false, fmt.Errorf("failed to check token support: %w", err)
	}

	return count > 0, nil
}

// Delete removes a supported token
func (r *SupportedTokenRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM supported_tokens WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete supported token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found: %d", id)
	}

	return nil
}

// validateToken validates the supported token data
func (r *SupportedTokenRepository) validateToken(token *types.SupportedToken) error {
	if token.ChainID == 0 {
		return fmt.Errorf("chain ID is required")
	}
	if token.TokenAddress == "" {
		return fmt.Errorf("token address is required")
	}
	if token.Name == "" {
		return fmt.Errorf("token name is required")
	}
	if token.Symbol == "" {
		return fmt.Errorf("token symbol is required")
	}
	if token.Decimals > 18 {
		return fmt.Errorf("decimals cannot exceed 18")
	}
	return nil
}