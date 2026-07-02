-- Migration for Sprint 3 Task 3 Hardening Patches
-- PATCH 1: Extend character_quests with version and progress
ALTER TABLE character_quests ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE character_quests ADD COLUMN IF NOT EXISTS progress TEXT;

-- PATCH 4: Create quest_rewards_claimed for idempotency
CREATE TABLE IF NOT EXISTS quest_rewards_claimed (
    character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    quest_id VARCHAR(64) NOT NULL,
    reward_hash TEXT NOT NULL,
    claimed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(character_id, quest_id)
);
