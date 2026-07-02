-- Migration 0006: Criar tabela de profissões e progressão (Professions & Progression)
CREATE TABLE IF NOT EXISTS character_professions (
    character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    profession VARCHAR(32) NOT NULL,
    level INT NOT NULL DEFAULT 1,
    experience INT NOT NULL DEFAULT 0,
    PRIMARY KEY(character_id, profession)
);
