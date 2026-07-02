-- Migration 0009: Criar tabelas do sistema de moradia (Housing System)
CREATE TABLE IF NOT EXISTS houses (
    house_id VARCHAR(64) PRIMARY KEY,
    owner_id INT UNIQUE REFERENCES characters(id) ON DELETE SET NULL,
    guild_id INT UNIQUE REFERENCES guilds(id) ON DELETE SET NULL,
    purchased_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_rent_paid_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    rent_status VARCHAR(32) DEFAULT 'active', -- 'active', 'warning', 'evicted'
    warning_sent_at TIMESTAMP WITH TIME ZONE,
    permissions VARCHAR(32) DEFAULT 'private',
    min_rank VARCHAR(32) DEFAULT 'member',
    CONSTRAINT chk_housing_owner CHECK (
        (owner_id IS NOT NULL AND guild_id IS NULL) OR
        (owner_id IS NULL AND guild_id IS NOT NULL) OR
        (owner_id IS NULL AND guild_id IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_houses_owner_id ON houses(owner_id);
CREATE INDEX IF NOT EXISTS idx_houses_guild_id ON houses(guild_id);

CREATE TABLE IF NOT EXISTS house_storage (
    house_id VARCHAR(64) NOT NULL REFERENCES houses(house_id) ON DELETE CASCADE,
    slot_id INT NOT NULL,
    item_id VARCHAR(64) NOT NULL,
    quantity INT NOT NULL,
    durability INT NOT NULL DEFAULT 100,
    PRIMARY KEY (house_id, slot_id)
);

CREATE TABLE IF NOT EXISTS house_decorations (
    id SERIAL PRIMARY KEY,
    house_id VARCHAR(64) NOT NULL REFERENCES houses(house_id) ON DELETE CASCADE,
    furniture_id VARCHAR(64) NOT NULL,
    x DOUBLE PRECISION NOT NULL,
    y DOUBLE PRECISION NOT NULL,
    z DOUBLE PRECISION NOT NULL,
    rotation DOUBLE PRECISION NOT NULL DEFAULT 0.0
);

CREATE INDEX IF NOT EXISTS idx_house_decorations_house_id ON house_decorations(house_id);

CREATE TABLE IF NOT EXISTS house_reclaim_storage (
    id SERIAL PRIMARY KEY,
    owner_id INT REFERENCES characters(id) ON DELETE CASCADE,
    guild_id INT REFERENCES guilds(id) ON DELETE CASCADE,
    item_id VARCHAR(64) NOT NULL,
    quantity INT NOT NULL,
    durability INT NOT NULL DEFAULT 100,
    reclaimed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_reclaim_owner CHECK (
        (owner_id IS NOT NULL AND guild_id IS NULL) OR
        (owner_id IS NULL AND guild_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_house_reclaim_owner ON house_reclaim_storage(owner_id);
CREATE INDEX IF NOT EXISTS idx_house_reclaim_guild ON house_reclaim_storage(guild_id);
