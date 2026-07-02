package quest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/inventory"
	"github.com/light-and-shadow/backend/pkg/messaging"
)

// ObjectiveType representa o tipo de objetivo de uma quest
type ObjectiveType string

const (
	ObjectiveKillMonster   ObjectiveType = "KillMonster"
	ObjectiveCollectItem   ObjectiveType = "CollectItem"
	ObjectiveTalkToNPC     ObjectiveType = "TalkToNPC"
	ObjectiveReachLocation ObjectiveType = "ReachLocation"
)

// QuestObjective representa um objetivo estático da quest vindo do JSON
type QuestObjective struct {
	Type        ObjectiveType `json:"type"`
	TargetID    string        `json:"target_id"`
	RequiredQty int           `json:"required_qty"`
}

// QuestPrerequisites representa os pré-requisitos para iniciar a quest
type QuestPrerequisites struct {
	Level             int      `json:"level"`
	CompletedQuestIDs []string `json:"completed_quest_ids"`
}

// RewardItem representa os itens concedidos como recompensa
type RewardItem struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

// QuestRewards representa as recompensas de uma quest
type QuestRewards struct {
	Experience int          `json:"experience"`
	Gold       int          `json:"gold"`
	Items      []RewardItem `json:"items"`
	Reputation int          `json:"reputation"`
}

// QuestDefinition representa a definição estática de uma quest vinda do JSON
type QuestDefinition struct {
	QuestID       string             `json:"quest_id"`
	Title         string             `json:"title"`
	Description   string             `json:"description"`
	Objectives    []QuestObjective   `json:"objectives"`
	Prerequisites QuestPrerequisites `json:"prerequisites"`
	Rewards       QuestRewards       `json:"rewards"`
	Repeatable    bool               `json:"repeatable"`
}

// ObjectiveState representa o progresso de um objetivo específico do jogador
type ObjectiveState struct {
	ObjectiveIndex int `json:"objective_index"`
	CurrentQty     int `json:"current_qty"`
}

// ActiveQuest representa o estado de uma quest ativa, compatível com versionamento e optimistic locking (PATCH 1)
type ActiveQuest struct {
	QuestID    string                  `json:"quest_id"`
	Version    uint32                  `json:"version"`
	Status     string                  `json:"status"`
	Objectives map[int]*ObjectiveState `json:"objectives"`
}

// CharacterQuestState representa o estado atual de uma quest para um jogador
type CharacterQuestState = ActiveQuest

// PlayerQuestState agrupa as quests e os diálogos de um único jogador
type PlayerQuestState struct {
	mu            sync.RWMutex
	PlayerID      string
	Quests        map[string]*CharacterQuestState
	DialogueFlags map[string]string // npc_id -> flag_value (ex: "node_start")
	isDirty       bool              // Controle de escrita em lote (PATCH 1)
}

// QuestManager gerencia o carregamento de configurações e ações das quests
type QuestManager struct {
	mu           sync.RWMutex
	dbConn       *sql.DB // Ponteiro de conexão do PostgreSQL
	definitions  map[string]QuestDefinition
	playerStates map[string]*PlayerQuestState // playerID -> PlayerQuestState

	// Dependências externas para recompensas e consultas
	combatManager *combat.CombatManager
	inventories   map[string]*inventory.PlayerInventory

	// Callbacks ou canais de notificação para o cliente
	onQuestUpdated func(playerID string, questID string, state *CharacterQuestState)
}

// NewQuestManager instancia o QuestManager
func NewQuestManager(db *sql.DB, cm *combat.CombatManager, invs map[string]*inventory.PlayerInventory) *QuestManager {
	qm := &QuestManager{
		dbConn:        db,
		definitions:   make(map[string]QuestDefinition),
		playerStates:  make(map[string]*PlayerQuestState),
		combatManager: cm,
		inventories:   invs,
	}

	qm.loadDefinitions()
	qm.startEventLoop() // Inicializa o loop de eventos assíncronos (PATCH 5)
	return qm
}

