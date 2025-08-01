package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"nexus-bridge/internal/cosmos/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	memKey     sdk.StoreKey
	paramstore paramtypes.Subspace

	accountKeeper   types.AccountKeeper
	bankKeeper      types.BankKeeper
	ibcKeeper       types.IBCKeeper
	capabilityKeeper *capabilitykeeper.Keeper
}

// NewKeeper creates a new bridge Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	ibcKeeper types.IBCKeeper,
	capabilityKeeper *capabilitykeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,

		accountKeeper:   accountKeeper,
		bankKeeper:      bankKeeper,
		ibcKeeper:       ibcKeeper,
		capabilityKeeper: capabilityKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams gets all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.paramstore.GetParamSet(ctx, &p)
	return p
}

// SetParams sets the parameters for the module
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// SetTransferRecord sets a transfer record in the store
func (k Keeper) SetTransferRecord(ctx sdk.Context, transfer types.TransferRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&transfer)
	store.Set(types.GetTransferKey(transfer.TransferID), bz)
}

// GetTransferRecord gets a transfer record from the store
func (k Keeper) GetTransferRecord(ctx sdk.Context, transferID string) (types.TransferRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTransferKey(transferID))
	if bz == nil {
		return types.TransferRecord{}, false
	}

	var transfer types.TransferRecord
	k.cdc.MustUnmarshal(bz, &transfer)
	return transfer, true
}

// DeleteTransferRecord deletes a transfer record from the store
func (k Keeper) DeleteTransferRecord(ctx sdk.Context, transferID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetTransferKey(transferID))
}

// GetAllTransferRecords returns all transfer records
func (k Keeper) GetAllTransferRecords(ctx sdk.Context) []types.TransferRecord {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.TransferKey)
	defer iterator.Close()

	var transfers []types.TransferRecord
	for ; iterator.Valid(); iterator.Next() {
		var transfer types.TransferRecord
		k.cdc.MustUnmarshal(iterator.Value(), &transfer)
		transfers = append(transfers, transfer)
	}

	return transfers
}

// SetTokenInfo sets token information in the registry
func (k Keeper) SetTokenInfo(ctx sdk.Context, tokenInfo types.TokenInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&tokenInfo)
	store.Set(types.GetTokenRegistryKey(tokenInfo.Denom), bz)
}

// GetTokenInfo gets token information from the registry
func (k Keeper) GetTokenInfo(ctx sdk.Context, denom string) (types.TokenInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTokenRegistryKey(denom))
	if bz == nil {
		return types.TokenInfo{}, false
	}

	var tokenInfo types.TokenInfo
	k.cdc.MustUnmarshal(bz, &tokenInfo)
	return tokenInfo, true
}

// DeleteTokenInfo deletes token information from the registry
func (k Keeper) DeleteTokenInfo(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetTokenRegistryKey(denom))
}

// GetAllTokenInfo returns all token information
func (k Keeper) GetAllTokenInfo(ctx sdk.Context) []types.TokenInfo {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.TokenRegistryKey)
	defer iterator.Close()

	var tokens []types.TokenInfo
	for ; iterator.Valid(); iterator.Next() {
		var tokenInfo types.TokenInfo
		k.cdc.MustUnmarshal(iterator.Value(), &tokenInfo)
		tokens = append(tokens, tokenInfo)
	}

	return tokens
}

// SetValidatorSet sets the validator set
func (k Keeper) SetValidatorSet(ctx sdk.Context, validatorSet types.ValidatorSet) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&validatorSet)
	store.Set(types.GetValidatorSetKey(), bz)
}

// GetValidatorSet gets the validator set
func (k Keeper) GetValidatorSet(ctx sdk.Context) (types.ValidatorSet, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValidatorSetKey())
	if bz == nil {
		return types.ValidatorSet{}, false
	}

	var validatorSet types.ValidatorSet
	k.cdc.MustUnmarshal(bz, &validatorSet)
	return validatorSet, true
}

