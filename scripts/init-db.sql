-- Initialize NexusBridge database schema

-- Create transfers table
CREATE TABLE IF NOT EXISTS transfers (
    id VARCHAR(66) PRIMARY KEY,
    source_chain INTEGER NOT NULL,
    destination_chain INTEGER NOT NULL,
    token VARCHAR(42) NOT NULL,
    amount DECIMAL(78,0) NOT NULL,
    sender VARCHAR(42) NOT NULL,
    recipient VARCHAR(42) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    source_tx_hash VARCHAR(66),
    destination_tx_hash VARCHAR(66),
    block_number BIGINT,
    confirmations INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create signatures table
CREATE TABLE IF NOT EXISTS signatures (
    id SERIAL PRIMARY KEY,
    transfer_id VARCHAR(66) REFERENCES transfers(id),
    relayer_address VARCHAR(42) NOT NULL,
    signature BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(transfer_id, relayer_address)
);

-- Create supported tokens table
CREATE TABLE IF NOT EXISTS supported_tokens (
    id SERIAL PRIMARY KEY,
    chain_id INTEGER NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    decimals INTEGER NOT NULL,
    is_native BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(chain_id, token_address)
);

-- Create relayers table
CREATE TABLE IF NOT EXISTS relayers (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) UNIQUE NOT NULL,
    public_key VARCHAR(132) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create chain configurations table
CREATE TABLE IF NOT EXISTS chain_configs (
    id SERIAL PRIMARY KEY,
    chain_id INTEGER UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(20) NOT NULL, -- ethereum, cosmos
    rpc_url VARCHAR(255) NOT NULL,
    wss_url VARCHAR(255),
    bridge_contract VARCHAR(42),
    required_confirmations INTEGER DEFAULT 12,
    block_time_ms INTEGER DEFAULT 12000,
    gas_limit BIGINT DEFAULT 21000,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default chain configurations
INSERT INTO chain_configs (chain_id, name, type, rpc_url, required_confirmations, block_time_ms) VALUES
(1, 'Ethereum Mainnet', 'ethereum', 'http://localhost:8545', 12, 12000),
(137, 'Polygon', 'ethereum', 'http://localhost:8546', 20, 2000),
(31337, 'Hardhat Local', 'ethereum', 'http://localhost:8545', 1, 1000)
ON CONFLICT (chain_id) DO NOTHING;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_transfers_status ON transfers(status);
CREATE INDEX IF NOT EXISTS idx_transfers_source_chain ON transfers(source_chain);
CREATE INDEX IF NOT EXISTS idx_transfers_destination_chain ON transfers(destination_chain);
CREATE INDEX IF NOT EXISTS idx_transfers_created_at ON transfers(created_at);
CREATE INDEX IF NOT EXISTS idx_signatures_transfer_id ON signatures(transfer_id);
CREATE INDEX IF NOT EXISTS idx_supported_tokens_chain_id ON supported_tokens(chain_id);
CREATE INDEX IF NOT EXISTS idx_supported_tokens_enabled ON supported_tokens(enabled);