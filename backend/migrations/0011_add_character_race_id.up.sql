-- Migration 0011: Adicionar a coluna race_id à tabela de personagens
-- Adiciona a coluna para armazenar a raça oficial do personagem, com um
-- default seguro ('human') para garantir compatibilidade com dados legados.
ALTER TABLE characters ADD COLUMN IF NOT EXISTS race_id VARCHAR(32) NOT NULL DEFAULT 'human';

CREATE INDEX IF NOT EXISTS idx_characters_race_id ON characters(race_id);