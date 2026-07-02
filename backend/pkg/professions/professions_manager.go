package professions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/inventory"
)

type RecipeMaterial struct {
	ItemID   string `json:"ItemID"`
	Quantity int    `json:"Quantity"`
}

type Recipe struct {
	RecipeID       string           `json:"RecipeID"`
	Profession     string           `json:"Profession"`
	MinLevel       int              `json:"MinLevel"`
	CharacterLevel int              `json:"CharacterLevel"`
	Duration       float64          `json:"Duration"`
	OutputItemID   string           `json:"OutputItemID"`
	OutputQuantity int              `json:"OutputQuantity"`
	SuccessRate    float64          `json:"SuccessRate"`
	XP             int              `json:"XP"`
	Materials      []RecipeMaterial `json:"Materials"`
}

type ResourceNodeState struct {
	NodeID          string    `json:"NodeID"`
	NodeType        string    `json:"NodeType"`
	ItemID          string    `json:"ItemID"`
	MinLevel        int       `json:"MinLevel"`
	Duration        float64   `json:"Duration"`
	XP              int       `json:"XP"`
	RespawnTime     int       `json:"RespawnTime"`
	X               float64   `json:"X"`
	Y               float64   `json:"Y"`
	Z               float64   `json:"Z"`
	Depleted        bool      `json:"-"`
	RespawnAt       time.Time `json:"-"`
	ReservedBy      string    `json:"-"`
	ReservationTime time.Time `json:"-"`
}

type ProfessionState struct {
	Profession string
	Level      int
	Experience int
}

func RequiredXP(level int) int {
	return 100 * level * level
}

func (p *ProfessionState) AddXP(amount int) bool {
	p.Experience += amount
	leveledUp := false
	for {
		req := RequiredXP(p.Level)
		if p.Experience >= req {
			p.Experience -= req
			p.Level++
			leveledUp = true
		} else {
			break
		}
	}
	return leveledUp
}

type ProfessionsManager struct {
	dbConn           *sql.DB
	combatManager    *combat.CombatManager
	recipes          map[string]Recipe
	nodes            map[string]*ResourceNodeState
	playerLastGather map[string]time.Time
	mu               sync.RWMutex
	onNodeStateChange func(nodeID string, depleted bool)
	onProfessionXP    func(playerID string, prof string, level int, xp int)
}

func NewProfessionsManager(dbConn *sql.DB, combatManager *combat.CombatManager) *ProfessionsManager {
	pm := &ProfessionsManager{
		dbConn:           dbConn,
		combatManager:    combatManager,
		recipes:          make(map[string]Recipe),
		nodes:            make(map[string]*ResourceNodeState),
		playerLastGather: make(map[string]time.Time),
	}

	pm.loadConfig()
	pm.initAuditLogs()
	return pm
}

func (pm *ProfessionsManager) RegisterCallbacks(nodeChange func(string, bool), profXP func(string, string, int, int)) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.onNodeStateChange = nodeChange
	pm.onProfessionXP = profXP
}

