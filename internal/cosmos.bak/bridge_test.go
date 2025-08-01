package cosmos

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nexus-bridge/pkg/types"
)

func TestCosmosBridge(t *testing.T) {
	// Create a new Cosmos bridge
	chainID := types.ChainCosmos
	config := types.ChainConfig{
		ChainID:               chainID,
		Name:                  "Cosmos Hub",
		Type:                  types.ChainTypeCosmos,
		RPC:                   "https://rpc.cosmos.network:443",
		BridgeContract:        "cosmos1bridge123",
		RequiredConfirmations: 6,
		BlockTime:             time.Second * 6,
		GasLimit:              200000,
		Enabled:               true,
	}

	bridge := NewCosmosBridge(chainID, config)
	require.NotNil(t, bridge)
	require.Equal(t, chainID, bridge.GetChainID())

	ctx := context.Background()

	// Test token registration
	err := bridge.RegisterToken(
		"uatom",
		"Cosmos Hub Atom",
		"ATOM",
		6,
		118,
		true,
		types.NewBigInt(types.NewBigIntFromInt64(1000)),
		types.NewBigInt(types.NewBigIntFromInt64(1000000000)),
	)
	require.NoError(t, err)

	// Verify token was registered
	tokenInfo, exists := bridge.GetTokenInfo("uatom")
	require.True(t, exists)
	require.Equal(t, "uatom", tokenInfo.Denom)
	require.Equal(t, "Cosmos Hub Atom", tokenInfo.Name)
	require.Equal(t, "ATOM", tokenInfo.Symbol)
	require.True(t, tokenInfo.Enabled)

	// Test relayer registration
	err = bridge.AddRelayer(
		"cosmos1relayer123",
		"pubkey123",
		types.NewBigInt(types.NewBigIntFromInt64(10000000000)),
	)
	require.NoError(t, err)

	// Verify relayer was added
	relayerInfo, exists := bridge.GetRelayerInfo("cosmos1relayer123")
	require.True(t, exists)
	require.Equal(t, "cosmos1relayer123", relayerInfo.Address)
	require.True(t, relayerInfo.IsActive)

	// Verify validator set was updated
	validatorSet := bridge.GetValidatorSet()
	require.Len(t, validatorSet.Relayers, 1)
	require.Equal(t, "cosmos1relayer123", validatorSet.Relayers[0].Address)
}

func TestLockTokens(t *testing.T) {
	bridge := NewCosmosBridge(types.ChainCosmos, types.ChainConfig{})
	ctx := context.Background()

	// Register a test token
	err := bridge.RegisterToken(
		"uatom",
		"Cosmos Hub Atom",
		"ATOM",
		6,
		118,
		true,
		types.NewBigInt(types.NewBigIntFromInt64(1000)),
		types.NewBigInt(types.NewBigIntFromInt64(1000000000)),
	)
	require.NoError(t, err)

	// Test successful token locking
	transferID := "test-transfer-1"
	sender := "cosmos1sender123"
	recipient := "0x1234567890123456789012345678901234567890"
	amount := types.NewBigInt(types.NewBigIntFromInt64(1000000))
	destinationChain := uint64(1) // Ethereum

	err = bridge.LockTokens(ctx, transferID, sender, "uatom", amount, destinationChain, recipient)
	require.NoError(t, err)

	// Verify transfer record was created
	transferRecord, exists := bridge.GetTransferRecord(transferID)
	require.True(t, exists)
	require.Equal(t, transferID, transferRecord.TransferID)
	require.Equal(t, sender, transferRecord.Sender)
	require.Equal(t, recipient, transferRecord.Recipient)
	require.Equal(t, "uatom", transferRecord.Denom)
	require.Equal(t, amount.Int, transferRecord.Amount.Int)
	require.Equal(t, TransferStatusLocked, transferRecord.Status)
}

func TestLockTokensValidation(t *testing.T) {
	bridge := NewCosmosBridge(types.ChainCosmos, types.ChainConfig{})
	ctx := context.Background()

	// Test with unsupported token
	err := bridge.LockTokens(ctx, "test-1", "cosmos1sender", "unsupported", types.NewBigInt(types.NewBigIntFromInt64(1000)), 1, "recipient")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not supported")

	// Register a test token
	err = bridge.RegisterToken(
		"uatom",
		"Cosmos Hub Atom",
		"ATOM",
		6,
		118,
		true,
		types.NewBigInt(types.NewBigIntFromInt64(1000)),
		types.NewBigInt(types.NewBigIntFromInt64(1000000)),
	)
	require.NoError(t, err)

	// Test with amount below minimum
	err = bridge.LockTokens(ctx, "test-2", "cosmos1sender", "uatom", types.NewBigInt(types.NewBigIntFromInt64(500)), 1, "recipient")
	require.Error(t, err)
	require.Contains(t, err.Error(), "below minimum")

	// Test with amount above maximum
	err = bridge.LockTokens(ctx, "test-3", "cosmos1sender", "uatom", types.NewBigInt(types.NewBigIntFromInt64(2000000)), 1, "recipient")
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds maximum")

	// Test duplicate transfer ID
	err = bridge.LockTokens(ctx, "test-4", "cosmos1sender", "uatom", types.NewBigInt(types.NewBigIntFromInt64(10000)), 1, "recipient")
	require.NoError(t, err)

	err = bridge.LockTokens(ctx, "test-4", "cosmos1sender2", "uatom", types.NewBigInt(types.NewBigIntFromInt64(20000)), 1, "recipient")
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}

