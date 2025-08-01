package adapters

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"nexus-bridge/pkg/types"
)

// EthereumAdapter implements the ChainAdapter interface for Ethereum-based chains
type EthereumAdapter struct {
	config       types.ChainConfig
	client       *ethclient.Client
	rpcClient    *rpc.Client
	privateKey   *ecdsa.PrivateKey
	bridgeABI    abi.ABI
	connected    bool
	mu           sync.RWMutex
	lastBlock    uint64
	eventFilters map[string]ethereum.FilterQuery
}

// NewEthereumAdapter creates a new Ethereum chain adapter
func NewEthereumAdapter(privateKey *ecdsa.PrivateKey) *EthereumAdapter {
	return &EthereumAdapter{
		privateKey:   privateKey,
		eventFilters: make(map[string]ethereum.FilterQuery),
	}
}

// Connect establishes connection to the Ethereum blockchain
func (e *EthereumAdapter) Connect(ctx context.Context, config types.ChainConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid chain config: %w", err)
	}

	// Connect to Ethereum client
	client, err := ethclient.DialContext(ctx, config.RPC)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}

	// Connect to RPC client for advanced operations
	rpcClient, err := rpc.DialContext(ctx, config.RPC)
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to connect to RPC client: %w", err)
	}

	// Verify connection by getting chain ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		rpcClient.Close()
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	if chainID.Uint64() != uint64(config.ChainID) {
		client.Close()
		rpcClient.Close()
		return fmt.Errorf("chain ID mismatch: expected %d, got %d", config.ChainID, chainID.Uint64())
	}

	// Load bridge contract ABI
	bridgeABI, err := e.loadBridgeABI()
	if err != nil {
		client.Close()
		rpcClient.Close()
		return fmt.Errorf("failed to load bridge ABI: %w", err)
	}

	// Get current block number
	currentBlock, err := client.BlockNumber(ctx)
	if err != nil {
		client.Close()
		rpcClient.Close()
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	e.config = config
	e.client = client
	e.rpcClient = rpcClient
	e.bridgeABI = bridgeABI
	e.lastBlock = currentBlock
	e.connected = true

	// Setup event filters
	e.setupEventFilters()

	return nil
}

// ListenForEvents starts listening for blockchain events
func (e *EthereumAdapter) ListenForEvents(ctx context.Context, eventChan chan<- types.Event) error {
	e.mu.RLock()
	if !e.connected {
		e.mu.RUnlock()
		return fmt.Errorf("adapter not connected")
	}
	e.mu.RUnlock()

	// Start event polling goroutine
	go e.pollEvents(ctx, eventChan)

	return nil
}

// SubmitTransaction submits a transaction to the blockchain
func (e *EthereumAdapter) SubmitTransaction(ctx context.Context, tx types.Transaction) (*types.TxResult, error) {
	e.mu.RLock()
	if !e.connected {
		e.mu.RUnlock()
		return nil, fmt.Errorf("adapter not connected")
	}
	client := e.client
	e.mu.RUnlock()

	// Get nonce for the transaction
	nonce, err := e.getNonce(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Estimate gas if not provided
	gasLimit := tx.GasLimit
	if gasLimit == 0 {
		gasLimit, err = e.estimateGas(ctx, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas: %w", err)
		}
	}

	// Get gas price if not provided
	gasPrice := tx.GasPrice
	if gasPrice == nil || gasPrice.Cmp(big.NewInt(0)) == 0 {
		gasPrice, err = e.getGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get gas price: %w", err)
		}
	}

	// Create Ethereum transaction
	value := big.NewInt(0)
	if tx.Value != nil {
		value = tx.Value.Int
	}

	ethTx := ethtypes.NewTransaction(
		nonce,
		common.HexToAddress(tx.To),
		value,
		gasLimit,
		gasPrice.Int,
		tx.Data,
	)

	// Sign transaction
	chainID := big.NewInt(int64(e.config.ChainID))
	signedTx, err := ethtypes.SignTx(ethTx, ethtypes.NewEIP155Signer(chainID), e.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return &types.TxResult{
		TxHash: signedTx.Hash().Hex(),
	}, nil
}

// GetBlockConfirmations returns the number of confirmations for a transaction
func (e *EthereumAdapter) GetBlockConfirmations(ctx context.Context, txHash string) (uint64, error) {
	e.mu.RLock()
	if !e.connected {
		e.mu.RUnlock()
		return 0, fmt.Errorf("adapter not connected")
	}
	client := e.client
	e.mu.RUnlock()

	// Get transaction receipt
	receipt, err := client.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return 0, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Get current block number
	currentBlock, err := client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get current block number: %w", err)
	}

	// Calculate confirmations
	if currentBlock < receipt.BlockNumber.Uint64() {
		return 0, nil
	}

	return currentBlock - receipt.BlockNumber.Uint64() + 1, nil
}

