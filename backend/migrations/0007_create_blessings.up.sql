-- Migration 0007: Criar tabela de bênçãos de personagens (Character Blessings)
CREATE TABLE IF NOT EXISTS character_blessings (
    character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    blessing_id VARCHAR(64) NOT NULL,
    acquired_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (character_id, blessing_id)
);

CREATE INDEX IF NOT EXISTS idx_character_blessings_char_id ON character_blessings(character_id);
