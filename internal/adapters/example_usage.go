package adapters

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"nexus-bridge/pkg/types"

	"github.com/ethereum/go-ethereum/crypto"
)

// ExampleEthereumAdapterUsage demonstrates how to use the EthereumAdapter
func ExampleEthereumAdapterUsage() {
	// Generate a private key for the relayer (in production, load from secure storage)
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// Create the adapter
	adapter := NewEthereumAdapter(privateKey)

	// Configure the chain connection
	config := types.ChainConfig{
		ChainID:               types.ChainEthereum,
		Name:                  "Ethereum Mainnet",
		Type:                  types.ChainTypeEthereum,
		RPC:                   "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
		WSS:                   "wss://mainnet.infura.io/ws/v3/YOUR_PROJECT_ID",
		BridgeContract:        "0x1234567890123456789012345678901234567890", // Your bridge contract address
		RequiredConfirmations: 12,
		BlockTime:             15 * time.Second,
		GasLimit:              300000,
		GasPrice:              types.NewBigInt(big.NewInt(20000000000)), // 20 gwei
		Enabled:               true,
	}

	// Connect to the blockchain
	ctx := context.Background()
	if err := adapter.Connect(ctx, config); err != nil {
		log.Fatalf("Failed to connect to Ethereum: %v", err)
	}
	defer adapter.Close()

	fmt.Printf("Connected to %s (Chain ID: %d)\n", config.Name, adapter.GetChainID())

	// Start listening for events
	eventChan := make(chan types.Event, 100)
	if err := adapter.ListenForEvents(ctx, eventChan); err != nil {
		log.Fatalf("Failed to start event listener: %v", err)
	}

	// Example: Submit a transaction
	tx := types.Transaction{
		To:       config.BridgeContract,
		Data:     []byte("example transaction data"),
		Value:    types.NewBigInt(big.NewInt(0)),
		GasLimit: 100000,
	}

	result, err := adapter.SubmitTransaction(ctx, tx)
	if err != nil {
		log.Printf("Failed to submit transaction: %v", err)
	} else {
		fmt.Printf("Transaction submitted: %s\n", result.TxHash)

		// Wait for confirmations
		for {
			confirmations, err := adapter.GetBlockConfirmations(ctx, result.TxHash)
			if err != nil {
				log.Printf("Failed to get confirmations: %v", err)
				break
			}

			fmt.Printf("Transaction %s has %d confirmations\n", result.TxHash, confirmations)

			if confirmations >= config.RequiredConfirmations {
				fmt.Printf("Transaction confirmed with %d confirmations\n", confirmations)
				break
			}

			time.Sleep(time.Duration(config.BlockTime))
		}
	}

	// Listen for events (in a real application, this would run in a goroutine)
	fmt.Println("Listening for bridge events...")
	select {
	case event := <-eventChan:
		fmt.Printf("Received event: %s (Type: %s, Chain: %d)\n", 
			event.ID, event.Type, event.ChainID)
		
		// Validate the event
		if err := adapter.ValidateEvent(ctx, event); err != nil {
			log.Printf("Event validation failed: %v", err)
		} else {
			fmt.Printf("Event validated successfully\n")
		}

	case <-time.After(30 * time.Second):
		fmt.Println("No events received within timeout")
	}
}

// ExampleCreateEthereumAdapterWithConfig shows how to create an adapter with custom configuration
func ExampleCreateEthereumAdapterWithConfig(privateKeyHex string, rpcURL string, bridgeContract string) (*EthereumAdapter, error) {
	// Parse private key from hex string
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Create adapter
	adapter := NewEthereumAdapter(privateKey)

	// Configure for testnet
	config := types.ChainConfig{
		ChainID:               types.ChainEthereum, // Use appropriate chain ID
		Name:                  "Ethereum Testnet",
		Type:                  types.ChainTypeEthereum,
		RPC:                   rpcURL,
		BridgeContract:        bridgeContract,
		RequiredConfirmations: 3, // Lower confirmations for testnet
		BlockTime:             15 * time.Second,
		GasLimit:              300000,
		GasPrice:              types.NewBigInt(big.NewInt(10000000000)), // 10 gwei
		Enabled:               true,
	}

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := adapter.Connect(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return adapter, nil
}

// ExampleEventProcessing shows how to process different types of bridge events
func ExampleEventProcessing(adapter *EthereumAdapter) {
	ctx := context.Background()
	eventChan := make(chan types.Event, 100)

	// Start listening
	if err := adapter.ListenForEvents(ctx, eventChan); err != nil {
		log.Fatalf("Failed to start event listener: %v", err)
	}

	// Process events
	for event := range eventChan {
		switch event.Type {
		case types.EventTypeLock:
			fmt.Printf("Lock event detected: %s\n", event.TransferID)
			fmt.Printf("  From: %s\n", event.Transfer.Sender)
			fmt.Printf("  To: %s (Chain %d)\n", event.Transfer.Recipient, event.Transfer.DestinationChain)
			fmt.Printf("  Amount: %s %s\n", event.Transfer.Amount.String(), event.Transfer.Token)

			// Validate the event
			if err := adapter.ValidateEvent(ctx, event); err != nil {
				log.Printf("Lock event validation failed: %v", err)
				continue
			}

			// Process the lock event (e.g., initiate cross-chain transfer)
			fmt.Printf("Processing lock event for transfer %s\n", event.TransferID)

		case types.EventTypeUnlock:
			fmt.Printf("Unlock event detected: %s\n", event.TransferID)
			fmt.Printf("  Recipient: %s\n", event.Transfer.Recipient)
			fmt.Printf("  Amount: %s %s\n", event.Transfer.Amount.String(), event.Transfer.Token)

			// Validate the event
			if err := adapter.ValidateEvent(ctx, event); err != nil {
				log.Printf("Unlock event validation failed: %v", err)
				continue
			}

			// Process the unlock event (e.g., mark transfer as completed)
			fmt.Printf("Processing unlock event for transfer %s\n", event.TransferID)

		default:
			fmt.Printf("Unknown event type: %s\n", event.Type)
		}
	}
}