// ValidateEvent validates the authenticity of a blockchain event
func (e *EthereumAdapter) ValidateEvent(ctx context.Context, event types.Event) error {
	e.mu.RLock()
	if !e.connected {
		e.mu.RUnlock()
		return fmt.Errorf("adapter not connected")
	}
	client := e.client
	e.mu.RUnlock()

	// Get transaction receipt to verify the event
	receipt, err := client.TransactionReceipt(ctx, common.HexToHash(event.TxHash))
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Verify transaction was successful
	if receipt.Status != 1 {
		return fmt.Errorf("transaction failed")
	}

	// Verify block number matches
	if receipt.BlockNumber.Uint64() != event.BlockNumber {
		return fmt.Errorf("block number mismatch")
	}

	// Verify the event exists in the transaction logs
	bridgeAddress := common.HexToAddress(e.config.BridgeContract)
	eventFound := false

	for _, log := range receipt.Logs {
		if log.Address == bridgeAddress {
			// Parse the log to verify it matches our event
			if e.validateEventLog(log, event) {
				eventFound = true
				break
			}
		}
	}

	if !eventFound {
		return fmt.Errorf("event not found in transaction logs")
	}

	return nil
}

// GetChainID returns the chain identifier
func (e *EthereumAdapter) GetChainID() types.ChainID {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config.ChainID
}

// IsConnected returns true if the adapter is connected to the blockchain
func (e *EthereumAdapter) IsConnected() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.connected
}

// Close closes the connection to the blockchain
func (e *EthereumAdapter) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.client != nil {
		e.client.Close()
	}
	if e.rpcClient != nil {
		e.rpcClient.Close()
	}

	e.connected = false
	return nil
}

// Private helper methods

// loadBridgeABI loads the bridge contract ABI
func (e *EthereumAdapter) loadBridgeABI() (abi.ABI, error) {
	// Bridge contract ABI (simplified for the core events)
	bridgeABIJSON := `[
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "transferId", "type": "bytes32"},
				{"indexed": true, "name": "user", "type": "address"},
				{"indexed": false, "name": "token", "type": "address"},
				{"indexed": false, "name": "amount", "type": "uint256"},
				{"indexed": false, "name": "destinationChain", "type": "uint256"},
				{"indexed": false, "name": "recipient", "type": "address"}
			],
			"name": "TokensLocked",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "transferId", "type": "bytes32"},
				{"indexed": true, "name": "recipient", "type": "address"},
				{"indexed": false, "name": "token", "type": "address"},
				{"indexed": false, "name": "amount", "type": "uint256"}
			],
			"name": "TokensUnlocked",
			"type": "event"
		}
	]`

	return abi.JSON(strings.NewReader(bridgeABIJSON))
}

// setupEventFilters sets up event filters for monitoring
func (e *EthereumAdapter) setupEventFilters() {
	bridgeAddress := common.HexToAddress(e.config.BridgeContract)

	// Filter for TokensLocked events
	e.eventFilters["TokensLocked"] = ethereum.FilterQuery{
		Addresses: []common.Address{bridgeAddress},
		Topics: [][]common.Hash{
			{crypto.Keccak256Hash([]byte("TokensLocked(bytes32,address,address,uint256,uint256,address)"))},
		},
	}

	// Filter for TokensUnlocked events
	e.eventFilters["TokensUnlocked"] = ethereum.FilterQuery{
		Addresses: []common.Address{bridgeAddress},
		Topics: [][]common.Hash{
			{crypto.Keccak256Hash([]byte("TokensUnlocked(bytes32,address,address,uint256)"))},
		},
	}
}

