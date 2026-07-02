package housing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/pkg/inventory"
)

// HouseConfig representa os metadados de uma moradia (carregados do JSON)
type HouseConfig struct {
	ID                  string  `json:"id"`
	Type                string  `json:"type"` // "player" ou "guild"
	Name                string  `json:"name"`
	Continent           string  `json:"continent"`
	X                   float64 `json:"x"`
	Y                   float64 `json:"y"`
	Z                   int     `json:"z"`
	Size                string  `json:"size"` // "small", "medium", "large"
	PurchaseCostBronze  int64   `json:"purchase_cost_bronze"`
	RentCostBronze      int64   `json:"rent_cost_bronze"`
	StorageCapacity     int     `json:"storage_capacity"`
}

// FurnitureConfig representa os metadados de uma mobília (carregados do JSON) (PATCH 8)
type FurnitureConfig struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Category       string `json:"category"` // "seating", "tables", "beds", "storage", "trophy", etc.
	CostBronze     int64  `json:"cost_bronze"`
	Functional     bool   `json:"functional,omitempty"`
	StorageBonus   int    `json:"storage_bonus,omitempty"`
	DecorationCost int    `json:"decoration_cost"`
}

// PlacedDecoration representa uma decoração colocada fisicamente dentro da casa
type PlacedDecoration struct {
	ID          int     `json:"id"`
	FurnitureID string  `json:"furniture_id"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Z           float64 `json:"z"`
	Rotation    float64 `json:"rotation"`
}

// HouseState representa o estado em memória ativo e dinâmico de uma moradia (PATCH 7)
type HouseState struct {
	HouseID        string
	OwnerID        int       // ID do personagem dono (0 se guild ou nenhum)
	OwnerName      string    // Nome do personagem dono
	GuildID        int       // ID da guilda dona (0 se player ou nenhum)
	GuildName      string    // Nome da guilda dona
	PurchasedAt    time.Time
	LastRentPaidAt time.Time
	RentStatus     string    // "active", "warning", "evicted"
	WarningSentAt  time.Time
	Storage        map[int]*inventory.InventoryItem // slotIndex -> item seguro
	Decorations    []PlacedDecoration
	Permissions    string    // ACL for player houses: "private", "friends", "guild", "public"
	MinRank        string    // Rank requirement for guild houses: "leader", "vice", "member", "recruit"
}

// HousingConfigWrapper é usado para carregar o arquivo JSON de moradias
type HousingConfigWrapper struct {
	Houses []HouseConfig `json:"houses"`
}

// FurnitureConfigWrapper é usado para carregar o arquivo JSON de mobílias
type FurnitureConfigWrapper struct {
	Furniture []FurnitureConfig `json:"furniture"`
}

// HousingManager coordena o ciclo de vida de moradias do mundo aberto (non-instanced)
type HousingManager struct {
	mu              sync.RWMutex
	db              *sql.DB
	houses          map[string]HouseConfig
	furniture       map[string]FurnitureConfig
	states          map[string]*HouseState
	gracePeriod     time.Duration
	warningInterval time.Duration
}

// NewHousingManager instancia e inicializa o HousingManager com carga de JSONs e sincronização ao Postgres
func NewHousingManager(db *sql.DB) *HousingManager {
	hm := &HousingManager{
		db:              db,
		houses:          make(map[string]HouseConfig),
		furniture:       make(map[string]FurnitureConfig),
		states:          make(map[string]*HouseState),
		gracePeriod:     7 * 24 * time.Hour, // 7 dias de tolerância após warning
		warningInterval: 30 * 24 * time.Hour, // Rent vence a cada 30 dias
	}

	hm.LoadConfigs()
	hm.SyncWithDatabase()

	// Inicia rotina periódica em background de processamento de aluguel/evicção (Sprint 5)
	go hm.startRentTicker(12 * time.Hour)

	return hm
}

// LoadConfigs carrega os JSONs de forma resiliente usando caminhos alternativos
func (hm *HousingManager) LoadConfigs() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	paths := []string{"backend/config/", "config/", "../config/", "../../config/"}
	
	// 1. Carrega houses.json
	var housesData []byte
	var err error
	var housesPath string
	for _, p := range paths {
		fp := p + "houses.json"
		if _, statErr := os.Stat(fp); statErr == nil {
			housesData, err = os.ReadFile(fp)
			if err == nil {
				housesPath = fp
				break
			}
		}
	}
	if err == nil && len(housesData) > 0 {
		var wrapper HousingConfigWrapper
		if jsonErr := json.Unmarshal(housesData, &wrapper); jsonErr == nil {
			for _, h := range wrapper.Houses {
				hm.houses[h.ID] = h
				// Inicializa estado padrão vazio em memória
				hm.states[h.ID] = &HouseState{
					HouseID:     h.ID,
					RentStatus:  "active",
					Storage:     make(map[int]*inventory.InventoryItem),
					Permissions: "private",
					MinRank:     "member",
				}
			}
			slog.Info("Successfully loaded open-world houses config", "count", len(hm.houses), "path", housesPath)
		} else {
			slog.Error("Failed to parse houses.json", "error", jsonErr)
		}
	} else {
		slog.Warn("Could not read houses.json, loading fallback houses")
		hm.loadFallbackHouses()
	}

	// 2. Carrega furniture.json
	var furnData []byte
	var furnPath string
	for _, p := range paths {
		fp := p + "furniture.json"
		if _, statErr := os.Stat(fp); statErr == nil {
			furnData, err = os.ReadFile(fp)
			if err == nil {
				furnPath = fp
				break
			}
		}
	}
	if err == nil && len(furnData) > 0 {
		var wrapper FurnitureConfigWrapper
		if jsonErr := json.Unmarshal(furnData, &wrapper); jsonErr == nil {
			for _, f := range wrapper.Furniture {
				hm.furniture[f.ID] = f
			}
			slog.Info("Successfully loaded furniture metadata config", "count", len(hm.furniture), "path", furnPath)
		} else {
			slog.Error("Failed to parse furniture.json", "error", jsonErr)
		}
	} else {
		slog.Warn("Could not read furniture.json, loading fallback furniture")
		hm.loadFallbackFurniture()
	}
}

func (hm *HousingManager) loadFallbackHouses() {
	fallbackList := []HouseConfig{
		{"house_beginner_01", "player", "Beginner Cozy Cottage", "Main Continent", 130.0, 140.0, 0, "small", 50000, 1000, 20},
		{"house_fire_01", "player", "Lava View Hearth Villa", "Fire Continent", 2150.0, 2150.0, 0, "medium", 250000, 5000, 50},
		{"guild_house_holy_01", "guild", "Sanctum Guild Citadel", "Holy Continent", 2550.0, 2550.0, 0, "large", 2000000, 50000, 100},
	}
	for _, h := range fallbackList {
		hm.houses[h.ID] = h
		hm.states[h.ID] = &HouseState{
			HouseID:     h.ID,
			RentStatus:  "active",
			Storage:     make(map[int]*inventory.InventoryItem),
			Permissions: "private",
			MinRank:     "member",
		}
	}
}

func (hm *HousingManager) loadFallbackFurniture() {
	fallbackList := []FurnitureConfig{
		{"furn_chair_wooden", "Rustic Wooden Chair", "seating", 1000, false, 0, 5},
		{"furn_table_wooden", "Oak Dining Table", "tables", 3000, false, 0, 10},
		{"furn_bed_comfy", "Royal Comfort Bed", "beds", 8000, false, 0, 20},
		{"furn_secure_chest", "Secure Iron-Bound Chest", "storage", 15000, true, 10, 15},
		{"furn_guild_vault_gold", "Golden Guild Vault", "guild_storage", 100000, true, 50, 40},
	}
	for _, f := range fallbackList {
		hm.furniture[f.ID] = f
	}
}

// SyncWithDatabase carrega do PostgreSQL os estados dinâmicos persistidos
func (hm *HousingManager) SyncWithDatabase() {
	if hm.db == nil {
		slog.Warn("PostgreSQL is in fallback mode. Housing state running strictly in-memory.")
		return
	}

	hm.mu.Lock()
	defer hm.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Carrega dados de posse de casas
	rows, err := hm.db.QueryContext(ctx, `
		SELECT house_id, owner_id, guild_id, purchased_at, last_rent_paid_at, rent_status, warning_sent_at, permissions, min_rank 
		FROM houses
	`)
	if err != nil {
		slog.Error("Failed to query houses states from database", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var houseID string
		var ownerID sql.NullInt32
		var guildID sql.NullInt32
		var purchasedAt sql.NullTime
		var lastRentPaidAt sql.NullTime
		var rentStatus string
		var warningSentAt sql.NullTime
		var permissions sql.NullString
		var minRank sql.NullString

		err := rows.Scan(&houseID, &ownerID, &guildID, &purchasedAt, &lastRentPaidAt, &rentStatus, &warningSentAt, &permissions, &minRank)
		if err == nil {
			state, exists := hm.states[houseID]
			if !exists {
				// Caso um id de casa não esteja mais no JSON mas esteja no banco de dados
				state = &HouseState{
					HouseID:     houseID,
					Storage:     make(map[int]*inventory.InventoryItem),
					Permissions: "private",
					MinRank:     "member",
				}
				hm.states[houseID] = state
			}
			if ownerID.Valid {
				state.OwnerID = int(ownerID.Int32)
			}
			if guildID.Valid {
				state.GuildID = int(guildID.Int32)
			}
			if purchasedAt.Valid {
				state.PurchasedAt = purchasedAt.Time
			}
			if permissions.Valid {
				state.Permissions = permissions.String
			} else {
				state.Permissions = "private"
			}
			if minRank.Valid {
				state.MinRank = minRank.String
			} else {
				state.MinRank = "member"
			}
			if lastRentPaidAt.Valid {
				state.LastRentPaidAt = lastRentPaidAt.Time
			}
			state.RentStatus = rentStatus
			if warningSentAt.Valid {
				state.WarningSentAt = warningSentAt.Time
			}
		}
	}

	// 2. Carrega caches de nomes de personagens/guildas
	for _, state := range hm.states {
		if state.OwnerID > 0 {
			_ = hm.db.QueryRowContext(ctx, "SELECT name FROM characters WHERE id = $1", state.OwnerID).Scan(&state.OwnerName)
		}
		if state.GuildID > 0 {
			_ = hm.db.QueryRowContext(ctx, "SELECT name FROM guilds WHERE id = $1", state.GuildID).Scan(&state.GuildName)
		}
	}

	// 3. Carrega itens armazenados nas casas
	storageRows, err := hm.db.QueryContext(ctx, "SELECT house_id, slot_id, item_id, quantity, durability FROM house_storage")
	if err == nil {
		defer storageRows.Close()
		for storageRows.Next() {
			var hID string
			var slotID int
			var itemID string
			var qty int
			var durability int
			if err := storageRows.Scan(&hID, &slotID, &itemID, &qty, &durability); err == nil {
				if state, ok := hm.states[hID]; ok {
					state.Storage[slotID] = &inventory.InventoryItem{
						ItemID:     itemID,
						Quantity:   qty,
						Durability: durability,
						SlotIndex:  slotID,
					}
				}
			}
		}
	}

	// 4. Carrega decorações colocadas
	decRows, err := hm.db.QueryContext(ctx, "SELECT id, house_id, furniture_id, x, y, z, rotation FROM house_decorations")
	if err == nil {
		defer decRows.Close()
		for decRows.Next() {
			var dID int
			var hID string
			var fID string
			var x, y, z, rot float64
			if err := decRows.Scan(&dID, &hID, &fID, &x, &y, &z, &rot); err == nil {
				if state, ok := hm.states[hID]; ok {
					state.Decorations = append(state.Decorations, PlacedDecoration{
						ID:          dID,
						FurnitureID: fID,
						X:           x,
						Y:           y,
						Z:           z,
						Rotation:    rot,
					})
				}
			}
		}
	}

	slog.Info("Successfully synchronized open-world Housing states from PostgreSQL")
}

// GetHouseConfig busca configuração de uma casa
func (hm *HousingManager) GetHouseConfig(houseID string) (HouseConfig, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	c, ok := hm.houses[houseID]
	return c, ok
}

// GetAllHouses retorna uma lista com todas as moradias configuradas (PATCH PvP)
func (hm *HousingManager) GetAllHouses() []HouseConfig {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	list := make([]HouseConfig, 0, len(hm.houses))
	for _, h := range hm.houses {
		list = append(list, h)
	}
	return list
}

// GetHouseState retorna cópia thread-safe do estado dinâmico de uma casa
func (hm *HousingManager) GetHouseState(houseID string) (HouseState, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	state, ok := hm.states[houseID]
	if !ok {
		return HouseState{}, false
	}

	// Clona para evitar race conditions de leitura externa
	copied := HouseState{
		HouseID:        state.HouseID,
		OwnerID:        state.OwnerID,
		OwnerName:      state.OwnerName,
		GuildID:        state.GuildID,
		GuildName:      state.GuildName,
		PurchasedAt:    state.PurchasedAt,
		LastRentPaidAt: state.LastRentPaidAt,
		RentStatus:     state.RentStatus,
		WarningSentAt:  state.WarningSentAt,
		Storage:        make(map[int]*inventory.InventoryItem),
		Decorations:    append([]PlacedDecoration(nil), state.Decorations...),
		Permissions:    state.Permissions,
		MinRank:        state.MinRank,
	}
	for k, v := range state.Storage {
		if v != nil {
			copied.Storage[k] = &inventory.InventoryItem{
				ItemID:     v.ItemID,
				Quantity:   v.Quantity,
				Durability: v.Durability,
				SlotIndex:  v.SlotIndex,
				ItemUUID:   v.ItemUUID,
			}
		}
	}
	return copied, true
}

// PurchaseHouse lida com a aquisição de moradia usando moedas bronze (PATCH 1 & PATCH 4)
func (hm *HousingManager) PurchaseHouse(playerID string, houseID string, playerInv *inventory.PlayerInventory) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// 1. Validação de configurações
	config, ok := hm.houses[houseID]
	if !ok {
		return fmt.Errorf("house with ID %s not found in coordinates registry", houseID)
	}

	state, ok := hm.states[houseID]
	if !ok {
		return fmt.Errorf("house state missing in registry")
	}

	// PATCH 1: Evita sobrescrever uma propriedade que já possui dono
	if state.OwnerID > 0 || state.GuildID > 0 {
		return fmt.Errorf("this physical open-world property is already owned and occupied")
	}

	// 2. Valida regras de limite de posse (Ex: Max 1 player house por jogador, Max 1 guild house por guilda)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var charID int
	var guildID int
	var err error

	if hm.db != nil {
		charID, guildID, err = hm.getPlayerDetails(ctx, nil, playerID)
		if err != nil {
			return fmt.Errorf("failed to retrieve character credentials: %w", err)
		}
	} else {
		// Modo fallback em memória
		charID = 1001 // Fallback mock id
		guildID = 0
	}

	if config.Type == "player" {
		// Verifica se o jogador já possui moradia
		for _, s := range hm.states {
			if s.OwnerID == charID {
				return fmt.Errorf("you already own a player house (%s). limit is 1 property per resident", s.HouseID)
			}
		}
	} else if config.Type == "guild" {
		if guildID == 0 {
			return fmt.Errorf("only guild members can purchase a guild house")
		}
		// Verifica se a guilda já possui moradia
		for _, s := range hm.states {
			if s.GuildID == guildID {
				return fmt.Errorf("your guild already owns a citadel (%s). limit is 1 stronghold per alliance", s.HouseID)
			}
		}
	}

	// 3. Verifica saldo em moedas de bronze (PATCH 4)
	cost := config.PurchaseCostBronze
	if playerInv == nil {
		return fmt.Errorf("player inventory not initialized")
	}

	playerInv.Lock()
	defer playerInv.Unlock()

	if playerInv.GetGold() < cost {
		return fmt.Errorf("insufficient funds in tiered currency: need %d copper/bronze, but have %d", cost, playerInv.GetGold())
	}

	// 4. Executa transação se banco ativo
	if hm.db != nil {
		tx, txErr := hm.db.BeginTx(ctx, nil)
		if txErr != nil {
			return fmt.Errorf("failed to start database transaction: %w", txErr)
		}
		defer tx.Rollback()

		// Dedução autoritativa do gold no banco
		var dbGold int64
		err = tx.QueryRowContext(ctx, "SELECT gold FROM characters WHERE id = $1 FOR UPDATE", charID).Scan(&dbGold)
		if err != nil {
			return fmt.Errorf("failed to lock character wallet: %w", err)
		}

		if dbGold < cost {
			return fmt.Errorf("insufficient database funds: need %d bronze, have %d", cost, dbGold)
		}

		newGold := dbGold - cost
		_, err = tx.ExecContext(ctx, "UPDATE characters SET gold = $1 WHERE id = $2", newGold, charID)
		if err != nil {
			return fmt.Errorf("failed to deduct currency for purchase: %w", err)
		}

		// Insere ou atualiza posse de moradia na tabela
		var query string
		if config.Type == "player" {
			query = `
				INSERT INTO houses (house_id, owner_id, guild_id, purchased_at, last_rent_paid_at, rent_status)
				VALUES ($1, $2, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'active')
				ON CONFLICT (house_id) DO UPDATE 
				SET owner_id = EXCLUDED.owner_id, guild_id = NULL, purchased_at = CURRENT_TIMESTAMP, last_rent_paid_at = CURRENT_TIMESTAMP, rent_status = 'active'
			`
			_, err = tx.ExecContext(ctx, query, houseID, charID)
		} else {
			query = `
				INSERT INTO houses (house_id, owner_id, guild_id, purchased_at, last_rent_paid_at, rent_status)
				VALUES ($1, NULL, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'active')
				ON CONFLICT (house_id) DO UPDATE 
				SET owner_id = NULL, guild_id = EXCLUDED.guild_id, purchased_at = CURRENT_TIMESTAMP, last_rent_paid_at = CURRENT_TIMESTAMP, rent_status = 'active'
			`
			_, err = tx.ExecContext(ctx, query, houseID, guildID)
		}

		if err != nil {
			return fmt.Errorf("failed to record house ownership in database: %w", err)
		}

		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("failed to commit housing transaction: %w", commitErr)
		}
	}

	// 5. Dedução em memória e atualização de estado local
	playerInv.RemoveGold(cost)
	playerInv.SetDirty(true)

	if config.Type == "player" {
		state.OwnerID = charID
		state.OwnerName = playerID
		state.GuildID = 0
		state.GuildName = ""
	} else {
		state.OwnerID = 0
		state.OwnerName = ""
		state.GuildID = guildID
		// Busca nome da guilda para o cache
		if hm.db != nil {
			_ = hm.db.QueryRowContext(ctx, "SELECT name FROM guilds WHERE id = $1", guildID).Scan(&state.GuildName)
		} else {
			state.GuildName = "Mock Guild"
		}
	}

	state.PurchasedAt = time.Now()
	state.LastRentPaidAt = time.Now()
	state.RentStatus = "active"
	state.WarningSentAt = time.Time{}
	state.Permissions = "private"
	state.MinRank = "member"

	slog.Info("House successfully purchased", "player", playerID, "house_id", houseID, "cost", cost, "type", config.Type)
	return nil
}

// PayRent lida com o pagamento recorrente de aluguel (PATCH 3 & PATCH 4)
func (hm *HousingManager) PayRent(playerID string, houseID string, playerInv *inventory.PlayerInventory) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	config, ok := hm.houses[houseID]
	if !ok {
		return fmt.Errorf("house %s not found", houseID)
	}

	state, ok := hm.states[houseID]
	if !ok {
		return fmt.Errorf("house state missing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var charID int
	var guildID int
	var err error
	if hm.db != nil {
		charID, guildID, err = hm.getPlayerDetails(ctx, nil, playerID)
		if err != nil {
			return fmt.Errorf("failed to verify credentials: %w", err)
		}
	} else {
		charID = state.OwnerID
		guildID = state.GuildID
	}

	// Valida se quem está pagando é o legítimo morador (ou da guilda correta)
	if config.Type == "player" && state.OwnerID != charID {
		return fmt.Errorf("only the property owner can pay the rent")
	} else if config.Type == "guild" && state.GuildID != guildID {
		return fmt.Errorf("only members of the owning guild can contribute to citadel rent")
	}

	cost := config.RentCostBronze
	if playerInv == nil {
		return fmt.Errorf("player inventory not initialized")
	}

	playerInv.Lock()
	defer playerInv.Unlock()

	if playerInv.GetGold() < cost {
		return fmt.Errorf("insufficient funds for rent: need %d bronze, have %d", cost, playerInv.GetGold())
	}

	if hm.db != nil {
		tx, txErr := hm.db.BeginTx(ctx, nil)
		if txErr != nil {
			return fmt.Errorf("failed to start database transaction: %w", txErr)
		}
		defer tx.Rollback()

		var dbGold int64
		err = tx.QueryRowContext(ctx, "SELECT gold FROM characters WHERE id = $1 FOR UPDATE", charID).Scan(&dbGold)
		if err != nil {
			return fmt.Errorf("failed to lock character wallet: %w", err)
		}

		if dbGold < cost {
			return fmt.Errorf("insufficient database funds: need %d bronze, have %d", cost, dbGold)
		}

		newGold := dbGold - cost
		_, err = tx.ExecContext(ctx, "UPDATE characters SET gold = $1 WHERE id = $2", newGold, charID)
		if err != nil {
			return fmt.Errorf("failed to deduct currency for rent: %w", err)
		}

		// Atualiza dados de rent na tabela
		_, err = tx.ExecContext(ctx, `
			UPDATE houses 
			SET last_rent_paid_at = CURRENT_TIMESTAMP, rent_status = 'active', warning_sent_at = NULL 
			WHERE house_id = $1
		`, houseID)
		if err != nil {
			return fmt.Errorf("failed to update rent status in DB: %w", err)
		}

		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("failed to commit rent transaction: %w", commitErr)
		}
	}

	playerInv.RemoveGold(cost)
	playerInv.SetDirty(true)

	state.LastRentPaidAt = time.Now()
	state.RentStatus = "active"
	state.WarningSentAt = time.Time{}

	slog.Info("Rent successfully paid", "player", playerID, "house_id", houseID, "amount_bronze", cost)
	return nil
}

// DepositStorage deposita um item do inventário do jogador no baú seguro da casa (PATCH 2 & PATCH 4)
func (hm *HousingManager) DepositStorage(playerID string, houseID string, slotIndex int, itemID string, qty int, playerInv *inventory.PlayerInventory) error {
	if qty <= 0 {
		return fmt.Errorf("invalid quantity")
	}

	// Adquire lock do HousingManager primeiro e PlayerInventory segundo (Evita Deadlock)
	hm.mu.Lock()
	defer hm.mu.Unlock()

	state, ok := hm.states[houseID]
	if !ok {
		return fmt.Errorf("house state not found")
	}

	config, ok := hm.houses[houseID]
	if !ok {
		return fmt.Errorf("house config not found")
	}

	if state.RentStatus == "evicted" {
		return fmt.Errorf("access blocked: this property has been evicted due to non-payment")
	}

	if playerInv == nil {
		return fmt.Errorf("player inventory not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var charID int
	var err error
	if hm.db != nil {
		charID, _, err = hm.getPlayerDetails(ctx, nil, playerID)
		if err != nil {
			return fmt.Errorf("failed to verify credentials: %w", err)
		}
	} else {
		charID = state.OwnerID
	}

	// Valida acesso seguro via ACL / Cargo de Guilda (PATCH 7)
	if err := hm.checkPermission(ctx, playerID, houseID, "storage_deposit"); err != nil {
		return err
	}

	playerInv.Lock()
	defer playerInv.Unlock()

	// Valida se o jogador possui o item e quantidade suficientes
	inventoryItem, hasSlotItem := playerInv.Items[slotIndex]
	if !hasSlotItem || inventoryItem == nil {
		return fmt.Errorf("item not found in inventory slot %d", slotIndex)
	}

	if inventoryItem.ItemID != itemID {
		return fmt.Errorf("item in slot %d does not match %s", slotIndex, itemID)
	}

	if inventoryItem.Quantity < qty {
		return fmt.Errorf("insufficient quantity: you have %d, requested %d", inventoryItem.Quantity, qty)
	}

	// Calcula capacidade de armazenamento dinâmica (com suporte a baús extras)
	bonusSlots := 0
	for _, dec := range state.Decorations {
		if fConfig, exists := hm.furniture[dec.FurnitureID]; exists {
			bonusSlots += fConfig.StorageBonus
		}
	}
	totalCapacity := config.StorageCapacity + bonusSlots

	// Verifica se a casa atingiu a capacidade máxima de slots ocupados
	currentOcupiedSlots := len(state.Storage)
	_, alreadyInTargetSlot := state.Storage[slotIndex]
	if currentOcupiedSlots >= totalCapacity && !alreadyInTargetSlot {
		return fmt.Errorf("secure storage at maximum capacity: limited to %d slots", totalCapacity)
	}

	// PATCH 2 Transactional database safety: Execute atomic operation
	if hm.db != nil {
		tx, txErr := hm.db.BeginTx(ctx, nil)
		if txErr != nil {
			return fmt.Errorf("failed to start database transaction: %w", txErr)
		}
		defer tx.Rollback()

		// 1. Remove do inventário do banco
		var dbInvQty int
		var dbInvDurability int
		dbQueryInv := "SELECT quantity, durability FROM character_inventory WHERE character_id = $1 AND slot_index = $2 FOR UPDATE"
		err = tx.QueryRowContext(ctx, dbQueryInv, charID, slotIndex).Scan(&dbInvQty, &dbInvDurability)
		if err != nil {
			return fmt.Errorf("failed to lock character inventory item in DB: %w", err)
		}

		if dbInvQty < qty {
			return fmt.Errorf("database validation failed: insufficient item quantity in DB")
		}

		if dbInvQty == qty {
			_, err = tx.ExecContext(ctx, "DELETE FROM character_inventory WHERE character_id = $1 AND slot_index = $2", charID, slotIndex)
		} else {
			_, err = tx.ExecContext(ctx, "UPDATE character_inventory SET quantity = quantity - $1 WHERE character_id = $2 AND slot_index = $3", qty, charID, slotIndex)
		}
		if err != nil {
			return fmt.Errorf("failed to update character inventory in database transaction: %w", err)
		}

		// 2. Adiciona ou une no baú seguro da casa no banco
		var existingHouseQty int
		dbQueryHouse := "SELECT quantity FROM house_storage WHERE house_id = $1 AND slot_id = $2 FOR UPDATE"
		rowErr := tx.QueryRowContext(ctx, dbQueryHouse, houseID, slotIndex).Scan(&existingHouseQty)

		if errors.Is(rowErr, sql.ErrNoRows) {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO house_storage (house_id, slot_id, item_id, quantity, durability)
				VALUES ($1, $2, $3, $4, $5)
			`, houseID, slotIndex, itemID, qty, dbInvDurability)
		} else if rowErr == nil {
			_, err = tx.ExecContext(ctx, `
				UPDATE house_storage SET quantity = quantity + $1 
				WHERE house_id = $2 AND slot_id = $3
			`, qty, houseID, slotIndex)
		} else {
			return fmt.Errorf("failed to query house secure storage: %w", rowErr)
		}

		if err != nil {
			return fmt.Errorf("failed to save house storage in database transaction: %w", err)
		}

		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("failed to commit secure deposit: %w", commitErr)
		}
	}

	// 3. Atualiza estado em memória de forma consistente
	durability := inventoryItem.Durability
	itemUUID := inventoryItem.ItemUUID

	// Dedução no inventário
	if inventoryItem.Quantity == qty {
		delete(playerInv.Items, slotIndex)
	} else {
		inventoryItem.Quantity -= qty
	}
	playerInv.SetDirty(true)

	// Adição ao baú seguro
	targetItem, exists := state.Storage[slotIndex]
	if !exists {
		state.Storage[slotIndex] = &inventory.InventoryItem{
			ItemID:     itemID,
			Quantity:   qty,
			Durability: durability,
			SlotIndex:  slotIndex,
			ItemUUID:   itemUUID,
		}
	} else {
		targetItem.Quantity += qty
	}

	slog.Info("Item successfully deposited in secure house storage", "player", playerID, "house_id", houseID, "item_id", itemID, "quantity", qty)
	return nil
}

