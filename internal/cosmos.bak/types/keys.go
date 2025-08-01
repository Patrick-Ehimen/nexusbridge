package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "bridge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_bridge"
)

// Store key prefixes
var (
	// TransferKey is the prefix for transfer records
	TransferKey = []byte{0x01}
	
	// TokenRegistryKey is the prefix for token registry
	TokenRegistryKey = []byte{0x02}
	
	// ValidatorSetKey is the prefix for validator set
	ValidatorSetKey = []byte{0x03}
	
	// RelayerKey is the prefix for relayer information
	RelayerKey = []byte{0x04}
	
	// IBCTransferKey is the prefix for IBC transfer records
	IBCTransferKey = []byte{0x05}
)

// GetTransferKey returns the store key for a transfer
func GetTransferKey(transferID string) []byte {
	return append(TransferKey, []byte(transferID)...)
}

// GetTokenRegistryKey returns the store key for a token
func GetTokenRegistryKey(denom string) []byte {
	return append(TokenRegistryKey, []byte(denom)...)
}

// GetValidatorSetKey returns the store key for validator set
func GetValidatorSetKey() []byte {
	return ValidatorSetKey
}

// GetRelayerKey returns the store key for a relayer
func GetRelayerKey(address string) []byte {
	return append(RelayerKey, []byte(address)...)
}

// GetIBCTransferKey returns the store key for an IBC transfer
func GetIBCTransferKey(packetSequence uint64) []byte {
	return append(IBCTransferKey, sdk.Uint64ToBigEndian(packetSequence)...)
}