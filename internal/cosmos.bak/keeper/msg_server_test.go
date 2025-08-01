package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	"nexus-bridge/internal/cosmos/keeper"
	"nexus-bridge/internal/cosmos/types"
)

type MsgServerTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	keeper    keeper.Keeper
	msgServer types.MsgServer
}

func (suite *MsgServerTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, sdk.StoreTypeMemory, nil)
	require.NoError(suite.T(), stateStore.LoadLatestVersion())

	// Create context
	suite.ctx = sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Create parameter subspace
	paramsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"BridgeParams",
	)

	// Create mock keepers
	accountKeeper := &MockAccountKeeper{}
	bankKeeper := &MockBankKeeper{}
	ibcKeeper := &MockIBCKeeper{}
	capabilityKeeper := &MockCapabilityKeeper{}

	// Create keeper
	suite.keeper = *keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		accountKeeper,
		bankKeeper,
		ibcKeeper,
		capabilityKeeper,
	)

	// Initialize params
	suite.keeper.SetParams(suite.ctx, types.DefaultParams())

	// Create message server
	suite.msgServer = keeper.NewMsgServerImpl(suite.keeper)
}

func (suite *MsgServerTestSuite) TestLockTokens() {
	// Setup test token
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
	suite.keeper.SetTokenInfo(suite.ctx, tokenInfo)

	// Test valid lock tokens message
	msg := &types.MsgLockTokens{
		Sender:           "cosmos1abc123",
		Denom:            "uatom",
		Amount:           "1000000",
		DestinationChain: 1,
		Recipient:        "0x1234567890123456789012345678901234567890",
		TransferID:       "test-transfer-1",
	}

	res, err := suite.msgServer.LockTokens(context.Background(), msg)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "test-transfer-1", res.TransferID)

	// Verify transfer record was created
	transfer, found := suite.keeper.GetTransferRecord(suite.ctx, "test-transfer-1")
	require.True(suite.T(), found)
	require.Equal(suite.T(), "test-transfer-1", transfer.TransferID)
	require.Equal(suite.T(), "cosmos1abc123", transfer.Sender)
	require.Equal(suite.T(), "uatom", transfer.Denom)
	require.Equal(suite.T(), sdk.NewInt(1000000), transfer.Amount)
	require.Equal(suite.T(), types.TransferStatusLocked, transfer.Status)
}

func (suite *MsgServerTestSuite) TestLockTokensInvalidToken() {
	// Test with unsupported token
	msg := &types.MsgLockTokens{
		Sender:           "cosmos1abc123",
		Denom:            "unsupported",
		Amount:           "1000000",
		DestinationChain: 1,
		Recipient:        "0x1234567890123456789012345678901234567890",
		TransferID:       "test-transfer-2",
	}

	_, err := suite.msgServer.LockTokens(context.Background(), msg)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "not supported")
}

func (suite *MsgServerTestSuite) TestLockTokensDuplicateTransferID() {
	// Setup test token
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
	suite.keeper.SetTokenInfo(suite.ctx, tokenInfo)

	// Create first transfer
	msg1 := &types.MsgLockTokens{
		Sender:           "cosmos1abc123",
		Denom:            "uatom",
		Amount:           "1000000",
		DestinationChain: 1,
		Recipient:        "0x1234567890123456789012345678901234567890",
		TransferID:       "duplicate-id",
	}

	_, err := suite.msgServer.LockTokens(context.Background(), msg1)
	require.NoError(suite.T(), err)

	// Try to create second transfer with same ID
	msg2 := &types.MsgLockTokens{
		Sender:           "cosmos1def456",
		Denom:            "uatom",
		Amount:           "2000000",
		DestinationChain: 1,
		Recipient:        "0x1234567890123456789012345678901234567890",
		TransferID:       "duplicate-id",
	}

	_, err = suite.msgServer.LockTokens(context.Background(), msg2)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "already exists")
}

func (suite *MsgServerTestSuite) TestUnlockTokens() {
	// Setup test token and relayer
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
	suite.keeper.SetTokenInfo(suite.ctx, tokenInfo)

	relayerInfo := types.RelayerInfo{
		Address:    "cosmos1relayer123",
		PubKey:     "pubkey123",
		IsActive:   true,
		Stake:      sdk.NewInt(10000000000),
		JoinedAt:   time.Now(),
		LastActive: time.Now(),
	}
	suite.keeper.SetRelayerInfo(suite.ctx, relayerInfo)

	// Create a locked transfer
	transferRecord := types.TransferRecord{
		TransferID:       "test-unlock-1",
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
	suite.keeper.SetTransferRecord(suite.ctx, transferRecord)

	// Test valid unlock tokens message
	msg := &types.MsgUnlockTokens{
		Relayer:     "cosmos1relayer123",
		TransferID:  "test-unlock-1",
		Denom:       "uatom",
		Amount:      "1000000",
		Recipient:   "cosmos1recipient123",
		Signatures:  []string{"sig1", "sig2", "sig3"},
		SourceChain: 1,
	}

	res, err := suite.msgServer.UnlockTokens(context.Background(), msg)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "test-unlock-1", res.TransferID)

	// Verify transfer record was updated
	transfer, found := suite.keeper.GetTransferRecord(suite.ctx, "test-unlock-1")
	require.True(suite.T(), found)
	require.Equal(suite.T(), types.TransferStatusUnlocked, transfer.Status)
}

func (suite *MsgServerTestSuite) TestUnlockTokensUnauthorizedRelayer() {
	// Create a locked transfer
	transferRecord := types.TransferRecord{
		TransferID:       "test-unlock-2",
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
	suite.keeper.SetTransferRecord(suite.ctx, transferRecord)

	// Test with unauthorized relayer
	msg := &types.MsgUnlockTokens{
		Relayer:     "cosmos1unauthorized",
		TransferID:  "test-unlock-2",
		Denom:       "uatom",
		Amount:      "1000000",
		Recipient:   "cosmos1recipient123",
		Signatures:  []string{"sig1", "sig2", "sig3"},
		SourceChain: 1,
	}

	_, err := suite.msgServer.UnlockTokens(context.Background(), msg)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "not authorized")
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}