// WithdrawStorage retira um item do baú seguro da casa e coloca de volta no inventário (PATCH 2 & PATCH 4)
func (hm *HousingManager) WithdrawStorage(playerID string, houseID string, slotIndex int, qty int, playerInv *inventory.PlayerInventory) error {
	if qty <= 0 {
		return fmt.Errorf("invalid quantity")
	}

	// Adquire lock do HousingManager primeiro e PlayerInventory segundo (Evita Deadlock)
	hm.mu.Lock()
	defer hm.mu.Unlock()

	state, ok := hm.states[houseID]
	if !ok {
		return fmt.Errorf("house state not found")
	}

	_, ok = hm.houses[houseID]
	if !ok {
		return fmt.Errorf("house config not found")
	}

	if state.RentStatus == "evicted" {
		return fmt.Errorf("access blocked: this property has been evicted")
	}

	if playerInv == nil {
		return fmt.Errorf("player inventory not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var charID int
	var err error
	if hm.db != nil {
		charID, _, err = hm.getPlayerDetails(ctx, nil, playerID)
		if err != nil {
			return fmt.Errorf("failed to verify credentials: %w", err)
		}
	} else {
		charID = state.OwnerID
	}

	// Valida permissão de acesso via ACL / Cargo de Guilda (PATCH 7)
	if err := hm.checkPermission(ctx, playerID, houseID, "storage_withdraw"); err != nil {
		return err
	}

	// Verifica se item existe no baú da casa
	houseItem, ok := state.Storage[slotIndex]
	if !ok || houseItem == nil {
		return fmt.Errorf("no item found in house storage slot %d", slotIndex)
	}

	if houseItem.Quantity < qty {
		return fmt.Errorf("insufficient quantity in house storage: requested %d, available %d", qty, houseItem.Quantity)
	}

	playerInv.Lock()
	defer playerInv.Unlock()

	// Verifica se há espaço no inventário do jogador (simplificado por adicionar em qualquer slot)
	// Vamos colocar no mesmo slotIndex no inventário se estiver vazio, ou buscar um slot vazio
	targetSlot := slotIndex
	if piItem, exists := playerInv.Items[targetSlot]; exists && piItem != nil {
		// Busca o primeiro slot vazio de 4 a 30 (mochila)
		found := false
		for i := 4; i < 30; i++ {
			if _, taken := playerInv.Items[i]; !taken {
				targetSlot = i
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("insufficient inventory slots to receive item")
		}
	}

	// PATCH 2 Transactional database safety: Execute atomic operation
	if hm.db != nil {
		tx, txErr := hm.db.BeginTx(ctx, nil)
		if txErr != nil {
			return fmt.Errorf("failed to start database transaction: %w", txErr)
		}
		defer tx.Rollback()

		// 1. Remove do baú do banco
		var dbHouseQty int
		var dbHouseDurability int
		dbQueryHouse := "SELECT quantity, durability FROM house_storage WHERE house_id = $1 AND slot_id = $2 FOR UPDATE"
		err = tx.QueryRowContext(ctx, dbQueryHouse, houseID, slotIndex).Scan(&dbHouseQty, &dbHouseDurability)
		if err != nil {
			return fmt.Errorf("failed to lock house storage item in DB: %w", err)
		}

		if dbHouseQty < qty {
			return fmt.Errorf("database validation failed: insufficient item quantity in DB storage")
		}

		if dbHouseQty == qty {
			_, err = tx.ExecContext(ctx, "DELETE FROM house_storage WHERE house_id = $1 AND slot_id = $2", houseID, slotIndex)
		} else {
			_, err = tx.ExecContext(ctx, "UPDATE house_storage SET quantity = quantity - $1 WHERE house_id = $2 AND slot_id = $3", qty, houseID, slotIndex)
		}
		if err != nil {
			return fmt.Errorf("failed to update house storage in DB transaction: %w", err)
		}

		// 2. Insere ou une no inventário do jogador no banco
		var existingInvQty int
		dbQueryInv := "SELECT quantity FROM character_inventory WHERE character_id = $1 AND slot_index = $2 FOR UPDATE"
		rowErr := tx.QueryRowContext(ctx, dbQueryInv, charID, targetSlot).Scan(&existingInvQty)

		if errors.Is(rowErr, sql.ErrNoRows) {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO character_inventory (character_id, slot_index, item_id, quantity, durability, item_uuid)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, charID, targetSlot, houseItem.ItemID, qty, dbHouseDurability, houseItem.ItemUUID)
		} else if rowErr == nil {
			_, err = tx.ExecContext(ctx, `
				UPDATE character_inventory SET quantity = quantity + $1 
				WHERE character_id = $2 AND slot_index = $3
			`, qty, charID, targetSlot)
		} else {
			return fmt.Errorf("failed to query character inventory: %w", rowErr)
		}

		if err != nil {
			return fmt.Errorf("failed to save character inventory in database transaction: %w", err)
		}

		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("failed to commit secure withdraw: %w", commitErr)
		}
	}

	// 3. Atualiza estado em memória de forma consistente
	durability := houseItem.Durability
	itemUUID := houseItem.ItemUUID
	itemID := houseItem.ItemID

	// Dedução no baú
	if houseItem.Quantity == qty {
		delete(state.Storage, slotIndex)
	} else {
		houseItem.Quantity -= qty
	}

	// Adição no inventário
	piItem, exists := playerInv.Items[targetSlot]
	if !exists {
		playerInv.Items[targetSlot] = &inventory.InventoryItem{
			ItemID:     itemID,
			Quantity:   qty,
			Durability: durability,
			SlotIndex:  targetSlot,
			ItemUUID:   itemUUID,
		}
	} else {
		piItem.Quantity += qty
	}
	playerInv.SetDirty(true)

	slog.Info("Item successfully withdrawn from secure house storage", "player", playerID, "house_id", houseID, "item_id", itemID, "quantity", qty)
	return nil
}

// PlaceFurniture coloca mobília ou decoração fisicamente na casa (PATCH 1 & PATCH 4)
func (hm *HousingManager) PlaceFurniture(playerID string, houseID string, furnitureID string, x, y, z, rotation float64, playerInv *inventory.PlayerInventory) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	fConfig, ok := hm.furniture[furnitureID]
	if !ok {
		return fmt.Errorf("furniture blueprint %s not found", furnitureID)
	}

	state, ok := hm.states[houseID]
	if !ok {
		return fmt.Errorf("house state missing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var charID int
	var err error
	if hm.db != nil {
		charID, _, err = hm.getPlayerDetails(ctx, nil, playerID)
		if err != nil {
			return fmt.Errorf("failed to verify credentials: %w", err)
		}
	} else {
		charID = state.OwnerID
	}

	// Valida permissão de decoração (PATCH 7)
	if err := hm.canEditDecorations(ctx, playerID, houseID); err != nil {
		return err
	}

	// Valida orçamento de decorações (PATCH 8)
	currentPoints := 0
	for _, dec := range state.Decorations {
		if fConf, exists := hm.furniture[dec.FurnitureID]; exists {
			currentPoints += fConf.DecorationCost
		}
	}
	config, ok := hm.houses[houseID]
	if !ok {
		return fmt.Errorf("house config missing")
	}
	budget := hm.GetDecorationBudget(config.Size)
	if currentPoints+fConfig.DecorationCost > budget {
		return fmt.Errorf("unable to place furniture: decoration budget exceeded (%d/%d points)", currentPoints+fConfig.DecorationCost, budget)
	}

	if playerInv == nil {
		return fmt.Errorf("player inventory not initialized")
	}

	playerInv.Lock()
	defer playerInv.Unlock()

	// Valida se o jogador possui o item de mobília na mochila
	foundSlot := -1
	for slot, item := range playerInv.Items {
		if item != nil && item.ItemID == furnitureID {
			foundSlot = slot
			break
		}
	}
	if foundSlot == -1 {
		return fmt.Errorf("you do not have %s in your inventory to decorate", fConfig.Name)
	}

	// Remove do inventário
	if playerInv.Items[foundSlot].Quantity == 1 {
		delete(playerInv.Items, foundSlot)
	} else {
		playerInv.Items[foundSlot].Quantity--
	}
	playerInv.SetDirty(true)

	// Persiste no banco de dados se conectado
	var placedID int = len(state.Decorations) + 1
	if hm.db != nil {
		tx, txErr := hm.db.BeginTx(ctx, nil)
		if txErr != nil {
			return fmt.Errorf("failed to start database transaction: %w", txErr)
		}
		defer tx.Rollback()

		// 1. Remove do inventário no banco
		_, err = tx.ExecContext(ctx, "DELETE FROM character_inventory WHERE character_id = $1 AND slot_index = $2", charID, foundSlot)
		if err != nil {
			return fmt.Errorf("failed to deduct furniture item from database inventory: %w", err)
		}

		// 2. Insere na tabela de decorações
		err = tx.QueryRowContext(ctx, `
			INSERT INTO house_decorations (house_id, furniture_id, x, y, z, rotation)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, houseID, furnitureID, x, y, z, rotation).Scan(&placedID)
		if err != nil {
			return fmt.Errorf("failed to record physical decoration placement: %w", err)
		}

		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("failed to commit decoration: %w", commitErr)
		}
	}

	// Adiciona na lista local em memória
	state.Decorations = append(state.Decorations, PlacedDecoration{
		ID:          placedID,
		FurnitureID: furnitureID,
		X:           x,
		Y:           y,
		Z:           z,
		Rotation:    rotation,
	})

	slog.Info("Furniture successfully placed in house", "player", playerID, "house_id", houseID, "furniture_id", furnitureID, "decoration_id", placedID)
	return nil
}

// RemoveFurniture recolhe uma mobília da casa e envia de volta ao inventário (PATCH 1 & PATCH 4)
func (hm *HousingManager) RemoveFurniture(playerID string, houseID string, decorationID int, playerInv *inventory.PlayerInventory) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	state, ok := hm.states[houseID]
	if !ok {
		return fmt.Errorf("house state missing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var charID int
	var err error
	if hm.db != nil {
		charID, _, err = hm.getPlayerDetails(ctx, nil, playerID)
		if err != nil {
			return fmt.Errorf("failed to verify credentials: %w", err)
		}
	} else {
		charID = state.OwnerID
	}

	// Valida permissão de decoração (PATCH 7)
	if err := hm.canEditDecorations(ctx, playerID, houseID); err != nil {
		return err
	}

	// Procura decoração correspondente
	foundIdx := -1
	var dec PlacedDecoration
	for idx, d := range state.Decorations {
		if d.ID == decorationID {
			foundIdx = idx
			dec = d
			break
		}
	}
	if foundIdx == -1 {
		return fmt.Errorf("decoration with ID %d not found in this property", decorationID)
	}

	if playerInv == nil {
		return fmt.Errorf("player inventory not initialized")
	}

	playerInv.Lock()
	defer playerInv.Unlock()

	// Encontra slot livre no inventário (slots 4 a 30)
	freeSlot := -1
	for i := 4; i < 30; i++ {
		if _, exists := playerInv.Items[i]; !exists {
			freeSlot = i
			break
		}
	}
	if freeSlot == -1 {
		return fmt.Errorf("no open slots in inventory to retrieve furniture")
	}

	// Persiste no banco de dados se conectado
	if hm.db != nil {
		tx, txErr := hm.db.BeginTx(ctx, nil)
		if txErr != nil {
			return fmt.Errorf("failed to start database transaction: %w", txErr)
		}
		defer tx.Rollback()

		// Delete from decorations table
		_, err = tx.ExecContext(ctx, "DELETE FROM house_decorations WHERE id = $1 AND house_id = $2", decorationID, houseID)
		if err != nil {
			return fmt.Errorf("failed to remove decoration from database: %w", err)
		}

		// Insert back into player inventory
		_, err = tx.ExecContext(ctx, `
			INSERT INTO character_inventory (character_id, slot_index, item_id, quantity, durability)
			VALUES ($1, $2, $3, 1, 100)
		`, charID, freeSlot, dec.FurnitureID)
		if err != nil {
			return fmt.Errorf("failed to restore furniture to player database inventory: %w", err)
		}

		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("failed to commit decoration removal: %w", commitErr)
		}
	}

	// Remove da lista em memória
	state.Decorations = append(state.Decorations[:foundIdx], state.Decorations[foundIdx+1:]...)

	// Restaura no inventário local
	playerInv.Items[freeSlot] = &inventory.InventoryItem{
		ItemID:     dec.FurnitureID,
		Quantity:   1,
		Durability: 100,
		SlotIndex:  freeSlot,
	}
	playerInv.SetDirty(true)

	slog.Info("Furniture successfully removed from house", "player", playerID, "house_id", houseID, "furniture_id", dec.FurnitureID, "decoration_id", decorationID)
	return nil
}

// EvictHouse processa a evicção de uma moradia devido ao aluguel atrasado, enviando itens para reclaim_storage (PATCH 3)
func (hm *HousingManager) EvictHouse(houseID string) error {
	state, ok := hm.states[houseID]
	if !ok {
		return fmt.Errorf("house state missing")
	}

	_, ok = hm.houses[houseID]
	if !ok {
		return fmt.Errorf("house config missing")
	}

	slog.Warn("Eviction triggered on open-world property due to rent non-payment", "house_id", houseID, "owner_id", state.OwnerID, "guild_id", state.GuildID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if hm.db != nil {
		tx, txErr := hm.db.BeginTx(ctx, nil)
		if txErr != nil {
			return fmt.Errorf("failed to begin eviction transaction: %w", txErr)
		}
		defer tx.Rollback()

		// 1. Move os itens do house_storage para house_reclaim_storage no banco
		if state.OwnerID > 0 {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO house_reclaim_storage (owner_id, guild_id, item_id, quantity, durability)
				SELECT $1, NULL, item_id, quantity, durability FROM house_storage WHERE house_id = $2
			`, state.OwnerID, houseID)
			if err != nil {
				return fmt.Errorf("failed to transfer secure items to character reclaim storage: %w", err)
			}
		} else if state.GuildID > 0 {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO house_reclaim_storage (owner_id, guild_id, item_id, quantity, durability)
				SELECT NULL, $1, item_id, quantity, durability FROM house_storage WHERE house_id = $2
			`, state.GuildID, houseID)
			if err != nil {
				return fmt.Errorf("failed to transfer secure items to guild reclaim storage: %w", err)
			}
		}

		// 2. Transfere mobílias colocadas (decorações) também para o reclaim do banco
		if state.OwnerID > 0 {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO house_reclaim_storage (owner_id, guild_id, item_id, quantity, durability)
				SELECT $1, NULL, furniture_id, 1, 100 FROM house_decorations WHERE house_id = $2
			`, state.OwnerID, houseID)
			if err != nil {
				return fmt.Errorf("failed to transfer decorations to character reclaim storage: %w", err)
			}
		}

		// 3. Deleta registros de armazenamento e decorações da casa
		_, _ = tx.ExecContext(ctx, "DELETE FROM house_storage WHERE house_id = $1", houseID)
		_, _ = tx.ExecContext(ctx, "DELETE FROM house_decorations WHERE house_id = $1", houseID)

		// 4. Limpa posse e marca estado na tabela houses
		_, err := tx.ExecContext(ctx, `
			UPDATE houses 
			SET owner_id = NULL, guild_id = NULL, last_rent_paid_at = CURRENT_TIMESTAMP, rent_status = 'active', warning_sent_at = NULL 
			WHERE house_id = $1
		`, houseID)
		if err != nil {
			return fmt.Errorf("failed to reset house ownership: %w", err)
		}

		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("failed to commit eviction transaction: %w", commitErr)
		}
	}

	// 5. Atualiza estado em memória
	state.OwnerID = 0
	state.OwnerName = ""
	state.GuildID = 0
	state.GuildName = ""
	state.Storage = make(map[int]*inventory.InventoryItem)
	state.Decorations = []PlacedDecoration{}
	state.RentStatus = "active" // Fica ativa/limpa para a próxima compra
	state.WarningSentAt = time.Time{}

	slog.Info("Eviction finalized successfully. All stored items and decorations moved to reclaim queue.", "house_id", houseID)
	return nil
}