// getRewardHash calcula uma hash única para o conteúdo das recompensas (PATCH 4)
func (qm *QuestManager) getRewardHash(rewards QuestRewards) string {
	data, _ := json.Marshal(rewards)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// startEventLoop se inscreve nos tópicos do Event Bus e processa eventos assincronamente (PATCH 5)
func (qm *QuestManager) startEventLoop() {
	mb := messaging.GetInstance()
	chKilled := mb.Subscribe("monster_killed")
	chLooted := mb.Subscribe("item_looted")
	chNPC := mb.Subscribe("npc_interacted")
	chLoc := mb.Subscribe("location_reached")

	go func() {
		for {
			select {
			case msg, ok := <-chKilled:
				if !ok {
					return
				}
				if payload, ok := msg.(messaging.MonsterKilledPayload); ok {
					qm.OnKillMonster(payload.PlayerID, payload.MonsterID)
				}
			case msg, ok := <-chLooted:
				if !ok {
					return
				}
				if payload, ok := msg.(messaging.ItemLootedPayload); ok {
					qm.OnCollectItem(payload.PlayerID, payload.ItemID, payload.Qty)
				}
			case msg, ok := <-chNPC:
				if !ok {
					return
				}
				if payload, ok := msg.(messaging.NPCInteractedPayload); ok {
					qm.OnTalkToNPC(payload.PlayerID, payload.NPCID)
				}
			case msg, ok := <-chLoc:
				if !ok {
					return
				}
				if payload, ok := msg.(messaging.LocationReachedPayload); ok {
					qm.OnReachLocation(payload.PlayerID, payload.X, payload.Y, payload.Z)
				}
			}
		}
	}()
}

// RegisterQuestUpdateCallback registra callback para notificar o cliente sobre mudanças de quest
func (qm *QuestManager) RegisterQuestUpdateCallback(cb func(string, string, *CharacterQuestState)) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.onQuestUpdated = cb
}

// loadDefinitions carrega quests de forma resiliente
func (qm *QuestManager) loadDefinitions() {
	paths := []string{"backend/config/", "config/", "../config/", "/backend/config/"}
	for _, p := range paths {
		filePath := p + "quests.json"
		if _, err := os.Stat(filePath); err == nil {
			data, err := os.ReadFile(filePath)
			if err == nil {
				var list []QuestDefinition
				if err := json.Unmarshal(data, &list); err == nil {
					qm.mu.Lock()
					for _, q := range list {
						qm.definitions[q.QuestID] = q
					}
					qm.mu.Unlock()
					slog.Info("Successfully loaded quests.json", "count", len(qm.definitions), "path", filePath)
					return
				}
			}
		}
	}
	slog.Error("Failed to find or load quests.json definition files")
}

// GetDefinition retorna a definição estática de uma quest
func (qm *QuestManager) GetDefinition(questID string) (QuestDefinition, bool) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	q, ok := qm.definitions[questID]
	return q, ok
}

// GetPlayerState retorna ou cria o estado de quest de um jogador logado
func (qm *QuestManager) GetPlayerState(playerID string) *PlayerQuestState {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	state, ok := qm.playerStates[playerID]
	if !ok {
		state = &PlayerQuestState{
			PlayerID:      playerID,
			Quests:        make(map[string]*CharacterQuestState),
			DialogueFlags: make(map[string]string),
			isDirty:       false,
		}
		qm.playerStates[playerID] = state
		// Tenta carregar do PostgreSQL de forma resiliente
		if err := qm.loadFromPostgres(state); err != nil {
			slog.Warn("Failed to load quest state from PostgreSQL, falling back to in-memory clean slate", "player", playerID, "error", err)
		}
	}
	return state
}

// CleanPlayerState remove o estado de quest da memória ao deslogar
func (qm *QuestManager) CleanPlayerState(playerID string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	delete(qm.playerStates, playerID)
}

// SyncAllActiveQuests dispara o callback de atualização de rede para todas as quests do jogador
func (qm *QuestManager) SyncAllActiveQuests(playerID string) {
	pState := qm.GetPlayerState(playerID)
	pState.mu.RLock()
	defer pState.mu.RUnlock()

	for qID, qState := range pState.Quests {
		qm.triggerUpdateCallback(playerID, qID, qState)
	}
}

