# Blockchain Adapters

This package contains blockchain adapter implementations for the Nexus Bridge relayer service. Each adapter implements the `ChainAdapter` interface defined in `pkg/types/interfaces.go`.

## Ethereum Adapter

The `EthereumAdapter` provides integration with Ethereum-based blockchains (Ethereum, Polygon, BSC, etc.).

### Features

- **Event Listening**: Monitors bridge contract events (Lock/Unlock) using `FilterLogs`
- **Transaction Submission**: Submits transactions with proper gas estimation and nonce management
- **Block Confirmation Tracking**: Tracks transaction confirmations and handles chain reorganizations
- **Event Validation**: Validates event authenticity by checking transaction receipts
- **Connection Management**: Handles RPC connections with automatic reconnection

### Usage

```go
import (
    "github.com/ethereum/go-ethereum/crypto"
    "nexus-bridge/internal/adapters"
    "nexus-bridge/pkg/types"
)

// Create adapter with private key
privateKey, _ := crypto.GenerateKey()
adapter := adapters.NewEthereumAdapter(privateKey)

// Configure connection
config := types.ChainConfig{
    ChainID:               types.ChainEthereum,
    Name:                  "Ethereum Mainnet",
    Type:                  types.ChainTypeEthereum,
    RPC:                   "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
    BridgeContract:        "0x1234567890123456789012345678901234567890",
    RequiredConfirmations: 12,
    BlockTime:             15 * time.Second,
    GasLimit:              300000,
    Enabled:               true,
}

// Connect and start listening
ctx := context.Background()
err := adapter.Connect(ctx, config)
if err != nil {
    log.Fatal(err)
}

eventChan := make(chan types.Event, 100)
err = adapter.ListenForEvents(ctx, eventChan)
if err != nil {
    log.Fatal(err)
}

// Process events
for event := range eventChan {
    fmt.Printf("Received event: %s\n", event.ID)
}
```

### Event Types

The adapter monitors the following bridge contract events:

#### TokensLocked Event

```solidity
event TokensLocked(
    bytes32 indexed transferId,
    address indexed user,
    address token,
    uint256 amount,
    uint256 destinationChain,
    address recipient
);
```

#### TokensUnlocked Event

```solidity
event TokensUnlocked(
    bytes32 indexed transferId,
    address indexed recipient,
    address token,
    uint256 amount
);
```

### Configuration

The adapter requires the following configuration parameters:

- `ChainID`: The blockchain network identifier
- `RPC`: HTTP RPC endpoint URL
- `WSS`: WebSocket endpoint URL (optional, for real-time events)
- `BridgeContract`: Address of the bridge contract
- `RequiredConfirmations`: Number of block confirmations required
- `BlockTime`: Average block time for the network
- `GasLimit`: Default gas limit for transactions
- `GasPrice`: Default gas price (optional, will use network suggestion if not provided)

### Error Handling

The adapter handles various error conditions:

- **Connection Failures**: Automatic reconnection attempts
- **Chain Reorganizations**: Detection and handling of block reorgs
- **Transaction Failures**: Proper error reporting with gas estimation
- **Event Validation**: Verification of event authenticity

### Testing

The adapter includes comprehensive unit tests with mocked Ethereum clients:

```bash
# Run tests
go test ./internal/adapters -v

# Run benchmarks
go test ./internal/adapters -bench=.

# Run with coverage
go test ./internal/adapters -cover
```

### Integration Tests

Integration tests are available but skipped by default (require real Ethereum node):

```bash
# Run integration tests (requires test network access)
go test ./internal/adapters -v -tags=integration
```

### Performance

The adapter is optimized for high-throughput scenarios:

- **Event Polling**: Configurable polling interval (default: half block time)
- **Batch Processing**: Processes multiple events in single RPC calls
- **Connection Pooling**: Reuses connections for multiple operations
- **Gas Optimization**: Intelligent gas estimation with safety buffers

### Security Considerations

- **Private Key Management**: Private keys should be stored securely (HSM, encrypted storage)
- **RPC Endpoints**: Use trusted RPC providers with proper authentication
- **Event Validation**: All events are validated against transaction receipts
- **Nonce Management**: Proper nonce tracking prevents transaction conflicts

### Monitoring

The adapter provides metrics and logging for operational monitoring:

- Connection status
- Event processing rates
- Transaction success/failure rates
- Gas usage statistics
- Block confirmation times

### Future Enhancements

Planned improvements include:

- WebSocket support for real-time events
- Multi-node failover for high availability
- Advanced gas price strategies (EIP-1559)
- Event replay capabilities for disaster recovery
- Prometheus metrics integration