// ReclaimStoredItems recolhe todos os itens sob custódia devido a evicção anterior e insere de volta no inventário do jogador (PATCH 3)
func (hm *HousingManager) ReclaimStoredItems(playerID string, playerInv *inventory.PlayerInventory) (int, error) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if playerInv == nil {
		return 0, fmt.Errorf("player inventory not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var charID int
	var err error
	if hm.db != nil {
		charID, _, err = hm.getPlayerDetails(ctx, nil, playerID)
		if err != nil {
			return 0, fmt.Errorf("failed to verify credentials: %w", err)
		}
	} else {
		// Mock ID fallback
		charID = 1001
	}

	playerInv.Lock()
	defer playerInv.Unlock()

	// 1. Busca itens sob custódia
	type ReclaimItem struct {
		ID         int
		ItemID     string
		Quantity   int
		Durability int
	}
	var list []ReclaimItem

	if hm.db != nil {
		rows, err := hm.db.QueryContext(ctx, "SELECT id, item_id, quantity, durability FROM house_reclaim_storage WHERE owner_id = $1", charID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r ReclaimItem
				if scanErr := rows.Scan(&r.ID, &r.ItemID, &r.Quantity, &r.Durability); scanErr == nil {
					list = append(list, r)
				}
			}
		}
	} else {
		// Fallback mock (adiciona um item mockado se o inventário estivesse vazio de reclaim)
		list = append(list, ReclaimItem{ID: 1, ItemID: "potion_heal", Quantity: 5, Durability: 100})
	}

	if len(list) == 0 {
		return 0, nil
	}

	// 2. Coloca os itens de volta no inventário do jogador (respeitando limites de slots)
	reclaimedCount := 0
	for _, rItem := range list {
		// Procura primeiro slot livre
		freeSlot := -1
		for i := 4; i < 30; i++ {
			if _, exists := playerInv.Items[i]; !exists {
				freeSlot = i
				break
			}
		}
		if freeSlot == -1 {
			// Sem espaço restante
			slog.Warn("Player inventory full during reclaim execution, stopping transfer", "player", playerID)
			break
		}

		// Transfere no banco se ativo
		if hm.db != nil {
			tx, txErr := hm.db.BeginTx(ctx, nil)
			if txErr != nil {
				return reclaimedCount, fmt.Errorf("failed to begin reclaim item transfer transaction: %w", txErr)
			}
			defer tx.Rollback()

			// Remove da tabela reclaim
			_, err = tx.ExecContext(ctx, "DELETE FROM house_reclaim_storage WHERE id = $1", rItem.ID)
			if err != nil {
				return reclaimedCount, fmt.Errorf("failed to delete reclaimed item from queue: %w", err)
			}

			// Insere no inventário do jogador
			_, err = tx.ExecContext(ctx, `
				INSERT INTO character_inventory (character_id, slot_index, item_id, quantity, durability)
				VALUES ($1, $2, $3, $4, $5)
			`, charID, freeSlot, rItem.ItemID, rItem.Quantity, rItem.Durability)
			if err != nil {
				return reclaimedCount, fmt.Errorf("failed to transfer reclaimed item to inventory in DB: %w", err)
			}

			if commitErr := tx.Commit(); commitErr != nil {
				return reclaimedCount, fmt.Errorf("failed to commit reclaim transfer: %w", commitErr)
			}
		}

		// Adiciona no inventário em memória
		playerInv.Items[freeSlot] = &inventory.InventoryItem{
			ItemID:     rItem.ItemID,
			Quantity:   rItem.Quantity,
			Durability: rItem.Durability,
			SlotIndex:  freeSlot,
		}
		playerInv.SetDirty(true)
		reclaimedCount++
	}

	slog.Info("Reclaimed stored items successfully retrieved", "player", playerID, "count", reclaimedCount)
	return reclaimedCount, nil
}

