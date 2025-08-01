package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TransferRecord represents a cross-chain transfer record in Cosmos
type TransferRecord struct {
	TransferID       string    `json:"transfer_id" yaml:"transfer_id"`
	Sender           string    `json:"sender" yaml:"sender"`
	Recipient        string    `json:"recipient" yaml:"recipient"`
	Denom            string    `json:"denom" yaml:"denom"`
	Amount           sdk.Int   `json:"amount" yaml:"amount"`
	SourceChain      uint64    `json:"source_chain" yaml:"source_chain"`
	DestinationChain uint64    `json:"destination_chain" yaml:"destination_chain"`
	Status           string    `json:"status" yaml:"status"`
	BlockHeight      int64     `json:"block_height" yaml:"block_height"`
	Timestamp        time.Time `json:"timestamp" yaml:"timestamp"`
	TxHash           string    `json:"tx_hash" yaml:"tx_hash"`
}

// TransferStatus represents the status of a transfer
const (
	TransferStatusPending   = "pending"
	TransferStatusLocked    = "locked"
	TransferStatusUnlocked  = "unlocked"
	TransferStatusCompleted = "completed"
	TransferStatusFailed    = "failed"
)

// TokenInfo represents information about a supported token
type TokenInfo struct {
	Denom        string `json:"denom" yaml:"denom"`
	Name         string `json:"name" yaml:"name"`
	Symbol       string `json:"symbol" yaml:"symbol"`
	Decimals     uint32 `json:"decimals" yaml:"decimals"`
	OriginChain  uint64 `json:"origin_chain" yaml:"origin_chain"`
	IsNative     bool   `json:"is_native" yaml:"is_native"`
	Enabled      bool   `json:"enabled" yaml:"enabled"`
	MinTransfer  sdk.Int `json:"min_transfer" yaml:"min_transfer"`
	MaxTransfer  sdk.Int `json:"max_transfer" yaml:"max_transfer"`
}

// Validate validates the token info
func (ti TokenInfo) Validate() error {
	if ti.Denom == "" {
		return fmt.Errorf("denom cannot be empty")
	}
	if ti.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if ti.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if ti.Decimals > 18 {
		return fmt.Errorf("decimals cannot exceed 18")
	}
	if ti.MinTransfer.IsNegative() {
		return fmt.Errorf("min transfer cannot be negative")
	}
	if ti.MaxTransfer.IsNegative() {
		return fmt.Errorf("max transfer cannot be negative")
	}
	if ti.MinTransfer.GT(ti.MaxTransfer) && !ti.MaxTransfer.IsZero() {
		return fmt.Errorf("min transfer cannot be greater than max transfer")
	}
	return nil
}

// RelayerInfo represents information about a relayer
type RelayerInfo struct {
	Address     string    `json:"address" yaml:"address"`
	PubKey      string    `json:"pub_key" yaml:"pub_key"`
	IsActive    bool      `json:"is_active" yaml:"is_active"`
	Stake       sdk.Int   `json:"stake" yaml:"stake"`
	JoinedAt    time.Time `json:"joined_at" yaml:"joined_at"`
	LastActive  time.Time `json:"last_active" yaml:"last_active"`
}

// ValidatorSet represents the set of authorized relayers
type ValidatorSet struct {
	Relayers  []RelayerInfo `json:"relayers" yaml:"relayers"`
	Threshold uint64        `json:"threshold" yaml:"threshold"`
	UpdatedAt time.Time     `json:"updated_at" yaml:"updated_at"`
}

// Validate validates the validator set
func (vs ValidatorSet) Validate() error {
	if len(vs.Relayers) == 0 {
		return fmt.Errorf("validator set cannot be empty")
	}
	if vs.Threshold == 0 {
		return fmt.Errorf("threshold cannot be zero")
	}
	if vs.Threshold > uint64(len(vs.Relayers)) {
		return fmt.Errorf("threshold cannot exceed number of relayers")
	}
	
	// Check for duplicate addresses
	seen := make(map[string]bool)
	for _, relayer := range vs.Relayers {
		if seen[relayer.Address] {
			return fmt.Errorf("duplicate relayer address: %s", relayer.Address)
		}
		seen[relayer.Address] = true
		
		// Validate relayer address
		if _, err := sdk.AccAddressFromBech32(relayer.Address); err != nil {
			return fmt.Errorf("invalid relayer address %s: %w", relayer.Address, err)
		}
	}
	
	return nil
}

// IBCTransferRecord represents an IBC transfer record
type IBCTransferRecord struct {
	PacketSequence   uint64    `json:"packet_sequence" yaml:"packet_sequence"`
	SourcePort       string    `json:"source_port" yaml:"source_port"`
	SourceChannel    string    `json:"source_channel" yaml:"source_channel"`
	DestinationPort  string    `json:"destination_port" yaml:"destination_port"`
	DestinationChannel string  `json:"destination_channel" yaml:"destination_channel"`
	TransferID       string    `json:"transfer_id" yaml:"transfer_id"`
	Sender           string    `json:"sender" yaml:"sender"`
	Receiver         string    `json:"receiver" yaml:"receiver"`
	Token            sdk.Coin  `json:"token" yaml:"token"`
	TimeoutHeight    uint64    `json:"timeout_height" yaml:"timeout_height"`
	TimeoutTimestamp uint64    `json:"timeout_timestamp" yaml:"timeout_timestamp"`
	Status           string    `json:"status" yaml:"status"`
	CreatedAt        time.Time `json:"created_at" yaml:"created_at"`
}

