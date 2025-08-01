package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	"nexus-bridge/internal/cosmos/keeper"
	"nexus-bridge/internal/cosmos/types"
)

type IBCHandlerTestSuite struct {
	suite.Suite

	ctx        sdk.Context
	keeper     keeper.Keeper
	ibcHandler keeper.IBCHandler
}

func (suite *IBCHandlerTestSuite) SetupTest() {
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

	// Create IBC handler
	suite.ibcHandler = keeper.NewIBCHandler(suite.keeper)
}

func (suite *IBCHandlerTestSuite) TestOnChanOpenInit() {
	// Test valid channel opening
	version, err := suite.ibcHandler.OnChanOpenInit(
		suite.ctx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		types.ModuleName,
		"channel-0",
		&channeltypes.Capability{},
		channeltypes.Counterparty{
			PortId:    types.ModuleName,
			ChannelId: "channel-1",
		},
		types.Version,
	)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), types.Version, version)
}

func (suite *IBCHandlerTestSuite) TestOnChanOpenInitInvalidPort() {
	// Test with invalid port
	_, err := suite.ibcHandler.OnChanOpenInit(
		suite.ctx,
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		"invalid-port",
		"channel-0",
		&channeltypes.Capability{},
		channeltypes.Counterparty{
			PortId:    types.ModuleName,
			ChannelId: "channel-1",
		},
		types.Version,
	)

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "invalid port")
}

func (suite *IBCHandlerTestSuite) TestOnChanOpenInitInvalidOrdering() {
	// Test with invalid channel ordering
	_, err := suite.ibcHandler.OnChanOpenInit(
		suite.ctx,
		channeltypes.ORDERED,
		[]string{"connection-0"},
		types.ModuleName,
		"channel-0",
		&channeltypes.Capability{},
		channeltypes.Counterparty{
			PortId:    types.ModuleName,
			ChannelId: "channel-1",
		},
		types.Version,
	)

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "invalid channel ordering")
}

func (suite *IBCHandlerTestSuite) TestOnRecvPacket() {
	// Create test packet data
	packetData := types.IBCTransferPacketData{
		TransferID: "test-transfer-1",
		Sender:     "cosmos1sender123",
		Receiver:   "cosmos1receiver123",
		Token:      sdk.NewCoin("uatom", sdk.NewInt(1000000)),
	}

	// Marshal packet data
	packetBytes, err := types.ModuleCdc.MarshalJSON(&packetData)
	require.NoError(suite.T(), err)

	// Create packet
	packet := channeltypes.NewPacket(
		packetBytes,
		1,
		types.ModuleName,
		"channel-0",
		types.ModuleName,
		"channel-1",
		channeltypes.NewHeight(0, 1000),
		0,
	)

	// Test packet reception
	ack := suite.ibcHandler.OnRecvPacket(
		suite.ctx,
		packet,
		sdk.AccAddress("relayer"),
	)

	require.True(suite.T(), ack.Success())

	// Verify IBC transfer record was created
	ibcTransfer, found := suite.keeper.GetIBCTransferRecord(suite.ctx, packet.GetSequence())
	require.True(suite.T(), found)
	require.Equal(suite.T(), "test-transfer-1", ibcTransfer.TransferID)
	require.Equal(suite.T(), types.IBCTransferStatusReceived, ibcTransfer.Status)
}

func (suite *IBCHandlerTestSuite) TestOnRecvPacketInvalidData() {
	// Create packet with invalid data
	packet := channeltypes.NewPacket(
		[]byte("invalid-json"),
		1,
		types.ModuleName,
		"channel-0",
		types.ModuleName,
		"channel-1",
		channeltypes.NewHeight(0, 1000),
		0,
	)

	// Test packet reception
	ack := suite.ibcHandler.OnRecvPacket(
		suite.ctx,
		packet,
		sdk.AccAddress("relayer"),
	)

	require.False(suite.T(), ack.Success())
}

func (suite *IBCHandlerTestSuite) TestOnTimeoutPacket() {
	// Create test packet data
	packetData := types.IBCTransferPacketData{
		TransferID: "test-transfer-timeout",
		Sender:     "cosmos1sender123",
		Receiver:   "cosmos1receiver123",
		Token:      sdk.NewCoin("uatom", sdk.NewInt(1000000)),
	}

	// Marshal packet data
	packetBytes, err := types.ModuleCdc.MarshalJSON(&packetData)
	require.NoError(suite.T(), err)

	// Create packet
	packet := channeltypes.NewPacket(
		packetBytes,
		1,
		types.ModuleName,
		"channel-0",
		types.ModuleName,
		"channel-1",
		channeltypes.NewHeight(0, 1000),
		0,
	)

	// Create IBC transfer record
	ibcTransfer := types.IBCTransferRecord{
		PacketSequence:     packet.GetSequence(),
		SourcePort:         packet.GetSourcePort(),
		SourceChannel:      packet.GetSourceChannel(),
		DestinationPort:    packet.GetDestPort(),
		DestinationChannel: packet.GetDestChannel(),
		TransferID:         packetData.TransferID,
		Sender:             packetData.Sender,
		Receiver:           packetData.Receiver,
		Token:              packetData.Token,
		TimeoutHeight:      packet.GetTimeoutHeight().GetRevisionHeight(),
		TimeoutTimestamp:   packet.GetTimeoutTimestamp(),
		Status:             types.IBCTransferStatusSent,
		CreatedAt:          suite.ctx.BlockTime(),
	}
	suite.keeper.SetIBCTransferRecord(suite.ctx, ibcTransfer)

	// Test timeout handling
	err = suite.ibcHandler.OnTimeoutPacket(
		suite.ctx,
		packet,
		sdk.AccAddress("relayer"),
	)

	require.NoError(suite.T(), err)

	// Verify IBC transfer record was updated
	updatedTransfer, found := suite.keeper.GetIBCTransferRecord(suite.ctx, packet.GetSequence())
	require.True(suite.T(), found)
	require.Equal(suite.T(), types.IBCTransferStatusTimeout, updatedTransfer.Status)
}

func (suite *IBCHandlerTestSuite) TestSendIBCTransfer() {
	// Test sending IBC transfer
	err := suite.keeper.SendIBCTransfer(
		suite.ctx,
		types.ModuleName,
		"channel-0",
		sdk.NewCoin("uatom", sdk.NewInt(1000000)),
		"cosmos1sender123",
		"cosmos1receiver123",
		1000,
		0,
		"test-transfer-send",
	)

	require.NoError(suite.T(), err)

	// Verify IBC transfer record was created
	ibcTransfer, found := suite.keeper.GetIBCTransferRecord(suite.ctx, 1)
	require.True(suite.T(), found)
	require.Equal(suite.T(), "test-transfer-send", ibcTransfer.TransferID)
	require.Equal(suite.T(), types.IBCTransferStatusSent, ibcTransfer.Status)
}

func TestIBCHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(IBCHandlerTestSuite))
}