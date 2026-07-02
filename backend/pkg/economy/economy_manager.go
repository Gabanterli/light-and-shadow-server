package economy

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"
	"github.com/light-and-shadow/backend/pkg/inventory"
	"github.com/light-and-shadow/backend/pkg/movement"
)

// ItemDefinition representa os metadados de itens carregados do config
type ItemDefinition struct {
	ItemID      string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	ValueGold   int64    `json:"value_gold"`
	RepairCost  int64    `json:"repair_cost"`
	MaxStack    int    `json:"max_stack,omitempty"`
}

// MarketOrder representa uma ordem ativa no leilão
type MarketOrder struct {
	OrderID         int64       `json:"order_id"`
	SellerName      string    `json:"seller_name"`
	ItemID          string    `json:"item_id"`
	ItemUUID        string    `json:"item_uuid"`
	Quantity        int       `json:"quantity"`
	PriceGold       int64       `json:"price_gold"`
	TaxGold         int64       `json:"tax_gold"`
	ExpiresAt       time.Time `json:"expires_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// TradeSession representa uma sessão ativa de troca entre dois jogadores
type TradeSession struct {
	TradeID   string
	PlayerA   string
	PlayerB   string
	GoldA     int64
	GoldB     int64
	ItemsA    map[int]*inventory.InventoryItem // Slot -> Item
	ItemsB    map[int]*inventory.InventoryItem // Slot -> Item
	LockedA   bool
	LockedB   bool
	AcceptedA bool
	AcceptedB bool
	CreatedAt time.Time
}

// EconomyManager centraliza todos os controles econômicos, comerciais e de leilão
type EconomyManager struct {
	mu             sync.RWMutex
	db             *sql.DB
	movementSystem *movement.MovementSystem
	itemDefs       map[string]ItemDefinition

	// Sessões ativas de trocas
	activeTrades   map[string]*TradeSession
	playerToTrade  map[string]string // PlayerID -> TradeID
	lockedSlots    map[string]map[int]bool // PlayerID -> SlotIndex -> Locked state (PATCH 2)
}

// NewEconomyManager inicializa o gerenciador da economia
func NewEconomyManager(db *sql.DB, ms *movement.MovementSystem, rawItemsJSON []byte) *EconomyManager {
	defs := make(map[string]ItemDefinition)

	if len(rawItemsJSON) == 0 {
		slog.Warn("No items.json data provided to economy manager, running with empty item definitions")
	} else {
		if err := json.Unmarshal(rawItemsJSON, &defs); err == nil {
			slog.Info("Loaded item definitions for authoritative economy shop validations from map format", "count", len(defs))
		} else {
			var inventoryItems []struct {
				ID              string `json:"ID"`
				Name            string `json:"Name"`
				Type            string `json:"Type"`
				MaxStack        int    `json:"MaxStack"`
				ValueGold       int64  `json:"ValueGold"`
				RepairCost      int64  `json:"RepairCost"`
				ValueGoldSnake  int64  `json:"value_gold"`
				RepairCostSnake int64  `json:"repair_cost"`
			}

			if listErr := json.Unmarshal(rawItemsJSON, &inventoryItems); listErr != nil {
				slog.Error("Failed to unmarshal item definitions for economy manager", "map_error", err, "list_error", listErr)
			} else {
				for _, item := range inventoryItems {
					valueGold := item.ValueGold
					if valueGold == 0 {
						valueGold = item.ValueGoldSnake
					}

					repairCost := item.RepairCost
					if repairCost == 0 {
						repairCost = item.RepairCostSnake
					}

					defs[item.ID] = ItemDefinition{
						ItemID:     item.ID,
						Name:       item.Name,
						Type:       item.Type,
						ValueGold:  valueGold,
						RepairCost: repairCost,
						MaxStack:   item.MaxStack,
					}
				}
				slog.Info("Loaded item definitions for authoritative economy shop validations from inventory list format", "count", len(defs))
			}
		}
	}

	return &EconomyManager{
		db:             db,
		movementSystem: ms,
		itemDefs:       defs,
		activeTrades:   make(map[string]*TradeSession),
		playerToTrade:  make(map[string]string),
		lockedSlots:    make(map[string]map[int]bool),
	}
}

// GetItemDefinition retorna metadados autoritativos do item
func (em *EconomyManager) GetItemDefinition(itemID string) (ItemDefinition, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()
	def, exists := em.itemDefs[itemID]
	return def, exists
}

// IsSlotLocked verifica se o slot do jogador está trancado em uma trade ativa (PATCH 2)
func (em *EconomyManager) IsSlotLocked(playerID string, slotIndex int) bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	slots, exists := em.lockedSlots[playerID]
	if !exists {
		return false
	}
	return slots[slotIndex]
}

// lockSlot tranca o slot para evitar alteração/descarte (PATCH 2)
func (em *EconomyManager) lockSlot(playerID string, slotIndex int) {
	if em.lockedSlots[playerID] == nil {
		em.lockedSlots[playerID] = make(map[int]bool)
	}
	em.lockedSlots[playerID][slotIndex] = true
}

// unlockSlot destranca o slot
func (em *EconomyManager) unlockSlot(playerID string, slotIndex int) {
	if slots, exists := em.lockedSlots[playerID]; exists {
		delete(slots, slotIndex)
	}
}

// clearLockedSlots remove todos os locks de slots do jogador
func (em *EconomyManager) clearLockedSlots(playerID string) {
	delete(em.lockedSlots, playerID)
}

// =========================================================================
// MODULE 1: NPC SHOP SYSTEM (Buy, Sell, Repair)
// =========================================================================

// BuyNPCItem processa a compra de itens de um NPC Shop de forma autoritativa
func (em *EconomyManager) BuyNPCItem(playerID string, itemID string, qty int, playerInv *inventory.PlayerInventory) (string, error) {
	if qty <= 0 {
		return "", errors.New("invalid quantity")
	}

	def, exists := em.GetItemDefinition(itemID)
	if !exists {
		return "", fmt.Errorf("item %s not found in catalog", itemID)
	}

	totalCost := def.ValueGold * int64(qty)
	if totalCost < 0 {
		return "", errors.New("negative overflow on cost calculation")
	}

	// Atômico: Verifica gold e remove gold (PATCH 1)
	if !playerInv.RemoveGold(int64(totalCost)) {
		return "", fmt.Errorf("insufficient gold: need %d, have %d", totalCost, playerInv.GetGold())
	}

	// Tenta adicionar ao inventário em slots livres ou stackáveis

	// Encontra slots para alocar os itens
	// Se for stackable, procura slot existente com o mesmo itemID que não esteja cheio
	maxStack := def.MaxStack
	if maxStack <= 0 {
		maxStack = 1
	}

	isStackable := maxStack > 1
	toAllocate := qty

	if isStackable {
		for _, item := range playerInv.Items {
			if item != nil && item.ItemID == itemID && item.Quantity < maxStack {
				spaceLeft := maxStack - item.Quantity
				if spaceLeft >= toAllocate {
					item.Quantity += toAllocate
					toAllocate = 0
					break
				} else {
					item.Quantity = maxStack
					toAllocate -= spaceLeft
				}
			}
		}
	}

	// Cria novos slots para os restantes
	if toAllocate > 0 {
		for slot := 0; slot < 30; slot++ {
			if playerInv.Items[slot] == nil {
				allocQty := toAllocate
				if allocQty > maxStack {
					allocQty = maxStack
				}

				// UUID único anti-duplicação se for item único/durável (PATCH 4)
				itemUUID := ""
				if maxStack == 1 {
					itemUUID = inventory.GenerateUUIDv4()
				}

				playerInv.Items[slot] = &inventory.InventoryItem{
					ItemID:     itemID,
					Quantity:   allocQty,
					Durability: 100,
					SlotIndex:  slot,
					ItemUUID:   itemUUID,
				}
				toAllocate -= allocQty
				if toAllocate <= 0 {
					break
				}
			}
		}
	}

	if toAllocate > 0 {
		// Sem espaço suficiente! Desfaz atômico devolvendo o gold
		playerInv.AddGold(totalCost)
		return "", errors.New("inventory full")
	}

	playerInv.SetDirty(true)
	slog.Info("NPC purchase executed successfully", "player", playerID, "item", itemID, "qty", qty, "cost", totalCost)
	return fmt.Sprintf("Bought %d %s", qty, def.Name), nil
}

// SellNPCItem processa a venda de itens para um NPC Shop de forma autoritativa
func (em *EconomyManager) SellNPCItem(playerID string, slotIndex int, qty int, playerInv *inventory.PlayerInventory) (string, error) {
	if qty <= 0 {
		return "", errors.New("invalid quantity")
	}

	if em.IsSlotLocked(playerID, slotIndex) {
		return "", errors.New("slot is currently locked in an active trade")
	}

	item, exists := playerInv.Items[slotIndex]
	if !exists || item == nil {
		return "", errors.New("item not found in selected slot")
	}

	if item.Quantity < qty {
		return "", errors.New("not enough items in slot to sell")
	}

	def, exists := em.itemDefs[item.ItemID]
	if !exists {
		return "", errors.New("invalid item definition")
	}

	// Venda dá 50% do valor de compra
	sellValue := (def.ValueGold * 50) / 100
	if sellValue < 1 {
		sellValue = 1
	}
	totalYield := sellValue * int64(qty)

	// Remove item do inventário
	item.Quantity -= qty
	if item.Quantity == 0 {
		delete(playerInv.Items, slotIndex)
	}
	playerInv.SetDirty(true)

	// Adiciona Gold atômico (PATCH 1)
	playerInv.AddGold(totalYield)

	slog.Info("NPC item sale executed successfully", "player", playerID, "item", def.ItemID, "qty", qty, "yield", totalYield)
	return fmt.Sprintf("Sold %d %s for %d Gold", qty, def.Name, totalYield), nil
}

// RepairItem NPC repara a durabilidade do equipamento selecionado se houver gold
func (em *EconomyManager) RepairItem(playerID string, slotIndex int, playerInv *inventory.PlayerInventory) (string, error) {
	if em.IsSlotLocked(playerID, slotIndex) {
		return "", errors.New("slot is locked in active trade")
	}

	item, exists := playerInv.Items[slotIndex]
	if !exists || item == nil {
		return "", errors.New("item not found in selected slot")
	}

	if item.Durability >= 100 {
		return "", errors.New("item already has maximum durability")
	}

	def, exists := em.itemDefs[item.ItemID]
	if !exists {
		return "", errors.New("invalid item metadata")
	}

	damage := 100 - item.Durability
	cost := def.RepairCost * int64(damage)
	if cost < 0 {
		cost = 0
	}

	// Atômico: Verifica se tem gold suficiente e remove (PATCH 1)
    if !playerInv.RemoveGold(cost) {
	return "", fmt.Errorf("insufficient gold: need %d, have %d", cost, playerInv.GetGold())
}

	item.Durability = 100
	playerInv.SetDirty(true)
	
	slog.Info("Equipment repaired successfully", "player", playerID, "item", def.ItemID, "cost", cost)
	return fmt.Sprintf("Repaired %s for %d Gold", def.Name, cost), nil
}

// =========================================================================
// MODULE 2: PLAYER TRADING (Request, Accept, Reject, Dual Lock/Confirm)
// =========================================================================

// StartTradeSession inicia uma nova proposta de troca se ambos jogadores estiverem livres
func (em *EconomyManager) StartTradeSession(playerA, playerB string) (string, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Valida se já estão em trade
	if _, inTrade := em.playerToTrade[playerA]; inTrade {
		return "", errors.New("you are already in an active trade session")
	}
	if _, inTrade := em.playerToTrade[playerB]; inTrade {
		return "", errors.New("the target player is already in an active trade session")
	}

	// Valida proximidade (< 3 tiles)
	ax, ay, _, aOk := em.movementSystem.GetPlayerPos(playerA)
	bx, by, _, bOk := em.movementSystem.GetPlayerPos(playerB)
	if !aOk || !bOk {
		return "", errors.New("both players must be logged and positioned in the world")
	}

	distance := math.Sqrt(math.Pow(ax-bx, 2) + math.Pow(ay-by, 2))
	if distance > 3.0 {
		return "", errors.New("players are too far apart to trade (must be within 3 tiles)")
	}

	tradeID := fmt.Sprintf("%d", time.Now().UnixNano())
	session := &TradeSession{
		TradeID:   tradeID,
		PlayerA:   playerA,
		PlayerB:   playerB,
		ItemsA:    make(map[int]*inventory.InventoryItem),
		ItemsB:    make(map[int]*inventory.InventoryItem),
		CreatedAt: time.Now(),
	}

	em.activeTrades[tradeID] = session
	em.playerToTrade[playerA] = tradeID
	em.playerToTrade[playerB] = tradeID

	slog.Info("Active trade session established", "trade_id", tradeID, "p_a", playerA, "p_b", playerB)
	return tradeID, nil
}

// OfferGold oferece gold para a troca. Só funciona se a sua parte não estiver bloqueada (Locked)
func (em *EconomyManager) OfferGold(playerID string, gold int64, playerInv *inventory.PlayerInventory) error {
	em.mu.Lock()
	tradeID, exists := em.playerToTrade[playerID]
	if !exists {
		em.mu.Unlock()
		return errors.New("no active trade session found")
	}
	session := em.activeTrades[tradeID]
	em.mu.Unlock()

	if gold < 0 {
		return errors.New("cannot offer negative gold")
	}

	if playerInv.GetGold() < int64(gold) {
	return errors.New("insufficient gold to offer")
    }

	em.mu.Lock()
	defer em.mu.Unlock()

	if playerID == session.PlayerA {
		if session.LockedA {
			return errors.New("cannot change offer when state is locked")
		}
		session.GoldA = gold
		session.AcceptedA = false
		session.AcceptedB = false
	} else {
		if session.LockedB {
			return errors.New("cannot change offer when state is locked")
		}
		session.GoldB = gold
		session.AcceptedA = false
		session.AcceptedB = false
	}

	return nil
}

// OfferItem oferece um item do slot do inventário para a troca
func (em *EconomyManager) OfferItem(playerID string, slotIndex int, qty int, playerInv *inventory.PlayerInventory) error {
	if qty <= 0 {
		return errors.New("invalid quantity")
	}

	em.mu.Lock()
	tradeID, exists := em.playerToTrade[playerID]
	if !exists {
		em.mu.Unlock()
		return errors.New("no active trade session found")
	}
	session := em.activeTrades[tradeID]
	em.mu.Unlock()

	item, existsItem := playerInv.Items[slotIndex]
	if !existsItem || item == nil {
		return errors.New("item not found in selected inventory slot")
	}
	if item.Quantity < qty {
	return errors.New("not enough items in slot to offer")
    }

	// Cria uma cópia do item com a quantidade oferecida
	offered := &inventory.InventoryItem{
		ItemID:     item.ItemID,
		Quantity:   qty,
		Durability: item.Durability,
		SlotIndex:  slotIndex, // Mapeia o slot original para evitar bypass
		ItemUUID:   item.ItemUUID,
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	if playerID == session.PlayerA {
		if session.LockedA {
			return errors.New("trade is locked")
		}
		session.ItemsA[slotIndex] = offered
		session.AcceptedA = false
		session.AcceptedB = false
		em.lockSlot(playerID, slotIndex) // LOCK NO SLOT DO INVENTÁRIO (PATCH 2)
	} else {
		if session.LockedB {
			return errors.New("trade is locked")
		}
		session.ItemsB[slotIndex] = offered
		session.AcceptedA = false
		session.AcceptedB = false
		em.lockSlot(playerID, slotIndex) // LOCK NO SLOT DO INVENTÁRIO (PATCH 2)
	}

	return nil
}

// LockTrade confirma a oferta e a trava. Após travar, nenhuma oferta pode ser alterada (Dual Confirm - Passo 1)
func (em *EconomyManager) LockTrade(playerID string) (*TradeSession, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	tradeID, exists := em.playerToTrade[playerID]
	if !exists {
		return nil, errors.New("no active trade session found")
	}
	session := em.activeTrades[tradeID]

	if playerID == session.PlayerA {
		session.LockedA = true
	} else {
		session.LockedB = true
	}

	slog.Info("Trade locked by player", "player", playerID, "locked_a", session.LockedA, "locked_b", session.LockedB)
	return session, nil
}

// CompleteTrade aceita e executa a troca definitiva se ambos travaram e aceitaram (Dual Confirm - Passo 2 - Rollback-safe)
func (em *EconomyManager) CompleteTrade(playerID string, invA, invB *inventory.PlayerInventory) (bool, string, error) {
	em.mu.Lock()
	tradeID, exists := em.playerToTrade[playerID]
	if !exists {
		em.mu.Unlock()
		return false, "", errors.New("no active trade session found")
	}
	session := em.activeTrades[tradeID]

	if !session.LockedA || !session.LockedB {
		em.mu.Unlock()
		return false, "", errors.New("both players must lock their offers before completing the trade")
	}

	if playerID == session.PlayerA {
		session.AcceptedA = true
	} else {
		session.AcceptedB = true
	}

	// Se ambos não aceitaram ainda, apenas salva o estado e retorna pendente
	if !session.AcceptedA || !session.AcceptedB {
		em.mu.Unlock()
		return false, "Trade acceptance recorded, waiting for second player confirmation", nil
	}
	em.mu.Unlock()

	// AMBOS ACEITARAM! Executa de forma transacional e segura em termos de threads e banco de dados
	// lock duplo hierárquico nos inventários para evitar deadlocks de concorrência
	var first, second *inventory.PlayerInventory

    if invA.PlayerID < invB.PlayerID {
	first = invA
	second = invB
    } else {
	first = invB
	second = invA
    }

    first.Lock()
	second.Lock()
	defer second.Unlock()
	defer first.Unlock()

	// 1. Validações adicionais finais antes da transferência de fundos
	if invA.Gold < int64(session.GoldA) {
		em.CancelTrade(playerID)
		return false, "", fmt.Errorf("player %s has insufficient gold", session.PlayerA)
	}
	if invB.Gold < int64(session.GoldB) {
		em.CancelTrade(playerID)
		return false, "", fmt.Errorf("player %s has insufficient gold", session.PlayerB)
	}

	// Valida se itens oferecidos por A ainda estão disponíveis e intocados
	for slot, itemOffered := range session.ItemsA {
		current, exists := invA.Items[slot]
		if !exists || current == nil || current.ItemID != itemOffered.ItemID || current.Quantity < itemOffered.Quantity || current.ItemUUID != itemOffered.ItemUUID {
			em.CancelTrade(playerID)
			return false, "", fmt.Errorf("player %s offered item has changed or is no longer in inventory", session.PlayerA)
		}
	}

	// Valida se itens oferecidos por B ainda estão disponíveis e intocados
	for slot, itemOffered := range session.ItemsB {
		current, exists := invB.Items[slot]
		if !exists || current == nil || current.ItemID != itemOffered.ItemID || current.Quantity < itemOffered.Quantity || current.ItemUUID != itemOffered.ItemUUID {
			em.CancelTrade(playerID)
			return false, "", fmt.Errorf("player %s offered item has changed or is no longer in inventory", session.PlayerB)
		}
	}

	// 2. Transferência atômica de Gold (PATCH 1)
	invA.Gold -= session.GoldA
	invA.Gold += session.GoldB

	invB.Gold -= session.GoldB
	invB.Gold += session.GoldA

	// 3. Transferência segura de Itens e slots
	// Remove os itens de A
	for slot, itemOffered := range session.ItemsA {
		current := invA.Items[slot]
		current.Quantity -= itemOffered.Quantity
		if current.Quantity == 0 {
			delete(invA.Items, slot)
		}
	}

	// Remove os itens de B
	for slot, itemOffered := range session.ItemsB {
		current := invB.Items[slot]
		current.Quantity -= itemOffered.Quantity
		if current.Quantity == 0 {
			delete(invB.Items, slot)
		}
	}

	// Adiciona itens de A para o inventário de B
	for _, itemOffered := range session.ItemsA {
		// Aloca de forma segura no inventário de B em slots livres
		allocated := false
		for s := 0; s < 30; s++ {
			if invB.Items[s] == nil {
				invB.Items[s] = &inventory.InventoryItem{
					ItemID:     itemOffered.ItemID,
					Quantity:   itemOffered.Quantity,
					Durability: itemOffered.Durability,
					SlotIndex:  s,
					ItemUUID:   itemOffered.ItemUUID, // Anti-duplicação preservada (PATCH 4)
				}
				allocated = true
				break
			}
		}
		if !allocated {
			// Desastre teórico: B não tem espaço! Cancela e rola de volta tudo.
			// Na prática isso deve ser mitigado. Como o lock duplo está segurando, desfazemos as mutações in-memory:
			invA.Gold += session.GoldA
			invA.Gold -= session.GoldB
			invB.Gold += session.GoldB
			invB.Gold -= session.GoldA
			// Restaura itens de A
			for slot, itemOffered := range session.ItemsA {
				if current, ok := invA.Items[slot]; ok {
					current.Quantity += itemOffered.Quantity
				} else {
					invA.Items[slot] = itemOffered
				}
			}
			// Restaura itens de B
			for slot, itemOffered := range session.ItemsB {
				if current, ok := invB.Items[slot]; ok {
					current.Quantity += itemOffered.Quantity
				} else {
					invB.Items[slot] = itemOffered
				}
			}
			em.CancelTrade(playerID)
			return false, "", errors.New("transaction aborted: buyer inventory full during commit phase")
		}
	}

	// Adiciona itens de B para o inventário de A
	for _, itemOffered := range session.ItemsB {
		allocated := false
		for s := 0; s < 30; s++ {
			if invA.Items[s] == nil {
				invA.Items[s] = &inventory.InventoryItem{
					ItemID:     itemOffered.ItemID,
					Quantity:   itemOffered.Quantity,
					Durability: itemOffered.Durability,
					SlotIndex:  s,
					ItemUUID:   itemOffered.ItemUUID,
				}
				allocated = true
				break
			}
		}
		if !allocated {
			// Rollback in-memory
			invA.Gold += session.GoldA
			invA.Gold -= session.GoldB
			invB.Gold += session.GoldB
			invB.Gold -= session.GoldA
			// Restores
			for slot, itemOffered := range session.ItemsA {
				if current, ok := invA.Items[slot]; ok {
					current.Quantity += itemOffered.Quantity
				} else {
					invA.Items[slot] = itemOffered
				}
			}
			for slot, itemOffered := range session.ItemsB {
				if current, ok := invB.Items[slot]; ok {
					current.Quantity += itemOffered.Quantity
				} else {
					invB.Items[slot] = itemOffered
				}
			}
			em.CancelTrade(playerID)
			return false, "", errors.New("transaction aborted: buyer inventory full during commit phase")
		}
	}

	// Marca ambos os inventários como Dirty para o autosave persistir no DB (PATCH 2)
	invA.SetDirty(true)
	invB.SetDirty(true)

	// 4. Log do Trade no PostgreSQL (Rollback-safe persistência se ativo - PATCH 5)
	if em.db != nil {
		go func() {
			itemsA_JSON, _ := json.Marshal(session.ItemsA)
			itemsB_JSON, _ := json.Marshal(session.ItemsB)
			_, err := em.db.Exec(`
				INSERT INTO trade_logs (player_a_name, player_b_name, gold_a, gold_b, items_a, items_b)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, session.PlayerA, session.PlayerB, session.GoldA, session.GoldB, string(itemsA_JSON), string(itemsB_JSON))
			if err != nil {
				slog.Error("Failed to persist transaction log to DB", "error", err)
			}
		}()
	}

	// Limpa sessão de trade
	em.mu.Lock()
	em.clearLockedSlots(session.PlayerA)
	em.clearLockedSlots(session.PlayerB)
	delete(em.playerToTrade, session.PlayerA)
	delete(em.playerToTrade, session.PlayerB)
	delete(em.activeTrades, session.TradeID)
	em.mu.Unlock()

	slog.Info("Trade transaction committed successfully!", "trade_id", session.TradeID, "p_a", session.PlayerA, "p_b", session.PlayerB)
	return true, fmt.Sprintf("Trade committed! Executed transfer of %d items", len(session.ItemsA)+len(session.ItemsB)), nil
}

