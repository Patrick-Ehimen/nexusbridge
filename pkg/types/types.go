package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// ChainID represents a blockchain network identifier
type ChainID uint64

const (
	ChainEthereum ChainID = 1
	ChainPolygon  ChainID = 137
	ChainCosmos   ChainID = 118
)

// String returns the string representation of ChainID
func (c ChainID) String() string {
	switch c {
	case ChainEthereum:
		return "ethereum"
	case ChainPolygon:
		return "polygon"
	case ChainCosmos:
		return "cosmos"
	default:
		return fmt.Sprintf("unknown-%d", uint64(c))
	}
}

// ChainType represents the type of blockchain
type ChainType string

const (
	ChainTypeEthereum ChainType = "ethereum"
	ChainTypeCosmos   ChainType = "cosmos"
)

// TransferStatus represents the status of a cross-chain transfer
type TransferStatus string

const (
	StatusPending     TransferStatus = "pending"
	StatusConfirming  TransferStatus = "confirming"
	StatusSigned      TransferStatus = "signed"
	StatusExecuting   TransferStatus = "executing"
	StatusCompleted   TransferStatus = "completed"
	StatusFailed      TransferStatus = "failed"
	StatusUnderReview TransferStatus = "under_review"
)

// EventType represents the type of blockchain event
type EventType string

const (
	EventTypeLock   EventType = "lock"
	EventTypeUnlock EventType = "unlock"
	EventTypeMint   EventType = "mint"
	EventTypeBurn   EventType = "burn"
)

// Transfer represents a cross-chain transfer
type Transfer struct {
	ID                string         `json:"id" db:"id" validate:"required"`
	SourceChain       ChainID        `json:"source_chain" db:"source_chain" validate:"required"`
	DestinationChain  ChainID        `json:"destination_chain" db:"destination_chain" validate:"required"`
	Token             string         `json:"token" db:"token" validate:"required,eth_addr"`
	Amount            *BigInt        `json:"amount" db:"amount" validate:"required"`
	Sender            string         `json:"sender" db:"sender" validate:"required"`
	Recipient         string         `json:"recipient" db:"recipient" validate:"required"`
	Status            TransferStatus `json:"status" db:"status" validate:"required"`
	SourceTxHash      string         `json:"source_tx_hash" db:"source_tx_hash"`
	DestinationTxHash string         `json:"destination_tx_hash" db:"destination_tx_hash"`
	BlockNumber       uint64         `json:"block_number" db:"block_number"`
	Confirmations     uint64         `json:"confirmations" db:"confirmations"`
	Fee               *BigInt        `json:"fee" db:"fee"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at" db:"updated_at"`
}

// Validate validates the transfer data
func (t *Transfer) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("transfer ID is required")
	}
	if t.SourceChain == t.DestinationChain {
		return fmt.Errorf("source and destination chains must be different")
	}
	if t.Amount == nil || t.Amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if t.Token == "" {
		return fmt.Errorf("token address is required")
	}
	if t.Sender == "" {
		return fmt.Errorf("sender address is required")
	}
	if t.Recipient == "" {
		return fmt.Errorf("recipient address is required")
	}
	return nil
}

// ChainConfig represents the configuration for a blockchain
type ChainConfig struct {
	ChainID               ChainID       `json:"chain_id" validate:"required"`
	Name                  string        `json:"name" validate:"required"`
	Type                  ChainType     `json:"type" validate:"required"`
	RPC                   string        `json:"rpc" validate:"required,url"`
	WSS                   string        `json:"wss,omitempty" validate:"omitempty,url"`
	BridgeContract        string        `json:"bridge_contract" validate:"required"`
	RequiredConfirmations uint64        `json:"required_confirmations" validate:"min=1"`
	BlockTime             time.Duration `json:"block_time" validate:"required"`
	GasLimit              uint64        `json:"gas_limit" validate:"min=21000"`
	GasPrice              *BigInt       `json:"gas_price"`
	Enabled               bool          `json:"enabled"`
}

