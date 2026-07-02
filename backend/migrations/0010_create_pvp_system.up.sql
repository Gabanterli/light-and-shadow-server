-- Migration 0010: PvP and Bounty System

CREATE TABLE IF NOT EXISTS pvp_skulls (
    character_id INT PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    skull_tier VARCHAR(16) DEFAULT 'none', -- 'none', 'white', 'red', 'black'
    unjust_kill_count INT DEFAULT 0,
    bounty_reward BIGINT DEFAULT 0,
    last_known_region VARCHAR(64) DEFAULT '',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pvp_unjust_kills (
    id SERIAL PRIMARY KEY,
    killer_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    victim_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    killed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexing for high-performance queries on kill history & skull lookups
CREATE INDEX IF NOT EXISTS idx_pvp_unjust_kills_killer_id ON pvp_unjust_kills(killer_id);
CREATE INDEX IF NOT EXISTS idx_pvp_unjust_kills_killed_at ON pvp_unjust_kills(killed_at);
