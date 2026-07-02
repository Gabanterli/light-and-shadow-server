package blessing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/pkg/inventory"
)

// BlessingDefinition representa os metadados de uma bênção primordial carregada do JSON
type BlessingDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AltarDefinition representa as configurações de um altar de bênção carregadas do JSON
type AltarDefinition struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	ContinentName    string  `json:"continent_name"`
	BlessingID       string  `json:"blessing_id"`
	X                float64 `json:"x"`
	Y                float64 `json:"y"`
	Z                int     `json:"z"`
	MinLevel         int     `json:"min_level"`
	AltarCostBronze  int64   `json:"altar_cost_bronze"`
}

// ConfigWrapper engloba as definições de bênçãos e altares
type ConfigWrapper struct {
	Blessings []BlessingDefinition `json:"blessings"`
	Altars    []AltarDefinition    `json:"altars"`
}

// BlessingManager centraliza as regras de negócio das Bênçãos Primordiais e Altares com cache de Bitmask
type BlessingManager struct {
	mu           sync.RWMutex
	db           *sql.DB
	blessings    map[string]BlessingDefinition
	altars       map[string]AltarDefinition
	blessingBits map[string]uint8           // blessing_id -> bit index (0..7)
	activeCache  map[string]map[string]bool // playerID -> set of blessingID
	playerMasks  map[string]uint8           // playerID -> BlessingMask (cache de bitmask em memória)
	loadedConfig bool
}

// NewBlessingManager inicializa e carrega a configuração externa de bênçãos de forma resiliente
func NewBlessingManager(db *sql.DB) *BlessingManager {
	bm := &BlessingManager{
		db:           db,
		blessings:    make(map[string]BlessingDefinition),
		altars:       make(map[string]AltarDefinition),
		blessingBits: make(map[string]uint8),
		activeCache:  make(map[string]map[string]bool),
		playerMasks:  make(map[string]uint8),
	}
	bm.LoadConfig()
	return bm
}

// LoadConfig carrega as configurações de bênçãos e altares do arquivo JSON de forma resiliente
func (bm *BlessingManager) LoadConfig() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	paths := []string{"backend/config/", "config/", "../config/", "../../config/"}
	fileName := "blessings_config.json"
	var data []byte
	var err error
	var finalPath string

	for _, p := range paths {
		filePath := p + fileName
		if _, statErr := os.Stat(filePath); statErr == nil {
			data, err = os.ReadFile(filePath)
			if err == nil {
				finalPath = filePath
				break
			}
		}
	}

	if err != nil || len(data) == 0 {
		slog.Warn("Could not find or read blessings_config.json, initializing default fallback blessings config")
		bm.loadFallbackConfig()
		return
	}

	var wrapper ConfigWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		slog.Error("Failed to parse blessings_config.json, using fallback", "error", err)
		bm.loadFallbackConfig()
		return
	}

	bm.blessings = make(map[string]BlessingDefinition)
	bm.altars = make(map[string]AltarDefinition)
	bm.blessingBits = make(map[string]uint8)

	var bitIdx uint8 = 0
	for _, b := range wrapper.Blessings {
		bm.blessings[b.ID] = b
		if bitIdx < 8 {
			bm.blessingBits[b.ID] = bitIdx
			bitIdx++
		}
	}
	for _, a := range wrapper.Altars {
		bm.altars[a.ID] = a
	}

	bm.loadedConfig = true
	slog.Info("Successfully loaded blessings and altars from JSON config", "blessings_count", len(bm.blessings), "altars_count", len(bm.altars), "path", finalPath)
}

