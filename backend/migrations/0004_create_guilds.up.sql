-- Migration 0004: Criar tabelas de Guildas e Associação de membros (Guilds)
CREATE TABLE IF NOT EXISTS guilds (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    leader_id INT NOT NULL REFERENCES characters(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS guild_members (
    guild_id INT NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,
    character_id INT NOT NULL UNIQUE REFERENCES characters(id) ON DELETE CASCADE,
    rank VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(guild_id, character_id)
);

CREATE INDEX IF NOT EXISTS idx_guild_members_character_id ON guild_members(character_id);
