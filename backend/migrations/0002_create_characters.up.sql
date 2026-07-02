-- Migration 0002: Criar tabela de personagens (Characters)
CREATE TABLE IF NOT EXISTS characters (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name VARCHAR(32) UNIQUE NOT NULL,
    class VARCHAR(20) NOT NULL,
    level INT DEFAULT 1,
    experience BIGINT DEFAULT 0,
    posX FLOAT DEFAULT 0.0,
    posY FLOAT DEFAULT 0.0,
    posZ FLOAT DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_characters_account_id ON characters(account_id);
CREATE INDEX IF NOT EXISTS idx_characters_name ON characters(name);
