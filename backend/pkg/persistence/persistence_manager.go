package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/db"
	"github.com/light-and-shadow/backend/pkg/inventory"
)

type PersistenceManager struct {
	pgPool *db.PostgresPool
}

type CharacterSummary struct {
	ID      int
	Name    string
	Class   string
	Level   int
	RaceID  string // (R1-I-B)
	Account int
}

func NewPersistenceManager(pgPool *db.PostgresPool) *PersistenceManager {
	return &PersistenceManager{
		pgPool: pgPool,
	}
}

func (pm *PersistenceManager) ListCharactersByAccount(accountID int) ([]CharacterSummary, error) {
	if pm.pgPool == nil || pm.pgPool.DB == nil {
		slog.Warn("PostgreSQL in fallback mode. Returning empty character list", "accountID", accountID)
		return []CharacterSummary{}, nil
	}

	dbConn := pm.pgPool.DB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := dbConn.QueryContext(ctx, `
        SELECT id, name, class, level, account_id, race_id
        FROM characters
        WHERE account_id = $1
        ORDER BY id ASC
    `, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list characters for account %d: %w", accountID, err)
	}
	defer rows.Close()

	characters := make([]CharacterSummary, 0)

	for rows.Next() {
		var ch CharacterSummary
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Class, &ch.Level, &ch.Account, &ch.RaceID); err != nil {
			return nil, fmt.Errorf("failed to scan character summary: %w", err)
		}
		characters = append(characters, ch)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating character summaries: %w", err)
	}

	return characters, nil
}

// CreateCharacterForAccount creates a new character and its initial inventory in a single transaction.
// This is the authoritative persistence method and should be called by a service layer that has already
// performed all business logic validations (name rules, race rules from Rule Registry, etc.).
// It returns a summary of the created character, a client-facing error code, or an internal error.
func (pm *PersistenceManager) CreateCharacterForAccount(ctx context.Context, accountID int, desiredName string, raceID string) (*CharacterSummary, string, error) {
	// 1. Perform minimal persistence-level validations.
	if accountID <= 0 {
		return nil, "not_authenticated", fmt.Errorf("invalid accountID: %d", accountID)
	}

	normalizedName := strings.TrimSpace(desiredName)
	if normalizedName == "" {
		return nil, "invalid_name", fmt.Errorf("desiredName cannot be empty")
	}

	normalizedRaceID := strings.TrimSpace(raceID)
	if normalizedRaceID == "" {
		return nil, "invalid_race", fmt.Errorf("raceID cannot be empty")
	}

	if pm.pgPool == nil || pm.pgPool.DB == nil {
		return nil, "persistence_error", fmt.Errorf("database is not available")
	}

	dbConn := pm.pgPool.DB
	tx, err := dbConn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, "internal_error", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback is a no-op if the transaction has been committed.

	// 2. Insert the new character row with authoritative defaults.
	// The query returns the newly generated ID for use in subsequent inserts.
	var newCharID int
	insertCharQuery := `
		INSERT INTO characters (account_id, name, class, level, race_id)
		VALUES ($1, $2, 'novice', 1, $3)
		RETURNING id
	`
	err = tx.QueryRowContext(ctx, insertCharQuery, accountID, normalizedName, normalizedRaceID).Scan(&newCharID)
	if err != nil {
		// Check for unique constraint violation on the name.
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // unique_violation
			return nil, "name_taken", fmt.Errorf("character name %q is already taken: %w", normalizedName, err)
		}
		return nil, "persistence_error", fmt.Errorf("failed to insert new character: %w", err)
	}

	// 3. Create the initial inventory for the new character.
	// We reuse the same logic from the LoadCharacter fallback for consistency.
	initialInventory := inventory.NewPlayerInventory(normalizedName)

	// 4. Insert the initial inventory items into the 'inventories' table.
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO inventories (character_id, slot_index, item_id, quantity, durability, item_uuid)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return nil, "internal_error", fmt.Errorf("failed to prepare inventory insert statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range initialInventory.Items {
		if item == nil {
			continue
		}
		_, err = stmt.ExecContext(ctx, newCharID, item.SlotIndex, item.ItemID, item.Quantity, item.Durability, item.ItemUUID)
		if err != nil {
			// The defer tx.Rollback() will handle the cleanup.
			return nil, "persistence_error", fmt.Errorf("failed to insert initial inventory item for slot %d: %w", item.SlotIndex, err)
		}
	}

	// 5. Commit the transaction if all inserts were successful.
	if err := tx.Commit(); err != nil {
		return nil, "internal_error", fmt.Errorf("failed to commit character creation transaction: %w", err)
	}

	slog.Info("Successfully created new character", "name", normalizedName, "race", normalizedRaceID, "accountID", accountID, "charID", newCharID)

	// 6. Return a summary of the newly created character.
	summary := &CharacterSummary{
		ID:      newCharID,
		Name:    normalizedName,
		Class:   "novice",
		Level:   1,
		RaceID:  normalizedRaceID,
		Account: accountID,
	}

	return summary, "", nil
}