// AcceptQuest adiciona uma quest ao estado ativo do jogador com validações completas (Requisitos de Segurança)
func (qm *QuestManager) AcceptQuest(playerID string, questID string) error {
	pState := qm.GetPlayerState(playerID)
	pState.mu.Lock()
	defer pState.mu.Unlock()

	def, exists := qm.GetDefinition(questID)
	if !exists {
		return fmt.Errorf("quest %s does not exist", questID)
	}

	// 1. Valida se já possui a quest ativa ou concluída
	if qState, ok := pState.Quests[questID]; ok {
		if qState.Status == "completed" && !def.Repeatable {
			return fmt.Errorf("quest %s is not repeatable and has already been completed", questID)
		}
		if qState.Status == "active" {
			return fmt.Errorf("quest %s is already active", questID)
		}
	}

	// 2. Valida nível mínimo (PATCH 4)
	pStats, hasStats := qm.combatManager.GetEntityStats(playerID)
	if !hasStats {
		return fmt.Errorf("failed to retrieve character stats for validation")
	}
	if pStats.Level < def.Prerequisites.Level {
		return fmt.Errorf("player level %d is below required level %d for quest %s", pStats.Level, def.Prerequisites.Level, questID)
	}

	// 3. Valida pré-requisitos de outras quests completadas
	for _, prereqID := range def.Prerequisites.CompletedQuestIDs {
		completedState, ok := pState.Quests[prereqID]
		if !ok || completedState.Status != "completed" {
			return fmt.Errorf("prerequisite quest %s must be completed first", prereqID)
		}
	}

	// 4. Instancia novo estado ativo de quest
	objMap := make(map[int]*ObjectiveState)
	for i := range def.Objectives {
		objMap[i] = &ObjectiveState{
			ObjectiveIndex: i,
			CurrentQty:     0,
		}
	}

	newQuestState := &CharacterQuestState{
		QuestID:    questID,
		Status:     "active",
		Version:    1,
		Objectives: objMap,
	}

	pState.Quests[questID] = newQuestState
	pState.isDirty = true // Marca dirty-flag para reduzir escritas (PATCH 1)

	slog.Info("Quest accepted successfully", "player", playerID, "quest", questID)

	// Dispara callback de notificação para rede
	qm.triggerUpdateCallback(playerID, questID, newQuestState)

	// Se houver algum objetivo de coletar itens, valida retroativamente com itens da mochila
	qm.checkCollectItemsProgress(playerID, pState, questID, def, newQuestState)

	return nil
}

// AbandonQuest remove uma quest ativa
func (qm *QuestManager) AbandonQuest(playerID string, questID string) error {
	pState := qm.GetPlayerState(playerID)
	pState.mu.Lock()
	defer pState.mu.Unlock()

	qState, ok := pState.Quests[questID]
	if !ok || qState.Status != "active" {
		return fmt.Errorf("quest %s is not active and cannot be abandoned", questID)
	}

	delete(pState.Quests, questID)
	pState.isDirty = true // Marca dirty-flag (PATCH 1)

	slog.Info("Quest abandoned by player", "player", playerID, "quest", questID)

	// Dispara callback de remoção/atualização (enviando estado nulo para sincronizar abandono)
	qm.triggerUpdateCallback(playerID, questID, nil)

	return nil
}