// IBCTransferStatus represents the status of an IBC transfer
const (
	IBCTransferStatusPending     = "pending"
	IBCTransferStatusSent        = "sent"
	IBCTransferStatusReceived    = "received"
	IBCTransferStatusTimeout     = "timeout"
	IBCTransferStatusFailed      = "failed"
)

// GenesisState defines the bridge module's genesis state
type GenesisState struct {
	Params           Params              `json:"params" yaml:"params"`
	Transfers        []TransferRecord    `json:"transfers" yaml:"transfers"`
	TokenRegistry    []TokenInfo         `json:"token_registry" yaml:"token_registry"`
	ValidatorSet     ValidatorSet        `json:"validator_set" yaml:"validator_set"`
	IBCTransfers     []IBCTransferRecord `json:"ibc_transfers" yaml:"ibc_transfers"`
}

// Params defines the parameters for the bridge module
type Params struct {
	MinTransferAmount     sdk.Int       `json:"min_transfer_amount" yaml:"min_transfer_amount"`
	MaxTransferAmount     sdk.Int       `json:"max_transfer_amount" yaml:"max_transfer_amount"`
	TransferFee           sdk.Dec       `json:"transfer_fee" yaml:"transfer_fee"`
	RelayerStakeRequired  sdk.Int       `json:"relayer_stake_required" yaml:"relayer_stake_required"`
	SignatureThreshold    uint64        `json:"signature_threshold" yaml:"signature_threshold"`
	TransferTimeout       time.Duration `json:"transfer_timeout" yaml:"transfer_timeout"`
	IBCTimeoutHeight      uint64        `json:"ibc_timeout_height" yaml:"ibc_timeout_height"`
	IBCTimeoutTimestamp   time.Duration `json:"ibc_timeout_timestamp" yaml:"ibc_timeout_timestamp"`
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MinTransferAmount:     sdk.NewInt(1),
		MaxTransferAmount:     sdk.NewInt(1000000000000), // 1 trillion base units
		TransferFee:           sdk.NewDecWithPrec(1, 3),  // 0.1%
		RelayerStakeRequired:  sdk.NewInt(10000000000),   // 10,000 base units
		SignatureThreshold:    2,                         // Require 2 signatures
		TransferTimeout:       time.Hour * 24,            // 24 hours
		IBCTimeoutHeight:      1000,                      // 1000 blocks
		IBCTimeoutTimestamp:   time.Hour * 1,             // 1 hour
	}
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.MinTransferAmount.IsNegative() {
		return fmt.Errorf("min transfer amount cannot be negative")
	}
	if p.MaxTransferAmount.IsNegative() {
		return fmt.Errorf("max transfer amount cannot be negative")
	}
	if p.MinTransferAmount.GT(p.MaxTransferAmount) {
		return fmt.Errorf("min transfer amount cannot be greater than max transfer amount")
	}
	if p.TransferFee.IsNegative() {
		return fmt.Errorf("transfer fee cannot be negative")
	}
	if p.TransferFee.GT(sdk.OneDec()) {
		return fmt.Errorf("transfer fee cannot exceed 100%%")
	}
	if p.RelayerStakeRequired.IsNegative() {
		return fmt.Errorf("relayer stake required cannot be negative")
	}
	if p.SignatureThreshold == 0 {
		return fmt.Errorf("signature threshold cannot be zero")
	}
	if p.TransferTimeout <= 0 {
		return fmt.Errorf("transfer timeout must be positive")
	}
	if p.IBCTimeoutHeight == 0 {
		return fmt.Errorf("IBC timeout height cannot be zero")
	}
	if p.IBCTimeoutTimestamp <= 0 {
		return fmt.Errorf("IBC timeout timestamp must be positive")
	}
	return nil
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:        DefaultParams(),
		Transfers:     []TransferRecord{},
		TokenRegistry: []TokenInfo{},
		ValidatorSet: ValidatorSet{
			Relayers:  []RelayerInfo{},
			Threshold: 2,
			UpdatedAt: time.Now(),
		},
		IBCTransfers: []IBCTransferRecord{},
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	
	if err := gs.ValidatorSet.Validate(); err != nil {
		return fmt.Errorf("invalid validator set: %w", err)
	}
	
	// Validate token registry
	for _, token := range gs.TokenRegistry {
		if err := token.Validate(); err != nil {
			return fmt.Errorf("invalid token %s: %w", token.Denom, err)
		}
	}
	
	// Check for duplicate transfer IDs
	transferIDs := make(map[string]bool)
	for _, transfer := range gs.Transfers {
		if transferIDs[transfer.TransferID] {
			return fmt.Errorf("duplicate transfer ID: %s", transfer.TransferID)
		}
		transferIDs[transfer.TransferID] = true
	}
	
	return nil
}
// Vers
ion defines the current version the IBC module supports
const Version = "bridge-1"

