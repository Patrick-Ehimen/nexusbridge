
package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"nexus-bridge/internal/cosmos/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// LockTokens implements the MsgServer interface for locking tokens
func (k msgServer) LockTokens(goCtx context.Context, msg *types.MsgLockTokens) (*types.MsgLockTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Check if transfer ID already exists
	if _, found := k.GetTransferRecord(ctx, msg.TransferID); found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "transfer ID %s already exists", msg.TransferID)
	}

	// Get token info from registry
	tokenInfo, found := k.GetTokenInfo(ctx, msg.Denom)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "token %s not supported", msg.Denom)
	}

	if !tokenInfo.Enabled {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "token %s is disabled", msg.Denom)
	}

	// Parse amount
	amount, ok := sdk.NewIntFromString(msg.Amount)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid amount %s", msg.Amount)
	}

	// Validate transfer amount limits
	if amount.LT(tokenInfo.MinTransfer) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "amount %s below minimum %s", amount, tokenInfo.MinTransfer)
	}

	if !tokenInfo.MaxTransfer.IsZero() && amount.GT(tokenInfo.MaxTransfer) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "amount %s exceeds maximum %s", amount, tokenInfo.MaxTransfer)
	}

	// Get sender address
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %s", err)
	}

	// Create coin to lock
	coin := sdk.NewCoin(msg.Denom, amount)

	// Lock tokens by sending them to the module account
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if err := k.bankKeeper.SendCoins(ctx, senderAddr, moduleAddr, sdk.NewCoins(coin)); err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to lock tokens")
	}

	// Create transfer record
	transferRecord := types.TransferRecord{
		TransferID:       msg.TransferID,
		Sender:           msg.Sender,
		Recipient:        msg.Recipient,
		Denom:            msg.Denom,
		Amount:           amount,
		SourceChain:      118, // Cosmos Hub chain ID
		DestinationChain: msg.DestinationChain,
		Status:           types.TransferStatusLocked,
		BlockHeight:      ctx.BlockHeight(),
		Timestamp:        ctx.BlockTime(),
		TxHash:           fmt.Sprintf("%X", ctx.TxBytes()),
	}

	// Store the transfer record
	k.SetTransferRecord(ctx, transferRecord)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"tokens_locked",
			sdk.NewAttribute("transfer_id", msg.TransferID),
			sdk.NewAttribute("sender", msg.Sender),
			sdk.NewAttribute("recipient", msg.Recipient),
			sdk.NewAttribute("denom", msg.Denom),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("destination_chain", fmt.Sprintf("%d", msg.DestinationChain)),
		),
	)

	return &types.MsgLockTokensResponse{
		TransferID: msg.TransferID,
	}, nil
}

// UnlockTokens implements the MsgServer interface for unlocking tokens
func (k msgServer) UnlockTokens(goCtx context.Context, msg *types.MsgUnlockTokens) (*types.MsgUnlockTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Check if relayer is authorized
	if !k.IsAuthorizedRelayer(ctx, msg.Relayer) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "relayer %s is not authorized", msg.Relayer)
	}

	// Check if transfer record exists
	transferRecord, found := k.GetTransferRecord(ctx, msg.TransferID)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "transfer ID %s not found", msg.TransferID)
	}

	// Check if transfer is already completed
	if transferRecord.Status == types.TransferStatusCompleted || transferRecord.Status == types.TransferStatusUnlocked {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "transfer %s already completed", msg.TransferID)
	}

	// Validate signatures
	if err := k.ValidateSignatures(ctx, msg.TransferID, msg.Signatures); err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signature validation failed: %s", err)
	}

	// Get token info
	tokenInfo, found := k.GetTokenInfo(ctx, msg.Denom)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "token %s not supported", msg.Denom)
	}

	// Parse amount
	amount, ok := sdk.NewIntFromString(msg.Amount)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid amount %s", msg.Amount)
	}

	// Verify amount matches the locked amount
	if !amount.Equal(transferRecord.Amount) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "amount mismatch: expected %s, got %s", transferRecord.Amount, amount)
	}

	// Get recipient address
	recipientAddr, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient address: %s", err)
	}

	// Create coin to unlock
	coin := sdk.NewCoin(msg.Denom, amount)

	// Unlock tokens by sending them from the module account to the recipient
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if err := k.bankKeeper.SendCoins(ctx, moduleAddr, recipientAddr, sdk.NewCoins(coin)); err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to unlock tokens")
	}

	// Update transfer record
	transferRecord.Status = types.TransferStatusUnlocked
	transferRecord.Timestamp = ctx.BlockTime()
	k.SetTransferRecord(ctx, transferRecord)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"tokens_unlocked",
			sdk.NewAttribute("transfer_id", msg.TransferID),
			sdk.NewAttribute("recipient", msg.Recipient),
			sdk.NewAttribute("denom", msg.Denom),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("relayer", msg.Relayer),
		),
	)

	return &types.MsgUnlockTokensResponse{
		TransferID: msg.TransferID,
	}, nil
}