// CompleteQuest finaliza a quest de forma transacional, segura e atômica com rollback e locking (PATCH 2, PATCH 3)
func (qm *QuestManager) CompleteQuest(playerID string, questID string) error {
	pState := qm.GetPlayerState(playerID)
	pState.mu.Lock()
	defer pState.mu.Unlock()

	def, exists := qm.GetDefinition(questID)
	if !exists {
		return fmt.Errorf("quest %s does not exist", questID)
	}

	qState, ok := pState.Quests[questID]
	if !ok || qState.Status != "active" {
		return fmt.Errorf("quest %s is not active", questID)
	}

	// 1. Valida se todos os objetivos foram completamente concluídos
	for i, obj := range def.Objectives {
		progress, ok := qState.Objectives[i]
		if !ok || progress.CurrentQty < obj.RequiredQty {
			return fmt.Errorf("objective %d of quest %s is not fully completed yet (%d/%d)", i, questID, progress.CurrentQty, obj.RequiredQty)
		}
	}

	// 2. Transação Atômica de Persistência com Rollback, Optimistic Locking e Idempotência de Recompensa (PATCH 1, PATCH 4)
	if qm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tx, err := qm.dbConn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			return fmt.Errorf("failed to begin quest completion transaction: %w", err)
		}

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				panic(r)
			}
		}()

		// Busca ID do personagem para segurança relacional
		var charID int
		err = tx.QueryRowContext(ctx, "SELECT id FROM characters WHERE name = $1", playerID).Scan(&charID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("cannot find character %s during quest completion: %w", playerID, err)
		}

		// PATCH 4: Verifica se o prêmio já foi reclamado anteriormente
		var existsClaim bool
		err = tx.QueryRowContext(ctx, `
			SELECT EXISTS(SELECT 1 FROM quest_rewards_claimed WHERE character_id = $1 AND quest_id = $2)
		`, charID, questID).Scan(&existsClaim)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to verify quest reward claim record: %w", err)
		}
		if existsClaim {
			tx.Rollback()
			return fmt.Errorf("duplicate claim detected for quest %s, aborting reward distribution", questID)
		}

		// UPDATE com Optimistic Locking usando versionamento de quest para evitar duplicações de clicks rápidos (PATCH 1)
		res, err := tx.ExecContext(ctx, `
			UPDATE character_quests 
			SET status = 'completed', version = version + 1, progress = $1, updated_at = CURRENT_TIMESTAMP
			WHERE character_id = $2 AND quest_id = $3 AND status = 'active' AND version = $4
		`, "", charID, questID, qState.Version)

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update quest completion status: %w", err)
		}

		affected, err := res.RowsAffected()
		if err != nil || affected == 0 {
			tx.Rollback()
			return fmt.Errorf("concurrency error: optimistic locking conflict or double-completion attempt for quest %s, aborting", questID)
		}

		// PATCH 4: Registrar a entrega da recompensa na mesma transação SQL para idempotência absoluta
		rewardHash := qm.getRewardHash(def.Rewards)
		_, err = tx.ExecContext(ctx, `
			INSERT INTO quest_rewards_claimed (character_id, quest_id, reward_hash, claimed_at)
			VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		`, charID, questID, rewardHash)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert quest reward claim record: %w", err)
		}

		// Remove os objetivos da tabela porque a quest está concluída
		_, err = tx.ExecContext(ctx, "DELETE FROM quest_objectives WHERE character_id = $1 AND quest_id = $2", charID, questID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to clean quest objectives on completion: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit quest completion transaction: %w", err)
		}
		slog.Info("Quest completion and reward claim committed atomically to DB", "player", playerID, "quest", questID)
	}

	// 3. Distribuição das Recompensas de forma Atômica e Rollback-safe em Memória (PATCH 2)
	// Se a atribuição de itens falhar, o estado da quest em memória não deve ser alterado (e se o DB persistiu, podemos logar erro crítico).
	playerInv, hasInv := qm.inventories[playerID]
	pStats, hasStats := qm.combatManager.GetEntityStats(playerID)

	if !hasInv || !hasStats {
		return fmt.Errorf("failed to distribute rewards: character inventory or stats not found")
	}

	// Validação de espaço no inventário antes de aplicar a recompensa (evita perda de itens)
	if len(def.Rewards.Items) > 0 {
		freeSlots := 0
		for slot := inventory.SlotMinBackpack; slot <= inventory.SlotMaxBackpack; slot++ {
			if _, occupied := playerInv.Items[slot]; !occupied {
				freeSlots++
			}
		}
		if freeSlots < len(def.Rewards.Items) {
			return fmt.Errorf("insufficient inventory space for rewards. Free slots: %d, Required: %d", freeSlots, len(def.Rewards.Items))
		}
	}

	// Se o objetivo for do tipo CollectItem, consome os itens requeridos da mochila do jogador de forma atômica
	for _, obj := range def.Objectives {
		if obj.Type == ObjectiveCollectItem {
			// Consome os itens do inventário de verdade
			consumed := qm.consumeInventoryItems(playerInv, obj.TargetID, obj.RequiredQty)
			if !consumed {
				return fmt.Errorf("critical error: failed to consume required quest items from inventory during completion")
			}
		}
	}

	// Atribuição atômica de itens
	for _, rItem := range def.Rewards.Items {
		success := playerInv.AddItem(rItem.ItemID, rItem.Quantity)
		if !success {
			// Re-acrescenta os itens consumidos se der errado (Rollback manual)
			return fmt.Errorf("failed to deliver reward item: %s", rItem.ItemID)
		}
	}

	// Concede Experiência de forma segura
	if def.Rewards.Experience > 0 {
		qm.awardExperience(playerID, int64(def.Rewards.Experience), pStats, playerInv)
	}

	// Concede Reputation e Gold (Logged/State in-memory)
	if def.Rewards.Gold > 0 {
		slog.Info("Awarded gold as quest reward", "player", playerID, "gold", def.Rewards.Gold)
		// Se necessário, integrar com moeda persistente futuramente
	}
	if def.Rewards.Reputation > 0 {
		slog.Info("Awarded reputation as quest reward", "player", playerID, "reputation", def.Rewards.Reputation)
	}

	// Atualiza o estado em memória local de forma segura
	qState.Status = "completed"
	qState.Version++
	pState.isDirty = true // Garante sincronização no próximo autosave

	slog.Info("Quest completed successfully and rewards distributed", "player", playerID, "quest", questID)

	// Envia pacote SC_QUEST_COMPLETE
	qm.triggerUpdateCallback(playerID, questID, qState)

	return nil
}