// CancelTrade cancela imediatamente e destranca todos os slots envolvidos (Rollback seguro)
func (em *EconomyManager) CancelTrade(playerID string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	tradeID, exists := em.playerToTrade[playerID]
	if !exists {
		return
	}

	session := em.activeTrades[tradeID]
	if session == nil {
		return
	}

	// Destranca os slots dos jogadores
	em.clearLockedSlots(session.PlayerA)
	em.clearLockedSlots(session.PlayerB)

	delete(em.playerToTrade, session.PlayerA)
	delete(em.playerToTrade, session.PlayerB)
	delete(em.activeTrades, tradeID)

	slog.Info("Trade session aborted and rolled back cleanly", "trade_id", tradeID, "player_triggered", playerID)
}

// ValidateProximityAndStatus garante o auto cancelamento autoritativo por distância, morte ou desconexão (PATCH 3)
func (em *EconomyManager) ValidateProximityAndStatus(playerID string, playerHP float64) bool {
	// Se o HP estiver zerado (morte), cancela trade
	if playerHP <= 0 {
		slog.Warn("Player died, auto cancelling active trade session", "player", playerID)
		em.CancelTrade(playerID)
		return false
	}

	em.mu.RLock()
	tradeID, exists := em.playerToTrade[playerID]
	if !exists {
		em.mu.RUnlock()
		return true
	}
	session := em.activeTrades[tradeID]
	em.mu.RUnlock()

	// Se ultrapassou 3 tiles de distância, auto cancela
	ax, ay, _, aOk := em.movementSystem.GetPlayerPos(session.PlayerA)
	bx, by, _, bOk := em.movementSystem.GetPlayerPos(session.PlayerB)
	if !aOk || !bOk {
		slog.Warn("Player disconnected or coordinate state lost, auto cancelling trade", "player", playerID)
		em.CancelTrade(playerID)
		return false
	}

	distance := math.Sqrt(math.Pow(ax-bx, 2) + math.Pow(ay-by, 2))
	if distance > 3.0 {
		slog.Warn("Players moved beyond 3 tiles of distance threshold, trade auto cancelled", "player", playerID, "dist", distance)
		em.CancelTrade(playerID)
		return false
	}

	return true
}