// Validate validates the chain configuration
func (c *ChainConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("chain name is required")
	}
	if c.RPC == "" {
		return fmt.Errorf("RPC URL is required")
	}
	if c.BridgeContract == "" {
		return fmt.Errorf("bridge contract address is required")
	}
	if c.RequiredConfirmations == 0 {
		return fmt.Errorf("required confirmations must be greater than 0")
	}
	if c.BlockTime <= 0 {
		return fmt.Errorf("block time must be positive")
	}
	return nil
}

// Event represents a blockchain event
type Event struct {
	ID          string    `json:"id"`
	Type        EventType `json:"type"`
	ChainID     ChainID   `json:"chain_id"`
	TxHash      string    `json:"tx_hash"`
	BlockNumber uint64    `json:"block_number"`
	TransferID  string    `json:"transfer_id"`
	Transfer    Transfer  `json:"transfer"`
	Timestamp   time.Time `json:"timestamp"`
	Raw         []byte    `json:"raw,omitempty"`
}

// Signature represents a cryptographic signature
type Signature struct {
	RelayerAddress string    `json:"relayer_address" db:"relayer_address" validate:"required"`
	Signature      []byte    `json:"signature" db:"signature" validate:"required"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Transaction represents a blockchain transaction
type Transaction struct {
	To       string  `json:"to" validate:"required"`
	Data     []byte  `json:"data"`
	Value    *BigInt `json:"value"`
	GasLimit uint64  `json:"gas_limit"`
	GasPrice *BigInt `json:"gas_price"`
	Nonce    uint64  `json:"nonce"`
}

// TxResult represents the result of a transaction submission
type TxResult struct {
	TxHash      string `json:"tx_hash"`
	BlockNumber uint64 `json:"block_number"`
	GasUsed     uint64 `json:"gas_used"`
	Status      bool   `json:"status"`
}

// SupportedToken represents a token supported by the bridge
type SupportedToken struct {
	ID           int     `json:"id" db:"id"`
	ChainID      ChainID `json:"chain_id" db:"chain_id" validate:"required"`
	TokenAddress string  `json:"token_address" db:"token_address" validate:"required"`
	Name         string  `json:"name" db:"name" validate:"required"`
	Symbol       string  `json:"symbol" db:"symbol" validate:"required"`
	Decimals     uint8   `json:"decimals" db:"decimals" validate:"max=18"`
	IsNative     bool    `json:"is_native" db:"is_native"`
	Enabled      bool    `json:"enabled" db:"enabled"`
}

// FeeEstimate represents fee estimation for a transfer
type FeeEstimate struct {
	SourceChainFee      *BigInt `json:"source_chain_fee"`
	DestinationChainFee *BigInt `json:"destination_chain_fee"`
	RelayerFee          *BigInt `json:"relayer_fee"`
	TotalFee            *BigInt `json:"total_fee"`
	EstimatedTime       int64   `json:"estimated_time_seconds"`
}

// BigInt wraps big.Int for JSON serialization and database storage
type BigInt struct {
	*big.Int
}

// NewBigInt creates a new BigInt
func NewBigInt(i *big.Int) *BigInt {
	if i == nil {
		return &BigInt{big.NewInt(0)}
	}
	return &BigInt{new(big.Int).Set(i)}
}

// NewBigIntFromInt64 creates a new BigInt from int64
func NewBigIntFromInt64(i int64) *big.Int {
	return big.NewInt(i)
}

// MarshalJSON implements json.Marshaler
func (b *BigInt) MarshalJSON() ([]byte, error) {
	if b.Int == nil {
		return json.Marshal("0")
	}
	return json.Marshal(b.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (b *BigInt) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return fmt.Errorf("invalid big integer: %s", s)
	}
	
	b.Int = i
	return nil
}

// Value implements driver.Valuer for database storage
func (b *BigInt) Value() (driver.Value, error) {
	if b.Int == nil {
		return "0", nil
	}
	return b.String(), nil
}

// Scan implements sql.Scanner for database retrieval
func (b *BigInt) Scan(value interface{}) error {
	if value == nil {
		b.Int = big.NewInt(0)
		return nil
	}
	
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("cannot scan %T into BigInt", value)
	}
	
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return fmt.Errorf("invalid big integer: %s", s)
	}
	
	b.Int = i
	return nil
}