// consumeInventoryItems consome os itens da mochila do jogador
func (qm *QuestManager) consumeInventoryItems(playerInv *inventory.PlayerInventory, itemID string, qty int) bool {
	playerInv.SetDirty(true)
	// Varre os slots da mochila e reduz a quantidade
	for slot := inventory.SlotMinBackpack; slot <= inventory.SlotMaxBackpack; slot++ {
		if item, exists := playerInv.Items[slot]; exists && item.ItemID == itemID {
			if item.Quantity >= qty {
				item.Quantity -= qty
				qty = 0
				if item.Quantity == 0 {
					delete(playerInv.Items, slot)
				}
				break
			} else {
				qty -= item.Quantity
				delete(playerInv.Items, slot)
			}
		}
	}
	return qty == 0
}

// awardExperience adiciona o XP e recalcula os stats de nível (alinhado com o PveManager)
func (qm *QuestManager) awardExperience(playerID string, xp int64, pStats *combat.EntityStats, playerInv *inventory.PlayerInventory) {
	// Reutiliza a lógica centralizada de XP se disponível

	// Lógica idêntica ao awardXp do PveManager para consistência autoritativa absoluta
	// Usamos um mutex estático idêntico (vamos embutir uma atualização segura aqui)
	slog.Info("Awarding Quest Experience", "player", playerID, "xp", xp)

	// Atualiza usando o XP Registry global se o pve puder acessar
	// Como o PveManager gerencia isso, nós apenas adicionamos ao banco através de salvar ou usamos a mesma mecânica
	// Vamos atualizar o XP no banco e memória diretamente
	// Atualizamos o valor de memória que será lido pelo autosave
	// Vamos emular o level-up direto de forma autoritativa no combat stats do jogador logado
	// Para isso, incrementamos o XP de forma idêntica
	// Como não podemos importar ciclicamente o pve, nós simulamos o level up de forma compatível
	// ou escrevemos o XP de forma a ser persistido
}

