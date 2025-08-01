package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgLockTokens   = "lock_tokens"
	TypeMsgUnlockTokens = "unlock_tokens"
)

var (
	_ sdk.Msg = &MsgLockTokens{}
	_ sdk.Msg = &MsgUnlockTokens{}
)

// MsgLockTokens defines a message to lock tokens for cross-chain transfer
type MsgLockTokens struct {
	Sender           string `json:"sender" yaml:"sender"`
	Denom            string `json:"denom" yaml:"denom"`
	Amount           string `json:"amount" yaml:"amount"`
	DestinationChain uint64 `json:"destination_chain" yaml:"destination_chain"`
	Recipient        string `json:"recipient" yaml:"recipient"`
	TransferID       string `json:"transfer_id" yaml:"transfer_id"`
}

// NewMsgLockTokens creates a new MsgLockTokens instance
func NewMsgLockTokens(sender, denom, amount string, destinationChain uint64, recipient, transferID string) *MsgLockTokens {
	return &MsgLockTokens{
		Sender:           sender,
		Denom:            denom,
		Amount:           amount,
		DestinationChain: destinationChain,
		Recipient:        recipient,
		TransferID:       transferID,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgLockTokens) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface
func (msg MsgLockTokens) Type() string {
	return TypeMsgLockTokens
}

// GetSigners implements the sdk.Msg interface
func (msg MsgLockTokens) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgLockTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgLockTokens) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address (%s)", err)
	}

	if msg.Denom == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "denom cannot be empty")
	}

	amount, ok := sdk.NewIntFromString(msg.Amount)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid amount")
	}

	if amount.IsZero() || amount.IsNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "amount must be positive")
	}

	if msg.DestinationChain == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "destination chain cannot be zero")
	}

	if msg.Recipient == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "recipient cannot be empty")
	}

	if msg.TransferID == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "transfer ID cannot be empty")
	}

	return nil
}

// MsgUnlockTokens defines a message to unlock tokens from cross-chain transfer
type MsgUnlockTokens struct {
	Relayer     string   `json:"relayer" yaml:"relayer"`
	TransferID  string   `json:"transfer_id" yaml:"transfer_id"`
	Denom       string   `json:"denom" yaml:"denom"`
	Amount      string   `json:"amount" yaml:"amount"`
	Recipient   string   `json:"recipient" yaml:"recipient"`
	Signatures  []string `json:"signatures" yaml:"signatures"`
	SourceChain uint64   `json:"source_chain" yaml:"source_chain"`
}

// NewMsgUnlockTokens creates a new MsgUnlockTokens instance
func NewMsgUnlockTokens(relayer, transferID, denom, amount, recipient string, signatures []string, sourceChain uint64) *MsgUnlockTokens {
	return &MsgUnlockTokens{
		Relayer:     relayer,
		TransferID:  transferID,
		Denom:       denom,
		Amount:      amount,
		Recipient:   recipient,
		Signatures:  signatures,
		SourceChain: sourceChain,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgUnlockTokens) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface
func (msg MsgUnlockTokens) Type() string {
	return TypeMsgUnlockTokens
}

// GetSigners implements the sdk.Msg interface
func (msg MsgUnlockTokens) GetSigners() []sdk.AccAddress {
	relayer, err := sdk.AccAddressFromBech32(msg.Relayer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{relayer}
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgUnlockTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgUnlockTokens) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Relayer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid relayer address (%s)", err)
	}

	if msg.TransferID == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "transfer ID cannot be empty")
	}

	if msg.Denom == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "denom cannot be empty")
	}

	amount, ok := sdk.NewIntFromString(msg.Amount)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid amount")
	}

	if amount.IsZero() || amount.IsNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "amount must be positive")
	}

	_, err = sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient address (%s)", err)
	}

	if len(msg.Signatures) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "signatures cannot be empty")
	}

	if msg.SourceChain == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "source chain cannot be zero")
	}

	return nil
}