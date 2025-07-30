-- Migration: 001_initial_schema.sql
-- Description: Create initial database schema for NexusBridge
-- Created: 2025-01-30

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
    fee DECIMAL(78,0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_different_chains CHECK (source_chain != destination_chain),
    CONSTRAINT chk_positive_amount CHECK (amount > 0),
    CONSTRAINT chk_valid_status CHECK (status IN ('pending', 'confirming', 'signed', 'executing', 'completed', 'failed', 'under_review')),
    CONSTRAINT chk_non_negative_confirmations CHECK (confirmations >= 0)
);

-- Create indexes for transfers table
CREATE INDEX IF NOT EXISTS idx_transfers_source_chain ON transfers(source_chain);
CREATE INDEX IF NOT EXISTS idx_transfers_destination_chain ON transfers(destination_chain);
CREATE INDEX IF NOT EXISTS idx_transfers_status ON transfers(status);
CREATE INDEX IF NOT EXISTS idx_transfers_created_at ON transfers(created_at);
CREATE INDEX IF NOT EXISTS idx_transfers_block_number ON transfers(source_chain, block_number);
CREATE INDEX IF NOT EXISTS idx_transfers_source_tx_hash ON transfers(source_tx_hash);
CREATE INDEX IF NOT EXISTS idx_transfers_destination_tx_hash ON transfers(destination_tx_hash);

-- Create signatures table
CREATE TABLE IF NOT EXISTS signatures (
    id SERIAL PRIMARY KEY,
    transfer_id VARCHAR(66) NOT NULL REFERENCES transfers(id) ON DELETE CASCADE,
    relayer_address VARCHAR(42) NOT NULL,
    signature BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure one signature per relayer per transfer
    CONSTRAINT uq_signatures_transfer_relayer UNIQUE (transfer_id, relayer_address)
);

-- Create indexes for signatures table
CREATE INDEX IF NOT EXISTS idx_signatures_transfer_id ON signatures(transfer_id);
CREATE INDEX IF NOT EXISTS idx_signatures_relayer_address ON signatures(relayer_address);
CREATE INDEX IF NOT EXISTS idx_signatures_created_at ON signatures(created_at);

-- Create supported_tokens table
CREATE TABLE IF NOT EXISTS supported_tokens (
    id SERIAL PRIMARY KEY,
    chain_id INTEGER NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    decimals INTEGER NOT NULL DEFAULT 18,
    is_native BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_valid_decimals CHECK (decimals >= 0 AND decimals <= 18),
    CONSTRAINT uq_supported_tokens_chain_address UNIQUE (chain_id, token_address)
);

-- Create indexes for supported_tokens table
CREATE INDEX IF NOT EXISTS idx_supported_tokens_chain_id ON supported_tokens(chain_id);
CREATE INDEX IF NOT EXISTS idx_supported_tokens_enabled ON supported_tokens(enabled);
CREATE INDEX IF NOT EXISTS idx_supported_tokens_symbol ON supported_tokens(symbol);

-- Create relayer_config table for storing relayer configuration
CREATE TABLE IF NOT EXISTS relayer_config (
    id SERIAL PRIMARY KEY,
    relayer_address VARCHAR(42) NOT NULL UNIQUE,
    public_key BYTEA NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    threshold_weight INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_positive_threshold_weight CHECK (threshold_weight > 0)
);

-- Create indexes for relayer_config table
CREATE INDEX IF NOT EXISTS idx_relayer_config_address ON relayer_config(relayer_address);
CREATE INDEX IF NOT EXISTS idx_relayer_config_active ON relayer_config(is_active);

-- Create chain_config table for storing chain configurations
CREATE TABLE IF NOT EXISTS chain_config (
    chain_id INTEGER PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(20) NOT NULL,
    rpc_url VARCHAR(255) NOT NULL,
    wss_url VARCHAR(255),
    bridge_contract VARCHAR(42) NOT NULL,
    required_confirmations INTEGER NOT NULL DEFAULT 12,
    block_time_ms INTEGER NOT NULL DEFAULT 12000,
    gas_limit BIGINT NOT NULL DEFAULT 21000,
    gas_price DECIMAL(78,0),
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_valid_chain_type CHECK (type IN ('ethereum', 'cosmos')),
    CONSTRAINT chk_positive_confirmations CHECK (required_confirmations > 0),
    CONSTRAINT chk_positive_block_time CHECK (block_time_ms > 0),
    CONSTRAINT chk_positive_gas_limit CHECK (gas_limit > 0)
);

-- Create indexes for chain_config table
CREATE INDEX IF NOT EXISTS idx_chain_config_enabled ON chain_config(enabled);
CREATE INDEX IF NOT EXISTS idx_chain_config_type ON chain_config(type);

-- Create audit_log table for tracking important events
CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(66) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    performed_by VARCHAR(42),
    performed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);

-- Create indexes for audit_log table
CREATE INDEX IF NOT EXISTS idx_audit_log_event_type ON audit_log(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_performed_at ON audit_log(performed_at);
CREATE INDEX IF NOT EXISTS idx_audit_log_performed_by ON audit_log(performed_by);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at columns
CREATE TRIGGER update_transfers_updated_at 
    BEFORE UPDATE ON transfers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_supported_tokens_updated_at 
    BEFORE UPDATE ON supported_tokens 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_relayer_config_updated_at 
    BEFORE UPDATE ON relayer_config 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chain_config_updated_at 
    BEFORE UPDATE ON chain_config 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default chain configurations
INSERT INTO chain_config (chain_id, name, type, rpc_url, bridge_contract, required_confirmations, block_time_ms, gas_limit) VALUES
(1, 'Ethereum Mainnet', 'ethereum', 'https://mainnet.infura.io/v3/YOUR_PROJECT_ID', '0x0000000000000000000000000000000000000000', 12, 12000, 21000),
(137, 'Polygon', 'ethereum', 'https://polygon-rpc.com', '0x0000000000000000000000000000000000000000', 20, 2000, 21000),
(118, 'Cosmos Hub', 'cosmos', 'https://rpc.cosmos.network:443', 'cosmos1bridge', 1, 6000, 200000)
ON CONFLICT (chain_id) DO NOTHING;

-- Insert some common tokens (these would be updated with real addresses)
INSERT INTO supported_tokens (chain_id, token_address, name, symbol, decimals, is_native) VALUES
(1, '0x0000000000000000000000000000000000000000', 'Ether', 'ETH', 18, true),
(1, '0xA0b86a33E6441E6C7D3E4C2C4C6C6C6C6C6C6C6C', 'USD Coin', 'USDC', 6, false),
(1, '0xdAC17F958D2ee523a2206206994597C13D831ec7', 'Tether USD', 'USDT', 6, false),
(137, '0x0000000000000000000000000000000000000000', 'Polygon', 'MATIC', 18, true),
(137, '0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174', 'USD Coin (PoS)', 'USDC', 6, false),
(118, 'uatom', 'Cosmos', 'ATOM', 6, true)
ON CONFLICT (chain_id, token_address) DO NOTHING;