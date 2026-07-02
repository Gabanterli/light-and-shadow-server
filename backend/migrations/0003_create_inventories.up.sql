-- Migration 0003: Criar tabela de inventários de itens (Inventories)
CREATE TABLE IF NOT EXISTS inventories (
    id SERIAL PRIMARY KEY,
    character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    slot_index INT NOT NULL,
    item_id VARCHAR(64) NOT NULL,
    quantity INT DEFAULT 1,
    durability INT DEFAULT 100,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(character_id, slot_index)
);

CREATE INDEX IF NOT EXISTS idx_inventories_character_id ON inventories(character_id);
