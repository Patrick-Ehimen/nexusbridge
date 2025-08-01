package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"

	"nexus-bridge/internal/cosmos/types"
)

// IBCHandler handles IBC packet processing for cross-chain transfers
type IBCHandler struct {
	keeper Keeper
}

// NewIBCHandler creates a new IBC handler
func NewIBCHandler(keeper Keeper) IBCHandler {
	return IBCHandler{
		keeper: keeper,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCHandler) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *channeltypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	// Validate port ID
	if portID != types.ModuleName {
		return "", sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, types.ModuleName)
	}

	// Validate ordering
	if order != channeltypes.UNORDERED {
		return "", sdkerrors.Wrapf(channeltypes.ErrInvalidChannelOrdering, "invalid channel ordering: %s, expected UNORDERED", order.String())
	}

	// Validate version
	if version != types.Version {
		return "", sdkerrors.Wrapf(types.ErrInvalidVersion, "invalid version: %s, expected %s", version, types.Version)
	}

	// Claim channel capability
	if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return "", err
	}

	return version, nil
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCHandler) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *channeltypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	// Validate port ID
	if portID != types.ModuleName {
		return "", sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, types.ModuleName)
	}

	// Validate ordering
	if order != channeltypes.UNORDERED {
		return "", sdkerrors.Wrapf(channeltypes.ErrInvalidChannelOrdering, "invalid channel ordering: %s, expected UNORDERED", order.String())
	}

	// Validate version
	if counterpartyVersion != types.Version {
		return "", sdkerrors.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: %s, expected %s", counterpartyVersion, types.Version)
	}

	// Claim channel capability
	if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return "", err
	}

	return types.Version, nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCHandler) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	if counterpartyVersion != types.Version {
		return sdkerrors.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: %s, expected %s", counterpartyVersion, types.Version)
	}
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCHandler) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCHandler) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for bridge channels
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCHandler) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface
func (im IBCHandler) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	var data types.IBCTransferPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Sprintf("cannot unmarshal IBC transfer packet data: %s", err.Error()))
	}

	// Validate packet data
	if err := data.ValidateBasic(); err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	// Process the incoming transfer
	if err := im.processIncomingTransfer(ctx, packet, data); err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	return channeltypes.NewResultAcknowledgement([]byte{byte(1)})
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCHandler) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	var data types.IBCTransferPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	if err := im.processAcknowledgement(ctx, packet, data, ack); err != nil {
		return err
	}

	return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCHandler) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	var data types.IBCTransferPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	// Refund tokens on timeout
	if err := im.refundTokens(ctx, packet, data); err != nil {
		return err
	}

	return nil
}

// processIncomingTransfer processes an incoming IBC transfer
func (im IBCHandler) processIncomingTransfer(ctx sdk.Context, packet channeltypes.Packet, data types.IBCTransferPacketData) error {
	// Create IBC transfer record
	ibcTransfer := types.IBCTransferRecord{
		PacketSequence:     packet.GetSequence(),
		SourcePort:         packet.GetSourcePort(),
		SourceChannel:      packet.GetSourceChannel(),
		DestinationPort:    packet.GetDestPort(),
		DestinationChannel: packet.GetDestChannel(),
		TransferID:         data.TransferID,
		Sender:             data.Sender,
		Receiver:           data.Receiver,
		Token:              data.Token,
		TimeoutHeight:      packet.GetTimeoutHeight().GetRevisionHeight(),
		TimeoutTimestamp:   packet.GetTimeoutTimestamp(),
		Status:             types.IBCTransferStatusReceived,
		CreatedAt:          ctx.BlockTime(),
	}

	// Store IBC transfer record
	im.keeper.SetIBCTransferRecord(ctx, ibcTransfer)

	// Get recipient address
	recipientAddr, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver address: %s", err)
	}

	// Mint tokens to the recipient (for incoming transfers from other chains)
	if err := im.keeper.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(data.Token)); err != nil {
		return sdkerrors.Wrapf(err, "failed to mint coins")
	}

	// Send minted tokens to recipient
	moduleAddr := im.keeper.accountKeeper.GetModuleAddress(types.ModuleName)
	if err := im.keeper.bankKeeper.SendCoins(ctx, moduleAddr, recipientAddr, sdk.NewCoins(data.Token)); err != nil {
		return sdkerrors.Wrapf(err, "failed to send coins to recipient")
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"ibc_transfer_received",
			sdk.NewAttribute("transfer_id", data.TransferID),
			sdk.NewAttribute("sender", data.Sender),
			sdk.NewAttribute("receiver", data.Receiver),
			sdk.NewAttribute("token", data.Token.String()),
			sdk.NewAttribute("source_channel", packet.GetSourceChannel()),
		),
	)

	return nil
}

