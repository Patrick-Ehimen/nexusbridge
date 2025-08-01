package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
)

// MockAccountKeeper implements the AccountKeeper interface for testing
type MockAccountKeeper struct{}

func (m *MockAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	return nil
}

func (m *MockAccountKeeper) SetAccount(ctx sdk.Context, acc authtypes.AccountI) {}

func (m *MockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return sdk.AccAddress("module_address")
}

func (m *MockAccountKeeper) GetModuleAccount(ctx sdk.Context, name string) authtypes.ModuleAccountI {
	return nil
}

// MockBankKeeper implements the BankKeeper interface for testing
type MockBankKeeper struct{}

func (m *MockBankKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdk.NewInt(1000000))
}

func (m *MockBankKeeper) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1000000)))
}

func (m *MockBankKeeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1000000)))
}

// MockChannelKeeper implements the ChannelKeeper interface for testing
type MockChannelKeeper struct{}

func (m *MockChannelKeeper) GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel ibcexported.ChannelI, found bool) {
	return nil, false
}

func (m *MockChannelKeeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) uint64 {
	return 1
}

func (m *MockChannelKeeper) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI) error {
	return nil
}

func (m *MockChannelKeeper) GetChannelCapability(ctx sdk.Context, portID, channelID string) *capabilitytypes.Capability {
	return &capabilitytypes.Capability{}
}

// MockPortKeeper implements the PortKeeper interface for testing
type MockPortKeeper struct{}

func (m *MockPortKeeper) BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability {
	return &capabilitytypes.Capability{}
}

// MockIBCKeeper implements the IBCKeeper interface for testing
type MockIBCKeeper struct {
	ChannelKeeper MockChannelKeeper
	PortKeeper    MockPortKeeper
}

// MockCapabilityKeeper implements the capability keeper interface for testing
type MockCapabilityKeeper struct{}

func (m *MockCapabilityKeeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return nil
}

func (m *MockCapabilityKeeper) GetCapability(ctx sdk.Context, name string) (*capabilitytypes.Capability, bool) {
	return &capabilitytypes.Capability{}, true
}