// GetPlayerActiveHouseLocation retorna as coordenadas da moradia do jogador se houver um ativador de respawn válido (PATCH 5 & PATCH 6)
func (hm *HousingManager) GetPlayerActiveHouseLocation(playerID string) (float64, float64, int, string, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var charID int
	if hm.db != nil {
		_ = hm.db.QueryRowContext(ctx, "SELECT id FROM characters WHERE name = $1", playerID).Scan(&charID)
	} else {
		charID = 1001
	}

	if charID <= 0 {
		return 0, 0, 0, "", false
	}

	// Varre todas as moradias em memória buscando a do jogador
	for houseID, s := range hm.states {
		if s.OwnerID == charID && s.RentStatus != "evicted" {
			// PATCH 6: Verifica se há mobília de ativação de respawn instalada (Bed, Soul Anchor, ou Hearth Sigil)
			hasRespawnFurniture := false
			for _, dec := range s.Decorations {
				if fConfig, exists := hm.furniture[dec.FurnitureID]; exists {
					cat := strings.ToLower(fConfig.Category)
					id := strings.ToLower(fConfig.ID)
					name := strings.ToLower(fConfig.Name)
					
					if cat == "beds" ||
						strings.Contains(id, "bed") || strings.Contains(id, "soul_anchor") || strings.Contains(id, "hearth_sigil") ||
						strings.Contains(name, "bed") || strings.Contains(name, "soul_anchor") || strings.Contains(name, "hearth_sigil") {
						hasRespawnFurniture = true
						break
					}
				}
			}

			if !hasRespawnFurniture {
				slog.Warn("Player has owned property but lacks active respawn activator (Bed, Soul Anchor, Hearth Sigil)", "player", playerID, "house_id", houseID)
				continue
			}

			if config, exists := hm.houses[houseID]; exists {
				return config.X, config.Y, config.Z, config.Name, true
			}
		}
	}

	return 0, 0, 0, "", false
}