// SetRelayerInfo sets relayer information
func (k Keeper) SetRelayerInfo(ctx sdk.Context, relayerInfo types.RelayerInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&relayerInfo)
	store.Set(types.GetRelayerKey(relayerInfo.Address), bz)
}

// GetRelayerInfo gets relayer information
func (k Keeper) GetRelayerInfo(ctx sdk.Context, address string) (types.RelayerInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRelayerKey(address))
	if bz == nil {
		return types.RelayerInfo{}, false
	}

	var relayerInfo types.RelayerInfo
	k.cdc.MustUnmarshal(bz, &relayerInfo)
	return relayerInfo, true
}

// DeleteRelayerInfo deletes relayer information
func (k Keeper) DeleteRelayerInfo(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetRelayerKey(address))
}

// GetAllRelayerInfo returns all relayer information
func (k Keeper) GetAllRelayerInfo(ctx sdk.Context) []types.RelayerInfo {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.RelayerKey)
	defer iterator.Close()

	var relayers []types.RelayerInfo
	for ; iterator.Valid(); iterator.Next() {
		var relayerInfo types.RelayerInfo
		k.cdc.MustUnmarshal(iterator.Value(), &relayerInfo)
		relayers = append(relayers, relayerInfo)
	}

	return relayers
}

// SetIBCTransferRecord sets an IBC transfer record
func (k Keeper) SetIBCTransferRecord(ctx sdk.Context, ibcTransfer types.IBCTransferRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&ibcTransfer)
	store.Set(types.GetIBCTransferKey(ibcTransfer.PacketSequence), bz)
}

// GetIBCTransferRecord gets an IBC transfer record
func (k Keeper) GetIBCTransferRecord(ctx sdk.Context, packetSequence uint64) (types.IBCTransferRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetIBCTransferKey(packetSequence))
	if bz == nil {
		return types.IBCTransferRecord{}, false
	}

	var ibcTransfer types.IBCTransferRecord
	k.cdc.MustUnmarshal(bz, &ibcTransfer)
	return ibcTransfer, true
}

// DeleteIBCTransferRecord deletes an IBC transfer record
func (k Keeper) DeleteIBCTransferRecord(ctx sdk.Context, packetSequence uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetIBCTransferKey(packetSequence))
}

// GetAllIBCTransferRecords returns all IBC transfer records
func (k Keeper) GetAllIBCTransferRecords(ctx sdk.Context) []types.IBCTransferRecord {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.IBCTransferKey)
	defer iterator.Close()

	var ibcTransfers []types.IBCTransferRecord
	for ; iterator.Valid(); iterator.Next() {
		var ibcTransfer types.IBCTransferRecord
		k.cdc.MustUnmarshal(iterator.Value(), &ibcTransfer)
		ibcTransfers = append(ibcTransfers, ibcTransfer)
	}

	return ibcTransfers
}

// IsAuthorizedRelayer checks if an address is an authorized relayer
func (k Keeper) IsAuthorizedRelayer(ctx sdk.Context, address string) bool {
	relayerInfo, found := k.GetRelayerInfo(ctx, address)
	return found && relayerInfo.IsActive
}

// ValidateSignatures validates a set of signatures for a transfer
func (k Keeper) ValidateSignatures(ctx sdk.Context, transferID string, signatures []string) error {
	params := k.GetParams(ctx)
	
	if uint64(len(signatures)) < params.SignatureThreshold {
		return fmt.Errorf("insufficient signatures: got %d, required %d", len(signatures), params.SignatureThreshold)
	}

	// TODO: Implement actual signature validation logic
	// This would involve:
	// 1. Reconstructing the message that was signed
	// 2. Verifying each signature against the relayer's public key
	// 3. Ensuring no duplicate signatures from the same relayer
	
	return nil
}
// Clai
mCapability claims a capability for the module
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.capabilityKeeper.ClaimCapability(ctx, cap, name)
}

// GetCapability gets a capability for the module
func (k Keeper) GetCapability(ctx sdk.Context, name string) (*capabilitytypes.Capability, bool) {
	return k.capabilityKeeper.GetCapability(ctx, name)
}