// =========================================================================
// MODULE 3: MARKETPLACE / AUCTION HOUSE (Escrowed Orders & Transactions)
// =========================================================================

// CreateMarketOrder cria uma ordem de leilão escrowando o item do inventário do jogador imediatamente (PATCH 3)
func (em *EconomyManager) CreateMarketOrder(playerID string, slotIndex int, qty int, priceGold int64, playerInv *inventory.PlayerInventory) (string, error) {
	if priceGold <= 0 || qty <= 0 {
		return "", errors.New("invalid quantity or price parameters")
	}

	if em.IsSlotLocked(playerID, slotIndex) {
		return "", errors.New("item slot is locked in active trade session")
	}

	if em.db == nil {
		return "", errors.New("marketplace database connection is in fallback offline mode")
	}

	item, exists := playerInv.Items[slotIndex]
	if !exists || item == nil {
		return "", errors.New("item not found in selected slot")
	}

	if item.Quantity < qty {
	return "", errors.New("insufficient item quantity in slot")
	}

	itemID := item.ItemID
	itemUUID := item.ItemUUID

	// Copia o item para salvar no banco de dados como ESCROW
	// Remove os itens do inventário de forma autoritativa IMEDIATAMENTE (PATCH 3)
	item.Quantity -= qty
	if item.Quantity == 0 {
		delete(playerInv.Items, slotIndex)
	}
	playerInv.SetDirty(true)

	// Taxa de listagem: 5% do valor do preço pedido
	tax := int64(math.Floor(float64(priceGold) * 0.05))
	if tax < 1 {
		tax = 1
	}

	// Remove a taxa de listagem de forma atômica do ouro do vendedor (PATCH 1)
	if !playerInv.RemoveGold(tax) {
		// Devolve os itens em caso de erro (Rollback-safe)
		if current, ok := playerInv.Items[slotIndex]; ok {
			current.Quantity += qty
		} else {
			playerInv.Items[slotIndex] = &inventory.InventoryItem{
				ItemID:     itemID,
				Quantity:   qty,
				Durability: 100,
				SlotIndex:  slotIndex,
				ItemUUID:   itemUUID,
			}
		}
		return "", fmt.Errorf("insufficient gold to cover listing tax: need %d", tax)
	}

	// Persiste a ordem no banco de dados como depósito em garantia (Escrow - PATCH 3)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := em.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		// Rollback in-memory
		playerInv.AddGold(tax)
		if current, ok := playerInv.Items[slotIndex]; ok {
			current.Quantity += qty
		} else {
			playerInv.Items[slotIndex] = &inventory.InventoryItem{
				ItemID:     itemID,
				Quantity:   qty,
				Durability: 100,
				SlotIndex:  slotIndex,
				ItemUUID:   itemUUID,
			}
		}
		return "", fmt.Errorf("failed to begin SQL transaction: %w", err)
	}
	defer tx.Rollback()

	// Pega ID do personagem do vendedor no PostgreSQL
	var sellerCharID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM characters WHERE name = $1", playerID).Scan(&sellerCharID)
	if err != nil {
		return "", fmt.Errorf("character %s not found in DB: %w", playerID, err)
	}

	// Expiração padrão de 24 horas (PATCH 3)
	expiresAt := time.Now().Add(24 * time.Hour)

	var orderID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO market_orders (seller_character_id, item_id, item_uuid, quantity, price_gold, tax_gold, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, sellerCharID, itemID, itemUUID, qty, priceGold, tax, expiresAt).Scan(&orderID)

	if err != nil {
		return "", fmt.Errorf("failed to persist escrow order: %w", err)
	}

	// Comita a ordem de listagem no PostgreSQL de forma rollback-safe (PATCH 5)
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit SQL transaction: %w", err)
	}

	slog.Info("Escrow market order successfully listed", "order_id", orderID, "seller", playerID, "item", itemID, "qty", qty, "price", priceGold)
	return fmt.Sprintf("Listed Order #%d: %d %s for %d Gold", orderID, qty, itemID, priceGold), nil
}