// CheckRentStatuses audita todas as moradias, mudando para warning ou disparando evicções imediatas (PATCH 3)
func (hm *HousingManager) CheckRentStatuses() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	now := time.Now()
	slog.Info("Running open-world housing rent audit...")

	for houseID, state := range hm.states {
		if state.OwnerID == 0 && state.GuildID == 0 {
			continue // Casa sem dono não paga aluguel
		}

		// Calcula tempo desde o último pagamento de aluguel
		elapsed := now.Sub(state.LastRentPaidAt)
		if elapsed > hm.warningInterval {
			// Rent expirou!
			if state.RentStatus == "active" {
				// Transiciona para warning e marca timestamp
				state.RentStatus = "warning"
				state.WarningSentAt = now
				slog.Warn("House rent expired! Set warning state. Owner has grace period.", "house_id", houseID)

				if hm.db != nil {
					_, _ = hm.db.Exec("UPDATE houses SET rent_status = 'warning', warning_sent_at = $1 WHERE house_id = $2", now, houseID)
				}
			} else if state.RentStatus == "warning" {
				// Se já em warning, valida se ultrapassou a carência (grace period)
				if now.Sub(state.WarningSentAt) > hm.gracePeriod {
					// Carência esgotada! EVICTA!
					slog.Error("House rent grace period expired! Triggering eviction.", "house_id", houseID)
					hm.mu.Unlock()
					_ = hm.EvictHouse(houseID)
					hm.mu.Lock()
				}
			}
		}
	}
}

