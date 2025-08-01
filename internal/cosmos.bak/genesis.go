package cosmos

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"nexus-bridge/internal/cosmos/keeper"
	"nexus-bridge/internal/cosmos/types"
)

// InitGenesis initializes the bridge module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set parameters
	k.SetParams(ctx, genState.Params)

	// Initialize transfer records
	for _, transfer := range genState.Transfers {
		k.SetTransferRecord(ctx, transfer)
	}

	// Initialize token registry
	for _, token := range genState.TokenRegistry {
		k.SetTokenInfo(ctx, token)
	}

	// Initialize validator set
	k.SetValidatorSet(ctx, genState.ValidatorSet)

	// Initialize relayer information
	for _, relayer := range genState.ValidatorSet.Relayers {
		k.SetRelayerInfo(ctx, relayer)
	}

	// Initialize IBC transfer records
	for _, ibcTransfer := range genState.IBCTransfers {
		k.SetIBCTransferRecord(ctx, ibcTransfer)
	}
}

// ExportGenesis returns the bridge module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	genesis := types.DefaultGenesis()

	// Export parameters
	genesis.Params = k.GetParams(ctx)

	// Export transfer records
	genesis.Transfers = k.GetAllTransferRecords(ctx)

	// Export token registry
	genesis.TokenRegistry = k.GetAllTokenInfo(ctx)

	// Export validator set
	if validatorSet, found := k.GetValidatorSet(ctx); found {
		genesis.ValidatorSet = validatorSet
	}

	// Export IBC transfer records
	genesis.IBCTransfers = k.GetAllIBCTransferRecords(ctx)

	return *genesis
}