// pollEvents polls for new events
func (e *EthereumAdapter) pollEvents(ctx context.Context, eventChan chan<- types.Event) {
	ticker := time.NewTicker(time.Duration(e.config.BlockTime) / 2) // Poll twice per block
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := e.fetchAndProcessEvents(ctx, eventChan); err != nil {
				// Log error but continue polling
				fmt.Printf("Error fetching events: %v\n", err)
			}
		}
	}
}

// fetchAndProcessEvents fetches and processes new events
func (e *EthereumAdapter) fetchAndProcessEvents(ctx context.Context, eventChan chan<- types.Event) error {
	e.mu.RLock()
	client := e.client
	lastBlock := e.lastBlock
	e.mu.RUnlock()

	// Get current block number
	currentBlock, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	// Check for chain reorganization
	if err := e.detectReorganization(ctx, lastBlock); err != nil {
		return fmt.Errorf("chain reorganization detected: %w", err)
	}

	// Process events from last processed block to current block
	fromBlock := lastBlock + 1
	if fromBlock > currentBlock {
		return nil // No new blocks
	}

	// Fetch events for each filter
	for eventType, filter := range e.eventFilters {
		filter.FromBlock = big.NewInt(int64(fromBlock))
		filter.ToBlock = big.NewInt(int64(currentBlock))

		logs, err := client.FilterLogs(ctx, filter)
		if err != nil {
			return fmt.Errorf("failed to filter logs for %s: %w", eventType, err)
		}

		// Process each log
		for _, log := range logs {
			event, err := e.parseLogToEvent(log, eventType)
			if err != nil {
				fmt.Printf("Error parsing log to event: %v\n", err)
				continue
			}

			select {
			case eventChan <- *event:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// Update last processed block
	e.mu.Lock()
	e.lastBlock = currentBlock
	e.mu.Unlock()

	return nil
}

// detectReorganization detects chain reorganizations
func (e *EthereumAdapter) detectReorganization(ctx context.Context, lastBlock uint64) error {
	if lastBlock == 0 {
		return nil // No previous block to check
	}

	// Get the block at lastBlock to verify it still exists
	block, err := e.client.BlockByNumber(ctx, big.NewInt(int64(lastBlock)))
	if err != nil {
		// Block might have been reorganized
		return fmt.Errorf("block %d not found, possible reorganization", lastBlock)
	}

	// Additional checks could be added here to verify block hash consistency
	_ = block

	return nil
}

// parseLogToEvent parses an Ethereum log to a bridge event
func (e *EthereumAdapter) parseLogToEvent(log ethtypes.Log, eventType string) (*types.Event, error) {
	switch eventType {
	case "TokensLocked":
		return e.parseTokensLockedEvent(log)
	case "TokensUnlocked":
		return e.parseTokensUnlockedEvent(log)
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
}

// parseTokensLockedEvent parses a TokensLocked event
func (e *EthereumAdapter) parseTokensLockedEvent(log ethtypes.Log) (*types.Event, error) {
	// Parse the event using ABI
	event := struct {
		TransferID       [32]byte
		User             common.Address
		Token            common.Address
		Amount           *big.Int
		DestinationChain *big.Int
		Recipient        common.Address
	}{}

	if err := e.bridgeABI.UnpackIntoInterface(&event, "TokensLocked", log.Data); err != nil {
		return nil, fmt.Errorf("failed to unpack TokensLocked event: %w", err)
	}

	// Extract indexed parameters from topics
	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("insufficient topics for TokensLocked event")
	}

	transferID := log.Topics[1]
	user := common.BytesToAddress(log.Topics[2].Bytes())

	// Create transfer object
	transfer := types.Transfer{
		ID:               transferID.Hex(),
		SourceChain:      e.config.ChainID,
		DestinationChain: types.ChainID(event.DestinationChain.Uint64()),
		Token:            event.Token.Hex(),
		Amount:           types.NewBigInt(event.Amount),
		Sender:           user.Hex(),
		Recipient:        event.Recipient.Hex(),
		Status:           types.StatusPending,
		SourceTxHash:     log.TxHash.Hex(),
		BlockNumber:      log.BlockNumber,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return &types.Event{
		ID:          fmt.Sprintf("%s-%d", log.TxHash.Hex(), log.Index),
		Type:        types.EventTypeLock,
		ChainID:     e.config.ChainID,
		TxHash:      log.TxHash.Hex(),
		BlockNumber: log.BlockNumber,
		TransferID:  transferID.Hex(),
		Transfer:    transfer,
		Timestamp:   time.Now(),
		Raw:         log.Data,
	}, nil
}

// parseTokensUnlockedEvent parses a TokensUnlocked event
func (e *EthereumAdapter) parseTokensUnlockedEvent(log ethtypes.Log) (*types.Event, error) {
	// Parse the event using ABI
	event := struct {
		TransferID [32]byte
		Recipient  common.Address
		Token      common.Address
		Amount     *big.Int
	}{}

	if err := e.bridgeABI.UnpackIntoInterface(&event, "TokensUnlocked", log.Data); err != nil {
		return nil, fmt.Errorf("failed to unpack TokensUnlocked event: %w", err)
	}

	// Extract indexed parameters from topics
	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("insufficient topics for TokensUnlocked event")
	}

	transferID := log.Topics[1]
	recipient := common.BytesToAddress(log.Topics[2].Bytes())

	// Create transfer object (partial, as this is the destination event)
	transfer := types.Transfer{
		ID:                transferID.Hex(),
		DestinationChain:  e.config.ChainID,
		Token:             event.Token.Hex(),
		Amount:            types.NewBigInt(event.Amount),
		Recipient:         recipient.Hex(),
		Status:            types.StatusCompleted,
		DestinationTxHash: log.TxHash.Hex(),
		BlockNumber:       log.BlockNumber,
		UpdatedAt:         time.Now(),
	}

	return &types.Event{
		ID:          fmt.Sprintf("%s-%d", log.TxHash.Hex(), log.Index),
		Type:        types.EventTypeUnlock,
		ChainID:     e.config.ChainID,
		TxHash:      log.TxHash.Hex(),
		BlockNumber: log.BlockNumber,
		TransferID:  transferID.Hex(),
		Transfer:    transfer,
		Timestamp:   time.Now(),
		Raw:         log.Data,
	}, nil
}

// validateEventLog validates that a log matches the expected event
func (e *EthereumAdapter) validateEventLog(log *ethtypes.Log, event types.Event) bool {
	// Basic validation - check if transaction hash matches
	return log.TxHash.Hex() == event.TxHash
}

// getNonce gets the next nonce for transactions
func (e *EthereumAdapter) getNonce(ctx context.Context) (uint64, error) {
	address := crypto.PubkeyToAddress(e.privateKey.PublicKey)
	return e.client.PendingNonceAt(ctx, address)
}

// estimateGas estimates gas for a transaction
func (e *EthereumAdapter) estimateGas(ctx context.Context, tx types.Transaction) (uint64, error) {
	value := big.NewInt(0)
	if tx.Value != nil {
		value = tx.Value.Int
	}

	msg := ethereum.CallMsg{
		To:   &common.Address{},
		Data: tx.Data,
		Value: value,
	}

	// Set the To address
	if tx.To != "" {
		to := common.HexToAddress(tx.To)
		msg.To = &to
	}

	gasLimit, err := e.client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, err
	}

	// Add 20% buffer to gas estimate
	return gasLimit * 120 / 100, nil
}

// getGasPrice gets the current gas price
func (e *EthereumAdapter) getGasPrice(ctx context.Context) (*types.BigInt, error) {
	gasPrice, err := e.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	return types.NewBigInt(gasPrice), nil
}