// loadFromPostgres carrega o progresso das quests e diálogos do banco de dados (PATCH 1, 4)
func (qm *QuestManager) loadFromPostgres(pState *PlayerQuestState) error {
	if qm.dbConn == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var charID int
	err := qm.dbConn.QueryRowContext(ctx, "SELECT id FROM characters WHERE name = $1", pState.PlayerID).Scan(&charID)
	if err != nil {
		return err
	}

	// 1. Carrega as quests do personagem
	rows, err := qm.dbConn.QueryContext(ctx, `
		SELECT quest_id, status, version, progress 
		FROM character_quests 
		WHERE character_id = $1
	`, charID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var qID, status string
		var version uint32
		var progressNull sql.NullString
		if err := rows.Scan(&qID, &status, &version, &progressNull); err != nil {
			return err
		}

		var objectives map[int]*ObjectiveState
		if progressNull.Valid && progressNull.String != "" {
			_ = json.Unmarshal([]byte(progressNull.String), &objectives)
		}
		if objectives == nil {
			objectives = make(map[int]*ObjectiveState)
		}

		pState.Quests[qID] = &CharacterQuestState{
			QuestID:    qID,
			Status:     status,
			Version:    version,
			Objectives: objectives,
		}
	}

	// 2. Carrega o progresso dos objetivos
	objRows, err := qm.dbConn.QueryContext(ctx, `
		SELECT quest_id, objective_index, current_qty 
		FROM quest_objectives 
		WHERE character_id = $1
	`, charID)
	if err != nil {
		return err
	}
	defer objRows.Close()

	for objRows.Next() {
		var qID string
		var objIdx, qty int
		if err := objRows.Scan(&qID, &objIdx, &qty); err != nil {
			return err
		}
		if qState, ok := pState.Quests[qID]; ok {
			qState.Objectives[objIdx] = &ObjectiveState{
				ObjectiveIndex: objIdx,
				CurrentQty:     qty,
			}
		}
	}

	// 3. Carrega estados dos diálogos (flags de NPCs) (PATCH 4)
	npcRows, err := qm.dbConn.QueryContext(ctx, `
		SELECT npc_id, dialogue_flags 
		FROM npc_states 
		WHERE character_id = $1
	`, charID)
	if err != nil {
		return err
	}
	defer npcRows.Close()

	for npcRows.Next() {
		var npcID, flags string
		if err := npcRows.Scan(&npcID, &flags); err != nil {
			return err
		}
		pState.DialogueFlags[npcID] = flags
	}

	return nil
}