func TestUnlockTokens(t *testing.T) {
	bridge := NewCosmosBridge(types.ChainCosmos, types.ChainConfig{})
	ctx := context.Background()

	// Register a test token
	err := bridge.RegisterToken(
		"uatom",
		"Cosmos Hub Atom",
		"ATOM",
		6,
		118,
		true,
		types.NewBigInt(types.NewBigIntFromInt64(1000)),
		types.NewBigInt(types.NewBigIntFromInt64(1000000000)),
	)
	require.NoError(t, err)

	// Add a relayer
	err = bridge.AddRelayer(
		"cosmos1relayer123",
		"pubkey123",
		types.NewBigInt(types.NewBigIntFromInt64(10000000000)),
	)
	require.NoError(t, err)

	// Create a locked transfer
	transferID := "test-unlock-1"
	amount := types.NewBigInt(types.NewBigIntFromInt64(1000000))
	err = bridge.LockTokens(ctx, transferID, "cosmos1sender", "uatom", amount, 1, "cosmos1recipient")
	require.NoError(t, err)

	// Test successful token unlocking
	signatures := []string{"sig1", "sig2", "sig3"}
	err = bridge.UnlockTokens(ctx, "cosmos1relayer123", transferID, "uatom", amount, "cosmos1recipient", signatures, 1)
	require.NoError(t, err)

	// Verify transfer record was updated
	transferRecord, exists := bridge.GetTransferRecord(transferID)
	require.True(t, exists)
	require.Equal(t, TransferStatusUnlocked, transferRecord.Status)
}

func TestUnlockTokensValidation(t *testing.T) {
	bridge := NewCosmosBridge(types.ChainCosmos, types.ChainConfig{})
	ctx := context.Background()

	// Test with unauthorized relayer
	err := bridge.UnlockTokens(ctx, "cosmos1unauthorized", "test-1", "uatom", types.NewBigInt(types.NewBigIntFromInt64(1000)), "recipient", []string{"sig1", "sig2"}, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not authorized")

	// Add a relayer
	err = bridge.AddRelayer(
		"cosmos1relayer123",
		"pubkey123",
		types.NewBigInt(types.NewBigIntFromInt64(10000000000)),
	)
	require.NoError(t, err)

	// Test with non-existent transfer
	err = bridge.UnlockTokens(ctx, "cosmos1relayer123", "non-existent", "uatom", types.NewBigInt(types.NewBigIntFromInt64(1000)), "recipient", []string{"sig1", "sig2"}, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestProcessIBCTransfer(t *testing.T) {
	bridge := NewCosmosBridge(types.ChainCosmos, types.ChainConfig{})
	ctx := context.Background()

	// Test IBC transfer processing
	err := bridge.ProcessIBCTransfer(
		ctx,
		1,                    // packet sequence
		"transfer",           // source port
		"channel-0",          // source channel
		"bridge",             // destination port
		"channel-1",          // destination channel
		"ibc-transfer-1",     // transfer ID
		"cosmos1sender123",   // sender
		"cosmos1receiver123", // receiver
		"uatom",              // token
		types.NewBigInt(types.NewBigIntFromInt64(1000000)), // amount
		1000, // timeout height
		0,    // timeout timestamp
	)
	require.NoError(t, err)

	// Verify IBC transfer record was created
	ibcTransfer, exists := bridge.ibcTransfers[1]
	require.True(t, exists)
	require.Equal(t, "ibc-transfer-1", ibcTransfer.TransferID)
	require.Equal(t, IBCTransferStatusReceived, ibcTransfer.Status)
}

func TestValidateSignatures(t *testing.T) {
	bridge := NewCosmosBridge(types.ChainCosmos, types.ChainConfig{})
	ctx := context.Background()

	// Test insufficient signatures
	err := bridge.ValidateSignatures(ctx, "test-transfer", []string{"sig1"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient signatures")

	// Test sufficient signatures
	err = bridge.ValidateSignatures(ctx, "test-transfer", []string{"sig1", "sig2", "sig3"})
	require.NoError(t, err)
}

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	require.NotNil(t, params.MinTransferAmount)
	require.NotNil(t, params.MaxTransferAmount)
	require.Equal(t, uint64(10), params.TransferFeePercent)
	require.Equal(t, uint64(2), params.SignatureThreshold)
	require.Equal(t, time.Hour*24, params.TransferTimeout)
	require.Equal(t, uint64(1000), params.IBCTimeoutHeight)
	require.Equal(t, time.Hour*1, params.IBCTimeoutTimestamp)
}