func (bm *BlessingManager) loadFallbackConfig() {
	fallbackBlessings := []BlessingDefinition{
		{"blessing_fire", "Primordial Blessing of Fire", "Protects against death penalties."},
		{"blessing_ice", "Primordial Blessing of Ice", "Protects against death penalties."},
		{"blessing_holy", "Primordial Blessing of Holy", "Protects against death penalties."},
		{"blessing_shadow", "Primordial Blessing of Shadow", "Protects against death penalties."},
		{"blessing_nature", "Primordial Blessing of Nature", "Protects against death penalties."},
		{"blessing_abyss", "Primordial Blessing of the Abyss", "Protects against death penalties."},
		{"blessing_light", "Primordial Blessing of Light", "Protects against death penalties."},
	}
	fallbackAltars := []AltarDefinition{
		{"altar_fire", "Hearth Altar", "Fire Continent", "blessing_fire", 2100.0, 2100.0, 0, 50, 50000},
		{"altar_ice", "Frost Altar", "Ice Continent", "blessing_ice", 2300.0, 2300.0, 0, 50, 100000},
		{"altar_holy", "Sanctum Altar", "Holy Continent", "blessing_holy", 2500.0, 2500.0, 0, 50, 150000},
		{"altar_shadow", "Eclipse Altar", "Shadow Continent", "blessing_shadow", 2700.0, 2700.0, 0, 50, 200000},
		{"altar_nature", "Wild Altar", "Nature Continent", "blessing_nature", 2900.0, 2900.0, 0, 50, 250000},
		{"altar_abyss", "Void Altar", "Abyssia", "blessing_abyss", 3500.0, 3500.0, 0, 200, 1000000},
		{"altar_light", "Aurora Altar", "Main Continent", "blessing_light", 100.0, 100.0, 0, 1, 10000},
	}

	bm.blessingBits = make(map[string]uint8)
	var bitIdx uint8 = 0
	for _, b := range fallbackBlessings {
		bm.blessings[b.ID] = b
		if bitIdx < 8 {
			bm.blessingBits[b.ID] = bitIdx
			bitIdx++
		}
	}
	for _, a := range fallbackAltars {
		bm.altars[a.ID] = a
	}
	bm.loadedConfig = true
}

// GetAltarByID retorna detalhes do altar
func (bm *BlessingManager) GetAltarByID(altarID string) (AltarDefinition, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	altar, ok := bm.altars[altarID]
	return altar, ok
}

// GetBlessingByID retorna detalhes da bênção
func (bm *BlessingManager) GetBlessingByID(blessingID string) (BlessingDefinition, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	blessing, ok := bm.blessings[blessingID]
	return blessing, ok
}

// GetAllBlessings retorna todas as bênçãos primordial configuradas
func (bm *BlessingManager) GetAllBlessings() []BlessingDefinition {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	list := make([]BlessingDefinition, 0, len(bm.blessings))
	for _, b := range bm.blessings {
		list = append(list, b)
	}
	return list
}

// GetAllAltars retorna todos os altares configurados
func (bm *BlessingManager) GetAllAltars() []AltarDefinition {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	list := make([]AltarDefinition, 0, len(bm.altars))
	for _, a := range bm.altars {
		list = append(list, a)
	}
	return list
}