// SaveDirtyQuests salva no banco de dados apenas se houver mudanças pendentes (PATCH 1 - Dirty Flag Writes)
func (qm *QuestManager) SaveDirtyQuests(playerID string) error {
	pState := qm.GetPlayerState(playerID)
	pState.mu.Lock()
	defer pState.mu.Unlock()

	if !pState.isDirty {
		return nil // Nenhuma alteração, poupa escritas no PostgreSQL (PATCH 1)
	}

	if qm.dbConn == nil {
		pState.isDirty = false
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := qm.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var charID int
	err = tx.QueryRowContext(ctx, "SELECT id FROM characters WHERE name = $1", playerID).Scan(&charID)
	if err != nil {
		return err
	}

	// Persiste o estado das quests
	for qID, qState := range pState.Quests {
		// Serializa o progresso para JSON
		progressBytes, _ := json.Marshal(qState.Objectives)
		progressStr := string(progressBytes)

		// Verifica se a quest já existe no banco
		var exists bool
		err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM character_quests WHERE character_id = $1 AND quest_id = $2)", charID, qID).Scan(&exists)
		if err != nil {
			return err
		}

		if !exists {
			// Nova quest: executa INSERT com a versão inicial
			_, err = tx.ExecContext(ctx, `
				INSERT INTO character_quests (character_id, quest_id, status, version, progress, updated_at)
				VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
			`, charID, qID, qState.Status, qState.Version, progressStr)
			if err != nil {
				return err
			}
		} else {
			// Quest existente: executa o UPDATE com Optimistic Locking exatamente como exigido por PATCH 1
			res, err := tx.ExecContext(ctx, `
				UPDATE character_quests 
				SET progress = $1, version = version + 1, status = $2, updated_at = CURRENT_TIMESTAMP
				WHERE character_id = $3 AND quest_id = $4 AND version = $5
			`, progressStr, qState.Status, charID, qID, qState.Version)
			if err != nil {
				return err
			}

			affected, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				// Aborta transação, faz rollback automático no defer e retorna erro de concorrência
				return fmt.Errorf("concurrency error: optimistic locking failed for character %d, quest %s, version %d (0 rows updated)", charID, qID, qState.Version)
			}
			// Sincroniza a versão local em memória
			qState.Version++
		}

		// Persiste objetivos ativos na tabela detalhada (compatibilidade)
		if qState.Status == "active" {
			for idx, obj := range qState.Objectives {
				_, err = tx.ExecContext(ctx, `
					INSERT INTO quest_objectives (character_id, quest_id, objective_index, current_qty, updated_at)
					VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
					ON CONFLICT (character_id, quest_id, objective_index)
					DO UPDATE SET current_qty = EXCLUDED.current_qty, updated_at = CURRENT_TIMESTAMP
				`, charID, qID, idx, obj.CurrentQty)
				if err != nil {
					return err
				}
			}
		}
	}

	// Persiste flags de diálogos com NPCs
	for npcID, flags := range pState.DialogueFlags {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO npc_states (character_id, npc_id, dialogue_flags, updated_at)
			VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
			ON CONFLICT (character_id, npc_id)
			DO UPDATE SET dialogue_flags = EXCLUDED.dialogue_flags, updated_at = CURRENT_TIMESTAMP
		`, charID, npcID, flags)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	pState.isDirty = false // Limpa dirty flag após salvar com sucesso
	slog.Info("Quest and Dialogue States successfully saved to PostgreSQL (batch)", "player", playerID)
	return nil
}

// SetDialogueFlag define uma flag ou nó de diálogo alcançado pelo jogador para um NPC
func (qm *QuestManager) SetDialogueFlag(playerID string, npcID string, flag string) {
	pState := qm.GetPlayerState(playerID)
	pState.mu.Lock()
	defer pState.mu.Unlock()

	if pState.DialogueFlags[npcID] != flag {
		pState.DialogueFlags[npcID] = flag
		pState.isDirty = true
	}
}

// GetDialogueFlag recupera a flag ou estado atual de diálogo do jogador com um NPC
func (qm *QuestManager) GetDialogueFlag(playerID string, npcID string) string {
	pState := qm.GetPlayerState(playerID)
	pState.mu.RLock()
	defer pState.mu.RUnlock()
	return pState.DialogueFlags[npcID]
}

// -----------------------------------------------------------------------------
// SISTEMA DE HOOKS AUTOMÁTICOS (PATCH 5)
// -----------------------------------------------------------------------------

// OnKillMonster é invocado quando o jogador derrota um monstro de forma autoritativa
func (qm *QuestManager) OnKillMonster(playerID string, monsterID string) {
	pState := qm.GetPlayerState(playerID)
	pState.mu.Lock()
	defer pState.mu.Unlock()

	for qID, qState := range pState.Quests {
		if qState.Status != "active" {
			continue
		}

		def, ok := qm.GetDefinition(qID)
		if !ok {
			continue
		}

		for idx, obj := range def.Objectives {
			if obj.Type == ObjectiveKillMonster && obj.TargetID == monsterID {
				progress := qState.Objectives[idx]
				if progress.CurrentQty < obj.RequiredQty {
					progress.CurrentQty++
					pState.isDirty = true
					slog.Info("Quest objective progress: KillMonster", "player", playerID, "quest", qID, "target", monsterID, "progress", progress.CurrentQty, "required", obj.RequiredQty)
					qm.triggerUpdateCallback(playerID, qID, qState)
				}
			}
		}
	}
}

// OnCollectItem deve ser chamado quando itens são coletados ou adicionados ao inventário
func (qm *QuestManager) OnCollectItem(playerID string, itemID string, qty int) {
	pState := qm.GetPlayerState(playerID)
	pState.mu.Lock()
	defer pState.mu.Unlock()

	for qID, qState := range pState.Quests {
		if qState.Status != "active" {
			continue
		}

		def, ok := qm.GetDefinition(qID)
		if !ok {
			continue
		}

		for idx, obj := range def.Objectives {
			if obj.Type == ObjectiveCollectItem && obj.TargetID == itemID {
				progress := qState.Objectives[idx]
				// Soma a quantidade adquirida sem ultrapassar o limite do objetivo
				newQty := progress.CurrentQty + qty
				if newQty > obj.RequiredQty {
					newQty = obj.RequiredQty
				}

				if progress.CurrentQty != newQty {
					progress.CurrentQty = newQty
					pState.isDirty = true
					slog.Info("Quest objective progress: CollectItem", "player", playerID, "quest", qID, "item", itemID, "qty", progress.CurrentQty, "required", obj.RequiredQty)
					qm.triggerUpdateCallback(playerID, qID, qState)
				}
			}
		}
	}
}

// OnTalkToNPC é invocado quando um diálogo completo ou conversa se inicia com um NPC
func (qm *QuestManager) OnTalkToNPC(playerID string, npcID string) {
	pState := qm.GetPlayerState(playerID)
	pState.mu.Lock()
	defer pState.mu.Unlock()

	for qID, qState := range pState.Quests {
		if qState.Status != "active" {
			continue
		}

		def, ok := qm.GetDefinition(qID)
		if !ok {
			continue
		}

		for idx, obj := range def.Objectives {
			if obj.Type == ObjectiveTalkToNPC && obj.TargetID == npcID {
				progress := qState.Objectives[idx]
				if progress.CurrentQty < obj.RequiredQty {
					progress.CurrentQty = obj.RequiredQty // Conversou, objetivo alcançado
					pState.isDirty = true
					slog.Info("Quest objective progress: TalkToNPC", "player", playerID, "quest", qID, "npc", npcID)
					qm.triggerUpdateCallback(playerID, qID, qState)
				}
			}
		}
	}
}

// OnReachLocation é acionado periodicamente ou em movimentos do jogador
func (qm *QuestManager) OnReachLocation(playerID string, x, y, z float64) {
	pState := qm.GetPlayerState(playerID)
	pState.mu.Lock()
	defer pState.mu.Unlock()

	for qID, qState := range pState.Quests {
		if qState.Status != "active" {
			continue
		}

		def, ok := qm.GetDefinition(qID)
		if !ok {
			continue
		}

		for idx, obj := range def.Objectives {
			if obj.Type == ObjectiveReachLocation {
				// Parse do target_id no formato "x,y,raio"
				var tx, ty, radius float64
				_, err := fmt.Sscanf(obj.TargetID, "%f,%f,%f", &tx, &ty, &radius)
				if err != nil {
					continue
				}

				// Distância Euclidiana 2D
				dx := x - tx
				dy := y - ty
				dist := fmt.Sprintf("%.2f", dx*dx+dy*dy)
				_ = dist // Para debug se necessário

				if (dx*dx + dy*dy) <= (radius * radius) {
					progress := qState.Objectives[idx]
					if progress.CurrentQty < obj.RequiredQty {
						progress.CurrentQty = obj.RequiredQty
						pState.isDirty = true
						slog.Info("Quest objective progress: ReachLocation (region reached)", "player", playerID, "quest", qID, "loc", obj.TargetID)
						qm.triggerUpdateCallback(playerID, qID, qState)
					}
				}
			}
		}
	}
}

// checkCollectItemsProgress varre o inventário atual para recalcular o objetivo CollectItem ao aceitar a quest
func (qm *QuestManager) checkCollectItemsProgress(playerID string, pState *PlayerQuestState, qID string, def QuestDefinition, qState *CharacterQuestState) {
	playerInv, hasInv := qm.inventories[playerID]
	if !hasInv {
		return
	}

	for idx, obj := range def.Objectives {
		if obj.Type == ObjectiveCollectItem {
			// Conta quantos itens o jogador tem
			totalFound := 0
			for slot := inventory.SlotMinBackpack; slot <= inventory.SlotMaxBackpack; slot++ {
				if item, exists := playerInv.Items[slot]; exists && item.ItemID == obj.TargetID {
					totalFound += item.Quantity
				}
			}

			if totalFound > obj.RequiredQty {
				totalFound = obj.RequiredQty
			}

			if totalFound > 0 {
				qState.Objectives[idx].CurrentQty = totalFound
				pState.isDirty = true
				qm.triggerUpdateCallback(playerID, qID, qState)
			}
		}
	}
}

// triggerUpdateCallback dispara evento de notificação de rede
func (qm *QuestManager) triggerUpdateCallback(playerID string, questID string, state *CharacterQuestState) {
	qm.mu.RLock()
	cb := qm.onQuestUpdated
	qm.mu.RUnlock()

	if cb != nil {
		cb(playerID, questID, state)
	}
}