// InitSchema cria tabelas e garante colunas necessÃ¡rias para atributos de combate no PostgreSQL
func (pm *PersistenceManager) InitSchema() error {
	if pm.pgPool == nil || pm.pgPool.DB == nil {
		slog.Warn("PostgreSQL is running in fallback mode, skipping schema initialization")
		return nil
	}

	dbConn := pm.pgPool.DB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("Initializing and verifying PostgreSQL schemas...")

	// 1. Cria tabela de contas
	_, err := dbConn.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS accounts (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_accounts_username ON accounts(username);
	`)
	if err != nil {
		return fmt.Errorf("failed to create accounts table: %w", err)
	}

	// 2. Garante conta padrÃ£o para testes
	_, err = dbConn.ExecContext(ctx, `
		INSERT INTO accounts (id, username, email, password_hash)
		VALUES (1, 'default_user', 'default@example.com', '$2a$10$/XE4ObyJ5H7yQn.ybMKfY.K7sGTu9xVjBUt6cR0pch3tJl9yBzAca')
		ON CONFLICT (id) DO NOTHING;
	`)
	if err != nil {
		slog.Warn("Could not insert default user account (might already exist or sequence conflict)", "error", err)
	}

	// 3. Cria tabela de personagens
	_, err = dbConn.ExecContext(ctx, `
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
	`)
	if err != nil {
		return fmt.Errorf("failed to create characters table: %w", err)
	}

	// 4. Garante colunas de atributos extras autoritativos de combate, facÃ§Ã£o e versÃ£o
	columnsToAdd := []struct {
		Name    string
		Type    string
		Default string
	}{
		{"health", "DOUBLE PRECISION", "600.0"},
		{"max_health", "DOUBLE PRECISION", "600.0"},
		{"mana", "DOUBLE PRECISION", "100.0"},
		{"max_mana", "DOUBLE PRECISION", "100.0"},
		{"base_attack", "DOUBLE PRECISION", "45.0"},
		{"weapon_damage", "DOUBLE PRECISION", "15.0"},
		{"defense", "DOUBLE PRECISION", "30.0"},
		{"resistance", "DOUBLE PRECISION", "15.0"},
		{"accuracy", "DOUBLE PRECISION", "95.0"},
		{"evasion", "DOUBLE PRECISION", "10.0"},
		{"crit_chance", "DOUBLE PRECISION", "0.10"},
		{"crit_multiplier", "DOUBLE PRECISION", "1.50"},
		{"armor_penetration", "DOUBLE PRECISION", "0.15"},
		{"element", "VARCHAR(20)", "'Light'"},
		{"element_attack_bonus", "DOUBLE PRECISION", "0.10"},
		{"element_def_bonus", "DOUBLE PRECISION", "0.05"},
		{"faction", "VARCHAR(32)", "'Alliance'"},
		{"version", "INT", "1"},                        // Controle de versÃ£o para optimistic locking (PATCH 4)
		{"gold", "INT", "1000"},                        // Gold do jogador (PATCH 1)
		{"subclass", "VARCHAR(50)", "''"},              // Subclasse do jogador (Sprint 3 Task 5)
		{"affinity_fire", "INT", "0"},                  // Afinidade de Fogo
		{"affinity_ice", "INT", "0"},                   // Afinidade de Gelo
		{"affinity_holy", "INT", "0"},                  // Afinidade Sagrada
		{"affinity_shadow", "INT", "0"},                // Afinidade Sombria
		{"affinity_nature", "INT", "0"},                // Afinidade Natural
		{"race_id", "VARCHAR(32) NOT NULL", "'human'"}, // Raça do personagem (R1-D)
	}

	for _, col := range columnsToAdd {
		alterQuery := fmt.Sprintf("ALTER TABLE characters ADD COLUMN IF NOT EXISTS %s %s DEFAULT %s", col.Name, col.Type, col.Default)
		_, err = dbConn.ExecContext(ctx, alterQuery)
		if err != nil {
			return fmt.Errorf("failed to add column %s to characters: %w", col.Name, err)
		}
	}

	// Garante o índice na coluna race_id para consistência com a migration 0011 (Task R1-F)
	_, err = dbConn.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_characters_race_id ON characters(race_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create race_id index on characters: %w", err)
	}

	// 5. Cria tabela de inventÃ¡rios
	_, err = dbConn.ExecContext(ctx, `
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
	`)
	if err != nil {
		return fmt.Errorf("failed to create inventories table: %w", err)
	}

	// 6. Cria tabelas do Sistema de Quests e DiÃ¡logos (Sprint 3 Task 3)
	_, err = dbConn.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS character_quests (
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			quest_id VARCHAR(64) NOT NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'active', -- active, completed
			version INT DEFAULT 1, -- Para Optimistic Locking (PATCH 3)
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(character_id, quest_id)
		);
		CREATE INDEX IF NOT EXISTS idx_character_quests_char_id ON character_quests(character_id);

		CREATE TABLE IF NOT EXISTS quest_objectives (
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			quest_id VARCHAR(64) NOT NULL,
			objective_index INT NOT NULL,
			current_qty INT NOT NULL DEFAULT 0,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(character_id, quest_id, objective_index),
			FOREIGN KEY(character_id, quest_id) REFERENCES character_quests(character_id, quest_id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS npc_states (
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			npc_id VARCHAR(64) NOT NULL,
			dialogue_flags TEXT NOT NULL DEFAULT '', -- Flags guardadas ou histÃ³rico em string
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(character_id, npc_id)
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create quest/npc state tables: %w", err)
	}

	// 7. PATCH 1 & PATCH 4 extra quest tables and versioning columns
	_, err = dbConn.ExecContext(ctx, `
		ALTER TABLE character_quests ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;
		ALTER TABLE character_quests ADD COLUMN IF NOT EXISTS progress TEXT;

		CREATE TABLE IF NOT EXISTS quest_rewards_claimed (
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			quest_id VARCHAR(64) NOT NULL,
			reward_hash TEXT NOT NULL,
			claimed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(character_id, quest_id)
		);

		CREATE TABLE IF NOT EXISTS guilds (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) UNIQUE NOT NULL,
			motd TEXT NOT NULL DEFAULT '',
			version INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS guild_members (
			guild_id INT REFERENCES guilds(id) ON DELETE CASCADE,
			character_name VARCHAR(32) PRIMARY KEY,
			role INT NOT NULL,
			joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS guild_audit_logs (
			id SERIAL PRIMARY KEY,
			guild_id INT REFERENCES guilds(id) ON DELETE CASCADE,
			action TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS guild_storage (
			guild_id INT REFERENCES guilds(id) ON DELETE CASCADE,
			slot INT NOT NULL,
			item_id VARCHAR(50) NOT NULL,
			quantity INT NOT NULL,
			PRIMARY KEY (guild_id, slot)
		);

		CREATE TABLE IF NOT EXISTS social_relations (
			character_name VARCHAR(32) NOT NULL,
			target_name VARCHAR(32) NOT NULL,
			relation_type VARCHAR(10) NOT NULL,
			PRIMARY KEY (character_name, target_name, relation_type)
		);

		-- Ensure pre-existing tables have all columns
		ALTER TABLE guilds ADD COLUMN IF NOT EXISTS motd TEXT NOT NULL DEFAULT '';
		ALTER TABLE guilds ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;
		ALTER TABLE guild_members ADD COLUMN IF NOT EXISTS character_name VARCHAR(32);
		ALTER TABLE guild_members ADD COLUMN IF NOT EXISTS role INTEGER NOT NULL DEFAULT 0;

		-- Economy, Trading & Marketplace System Tables (Sprint 3 Task 4)
		ALTER TABLE inventories ADD COLUMN IF NOT EXISTS item_uuid VARCHAR(64);

		CREATE TABLE IF NOT EXISTS item_instances (
			id VARCHAR(64) PRIMARY KEY,
			item_id VARCHAR(64) NOT NULL,
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			durability INT NOT NULL DEFAULT 100,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS market_orders (
			id SERIAL PRIMARY KEY,
			seller_character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			item_id VARCHAR(64) NOT NULL,
			item_uuid VARCHAR(64),
			quantity INT NOT NULL DEFAULT 1,
			price_gold INT NOT NULL,
			tax_gold INT NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS market_history (
			id SERIAL PRIMARY KEY,
			seller_character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			buyer_character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			item_id VARCHAR(64) NOT NULL,
			quantity INT NOT NULL,
			price_gold INT NOT NULL,
			tax_gold INT NOT NULL,
			sold_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS trade_logs (
			id SERIAL PRIMARY KEY,
			player_a_name VARCHAR(32) NOT NULL,
			player_b_name VARCHAR(32) NOT NULL,
			gold_a INT NOT NULL,
			gold_b INT NOT NULL,
			items_a TEXT NOT NULL,
			items_b TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		-- Professions System Tables (Sprint 4 Task 1)
		CREATE TABLE IF NOT EXISTS character_professions (
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			profession VARCHAR(32) NOT NULL,
			level INT NOT NULL DEFAULT 1,
			experience INT NOT NULL DEFAULT 0,
			PRIMARY KEY(character_id, profession)
		);

		-- Dungeon, Raid & Boss Systems (Sprint 4 Task 2)
		CREATE TABLE IF NOT EXISTS raid_lockouts (
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			boss_id VARCHAR(64) NOT NULL,
			locked_until TIMESTAMP WITH TIME ZONE NOT NULL,
			PRIMARY KEY (character_id, boss_id)
		);

		CREATE TABLE IF NOT EXISTS dungeon_checkpoints (
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			dungeon_id VARCHAR(64) NOT NULL,
			checkpoint_id VARCHAR(64) NOT NULL,
			saved_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (character_id, dungeon_id)
		);

		CREATE TABLE IF NOT EXISTS boss_kill_states (
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			boss_id VARCHAR(64) NOT NULL,
			killed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (character_id, boss_id)
		);

		CREATE TABLE IF NOT EXISTS contribution_audit_logs (
			id SERIAL PRIMARY KEY,
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			boss_id VARCHAR(64) NOT NULL,
			damage_dealt DOUBLE PRECISION NOT NULL,
			healing_done DOUBLE PRECISION NOT NULL,
			contribution_score DOUBLE PRECISION NOT NULL,
			reward_item_id VARCHAR(64) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		-- Blessing System Table (Sprint 5)
		CREATE TABLE IF NOT EXISTS character_blessings (
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			blessing_id VARCHAR(64) NOT NULL,
			acquired_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (character_id, blessing_id)
		);
		CREATE INDEX IF NOT EXISTS idx_character_blessings_char_id ON character_blessings(character_id);

		-- Upgrade Currency columns to support uint64 (BIGINT)
		ALTER TABLE characters ALTER COLUMN gold TYPE BIGINT;
		ALTER TABLE market_orders ALTER COLUMN price_gold TYPE BIGINT;
		ALTER TABLE market_orders ALTER COLUMN tax_gold TYPE BIGINT;
		ALTER TABLE market_history ALTER COLUMN price_gold TYPE BIGINT;
		ALTER TABLE market_history ALTER COLUMN tax_gold TYPE BIGINT;
		ALTER TABLE trade_logs ALTER COLUMN gold_a TYPE BIGINT;
		ALTER TABLE trade_logs ALTER COLUMN gold_b TYPE BIGINT;

		-- Housing System Tables (Sprint 5)
		CREATE TABLE IF NOT EXISTS houses (
			house_id VARCHAR(64) PRIMARY KEY,
			owner_id INT UNIQUE REFERENCES characters(id) ON DELETE SET NULL,
			guild_id INT UNIQUE REFERENCES guilds(id) ON DELETE SET NULL,
			purchased_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			last_rent_paid_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			rent_status VARCHAR(32) DEFAULT 'active',
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

		ALTER TABLE houses ADD COLUMN IF NOT EXISTS permissions VARCHAR(32) DEFAULT 'private';
		ALTER TABLE houses ADD COLUMN IF NOT EXISTS min_rank VARCHAR(32) DEFAULT 'member';

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

		-- PvP and Bounty System (Sprint 6/PvP System)
		CREATE TABLE IF NOT EXISTS pvp_skulls (
			character_id INT PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
			skull_tier VARCHAR(16) DEFAULT 'none',
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

		CREATE INDEX IF NOT EXISTS idx_pvp_unjust_kills_killer_id ON pvp_unjust_kills(killer_id);
		CREATE INDEX IF NOT EXISTS idx_pvp_unjust_kills_killed_at ON pvp_unjust_kills(killed_at);
	`)
	if err != nil {
		return fmt.Errorf("failed to run PATCH 1, 4 & PvP system tables: %w", err)
	}

	slog.Info("PostgreSQL schema validated and upgraded successfully")
	return nil
}

func (pm *PersistenceManager) CharacterBelongsToAccount(accountID int, characterName string) (bool, error) {
	if characterName == "" {
		return false, nil
	}

	if pm.pgPool == nil || pm.pgPool.DB == nil {
		slog.Warn("PostgreSQL in fallback mode. Allowing character selection", "accountID", accountID, "character", characterName)
		return true, nil
	}

	dbConn := pm.pgPool.DB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	err := dbConn.QueryRowContext(ctx, `
        SELECT EXISTS (
            SELECT 1
            FROM characters
            WHERE account_id = $1 AND name = $2
        )
    `, accountID, characterName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to validate character ownership for account %d and character %s: %w", accountID, characterName, err)
	}

	return exists, nil
}

// LoadCharacter carrega os atributos autoritativos de um personagem, versÃ£o e inventÃ¡rio correspondente do PostgreSQL (PATCH 4)
func (pm *PersistenceManager) LoadCharacter(playerID string) (*combat.EntityStats, map[int]*inventory.InventoryItem, float64, float64, float64, int, int64, int64, error) {
	if pm.pgPool == nil || pm.pgPool.DB == nil {
		slog.Warn("PostgreSQL in fallback mode. Loading default state in-memory", "playerID", playerID)
		defaultInv := inventory.NewPlayerInventory(playerID)
		defaultInv.BaseStats.RaceID = "human" // (R1-H)
		return &defaultInv.BaseStats, defaultInv.Items, 100.0, 100.0, 0.0, 1, 0, int64(1000), nil
	}

	dbConn := pm.pgPool.DB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var charID int
	var accountID int
	var className string
	var level int
	var experience int64
	var posX, posY, posZ float64
	var gold int64

	// Atributos de combate
	var health, maxHealth, mana, maxMana, baseAttack, weaponDamage, defense, resistance float64
	var accuracy, evasion, critChance, critMultiplier, armorPenetration float64
	var element string
	var raceID string // (R1-H)
	var elementAttackBonus, elementDefBonus float64
	var faction string
	var version int // (PATCH 4)
	var subclass string
	var affFire, affIce, affHoly, affShadow, affNature int

	// Busca o personagem na tabela (inclui gold e progresso da Sprint 3 Task 5)
	row := dbConn.QueryRowContext(ctx, `
		SELECT id, account_id, class, level, experience, posX, posY, posZ,
			health, max_health, mana, max_mana, base_attack, weapon_damage, defense, resistance, race_id,
			accuracy, evasion, crit_chance, crit_multiplier, armor_penetration, element, element_attack_bonus, element_def_bonus, faction, version, gold,
			subclass, affinity_fire, affinity_ice, affinity_holy, affinity_shadow, affinity_nature
		FROM characters WHERE name = $1`, playerID)

	err := row.Scan(&charID, &accountID, &className, &level, &experience, &posX, &posY, &posZ,
		&health, &maxHealth, &mana, &maxMana, &baseAttack, &weaponDamage, &defense, &resistance, &raceID,
		&accuracy, &evasion, &critChance, &critMultiplier, &armorPenetration, &element, &elementAttackBonus, &elementDefBonus, &faction, &version, &gold,
		&subclass, &affFire, &affIce, &affHoly, &affShadow, &affNature)

	if errors.Is(err, sql.ErrNoRows) {
		slog.Info("Character not found in database. Creating a new persistent character row", "playerID", playerID)

		// Insere novo personagem padrÃ£o com gold inicial no estado Novice (Level 1)
		err = dbConn.QueryRowContext(ctx, `
			INSERT INTO characters (account_id, name, class, level, experience, posX, posY, posZ,
				health, max_health, mana, max_mana, base_attack, weapon_damage, defense, resistance, race_id,
				accuracy, evasion, crit_chance, crit_multiplier, armor_penetration, element, element_attack_bonus, element_def_bonus, faction, version, gold,
				subclass, affinity_fire, affinity_ice, affinity_holy, affinity_shadow, affinity_nature)
			VALUES (1, $1, 'Novice', 1, 0, 100.0, 100.0, 0.0,
				200.0, 200.0, 50.0, 50.0, 10.0, 15.0, 10.0, 0.0, 'human',
				95.0, 5.0, 0.05, 1.50, 0.05, 'None', 0.0, 0.0, 'Alliance', 1, 1000,
				'', 0, 0, 0, 0, 0)
			RETURNING id, account_id, class, level, experience, posX, posY, posZ,
				health, max_health, mana, max_mana, base_attack, weapon_damage, defense, resistance, race_id,
				accuracy, evasion, crit_chance, crit_multiplier, armor_penetration, element, element_attack_bonus, element_def_bonus, faction, version, gold,
				subclass, affinity_fire, affinity_ice, affinity_holy, affinity_shadow, affinity_nature
		`, playerID).Scan(&charID, &accountID, &className, &level, &experience, &posX, &posY, &posZ,
			&health, &maxHealth, &mana, &maxMana, &baseAttack, &weaponDamage, &defense, &resistance, &raceID,
			&accuracy, &evasion, &critChance, &critMultiplier, &armorPenetration, &element, &elementAttackBonus, &elementDefBonus, &faction, &version, &gold,
			&subclass, &affFire, &affIce, &affHoly, &affShadow, &affNature)

		if err != nil {
			return nil, nil, 0, 0, 0, 1, 0, 1000, fmt.Errorf("failed to create default character row: %w", err)
		}

		// Cria inventÃ¡rio padrÃ£o inicial na tabela inventories
		defaultInv := inventory.NewPlayerInventory(playerID)
		tx, err := dbConn.BeginTx(ctx, nil)
		if err != nil {
			return nil, nil, 0, 0, 0, 1, 0, 1000, err
		}
		defer tx.Rollback()

		for _, item := range defaultInv.Items {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO inventories (character_id, slot_index, item_id, quantity, durability, item_uuid)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, charID, item.SlotIndex, item.ItemID, item.Quantity, item.Durability, item.ItemUUID)
			if err != nil {
				return nil, nil, 0, 0, 0, 1, 0, 1000, fmt.Errorf("failed to populate default inventory: %w", err)
			}
		}

		if err := tx.Commit(); err != nil {
			return nil, nil, 0, 0, 0, 1, 0, 1000, err
		}

		defaultInv.BaseStats.Class = className
		defaultInv.BaseStats.Subclass = subclass
		defaultInv.BaseStats.AffinityFire = affFire
		defaultInv.BaseStats.AffinityIce = affIce
		defaultInv.BaseStats.AffinityHoly = affHoly
		defaultInv.BaseStats.AffinityShadow = affShadow
		defaultInv.BaseStats.AffinityNature = affNature
		defaultInv.BaseStats.RaceID = raceID // (R1-H)
		defaultInv.BaseStats.Level = level

		slog.Info("New character and default items persisted successfully", "player", playerID, "char_id", charID)
		return &defaultInv.BaseStats, defaultInv.Items, posX, posY, posZ, 1, 0, 1000, nil
	} else if err != nil {
		return nil, nil, 0, 0, 0, 1, 0, 1000, fmt.Errorf("failed to query character: %w", err)
	}

	// Carrega itens de inventÃ¡rio
	items := make(map[int]*inventory.InventoryItem)
	rows, err := dbConn.QueryContext(ctx, `
		SELECT slot_index, item_id, quantity, durability, item_uuid 
		FROM inventories WHERE character_id = $1`, charID)
	if err != nil {
		return nil, nil, 0, 0, 0, version, 0, gold, fmt.Errorf("failed to query inventories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var slotIndex int
		var itemID string
		var quantity int
		var durability int
		var itemUUID sql.NullString
		if err := rows.Scan(&slotIndex, &itemID, &quantity, &durability, &itemUUID); err != nil {
			return nil, nil, 0, 0, 0, version, 0, gold, fmt.Errorf("failed to scan inventory row: %w", err)
		}
		items[slotIndex] = &inventory.InventoryItem{
			ItemID:     itemID,
			Quantity:   quantity,
			Durability: durability,
			SlotIndex:  slotIndex,
			ItemUUID:   itemUUID.String,
		}
	}

	stats := &combat.EntityStats{
		ID:                 playerID,
		Name:               playerID,
		IsPlayer:           true,
		Faction:            faction,
		RaceID:             raceID, // (R1-H)
		Level:              level,
		BaseAttack:         baseAttack,
		WeaponDamage:       weaponDamage,
		Defense:            defense,
		Resistance:         resistance,
		Accuracy:           accuracy,
		Evasion:            evasion,
		CritChance:         critChance,
		CritMultiplier:     critMultiplier,
		ArmorPenetration:   armorPenetration,
		Element:            element,
		ElementAttackBonus: elementAttackBonus,
		ElementDefBonus:    elementDefBonus,
		Health:             health,
		MaxHealth:          maxHealth,
		Mana:               mana,
		MaxMana:            maxMana,
		Class:              className,
		Subclass:           subclass,
		AffinityFire:       affFire,
		AffinityIce:        affIce,
		AffinityHoly:       affHoly,
		AffinityShadow:     affShadow,
		AffinityNature:     affNature,
	}

	slog.Info("Character and inventory loaded successfully from PostgreSQL", "player", playerID, "lvl", level, "items_count", len(items), "version", version, "gold", gold)
	return stats, items, posX, posY, posZ, version, experience, gold, nil
}

// SaveCharacter salva de forma atÃ´mica, transacional e segura o estado do personagem e seu inventÃ¡rio no PostgreSQL (PATCH 1, 3, 4)
func (pm *PersistenceManager) SaveCharacter(playerID string, stats *combat.EntityStats, items map[int]inventory.InventoryItem, posX, posY, posZ float64, currentVersion int, experience int64, gold int64) (err error) {
	if pm.pgPool == nil || pm.pgPool.DB == nil {
		slog.Warn("PostgreSQL in fallback mode. Skipping save to DB", "playerID", playerID)
		return nil
	}

	dbConn := pm.pgPool.DB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// EnforÃ§a Isolation Level estrito de forma segura (PATCH 3)
	tx, err := dbConn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return fmt.Errorf("failed to begin save transaction: %w", err)
	}

	// Garante Rollback automÃ¡tico em caso de pÃ¢nico ou erro de execuÃ§Ã£o (PATCH 3)
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Busca ID do personagem para garantir chave estrangeira correta
	var charID int
	err = tx.QueryRowContext(ctx, "SELECT id FROM characters WHERE name = $1", playerID).Scan(&charID)
	if err != nil {
		return fmt.Errorf("cannot find character %s during save: %w", playerID, err)
	}

	// 2. Atualiza estado do personagem aplicando Optimistic Locking na coluna version (PATCH 4)
	res, err := tx.ExecContext(ctx, `
		UPDATE characters SET
			level = $1,
			posX = $2,
			posY = $3,
			posZ = $4,
			health = $5,
			max_health = $6,
			mana = $7,
			max_mana = $8,
			base_attack = $9,
			weapon_damage = $10,
			defense = $11,
			resistance = $12,
			accuracy = $13,
			evasion = $14,
			crit_chance = $15,
			crit_multiplier = $16,
			armor_penetration = $17,
			element = $18,
			element_attack_bonus = $19,
			element_def_bonus = $20,
			faction = $21,
			experience = $22,
			gold = $23,
			class = $24,
			subclass = $25,
			affinity_fire = $26,
			affinity_ice = $27,
			affinity_holy = $28,
			affinity_shadow = $29,
			affinity_nature = $30,
			version = version + 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $31 AND version = $32
	`, stats.Level, posX, posY, posZ, stats.Health, stats.MaxHealth, stats.Mana, stats.MaxMana,
		stats.BaseAttack, stats.WeaponDamage, stats.Defense, stats.Resistance, stats.Accuracy,
		stats.Evasion, stats.CritChance, stats.CritMultiplier, stats.ArmorPenetration,
		stats.Element, stats.ElementAttackBonus, stats.ElementDefBonus, stats.Faction,
		experience, gold, stats.Class, stats.Subclass,
		stats.AffinityFire, stats.AffinityIce, stats.AffinityHoly, stats.AffinityShadow, stats.AffinityNature,
		charID, currentVersion)

	if err != nil {
		return fmt.Errorf("failed to update character state: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	// Se RowsAffected for 0, ocorreu um conflito de concorrÃªncia ou tentativa de exploit de duplicaÃ§Ã£o (PATCH 4)
	if rowsAffected == 0 {
		return fmt.Errorf("optimistic locking conflict: character %s has been modified by another transaction or version mismatch (expected version %d)", playerID, currentVersion)
	}

	// 3. Deleta inventÃ¡rio atual para salvar os novos slots limpos
	_, err = tx.ExecContext(ctx, "DELETE FROM inventories WHERE character_id = $1", charID)
	if err != nil {
		return fmt.Errorf("failed to clean inventory before save: %w", err)
	}

	// 4. Insere todos os itens do inventÃ¡rio a partir do snapshot de forma segura
	for _, item := range items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO inventories (character_id, slot_index, item_id, quantity, durability, item_uuid)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, charID, item.SlotIndex, item.ItemID, item.Quantity, item.Durability, item.ItemUUID)
		if err != nil {
			return fmt.Errorf("failed to insert inventory slot %d: %w", item.SlotIndex, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit save transaction: %w", err)
	}

	slog.Info("Character state, version, gold and inventory successfully persisted to PostgreSQL", "player", playerID, "lvl", stats.Level, "x", posX, "y", posY, "gold", gold, "new_version", currentVersion+1)
	return nil
}