// BuyMarketItem processa a compra de uma ordem escrowed do mercado com taxação e transferências atômicas (PATCH 1, 3, 5)
func (em *EconomyManager) BuyMarketItem(buyerID string, orderID int64, buyerInv *inventory.PlayerInventory) (string, error) {
	if em.db == nil {
		return "", errors.New("marketplace database offline")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Inicia transação SQL robusta com isolamento serializável absoluto (PATCH 5)
	tx, err := em.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return "", fmt.Errorf("failed to start purchase transaction: %w", err)
	}
	defer tx.Rollback()

	// Busca detalhes da ordem bloqueando a linha correspondente de forma segura (Lock)
	var sellerCharID int64
	var sellerName string
	var itemID string
	var itemUUID sql.NullString
	var qty int
	var priceGold int64
	var taxGold int64
	var expiresAt time.Time

	err = tx.QueryRowContext(ctx, `
		SELECT mo.id, mo.seller_character_id, c.name, mo.item_id, mo.item_uuid, mo.quantity, mo.price_gold, mo.tax_gold, mo.expires_at
		FROM market_orders mo
		JOIN characters c ON mo.seller_character_id = c.id
		WHERE mo.id = $1 FOR UPDATE
	`, orderID).Scan(&orderID, &sellerCharID, &sellerName, &itemID, &itemUUID, &qty, &priceGold, &taxGold, &expiresAt)

	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("this market order does not exist or has already been purchased")
	} else if err != nil {
		return "", fmt.Errorf("failed to scan market order details: %w", err)
	}

	// Verifica se a ordem expirou
	if time.Now().After(expiresAt) {
		return "", errors.New("this market order has expired and cannot be purchased")
	}

	if sellerName == buyerID {
		return "", errors.New("you cannot purchase your own listed market order")
	}

	// Verifica fundos do comprador de forma segura
	if buyerInv.GetGold() < priceGold {
		return "", fmt.Errorf("insufficient gold to purchase: need %d", priceGold)
	}

	// Dedução atômica de gold do comprador (PATCH 1)
	buyerInv.RemoveGold(priceGold)

	// Adiciona os itens escrowed no inventário do comprador de forma autoritativa
	buyerInv.Lock()
	allocated := false
	for s := 0; s < 30; s++ {
		if buyerInv.Items[s] == nil {
			buyerInv.Items[s] = &inventory.InventoryItem{
				ItemID:     itemID,
				Quantity:   qty,
				Durability: 100,
				SlotIndex:  s,
				ItemUUID:   itemUUID.String, // Preserva UUID único para evitar duplicações (PATCH 4)
			}
			allocated = true
			break
		}
	}
	buyerInv.Unlock()

	if !allocated {
		// Devolve o gold do comprador (Rollback)
		buyerInv.AddGold(priceGold)
		return "", errors.New("purchase aborted: buyer inventory has no free slots")
	}

	buyerInv.SetDirty(true)

	// Atualiza gold do vendedor no PostgreSQL de forma direta e segura no banco
	// Vendedor recebe: Preço - Taxa (tax_gold é recolhido pelo leilão)
	payout := priceGold // Taxa de listagem (taxGold) já foi cobrada na postagem da ordem!
	_, err = tx.ExecContext(ctx, "UPDATE characters SET gold = gold + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", payout, sellerCharID)
	if err != nil {
		buyerInv.Lock()
		// Remove item do inventário do comprador (Rollback)
		for s, item := range buyerInv.Items {
			if item != nil && item.ItemUUID == itemUUID.String && item.ItemID == itemID {
				delete(buyerInv.Items, s)
				break
			}
		}
		buyerInv.Unlock()
		buyerInv.AddGold(priceGold)
		return "", fmt.Errorf("failed to transfer gold to seller: %w", err)
	}

	// Insere no histórico de leilão
	var buyerCharID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM characters WHERE name = $1", buyerID).Scan(&buyerCharID)
	if err != nil {
		buyerInv.Lock()
		for s, item := range buyerInv.Items {
			if item != nil && item.ItemUUID == itemUUID.String && item.ItemID == itemID {
				delete(buyerInv.Items, s)
				break
			}
		}
		buyerInv.Unlock()
		buyerInv.AddGold(priceGold)
		return "", fmt.Errorf("failed to retrieve buyer identity from DB: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO market_history (seller_character_id, buyer_character_id, item_id, quantity, price_gold, tax_gold)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, sellerCharID, buyerCharID, itemID, qty, priceGold, taxGold)
	if err != nil {
		slog.Error("Failed to record market history", "error", err)
	}

	// Deleta a ordem ativa no banco de dados (Remoção do Escrow)
	_, err = tx.ExecContext(ctx, "DELETE FROM market_orders WHERE id = $1", orderID)
	if err != nil {
		return "", fmt.Errorf("failed to clear market order from active lists: %w", err)
	}

	// Commit transacional definitivo (PATCH 5)
	if err := tx.Commit(); err != nil {
		buyerInv.Lock()
		for s, item := range buyerInv.Items {
			if item != nil && item.ItemUUID == itemUUID.String && item.ItemID == itemID {
				delete(buyerInv.Items, s)
				break
			}
		}
		buyerInv.Unlock()
		buyerInv.AddGold(priceGold)
		return "", fmt.Errorf("transaction commit failed: %w", err)
	}

	slog.Info("Escrow market purchase completed successfully", "order_id", orderID, "buyer", buyerID, "seller", sellerName, "price", priceGold)
	return fmt.Sprintf("Purchased Order #%d for %d Gold", orderID, priceGold), nil
}

// CancelMarketOrder cancela uma ordem de leilão devolvendo o item escrowed para o inventário do vendedor
func (em *EconomyManager) CancelMarketOrder(playerID string, orderID int64, playerInv *inventory.PlayerInventory) (string, error) {
	if em.db == nil {
		return "", errors.New("marketplace database offline")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := em.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var sellerName string
	var itemID string
	var itemUUID sql.NullString
	var qty int

	err = tx.QueryRowContext(ctx, `
		SELECT c.name, mo.item_id, mo.item_uuid, mo.quantity
		FROM market_orders mo
		JOIN characters c ON mo.seller_character_id = c.id
		WHERE mo.id = $1 FOR UPDATE
	`, orderID).Scan(&sellerName, &itemID, &itemUUID, &qty)

	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("market order not found")
	} else if err != nil {
		return "", err
	}

	if sellerName != playerID {
		return "", errors.New("you are not authorized to cancel this order")
	}

	// Devolve os itens do Escrow de volta ao inventário (atômico)
	allocated := false
	for s := 0; s < 30; s++ {
		if playerInv.Items[s] == nil {
			playerInv.Items[s] = &inventory.InventoryItem{
				ItemID:     itemID,
				Quantity:   qty,
				Durability: 100,
				SlotIndex:  s,
				ItemUUID:   itemUUID.String,
			}
			allocated = true
			break
		}
	}

	if !allocated {
		return "", errors.New("cancel aborted: your inventory has no free slots to receive the escrowed items back")
	}

	playerInv.SetDirty(true)

	// Remove ordem do PostgreSQL
	_, err = tx.ExecContext(ctx, "DELETE FROM market_orders WHERE id = $1", orderID)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	slog.Info("Escrow market order cancelled cleanly", "order_id", orderID, "player", playerID)
	return fmt.Sprintf("Cancelled Order #%d and returned item", orderID), nil
}

// SearchMarketOrders retorna as ordens de leilão ativas filtradas por item_id
func (em *EconomyManager) SearchMarketOrders(filterItemID string) ([]MarketOrder, error) {
	if em.db == nil {
		return nil, errors.New("marketplace database offline")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string
	var args []interface{}

	if filterItemID != "" {
		query = `
			SELECT mo.id, c.name, mo.item_id, mo.item_uuid, mo.quantity, mo.price_gold, mo.tax_gold, mo.expires_at, mo.created_at
			FROM market_orders mo
			JOIN characters c ON mo.seller_character_id = c.id
			WHERE mo.item_id ILIKE $1 AND mo.expires_at > CURRENT_TIMESTAMP
			ORDER BY mo.created_at DESC
		`
		args = append(args, "%"+filterItemID+"%")
	} else {
		query = `
			SELECT mo.id, c.name, mo.item_id, mo.item_uuid, mo.quantity, mo.price_gold, mo.tax_gold, mo.expires_at, mo.created_at
			FROM market_orders mo
			JOIN characters c ON mo.seller_character_id = c.id
			WHERE mo.expires_at > CURRENT_TIMESTAMP
			ORDER BY mo.created_at DESC
		`
	}

	rows, err := em.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []MarketOrder
	for rows.Next() {
		var o MarketOrder
		var itemUUID sql.NullString
		err := rows.Scan(&o.OrderID, &o.SellerName, &o.ItemID, &itemUUID, &o.Quantity, &o.PriceGold, &o.TaxGold, &o.ExpiresAt, &o.CreatedAt)
		if err != nil {
			return nil, err
		}
		o.ItemUUID = itemUUID.String
		orders = append(orders, o)
	}

	return orders, nil
}

// GetTradeSession retorna uma cópia segura do estado de troca de um jogador
func (em *EconomyManager) GetTradeSession(playerID string) (TradeSession, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()
	tradeID, exists := em.playerToTrade[playerID]
	if !exists {
		return TradeSession{}, false
	}
	session := em.activeTrades[tradeID]
	if session == nil {
		return TradeSession{}, false
	}
	return *session, true
}