// startRentTicker roda a tarefa periódica de auditoria de aluguel em background
func (hm *HousingManager) startRentTicker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		hm.CheckRentStatuses()
	}
}

// Helper para buscar informações de personagens de forma centralizada e resiliente (corrigido para character_name no guild_members)
func (hm *HousingManager) getPlayerDetails(ctx context.Context, tx *sql.Tx, playerID string) (charID int, guildID int, err error) {
	var row *sql.Row
	queryChar := "SELECT id FROM characters WHERE name = $1"
	if tx != nil {
		row = tx.QueryRowContext(ctx, queryChar, playerID)
	} else {
		row = hm.db.QueryRowContext(ctx, queryChar, playerID)
	}
	err = row.Scan(&charID)
	if err != nil {
		return 0, 0, err
	}

	queryGuild := "SELECT guild_id FROM guild_members WHERE character_name = $1"
	if tx != nil {
		row = tx.QueryRowContext(ctx, queryGuild, playerID)
	} else {
		row = hm.db.QueryRowContext(ctx, queryGuild, playerID)
	}
	err = row.Scan(&guildID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return charID, 0, nil
		}
		return charID, 0, err
	}
	return charID, guildID, nil
}

// GetDecorationBudget retorna o orçamento de decoração por tamanho de moradia (PATCH 8)
func (hm *HousingManager) GetDecorationBudget(size string) int {
	switch strings.ToLower(size) {
	case "small":
		return 50
	case "medium":
		return 100
	case "large":
		return 200
	default:
		return 50
	}
}