// processAcknowledgement processes an acknowledgement for an outgoing transfer
func (im IBCHandler) processAcknowledgement(ctx sdk.Context, packet channeltypes.Packet, data types.IBCTransferPacketData, ack channeltypes.Acknowledgement) error {
	// Update IBC transfer record status
	if ibcTransfer, found := im.keeper.GetIBCTransferRecord(ctx, packet.GetSequence()); found {
		if ack.Success() {
			ibcTransfer.Status = types.IBCTransferStatusReceived
		} else {
			ibcTransfer.Status = types.IBCTransferStatusFailed
		}
		im.keeper.SetIBCTransferRecord(ctx, ibcTransfer)
	}

	// If acknowledgement indicates failure, refund tokens
	if !ack.Success() {
		return im.refundTokens(ctx, packet, data)
	}

	return nil
}

// refundTokens refunds tokens when a transfer fails or times out
func (im IBCHandler) refundTokens(ctx sdk.Context, packet channeltypes.Packet, data types.IBCTransferPacketData) error {
	// Get sender address
	senderAddr, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %s", err)
	}

	// Refund tokens to sender
	moduleAddr := im.keeper.accountKeeper.GetModuleAddress(types.ModuleName)
	if err := im.keeper.bankKeeper.SendCoins(ctx, moduleAddr, senderAddr, sdk.NewCoins(data.Token)); err != nil {
		return sdkerrors.Wrapf(err, "failed to refund tokens")
	}

	// Update IBC transfer record status
	if ibcTransfer, found := im.keeper.GetIBCTransferRecord(ctx, packet.GetSequence()); found {
		ibcTransfer.Status = types.IBCTransferStatusTimeout
		im.keeper.SetIBCTransferRecord(ctx, ibcTransfer)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"ibc_transfer_refunded",
			sdk.NewAttribute("transfer_id", data.TransferID),
			sdk.NewAttribute("sender", data.Sender),
			sdk.NewAttribute("token", data.Token.String()),
			sdk.NewAttribute("reason", "timeout_or_failure"),
		),
	)

	return nil
}

// SendIBCTransfer sends an IBC transfer to another chain
func (k Keeper) SendIBCTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	token sdk.Coin,
	sender,
	receiver string,
	timeoutHeight uint64,
	timeoutTimestamp uint64,
	transferID string,
) error {
	// Create packet data
	packetData := types.IBCTransferPacketData{
		TransferID: transferID,
		Sender:     sender,
		Receiver:   receiver,
		Token:      token,
	}

	// Marshal packet data
	packetBytes, err := types.ModuleCdc.MarshalJSON(&packetData)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to marshal packet data")
	}

	// Create packet
	packet := channeltypes.NewPacket(
		packetBytes,
		k.ibcKeeper.ChannelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel),
		sourcePort,
		sourceChannel,
		types.ModuleName, // destination port
		"",               // destination channel will be filled by IBC
		channeltypes.NewHeight(0, timeoutHeight),
		timeoutTimestamp,
	)

	// Send packet
	if err := k.ibcKeeper.ChannelKeeper.SendPacket(ctx, k.ibcKeeper.ChannelKeeper.GetChannelCapability(ctx, sourcePort, sourceChannel), packet); err != nil {
		return sdkerrors.Wrapf(err, "failed to send IBC packet")
	}

	// Create IBC transfer record
	ibcTransfer := types.IBCTransferRecord{
		PacketSequence:     packet.GetSequence(),
		SourcePort:         sourcePort,
		SourceChannel:      sourceChannel,
		DestinationPort:    types.ModuleName,
		DestinationChannel: "", // Will be filled when packet is received
		TransferID:         transferID,
		Sender:             sender,
		Receiver:           receiver,
		Token:              token,
		TimeoutHeight:      timeoutHeight,
		TimeoutTimestamp:   timeoutTimestamp,
		Status:             types.IBCTransferStatusSent,
		CreatedAt:          ctx.BlockTime(),
	}

	// Store IBC transfer record
	k.SetIBCTransferRecord(ctx, ibcTransfer)

	return nil
}