func (pm *ProfessionsManager) loadConfig() {
	recipesPaths := []string{"backend/config/recipes.json", "config/recipes.json", "../config/recipes.json"}
	var recipesData []byte
	var err error
	for _, p := range recipesPaths {
		recipesData, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err == nil {
		var list []Recipe
		if err := json.Unmarshal(recipesData, &list); err == nil {
			for _, r := range list {
				pm.recipes[r.RecipeID] = r
			}
			slog.Info("Loaded crafting recipes", "count", len(pm.recipes))
		} else {
			slog.Error("Failed to unmarshal recipes.json", "error", err)
		}
	} else {
		slog.Warn("Could not find recipes.json, running without pre-configured recipes")
	}

	nodesPaths := []string{"backend/config/resource_nodes.json", "config/resource_nodes.json", "../config/resource_nodes.json"}
	var nodesData []byte
	for _, p := range nodesPaths {
		nodesData, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err == nil {
		var list []ResourceNodeState
		if err := json.Unmarshal(nodesData, &list); err == nil {
			for _, n := range list {
				nodeCopy := n
				pm.nodes[n.NodeID] = &nodeCopy
			}
			slog.Info("Loaded resource spawn nodes", "count", len(pm.nodes))
		} else {
			slog.Error("Failed to unmarshal resource_nodes.json", "error", err)
		}
	} else {
		slog.Warn("Could not find resource_nodes.json, running without pre-configured nodes")
	}
}

func (pm *ProfessionsManager) initAuditLogs() {
	if pm.dbConn == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := pm.dbConn.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS craft_audit_logs (
			id SERIAL PRIMARY KEY,
			player_id VARCHAR(64) NOT NULL,
			recipe_id VARCHAR(64) NOT NULL,
			success_rate DOUBLE PRECISION NOT NULL,
			roll DOUBLE PRECISION NOT NULL,
			outcome VARCHAR(16) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		slog.Error("Failed to create craft_audit_logs table", "error", err)
	}
}

func (pm *ProfessionsManager) getCharacterID(playerID string) (int, error) {
	var charID int
	err := pm.dbConn.QueryRow("SELECT id FROM characters WHERE name = $1", playerID).Scan(&charID)
	if err != nil {
		return 0, err
	}
	return charID, nil
}

func (pm *ProfessionsManager) LoadProfessions(playerID string) (map[string]*ProfessionState, error) {
	professions := make(map[string]*ProfessionState)
	allProfs := []string{"mining", "woodcutting", "herbalism", "fishing", "blacksmithing", "alchemy", "enchanting"}
	for _, p := range allProfs {
		professions[p] = &ProfessionState{
			Profession: p,
			Level:      1,
			Experience: 0,
		}
	}

	if pm.dbConn == nil {
		return professions, nil // Fallback safe
	}

	charID, err := pm.getCharacterID(playerID)
	if err != nil {
		return professions, nil // Novo personagem ou fallback
	}

	rows, err := pm.dbConn.Query("SELECT profession, level, experience FROM character_professions WHERE character_id = $1", charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var prof string
		var lvl, xp int
		if err := rows.Scan(&prof, &lvl, &xp); err == nil {
			if state, ok := professions[prof]; ok {
				state.Level = lvl
				state.Experience = xp
			}
		}
	}

	return professions, nil
}

func (pm *ProfessionsManager) SaveProfession(playerID string, prof string, level int, xp int) error {
	if pm.dbConn == nil {
		return nil
	}
	charID, err := pm.getCharacterID(playerID)
	if err != nil {
		return err
	}

	_, err = pm.dbConn.Exec(`
		INSERT INTO character_professions (character_id, profession, level, experience)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (character_id, profession)
		DO UPDATE SET level = EXCLUDED.level, experience = EXCLUDED.experience
	`, charID, prof, level, xp)
	return err
}

func (pm *ProfessionsManager) writeAuditLog(playerID, recipeID string, successRate, roll float64, outcome string) {
	slog.Info("RNG Craft Audit Log", "player", playerID, "recipe", recipeID, "success_rate", successRate, "roll", roll, "outcome", outcome)
	if pm.dbConn == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := pm.dbConn.ExecContext(ctx, `
		INSERT INTO craft_audit_logs (player_id, recipe_id, success_rate, roll, outcome)
		VALUES ($1, $2, $3, $4, $5)
	`, playerID, recipeID, successRate, roll, outcome)
	if err != nil {
		slog.Error("Failed to write craft audit log to database", "error", err)
	}
}

// GetNodesState retorna uma cópia de todos os nós de recurso e seus estados para sincronização com o cliente
func (pm *ProfessionsManager) GetNodesState() []ResourceNodeState {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	res := make([]ResourceNodeState, 0, len(pm.nodes))
	for _, node := range pm.nodes {
		res = append(res, *node)
	}
	return res
}

// StartGathering inicia a coleta em um nó de recurso autoritativo
func (pm *ProfessionsManager) StartGathering(playerID string, nodeID string, playerInv *inventory.PlayerInventory) (float64, error) {
	pm.mu.Lock()

	// 1. Valida se o nó existe
	node, exists := pm.nodes[nodeID]
	if !exists {
		pm.mu.Unlock()
		return 0, fmt.Errorf("Resource node not found")
	}

	// 2. Valida se está esgotado
	if node.Depleted {
		pm.mu.Unlock()
		return 0, fmt.Errorf("Resource node is currently depleted")
	}

	// 3. PATCH 1: Node reservation locking
	now := time.Now()
	isReserved := node.ReservedBy != "" && now.Sub(node.ReservationTime) < time.Duration(node.Duration)*time.Second+2*time.Second
	if isReserved && node.ReservedBy != playerID {
		pm.mu.Unlock()
		return 0, fmt.Errorf("Resource node already reserved by another player")
	}

	// 4. PATCH 3: Anti-bot gathering cooldown (Strict 1.5 seconds cooldown after gather finished or started)
	lastGather, hasLast := pm.playerLastGather[playerID]
	if hasLast && now.Sub(lastGather) < 1500*time.Millisecond {
		pm.mu.Unlock()
		return 0, fmt.Errorf("Anti-bot: Cooldown in progress between gathering actions")
	}

	// 5. Carrega profissões do jogador
	pm.mu.Unlock() // Destrava temporariamente para fazer a operação do DB sem travar os nós
	profs, err := pm.LoadProfessions(playerID)
	if err != nil {
		return 0, fmt.Errorf("Failed to load player professions: %w", err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Re-valida tudo pós unlock do DB
	if node.Depleted {
		return 0, fmt.Errorf("Resource node became depleted")
	}
	if node.ReservedBy != "" && node.ReservedBy != playerID && now.Sub(node.ReservationTime) < time.Duration(node.Duration)*time.Second+2*time.Second {
		return 0, fmt.Errorf("Resource node was reserved during database query")
	}

	pProf, ok := profs[node.NodeType]
	if !ok || pProf.Level < node.MinLevel {
		return 0, fmt.Errorf("Insufficient profession level. Required: %s level %d", node.NodeType, node.MinLevel)
	}

	// Pre-valida espaço na mochila
	if !hasSpaceFor(playerInv, node.ItemID, 1) {
		return 0, fmt.Errorf("Your inventory is full")
	}

	// Reserva o nó e atualiza cooldown de coleta
	node.ReservedBy = playerID
	node.ReservationTime = now
	pm.playerLastGather[playerID] = now

	slog.Info("Player started gathering resource node", "player", playerID, "node", nodeID, "type", node.NodeType)

	// Dispara o callback de início de progresso (opcional, para atualizar UI se houver)
	if pm.onNodeStateChange != nil {
		// Envia que o nó foi reservado temporariamente
		pm.onNodeStateChange(nodeID, false)
	}

	return node.Duration, nil
}

// CompleteGathering finaliza a coleta e concede os prêmios e XP de forma transacional segura
func (pm *ProfessionsManager) CompleteGathering(playerID string, nodeID string, playerInv *inventory.PlayerInventory) (string, int, error) {
	pm.mu.Lock()
	node, exists := pm.nodes[nodeID]
	if !exists {
		pm.mu.Unlock()
		return "", 0, fmt.Errorf("Resource node not found")
	}

	if node.Depleted {
		pm.mu.Unlock()
		return "", 0, fmt.Errorf("Resource node is already depleted")
	}

	if node.ReservedBy != playerID {
		pm.mu.Unlock()
		return "", 0, fmt.Errorf("Reservation mismatch or expired")
	}

	// Libera reserva e esgota o nó
	node.Depleted = true
	node.ReservedBy = ""
	node.RespawnAt = time.Now().Add(time.Duration(node.RespawnTime) * time.Second)
	nodeIDToRespawn := nodeID
	respawnDur := time.Duration(node.RespawnTime) * time.Second

	pm.mu.Unlock()

	// Concede o item
	if !playerInv.AddItem(node.ItemID, 1) {
		// Inventário encheu bem no momento de receber
		pm.mu.Lock()
		node.Depleted = false // Restaura o nó
		pm.mu.Unlock()
		return "", 0, fmt.Errorf("Failed to add gathered item (inventory full)")
	}

	// Concede XP da profissão
	profs, err := pm.LoadProfessions(playerID)
	if err == nil {
		if pProf, ok := profs[node.NodeType]; ok {
			leveledUp := pProf.AddXP(node.XP)
			_ = pm.SaveProfession(playerID, node.NodeType, pProf.Level, pProf.Experience)
			if pm.onProfessionXP != nil {
				pm.onProfessionXP(playerID, node.NodeType, pProf.Level, pProf.Experience)
			}
			if leveledUp {
				slog.Info("Player leveled up gathering profession!", "player", playerID, "profession", node.NodeType, "new_level", pProf.Level)
			}
		}
	}

	// Callback de mudança de estado do nó (esgotado)
	if pm.onNodeStateChange != nil {
		pm.onNodeStateChange(nodeID, true)
	}

	// Agenda respawn do nó
	go func() {
		time.Sleep(respawnDur)
		pm.mu.Lock()
		if n, ok := pm.nodes[nodeIDToRespawn]; ok {
			n.Depleted = false
			n.ReservedBy = ""
			slog.Info("Resource node respawned", "node", nodeIDToRespawn)
		}
		pm.mu.Unlock()

		if pm.onNodeStateChange != nil {
			pm.onNodeStateChange(nodeIDToRespawn, false)
		}
	}()

	return node.ItemID, node.XP, nil
}

// CancelGathering cancela a reserva se o jogador se mover ou cancelar a ação
func (pm *ProfessionsManager) CancelGathering(playerID string, nodeID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if node, ok := pm.nodes[nodeID]; ok && node.ReservedBy == playerID {
		node.ReservedBy = ""
		slog.Info("Player cancelled gathering or moved away", "player", playerID, "node", nodeID)
		if pm.onNodeStateChange != nil {
			pm.onNodeStateChange(nodeID, false)
		}
	}
}

// PerformCraft realiza uma síntese autoritativa atômica com validações completas (PATCH 2, PATCH 4, PATCH 5)
func (pm *ProfessionsManager) PerformCraft(playerID string, recipeID string, playerInv *inventory.PlayerInventory) (string, int, bool, error) {
	pm.mu.Lock()
	recipe, exists := pm.recipes[recipeID]
	if !exists {
		pm.mu.Unlock()
		return "", 0, false, fmt.Errorf("Recipe not found")
	}
	pm.mu.Unlock()

	// Carrega profissões do jogador
	profs, err := pm.LoadProfessions(playerID)
	if err != nil {
		return "", 0, false, fmt.Errorf("Failed to load player professions: %w", err)
	}

	// Get player character stats to check level requirement
	pStats, hasStats := pm.combatManager.GetEntityStats(playerID)
	if !hasStats {
		return "", 0, false, fmt.Errorf("Character stats not found")
	}

	// 1. PATCH 4: Server-side recipe validation
	// Valida nível do personagem
	if pStats.Level < recipe.CharacterLevel {
		return "", 0, false, fmt.Errorf("Insufficient character level. Required: %d", recipe.CharacterLevel)
	}

	// Valida nível de profissão
	pProf, ok := profs[recipe.Profession]
	if !ok || pProf.Level < recipe.MinLevel {
		return "", 0, false, fmt.Errorf("Insufficient profession level. Required: %s level %d", recipe.Profession, recipe.MinLevel)
	}

	// Valida se possui todos os materiais requeridos
	for _, m := range recipe.Materials {
		if !playerInv.HasItem(m.ItemID, m.Quantity) {
			return "", 0, false, fmt.Errorf("Missing materials for recipe. ItemID: %s, Qty: %d", m.ItemID, m.Quantity)
		}
	}

	// Valida espaço no inventário antes de consumir materiais
	if !hasSpaceFor(playerInv, recipe.OutputItemID, recipe.OutputQuantity) {
		return "", 0, false, fmt.Errorf("Your inventory has no space for the crafted item")
	}

	// 2. PATCH 2: Atomic craft transactions (Consome materiais e decide sucesso)
	// Como o inventário possui travas próprias internas de forma segura, garantimos a atomicidade aqui:
	for _, m := range recipe.Materials {
		removed := playerInv.RemoveItemByID(m.ItemID, m.Quantity)
		if !removed {
			return "", 0, false, fmt.Errorf("Failed to safely consume materials (concurrency conflict)")
		}
	}

	// Decide sucesso ou falha através de RNG
	roll := rand.Float64()
	success := roll <= recipe.SuccessRate

	// 3. PATCH 5: RNG audit trail for rare crafts (Grava no banco para auditoria)
	outcome := "success"
	if !success {
		outcome = "failure"
	}
	pm.writeAuditLog(playerID, recipeID, recipe.SuccessRate, roll, outcome)

	if !success {
		// Falha no craft: materiais perdidos, mas ganha um pouco de XP de consolação!
		pProf.AddXP(recipe.XP / 4) // 25% de XP de consolação
		_ = pm.SaveProfession(playerID, recipe.Profession, pProf.Level, pProf.Experience)
		if pm.onProfessionXP != nil {
			pm.onProfessionXP(playerID, recipe.Profession, pProf.Level, pProf.Experience)
		}
		return recipe.OutputItemID, recipe.XP / 4, false, nil
	}

	// Sucesso no craft: adiciona item gerado
	added := playerInv.AddItem(recipe.OutputItemID, recipe.OutputQuantity)
	if !added {
		// Caso de segurança extremo (por exemplo se espaço foi preenchido de última hora)
		// Devolve materiais
		for _, m := range recipe.Materials {
			playerInv.AddItem(m.ItemID, m.Quantity)
		}
		return "", 0, false, fmt.Errorf("Failed to deliver crafted item, materials returned")
	}

	// Adiciona XP completo de profissão
	leveledUp := pProf.AddXP(recipe.XP)
	_ = pm.SaveProfession(playerID, recipe.Profession, pProf.Level, pProf.Experience)
	if pm.onProfessionXP != nil {
		pm.onProfessionXP(playerID, recipe.Profession, pProf.Level, pProf.Experience)
	}

	if leveledUp {
		slog.Info("Player leveled up crafting profession!", "player", playerID, "profession", recipe.Profession, "new_level", pProf.Level)
	}

	return recipe.OutputItemID, recipe.XP, true, nil
}

// CheckSpaceForHelper expõe a validação local
func hasSpaceFor(pi *inventory.PlayerInventory, itemID string, qty int) bool {
	// Chamamos a função interna robusta
	return HasSpaceForHelper(pi, itemID, qty)
}

func HasSpaceForHelper(pi *inventory.PlayerInventory, itemID string, qty int) bool {
	items := pi.GetItems()

	def, exists := inventory.GetItemDef(itemID)
	if !exists {
		return false
	}

	remaining := qty
	if def.Stackable {
		for _, item := range items {
			if item.ItemID == itemID && item.Quantity < def.MaxStack {
				spaceLeft := def.MaxStack - item.Quantity
				if remaining <= spaceLeft {
					return true
				}
				remaining -= spaceLeft
			}
		}
	}

	emptySlots := 0
	for slot := inventory.SlotMinBackpack; slot <= inventory.SlotMaxBackpack; slot++ {
		if _, occupied := items[slot]; !occupied {
			emptySlots++
		}
	}

	slotsNeeded := 0
	if def.Stackable {
		slotsNeeded = (remaining + def.MaxStack - 1) / def.MaxStack
	} else {
		slotsNeeded = remaining
	}

	return emptySlots >= slotsNeeded
}
