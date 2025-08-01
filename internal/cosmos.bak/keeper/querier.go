package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"nexus-bridge/internal/cosmos/types"
)

// NewQuerier creates a new querier for bridge clients.
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case "transfer":
			return queryTransfer(ctx, path[1:], req, k, legacyQuerierCdc)
		case "token":
			return queryToken(ctx, path[1:], req, k, legacyQuerierCdc)
		case "validator-set":
			return queryValidatorSet(ctx, req, k, legacyQuerierCdc)
		case "relayer":
			return queryRelayer(ctx, path[1:], req, k, legacyQuerierCdc)
		case "params":
			return queryParams(ctx, req, k, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}
	}
}

func queryTransfer(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) == 0 {
		// Return all transfers
		transfers := keeper.GetAllTransferRecords(ctx)
		res, err := codec.MarshalJSONIndent(legacyQuerierCdc, transfers)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
		}
		return res, nil
	}

	// Return specific transfer
	transferID := path[0]
	transfer, found := keeper.GetTransferRecord(ctx, transferID)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "transfer %s not found", transferID)
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, transfer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryToken(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) == 0 {
		// Return all tokens
		tokens := keeper.GetAllTokenInfo(ctx)
		res, err := codec.MarshalJSONIndent(legacyQuerierCdc, tokens)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
		}
		return res, nil
	}

	// Return specific token
	denom := path[0]
	token, found := keeper.GetTokenInfo(ctx, denom)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "token %s not found", denom)
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, token)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryValidatorSet(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	validatorSet, found := keeper.GetValidatorSet(ctx)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "validator set not found")
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, validatorSet)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryRelayer(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) == 0 {
		// Return all relayers
		relayers := keeper.GetAllRelayerInfo(ctx)
		res, err := codec.MarshalJSONIndent(legacyQuerierCdc, relayers)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
		}
		return res, nil
	}

	// Return specific relayer
	address := path[0]
	relayer, found := keeper.GetRelayerInfo(ctx, address)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "relayer %s not found", address)
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, relayer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := keeper.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}