// IBCTransferPacketData defines the packet data for IBC transfers
type IBCTransferPacketData struct {
	TransferID string   `json:"transfer_id" yaml:"transfer_id"`
	Sender     string   `json:"sender" yaml:"sender"`
	Receiver   string   `json:"receiver" yaml:"receiver"`
	Token      sdk.Coin `json:"token" yaml:"token"`
}

// ValidateBasic validates the packet data
func (pd IBCTransferPacketData) ValidateBasic() error {
	if pd.TransferID == "" {
		return fmt.Errorf("transfer ID cannot be empty")
	}
	if pd.Sender == "" {
		return fmt.Errorf("sender cannot be empty")
	}
	if pd.Receiver == "" {
		return fmt.Errorf("receiver cannot be empty")
	}
	if !pd.Token.IsValid() {
		return fmt.Errorf("invalid token")
	}
	if pd.Token.IsZero() {
		return fmt.Errorf("token amount cannot be zero")
	}
	return nil
}

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, acc authtypes.AccountI)
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) authtypes.ModuleAccountI
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// IBCKeeper defines the expected IBC keeper
type IBCKeeper interface {
	ChannelKeeper interface {
		GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel ibcexported.ChannelI, found bool)
		GetNextSequenceSend(ctx sdk.Context, portID, channelID string) uint64
		SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI) error
		GetChannelCapability(ctx sdk.Context, portID, channelID string) *capabilitytypes.Capability
	}
	PortKeeper interface {
		BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability
	}
}

// ParamKeyTable returns the parameter key table for the bridge module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair([]byte("MinTransferAmount"), &p.MinTransferAmount, validateMinTransferAmount),
		paramtypes.NewParamSetPair([]byte("MaxTransferAmount"), &p.MaxTransferAmount, validateMaxTransferAmount),
		paramtypes.NewParamSetPair([]byte("TransferFee"), &p.TransferFee, validateTransferFee),
		paramtypes.NewParamSetPair([]byte("RelayerStakeRequired"), &p.RelayerStakeRequired, validateRelayerStakeRequired),
		paramtypes.NewParamSetPair([]byte("SignatureThreshold"), &p.SignatureThreshold, validateSignatureThreshold),
		paramtypes.NewParamSetPair([]byte("TransferTimeout"), &p.TransferTimeout, validateTransferTimeout),
		paramtypes.NewParamSetPair([]byte("IBCTimeoutHeight"), &p.IBCTimeoutHeight, validateIBCTimeoutHeight),
		paramtypes.NewParamSetPair([]byte("IBCTimeoutTimestamp"), &p.IBCTimeoutTimestamp, validateIBCTimeoutTimestamp),
	}
}

// Parameter validation functions
func validateMinTransferAmount(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("min transfer amount cannot be negative")
	}
	return nil
}

func validateMaxTransferAmount(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("max transfer amount cannot be negative")
	}
	return nil
}

func validateTransferFee(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("transfer fee cannot be negative")
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("transfer fee cannot exceed 100%%")
	}
	return nil
}

func validateRelayerStakeRequired(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("relayer stake required cannot be negative")
	}
	return nil
}

func validateSignatureThreshold(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("signature threshold cannot be zero")
	}
	return nil
}

func validateTransferTimeout(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("transfer timeout must be positive")
	}
	return nil
}

func validateIBCTimeoutHeight(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("IBC timeout height cannot be zero")
	}
	return nil
}

func validateIBCTimeoutTimestamp(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("IBC timeout timestamp must be positive")
	}
	return nil
}

// Error definitions
var (
	ErrInvalidVersion = fmt.Errorf("invalid version")
)

// Message response types
type MsgLockTokensResponse struct {
	TransferID string `json:"transfer_id" yaml:"transfer_id"`
}

type MsgUnlockTokensResponse struct {
	TransferID string `json:"transfer_id" yaml:"transfer_id"`
}

// MsgServer interface
type MsgServer interface {
	LockTokens(ctx context.Context, msg *MsgLockTokens) (*MsgLockTokensResponse, error)
	UnlockTokens(ctx context.Context, msg *MsgUnlockTokens) (*MsgUnlockTokensResponse, error)
}

// QueryServer interface
type QueryServer interface {
	// Add query methods as needed
}

// Service descriptor placeholder
var _Msg_serviceDesc = struct{}{}

// RegisterMsgServer registers the msg server
func RegisterMsgServer(s interface{}, srv MsgServer) {
	// Implementation would be generated by protobuf
}

// RegisterQueryServer registers the query server
func RegisterQueryServer(s interface{}, srv QueryServer) {
	// Implementation would be generated by protobuf
}

// RegisterQueryHandlerClient registers the query handler client
func RegisterQueryHandlerClient(ctx context.Context, mux interface{}, client interface{}) error {
	// Implementation would be generated by protobuf
	return nil
}

// NewQueryClient creates a new query client
func NewQueryClient(clientCtx interface{}) interface{} {
	// Implementation would be generated by protobuf
	return nil
}