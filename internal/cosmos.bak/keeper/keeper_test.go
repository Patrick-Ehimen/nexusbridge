package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"nexus-bridge/internal/cosmos/types"
)

func TestTransferRecord(t *testing.T) {
	// Test transfer record creation and validation
	transferRecord := types.TransferRecord{
		TransferID:       "test-transfer-1",
		Sender:           "cosmos1sender123",
		Recipient:        "cosmos1recipient123",
		Denom:            "uatom",
		Amount:           sdk.NewInt(1000000),
		SourceChain:      1,
		DestinationChain: 118,
		Status:           types.TransferStatusLocked,
		BlockHeight:      100,
		Timestamp:        time.Now(),
		TxHash:           "hash123",
	}

	require.Equal(t, "test-transfer-1", transferRecord.TransferID)
	require.Equal(t, "cosmos1sender123", transferRecord.Sender)
	require.Equal(t, types.TransferStatusLocked, transferRecord.Status)
}

func TestTokenInfo(t *testing.T) {
	// Test token info validation
	tokenInfo := types.TokenInfo{
		Denom:       "uatom",
		Name:        "Cosmos Hub Atom",
		Symbol:      "ATOM",
		Decimals:    6,
		OriginChain: 118,
		IsNative:    true,
		Enabled:     true,
		MinTransfer: sdk.NewInt(1000),
		MaxTransfer: sdk.NewInt(1000000000),
	}

	err := tokenInfo.Validate()
	require.NoError(t, err)

	// Test invalid token info
	invalidToken := types.TokenInfo{
		Denom:       "", // Empty denom should fail
		Name:        "Test Token",
		Symbol:      "TEST",
		Decimals:    6,
		OriginChain: 1,
		IsNative:    false,
		Enabled:     true,
		MinTransfer: sdk.NewInt(1000),
		MaxTransfer: sdk.NewInt(1000000000),
	}

	err = invalidToken.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "denom cannot be empty")
}

func TestValidatorSet(t *testing.T) {
	// Test validator set validation
	relayers := []types.RelayerInfo{
		{
			Address:    "cosmos1relayer1",
			PubKey:     "pubkey1",
			IsActive:   true,
			Stake:      sdk.NewInt(10000000000),
			JoinedAt:   time.Now(),
			LastActive: time.Now(),
		},
		{
			Address:    "cosmos1relayer2",
			PubKey:     "pubkey2",
			IsActive:   true,
			Stake:      sdk.NewInt(10000000000),
			JoinedAt:   time.Now(),
			LastActive: time.Now(),
		},
	}

	validatorSet := types.ValidatorSet{
		Relayers:  relayers,
		Threshold: 2,
		UpdatedAt: time.Now(),
	}

	err := validatorSet.Validate()
	require.NoError(t, err)

	// Test invalid validator set (threshold too high)
	invalidValidatorSet := types.ValidatorSet{
		Relayers:  relayers,
		Threshold: 3, // Higher than number of relayers
		UpdatedAt: time.Now(),
	}

	err = invalidValidatorSet.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "threshold cannot exceed number of relayers")
}

func TestParams(t *testing.T) {
	// Test default params
	params := types.DefaultParams()
	err := params.Validate()
	require.NoError(t, err)

	// Test invalid params
	invalidParams := types.Params{
		MinTransferAmount:     sdk.NewInt(-1), // Negative amount should fail
		MaxTransferAmount:     sdk.NewInt(1000000000000),
		TransferFee:           sdk.NewDecWithPrec(1, 3),
		RelayerStakeRequired:  sdk.NewInt(10000000000),
		SignatureThreshold:    2,
		TransferTimeout:       time.Hour * 24,
		IBCTimeoutHeight:      1000,
		IBCTimeoutTimestamp:   time.Hour * 1,
	}

	err = invalidParams.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "min transfer amount cannot be negative")
}

func TestIBCTransferPacketData(t *testing.T) {
	// Test valid packet data
	packetData := types.IBCTransferPacketData{
		TransferID: "test-transfer-1",
		Sender:     "cosmos1sender123",
		Receiver:   "cosmos1receiver123",
		Token:      sdk.NewCoin("uatom", sdk.NewInt(1000000)),
	}

	err := packetData.ValidateBasic()
	require.NoError(t, err)

	// Test invalid packet data
	invalidPacketData := types.IBCTransferPacketData{
		TransferID: "", // Empty transfer ID should fail
		Sender:     "cosmos1sender123",
		Receiver:   "cosmos1receiver123",
		Token:      sdk.NewCoin("uatom", sdk.NewInt(1000000)),
	}

	err = invalidPacketData.ValidateBasic()
	require.Error(t, err)
	require.Contains(t, err.Error(), "transfer ID cannot be empty")
}

func TestMsgLockTokens(t *testing.T) {
	// Test valid message
	msg := types.NewMsgLockTokens(
		"cosmos1sender123",
		"uatom",
		"1000000",
		1,
		"0x1234567890123456789012345678901234567890",
		"test-transfer-1",
	)

	err := msg.ValidateBasic()
	require.NoError(t, err)
	require.Equal(t, types.TypeMsgLockTokens, msg.Type())
	require.Equal(t, types.RouterKey, msg.Route())

	// Test invalid message
	invalidMsg := types.NewMsgLockTokens(
		"", // Empty sender should fail
		"uatom",
		"1000000",
		1,
		"0x1234567890123456789012345678901234567890",
		"test-transfer-1",
	)

	err = invalidMsg.ValidateBasic()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid sender address")
}

func TestMsgUnlockTokens(t *testing.T) {
	// Test valid message
	msg := types.NewMsgUnlockTokens(
		"cosmos1relayer123",
		"test-transfer-1",
		"uatom",
		"1000000",
		"cosmos1recipient123",
		[]string{"sig1", "sig2", "sig3"},
		1,
	)

	err := msg.ValidateBasic()
	require.NoError(t, err)
	require.Equal(t, types.TypeMsgUnlockTokens, msg.Type())
	require.Equal(t, types.RouterKey, msg.Route())

	// Test invalid message
	invalidMsg := types.NewMsgUnlockTokens(
		"cosmos1relayer123",
		"", // Empty transfer ID should fail
		"uatom",
		"1000000",
		"cosmos1recipient123",
		[]string{"sig1", "sig2", "sig3"},
		1,
	)

	err = invalidMsg.ValidateBasic()
	require.Error(t, err)
	require.Contains(t, err.Error(), "transfer ID cannot be empty")
}