// LoadPlayerBlessings busca do banco de dados e carrega no cache as bênçãos do jogador
func (bm *BlessingManager) LoadPlayerBlessings(playerID string) ([]string, error) {
	if bm.db == nil {
		bm.mu.Lock()
		defer bm.mu.Unlock()
		if bm.activeCache[playerID] == nil {
			bm.activeCache[playerID] = make(map[string]bool)
		}
		if bm.playerMasks == nil {
			bm.playerMasks = make(map[string]uint8)
		}
		var mask uint8 = 0
		active := make([]string, 0)
		for k := range bm.activeCache[playerID] {
			active = append(active, k)
			if bit, exists := bm.blessingBits[k]; exists {
				mask |= (1 << bit)
			}
		}
		bm.playerMasks[playerID] = mask
		return active, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Busca o character_id do jogador
	var charID int
	err := bm.db.QueryRowContext(ctx, "SELECT id FROM characters WHERE name = $1", playerID).Scan(&charID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to retrieve character for blessings load: %w", err)
	}

	rows, err := bm.db.QueryContext(ctx, "SELECT blessing_id FROM character_blessings WHERE character_id = $1", charID)
	if err != nil {
		return nil, fmt.Errorf("failed to load character blessings from DB: %w", err)
	}
	defer rows.Close()

	blessingsSet := make(map[string]bool)
	activeList := make([]string, 0)
	var mask uint8 = 0

	for rows.Next() {
		var bid string
		if err := rows.Scan(&bid); err == nil {
			blessingsSet[bid] = true
			activeList = append(activeList, bid)
			if bit, exists := bm.blessingBits[bid]; exists {
				mask |= (1 << bit)
			}
		}
	}

	bm.mu.Lock()
	if bm.playerMasks == nil {
		bm.playerMasks = make(map[string]uint8)
	}
	bm.activeCache[playerID] = blessingsSet
	bm.playerMasks[playerID] = mask
	bm.mu.Unlock()

	return activeList, nil
}

// HasBlessing verifica se o jogador possui uma determinada bênção usando bitwise O(1) checks
func (bm *BlessingManager) HasBlessing(playerID string, blessingID string) bool {
	bm.mu.RLock()
	bit, bitExists := bm.blessingBits[blessingID]
	mask, exists := bm.playerMasks[playerID]
	bm.mu.RUnlock()

	if !bitExists {
		return false
	}

	if !exists {
		// Carrega sob demanda de forma resiliente
		_, err := bm.LoadPlayerBlessings(playerID)
		if err != nil {
			return false
		}
		bm.mu.RLock()
		mask = bm.playerMasks[playerID]
		bm.mu.RUnlock()
	}

	return (mask & (1 << bit)) != 0
}

// IsFullyBlessed verifica se o jogador obteve as 7 bênçãos primordiais (ou todas as configuradas) via bitwise operations
func (bm *BlessingManager) IsFullyBlessed(playerID string) bool {
	bm.mu.RLock()
	totalConfigured := len(bm.blessings)
	var fullMask uint8 = 0
	for _, bit := range bm.blessingBits {
		fullMask |= (1 << bit)
	}
	bm.mu.RUnlock()

	if totalConfigured == 0 {
		return false
	}

	bm.mu.RLock()
	mask, exists := bm.playerMasks[playerID]
	bm.mu.RUnlock()

	if !exists {
		_, err := bm.LoadPlayerBlessings(playerID)
		if err != nil {
			return false
		}
		bm.mu.RLock()
		mask = bm.playerMasks[playerID]
		bm.mu.RUnlock()
	}

	return mask == fullMask
}

// GetActiveBlessingsCount returns the count of active blessings for a player
func (bm *BlessingManager) GetActiveBlessingsCount(playerID string) int {
	bm.mu.RLock()
	mask, exists := bm.playerMasks[playerID]
	bm.mu.RUnlock()

	if !exists {
		_, err := bm.LoadPlayerBlessings(playerID)
		if err != nil {
			return 0
		}
		bm.mu.RLock()
		mask = bm.playerMasks[playerID]
		bm.mu.RUnlock()
	}

	// Count set bits in the mask
	count := 0
	for i := 0; i < 8; i++ {
		if (mask & (1 << i)) != 0 {
			count++
		}
	}
	return count
}

// AcquireBlessing concede uma bênção a um jogador ao interagir com o altar correspondente e cobrar o custo em moedas
func (bm *BlessingManager) AcquireBlessing(playerID string, altarID string, playerX, playerY float64, playerLevel int, playerInv *inventory.PlayerInventory) error {
	bm.mu.RLock()
	altar, ok := bm.altars[altarID]
	bit, bitExists := bm.blessingBits[altar.BlessingID]
	bm.mu.RUnlock()

	if !ok || !bitExists {
		return fmt.Errorf("altar %s not found in configurations", altarID)
	}

	// 1. Validação de Nível Mínimo
	if playerLevel < altar.MinLevel {
		return fmt.Errorf("you must be at least level %d to interact with the %s Altar", altar.MinLevel, altar.Name)
	}

	// 2. Validação de Proximidade Espacial (dentro de 10.0 metros)
	distance := math.Sqrt(math.Pow(playerX-altar.X, 2) + math.Pow(playerY-altar.Y, 2))
	if distance > 10.0 {
		return fmt.Errorf("you are too far from the %s Altar to receive its blessing", altar.Name)
	}

	// Verifica se já possui
	if bm.HasBlessing(playerID, altar.BlessingID) {
		return fmt.Errorf("you have already received the %s from this Altar", altar.Name)
	}

	// 3. Validação de Custo e Dedução de Moeda (PATCH 4 & PATCH 1 Server-side Validation)
	cost := altar.AltarCostBronze
	if cost < 0 {
		return fmt.Errorf("invalid blessing cost configuration")
	}

	if playerInv != nil {
		if playerInv.GetGold() < cost {
			return fmt.Errorf("insufficient funds: you need %d bronze, but you have %d bronze", cost, playerInv.GetGold())
		}
	}

	// 4. Persistência de forma atômica no Banco de Dados
	if bm.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tx, err := bm.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback()

		var charID int
		var currentGold int64
		err = tx.QueryRowContext(ctx, "SELECT id, gold FROM characters WHERE name = $1 FOR UPDATE", playerID).Scan(&charID, &currentGold)
		if err != nil {
			return fmt.Errorf("failed to lock character: %w", err)
		}

		if currentGold < cost {
			return fmt.Errorf("insufficient database funds: you need %d bronze, but you have %d bronze", cost, currentGold)
		}

		newGold := currentGold - cost

		// Atualiza gold do personagem
		_, err = tx.ExecContext(ctx, "UPDATE characters SET gold = $1 WHERE id = $2", newGold, charID)
		if err != nil {
			return fmt.Errorf("failed to deduct blessing cost: %w", err)
		}

		// Insere bênção
		_, err = tx.ExecContext(ctx, `
			INSERT INTO character_blessings (character_id, blessing_id)
			VALUES ($1, $2)
			ON CONFLICT (character_id, blessing_id) DO NOTHING
		`, charID, altar.BlessingID)
		if err != nil {
			return fmt.Errorf("failed to persist acquired blessing: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit acquired blessing: %w", err)
		}
	}

	// 5. Deduz moeda em memória se playerInv fornecido
	if playerInv != nil {
		playerInv.RemoveGold(cost)
	}

	// 6. Atualização segura do cache de bitmask em memória
	bm.mu.Lock()
	if bm.activeCache[playerID] == nil {
		bm.activeCache[playerID] = make(map[string]bool)
	}
	bm.activeCache[playerID][altar.BlessingID] = true

	if bm.playerMasks == nil {
		bm.playerMasks = make(map[string]uint8)
	}
	bm.playerMasks[playerID] |= (1 << bit)
	bm.mu.Unlock()

	slog.Info("Player successfully acquired primordial blessing", "player", playerID, "blessing", altar.BlessingID, "altar", altarID, "cost_bronze", cost)
	return nil
}

