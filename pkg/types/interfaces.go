package types

import (
	"context"
	"math/big"
)

// ChainAdapter defines the interface for blockchain interactions
type ChainAdapter interface {
	// Connect establishes connection to the blockchain
	Connect(ctx context.Context, config ChainConfig) error
	
	// ListenForEvents starts listening for blockchain events
	ListenForEvents(ctx context.Context, eventChan chan<- Event) error
	
	// SubmitTransaction submits a transaction to the blockchain
	SubmitTransaction(ctx context.Context, tx Transaction) (*TxResult, error)
	
	// GetBlockConfirmations returns the number of confirmations for a transaction
	GetBlockConfirmations(ctx context.Context, txHash string) (uint64, error)
	
	// ValidateEvent validates the authenticity of a blockchain event
	ValidateEvent(ctx context.Context, event Event) error
	
	// GetChainID returns the chain identifier
	GetChainID() ChainID
	
	// IsConnected returns true if the adapter is connected to the blockchain
	IsConnected() bool
	
	// Close closes the connection to the blockchain
	Close() error
}

// SignatureValidator handles multi-signature operations
type SignatureValidator interface {
	// SignTransfer generates a signature for a transfer
	SignTransfer(transferID string, transfer Transfer) (*Signature, error)
	
	// ValidateSignatures validates a collection of signatures for a transfer
	ValidateSignatures(transferID string, signatures []Signature) error
	
	// GetRequiredThreshold returns the minimum number of signatures required
	GetRequiredThreshold() uint64
	
	// GetRelayerAddress returns the address of this relayer
	GetRelayerAddress() string
	
	// IsAuthorizedRelayer checks if an address is an authorized relayer
	IsAuthorizedRelayer(address string) bool
	
	// RotateKey rotates the relayer's signing key
	RotateKey(newPrivateKey []byte) error
}

// StateManager tracks transfer states and prevents double-spending
type StateManager interface {
	// RecordTransfer records a new transfer in the database
	RecordTransfer(ctx context.Context, transfer Transfer) error
	
	// GetTransferStatus returns the current status of a transfer
	GetTransferStatus(ctx context.Context, transferID string) (*TransferStatus, error)
	
	// UpdateTransferStatus updates the status of a transfer
	UpdateTransferStatus(ctx context.Context, transferID string, status TransferStatus) error
	
	// MarkTransferComplete marks a transfer as completed
	MarkTransferComplete(ctx context.Context, transferID string, destinationTxHash string) error
	
	// IsTransferProcessed checks if a transfer has already been processed
	IsTransferProcessed(ctx context.Context, transferID string) (bool, error)
	
	// GetTransfersInBlockRange returns transfers affected by a block range (for reorg handling)
	GetTransfersInBlockRange(ctx context.Context, chainID ChainID, fromBlock, toBlock uint64) ([]Transfer, error)
	
	// RecordSignature records a signature for a transfer
	RecordSignature(ctx context.Context, transferID string, signature Signature) error
	
	// GetSignatures returns all signatures for a transfer
	GetSignatures(ctx context.Context, transferID string) ([]Signature, error)
	
	// MarkTransferForReview marks a transfer for manual review
	MarkTransferForReview(ctx context.Context, transferID string, reason string) error
}

// EventListener defines the interface for event processing
type EventListener interface {
	// Start begins listening for events
	Start(ctx context.Context) error
	
	// Stop stops listening for events
	Stop() error
	
	// RegisterHandler registers an event handler
	RegisterHandler(eventType EventType, handler EventHandler) error
}

// EventHandler defines the interface for handling specific events
type EventHandler interface {
	// Handle processes an event
	Handle(ctx context.Context, event Event) error
	
	// GetEventType returns the type of events this handler processes
	GetEventType() EventType
}

// FeeCalculator defines the interface for fee calculation
type FeeCalculator interface {
	// EstimateFee estimates the fee for a cross-chain transfer
	EstimateFee(ctx context.Context, transfer Transfer) (*FeeEstimate, error)
	
	// GetGasPrice returns the current gas price for a chain
	GetGasPrice(ctx context.Context, chainID ChainID) (*big.Int, error)
	
	// ValidateFee validates that the provided fee is sufficient
	ValidateFee(ctx context.Context, transfer Transfer, providedFee *big.Int) error
}