// checkPermission valida o acesso de um jogador à casa de acordo com ACL/rank (PATCH 7)
func (hm *HousingManager) checkPermission(ctx context.Context, playerID string, houseID string, action string) error {
	state, ok := hm.states[houseID]
	if !ok {
		return fmt.Errorf("house state missing")
	}
	config, ok := hm.houses[houseID]
	if !ok {
		return fmt.Errorf("house config missing")
	}

	var charID int
	var playerGuildID int
	var playerRole int = -1
	var err error

	if hm.db != nil {
		charID, playerGuildID, err = hm.getPlayerDetails(ctx, nil, playerID)
		if err != nil {
			return fmt.Errorf("failed to retrieve character credentials: %w", err)
		}
		if playerGuildID > 0 {
			_ = hm.db.QueryRowContext(ctx, "SELECT role FROM guild_members WHERE character_name = $1 AND guild_id = $2", playerID, playerGuildID).Scan(&playerRole)
		}
	} else {
		if playerID == state.OwnerName {
			charID = state.OwnerID
		} else {
			charID = 9999
		}
		playerGuildID = state.GuildID
		playerRole = 0
	}

	// Proprietário direto da player house sempre tem acesso irrestrito
	if config.Type == "player" && charID > 0 && state.OwnerID == charID {
		return nil
	}

	// Validação de moradia de guilda
	if config.Type == "guild" {
		if playerGuildID == 0 || playerGuildID != state.GuildID {
			return fmt.Errorf("unauthorized access: your guild does not own this citadel")
		}

		// Valida nível de cargo mínimo configurado (leader, vice, member, recruit)
		requiredRole := 0
		switch strings.ToLower(state.MinRank) {
		case "leader":
			requiredRole = 2
		case "vice":
			requiredRole = 1
		case "member":
			requiredRole = 0
		case "recruit":
			requiredRole = 0 // Qualquer membro da guilda
		default:
			requiredRole = 0
		}

		if playerRole < requiredRole {
			return fmt.Errorf("unauthorized access: your guild rank (%d) is lower than required level (%s)", playerRole, state.MinRank)
		}
		return nil
	}

	// Validação de moradia de jogador individual (ACL base)
	perm := strings.ToLower(state.Permissions)
	if perm == "" {
		perm = "private"
	}

	switch perm {
	case "private":
		return fmt.Errorf("unauthorized access: this property storage is set to private")

	case "friends":
		if hm.db != nil {
			var isFriend bool
			err = hm.db.QueryRowContext(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM social_relations 
					WHERE character_name = $1 AND target_name = $2 AND relation_type = 'friend'
				)
			`, state.OwnerName, playerID).Scan(&isFriend)
			if err != nil || !isFriend {
				return fmt.Errorf("unauthorized access: only friends of the owner can access this storage")
			}
		} else {
			return fmt.Errorf("unauthorized access: friends permission requires DB connection")
		}

	case "guild":
		if hm.db != nil {
			var ownerGuildID int
			_ = hm.db.QueryRowContext(ctx, "SELECT guild_id FROM guild_members WHERE character_name = $1", state.OwnerName).Scan(&ownerGuildID)
			if ownerGuildID == 0 || playerGuildID == 0 || ownerGuildID != playerGuildID {
				return fmt.Errorf("unauthorized access: only guild members of the owner can access this storage")
			}
		} else {
			return fmt.Errorf("unauthorized access: guild permission requires DB connection")
		}

	case "public":
		return nil

	default:
		return fmt.Errorf("unauthorized access: invalid permission config")
	}

	return nil
}

// canEditDecorations valida se o jogador pode colocar/remover móveis (PATCH 7)
func (hm *HousingManager) canEditDecorations(ctx context.Context, playerID string, houseID string) error {
	state, ok := hm.states[houseID]
	if !ok {
		return fmt.Errorf("house state missing")
	}
	config, ok := hm.houses[houseID]
	if !ok {
		return fmt.Errorf("house config missing")
	}

	var charID int
	var playerGuildID int
	var playerRole int = -1
	var err error

	if hm.db != nil {
		charID, playerGuildID, err = hm.getPlayerDetails(ctx, nil, playerID)
		if err != nil {
			return fmt.Errorf("failed to retrieve character credentials: %w", err)
		}
		if playerGuildID > 0 {
			_ = hm.db.QueryRowContext(ctx, "SELECT role FROM guild_members WHERE character_name = $1 AND guild_id = $2", playerID, playerGuildID).Scan(&playerRole)
		}
	} else {
		if playerID == state.OwnerName {
			charID = state.OwnerID
		} else {
			charID = 9999
		}
		playerGuildID = state.GuildID
		playerRole = 0
	}

	if config.Type == "player" {
		if charID <= 0 || state.OwnerID != charID {
			return fmt.Errorf("only the house owner can rearrange or remove decoration setups")
		}
		return nil
	}

	// Guild house decoration editing requirements: Leader or Vice (role >= 1)
	if playerGuildID == 0 || playerGuildID != state.GuildID {
		return fmt.Errorf("your guild does not own this citadel")
	}
	if playerRole < 1 {
		return fmt.Errorf("only guild leaders or vices can place/remove decorative arrangements")
	}
	return nil
}

// UpdatePermissions altera a permissão (ACL) de uma moradia de jogador (PATCH 7)
func (hm *HousingManager) UpdatePermissions(playerID string, houseID string, permissions string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	state, ok := hm.states[houseID]
	if !ok {
		return fmt.Errorf("house state not found")
	}

	config, ok := hm.houses[houseID]
	if !ok {
		return fmt.Errorf("house config not found")
	}

	if config.Type != "player" {
		return fmt.Errorf("only player houses support custom ACL permissions")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	charID, _, err := hm.getPlayerDetails(ctx, nil, playerID)
	if err != nil || state.OwnerID != charID {
		return fmt.Errorf("unauthorized: only the property owner can modify access controls")
	}

	permissions = strings.ToLower(strings.TrimSpace(permissions))
	switch permissions {
	case "private", "friends", "guild", "public":
		// Válido
	default:
		return fmt.Errorf("invalid permission type: must be private, friends, guild, or public")
	}

	state.Permissions = permissions

	if hm.db != nil {
		_, err = hm.db.ExecContext(ctx, "UPDATE houses SET permissions = $1 WHERE house_id = $2", permissions, houseID)
		if err != nil {
			return fmt.Errorf("failed to persist permission changes to database: %w", err)
		}
	}

	slog.Info("Property permissions updated", "house_id", houseID, "permissions", permissions)
	return nil
}

// UpdateMinRank altera o rank mínimo para interagir com uma casa de guilda (PATCH 7)
func (hm *HousingManager) UpdateMinRank(playerID string, houseID string, minRank string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	state, ok := hm.states[houseID]
	if !ok {
		return fmt.Errorf("house state not found")
	}

	config, ok := hm.houses[houseID]
	if !ok {
		return fmt.Errorf("house config not found")
	}

	if config.Type != "guild" {
		return fmt.Errorf("only guild houses support minimum rank permissions")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, playerGuildID, err := hm.getPlayerDetails(ctx, nil, playerID)
	if err != nil || playerGuildID == 0 || playerGuildID != state.GuildID {
		return fmt.Errorf("unauthorized: your guild does not own this stronghold")
	}

	// Verifica se é líder da guilda
	var role int
	err = hm.db.QueryRowContext(ctx, "SELECT role FROM guild_members WHERE character_name = $1 AND guild_id = $2", playerID, playerGuildID).Scan(&role)
	if err != nil || role < 2 { // 2 = Leader
		return fmt.Errorf("unauthorized: only the guild leader can change rank permissions")
	}

	minRank = strings.ToLower(strings.TrimSpace(minRank))
	switch minRank {
	case "leader", "vice", "member", "recruit":
		// Válido
	default:
		return fmt.Errorf("invalid rank level: must be leader, vice, member, or recruit")
	}

	state.MinRank = minRank

	if hm.db != nil {
		_, err = hm.db.ExecContext(ctx, "UPDATE houses SET min_rank = $1 WHERE house_id = $2", minRank, houseID)
		if err != nil {
			return fmt.Errorf("failed to persist rank permission changes to database: %w", err)
		}
	}

	slog.Info("Guild property minimum rank permissions updated", "house_id", houseID, "min_rank", minRank)
	return nil
}

// =============================================================================
// PATCH 9 — FUTURE HOUSING MARKET HOOKS
// =============================================================================

// HousingMarketHook representa o contrato para extensões futuras do mercado de moradias (Leilões e Revendas)
type HousingMarketHook interface {
	RegisterPropertyForAuction(ctx context.Context, houseID string, startingBid int64, duration time.Duration) error
	PlaceBidOnProperty(ctx context.Context, playerID string, houseID string, bidAmount int64) error
	FinalizeAuction(ctx context.Context, houseID string) error
	TransferOwnershipDirect(ctx context.Context, houseID string, newOwnerID int, salePrice int64) error
}

// PropertyListing representa um anúncio ativo de leilão ou venda direta de moradia
type PropertyListing struct {
	HouseID       string    `json:"house_id"`
	SellerID      int       `json:"seller_id"`
	ListingType   string    `json:"listing_type"` // "auction" ou "direct_sale"
	BuyoutPrice   int64     `json:"buyout_price,omitempty"`
	CurrentBid    int64     `json:"current_bid,omitempty"`
	HighestBidder int       `json:"highest_bidder,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
}