// ConsumeAllBlessings remove todas as bênçãos adquiridas pelo jogador. Usado sob morte se "Fully Blessed".
func (bm *BlessingManager) ConsumeAllBlessings(playerID string) bool {
	if !bm.IsFullyBlessed(playerID) {
		return false
	}

	slog.Info("Player died fully blessed! Consuming blessings to protect against all death penalties.", "player", playerID)

	if bm.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var charID int
		err := bm.db.QueryRowContext(ctx, "SELECT id FROM characters WHERE name = $1", playerID).Scan(&charID)
		if err == nil {
			_, err = bm.db.ExecContext(ctx, "DELETE FROM character_blessings WHERE character_id = $1", charID)
			if err != nil {
				slog.Error("Failed to remove player blessings from database on death", "player", playerID, "error", err)
			}
		}
	}

	bm.mu.Lock()
	bm.activeCache[playerID] = make(map[string]bool)
	if bm.playerMasks != nil {
		bm.playerMasks[playerID] = 0
	}
	bm.mu.Unlock()

	return true
}

// ClearCache limpa cache de um jogador desconectado
func (bm *BlessingManager) ClearCache(playerID string) {
	bm.mu.Lock()
	delete(bm.activeCache, playerID)
	if bm.playerMasks != nil {
		delete(bm.playerMasks, playerID)
	}
	bm.mu.Unlock()
}
