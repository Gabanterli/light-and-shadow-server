package inventory

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/pkg/combat"
)

// GenerateUUIDv4 gera um UUIDv4 de forma criptograficamente segura (anti-duplicação - PATCH 4)
func GenerateUUIDv4() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	// Ajusta bits para a versão 4
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

const (
	SlotMinBackpack = 0
	SlotMaxBackpack = 29
	SlotWeapon      = 30
	SlotArmor       = 31
	SlotAccessory   = 32
	TotalSlots      = 33
)

// ItemDef define a estrutura estática de um item
type ItemDef struct {
	ID         string  `json:"ID"`
	Name       string  `json:"Name"`
	Type       string  `json:"Type"` // "weapon", "armor", "accessory", "consumable", "material"
	Stackable  bool    `json:"Stackable"`
	MaxStack   int     `json:"MaxStack"`
	BaseDamage float64 `json:"BaseDamage"` // Para armas
	BaseDef    float64 `json:"BaseDef"`    // Para armaduras
	BaseRes    float64 `json:"BaseRes"`    // Para acessórios (Resistência adicional)
	CritBonus  float64 `json:"CritBonus"`  // Para acessórios (CritChance adicional)
	Tier       int     `json:"Tier"`       // Tier do item para restrições de nível (Sprint 3 Task 5)
}

// ItemDictionary e itemDictMu controlam a biblioteca autoritativa e mutável de itens de forma thread-safe (PATCH 5)
var (
	itemDictMu     sync.RWMutex
	ItemDictionary = make(map[string]ItemDef)
)

// LoadItemDefinitions carrega de forma atômica e thread-safe a definição dos itens a partir do arquivo JSON (PATCH 5)
func LoadItemDefinitions(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read item definitions file: %w", err)
	}

	var itemsList []ItemDef
	if err := json.Unmarshal(data, &itemsList); err != nil {
		return fmt.Errorf("failed to unmarshal item definitions: %w", err)
	}

	newDict := make(map[string]ItemDef)
	for _, item := range itemsList {
		newDict[item.ID] = item
	}

	itemDictMu.Lock()
	ItemDictionary = newDict
	itemDictMu.Unlock()

	slog.Info("Successfully loaded and parsed item definitions", "count", len(newDict), "path", filePath)
	return nil
}

// SetupItemHotReload monitora o arquivo JSON em segundo plano para hot-reload sem reiniciar o servidor (PATCH 5)
func SetupItemHotReload(filePath string) {
	go func() {
		lastMod := time.Time{}
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			info, err := os.Stat(filePath)
			if err != nil {
				continue
			}
			if info.ModTime().After(lastMod) {
				if !lastMod.IsZero() {
					slog.Info("Hot-reloading item definitions due to file change", "path", filePath)
				}
				if err := LoadItemDefinitions(filePath); err == nil {
					lastMod = info.ModTime()
				} else {
					slog.Error("Failed to hot-reload item definitions", "error", err)
				}
			}
		}
	}()
}

// InventoryItem representa um item instanciado no inventário do jogador
type InventoryItem struct {
	ItemID     string `json:"item_id"`
	Quantity   int    `json:"quantity"`
	Durability int    `json:"durability"`
	SlotIndex  int    `json:"slot_index"`
	ItemUUID   string `json:"item_uuid"` // Identificador único (anti-duplicação - PATCH 4)
}

// PlayerInventory armazena os itens de um jogador com exclusão mútua thread-safe
type PlayerInventory struct {
	mu        sync.RWMutex
	PlayerID  string
	Items     map[int]*InventoryItem
	BaseStats combat.EntityStats // Atributos base do jogador sem equipamentos
	isDirty   bool               // Flag de alteração de estado (PATCH 2)
	Version   int                // Controle de versão para optimistic locking (PATCH 4)
	Gold      int64              // Gold do jogador (PATCH 1)
}

// PlayerSnapshot representa o snapshot imutável para evitar partial-state races (PATCH 1 & 4)
type PlayerSnapshot struct {
	PlayerID string
	Stats    combat.EntityStats
	Items    map[int]InventoryItem
	PosX     float64
	PosY     float64
	PosZ     float64
	Version  int
	IsDirty  bool
	Gold     int64
}

// NewPlayerInventory inicializa um inventário com itens padrão para demonstração
func NewPlayerInventory(playerID string) *PlayerInventory {
	pi := &PlayerInventory{
		PlayerID: playerID,
		Items:    make(map[int]*InventoryItem),
		BaseStats: combat.EntityStats{
			ID:                 playerID,
			Name:               playerID + " (Paladin)",
			IsPlayer:           true,
			Faction:            "Alliance",
			Level:              45,
			BaseAttack:         45.0,
			WeaponDamage:       15.0, // Bare hands fallback
			Defense:            30.0,
			Resistance:         15.0,
			Accuracy:           95.0,
			Evasion:            10.0,
			CritChance:         0.10,
			CritMultiplier:     1.50,
			ArmorPenetration:   0.15,
			Element:            "Light",
			ElementAttackBonus: 0.10,
			ElementDefBonus:    0.05,
			Health:             600.0,
			MaxHealth:          600.0,
			Mana:               100.0,
			MaxMana:            100.0,
		},
		isDirty: false,
		Version: 1,
		Gold:    1000, // Gold inicial padrão (PATCH 1)
	}

	// Adiciona itens iniciais padrões para demonstração com UUIDs únicos (anti-duplicação - PATCH 4)
	pi.Items[0] = &InventoryItem{ItemID: "sword_basic", Quantity: 1, Durability: 100, SlotIndex: 0, ItemUUID: GenerateUUIDv4()}
	pi.Items[1] = &InventoryItem{ItemID: "sword_excalibur", Quantity: 1, Durability: 100, SlotIndex: 1, ItemUUID: GenerateUUIDv4()}
	pi.Items[2] = &InventoryItem{ItemID: "armor_plate", Quantity: 1, Durability: 100, SlotIndex: 2, ItemUUID: GenerateUUIDv4()}
	pi.Items[3] = &InventoryItem{ItemID: "ring_crit", Quantity: 1, Durability: 100, SlotIndex: 3, ItemUUID: GenerateUUIDv4()}
	pi.Items[4] = &InventoryItem{ItemID: "potion_heal", Quantity: 5, Durability: 100, SlotIndex: 4, ItemUUID: ""} // Stackable item needs no unique UUID
	pi.Items[5] = &InventoryItem{ItemID: "iron_ore", Quantity: 10, Durability: 100, SlotIndex: 5, ItemUUID: ""}

	return pi
}

// GetItems retorna uma cópia dos itens do inventário para sincronização
func (pi *PlayerInventory) GetItems() map[int]InventoryItem {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	copied := make(map[int]InventoryItem)
	for slot, item := range pi.Items {
		if item != nil {
			copied[slot] = *item
		}
	}
	return copied
}

// SetItems define os itens carregados do banco de dados
func (pi *PlayerInventory) SetItems(items map[int]*InventoryItem) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.Items = items
}

// Lock adquire o lock de exclusão mútua do inventário
func (pi *PlayerInventory) Lock() {
	pi.mu.Lock()
}

// Unlock libera o lock de exclusão mútua do inventário
func (pi *PlayerInventory) Unlock() {
	pi.mu.Unlock()
}

// RLock adquire o lock de leitura do inventário
func (pi *PlayerInventory) RLock() {
	pi.mu.RLock()
}

// RUnlock libera o lock de leitura do inventário
func (pi *PlayerInventory) RUnlock() {
	pi.mu.RUnlock()
}

// Métodos de Controle Dirty e Version (PATCH 2 & 4)
func (pi *PlayerInventory) IsDirty() bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.isDirty
}

func (pi *PlayerInventory) SetDirty(dirty bool) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.isDirty = dirty
}

func (pi *PlayerInventory) GetVersion() int {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.Version
}

func (pi *PlayerInventory) SetVersion(version int) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.Version = version
}

// CreateSnapshot tira um snapshot imutável de forma atômica para persistência (PATCH 1)
func (pi *PlayerInventory) CreateSnapshot(stats combat.EntityStats, posX, posY, posZ float64) *PlayerSnapshot {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	itemsCopy := make(map[int]InventoryItem)
	for slot, item := range pi.Items {
		if item != nil {
			itemsCopy[slot] = *item
		}
	}

	return &PlayerSnapshot{
		PlayerID: pi.PlayerID,
		Stats:    stats, // Passada por valor (já é cópia imutável)
		Items:    itemsCopy,
		PosX:     posX,
		PosY:     posY,
		PosZ:     posZ,
		Version:  pi.Version,
		IsDirty:  pi.isDirty,
		Gold:     pi.Gold,
	}
}

// Métodos Thread-Safe de Gestão de Gold e Itens (PATCH 1 & 2)
func (pi *PlayerInventory) GetGold() int64 {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.Gold
}

func (pi *PlayerInventory) SetGold(gold int64) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	if gold < 0 {
		gold = 0
	}
	pi.Gold = gold
	pi.isDirty = true
}

func (pi *PlayerInventory) AddGold(amount int64) bool {
	if amount < 0 {
		return false
	}
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.Gold += amount
	pi.isDirty = true
	return true
}

func (pi *PlayerInventory) RemoveGold(amount int64) bool {
	if amount < 0 {
		return false
	}
	pi.mu.Lock()
	defer pi.mu.Unlock()
	if pi.Gold < amount {
		return false
	}
	pi.Gold -= amount
	pi.isDirty = true
	return true
}

func (pi *PlayerInventory) HasItem(itemID string, qty int) bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	accum := 0
	for _, item := range pi.Items {
		if item != nil && item.ItemID == itemID {
			accum += item.Quantity
		}
	}
	return accum >= qty
}

func (pi *PlayerInventory) RemoveItemBySlot(slotIndex int, qty int) bool {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	item, exists := pi.Items[slotIndex]
	if !exists || item == nil {
		return false
	}
	if item.Quantity < qty {
		return false
	}
	item.Quantity -= qty
	if item.Quantity == 0 {
		delete(pi.Items, slotIndex)
	}
	pi.isDirty = true
	return true
}

func (pi *PlayerInventory) RemoveItemByID(itemID string, qty int) bool {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	// Verifica primeiro se tem a quantidade
	accum := 0
	for _, item := range pi.Items {
		if item != nil && item.ItemID == itemID {
			accum += item.Quantity
		}
	}
	if accum < qty {
		return false
	}

	// Remove incrementalmente
	toRemove := qty
	for slot, item := range pi.Items {
		if item != nil && item.ItemID == itemID {
			if item.Quantity <= toRemove {
				toRemove -= item.Quantity
				delete(pi.Items, slot)
			} else {
				item.Quantity -= toRemove
				toRemove = 0
			}
			if toRemove == 0 {
				break
			}
		}
	}
	pi.isDirty = true
	return true
}

// RecalculateStats reconstrói os atributos dinâmicos do personagem somando os bônus dos equipamentos equipados
func (pi *PlayerInventory) RecalculateStats(currentStats *combat.EntityStats) {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	// Preserva os atributos mutáveis de tempo real (vida/mana atuais e time de combate)
	currentHp := currentStats.Health
	currentMana := currentStats.Mana
	lastCombat := currentStats.LastCombatTime

	// Restaura atributos para a base
	*currentStats = pi.BaseStats

	itemDictMu.RLock()
	defer itemDictMu.RUnlock()

	// Aplica bônus do slot de Arma
	if weaponItem, ok := pi.Items[SlotWeapon]; ok && weaponItem != nil {
		if def, exists := ItemDictionary[weaponItem.ItemID]; exists {
			currentStats.WeaponDamage = def.BaseDamage
		}
	} else {
		currentStats.WeaponDamage = 15.0 // Dano desarmado padrão
	}

	// Aplica bônus do slot de Armadura
	if armorItem, ok := pi.Items[SlotArmor]; ok && armorItem != nil {
		if def, exists := ItemDictionary[armorItem.ItemID]; exists {
			currentStats.Defense = pi.BaseStats.Defense + def.BaseDef
		}
	}

	// Aplica bônus do slot de Acessório
	if accItem, ok := pi.Items[SlotAccessory]; ok && accItem != nil {
		if def, exists := ItemDictionary[accItem.ItemID]; exists {
			currentStats.Resistance = pi.BaseStats.Resistance + def.BaseRes
			currentStats.CritChance = pi.BaseStats.CritChance + def.CritBonus
		}
	}

	// Restaura vida e mana atuais respeitando os novos limites máximos recalculados
	if currentHp > currentStats.MaxHealth {
		currentStats.Health = currentStats.MaxHealth
	} else {
		currentStats.Health = currentHp
	}

	if currentMana > currentStats.MaxMana {
		currentStats.Mana = currentStats.MaxMana
	} else {
		currentStats.Mana = currentMana
	}

	currentStats.LastCombatTime = lastCombat

	slog.Info("Stats autoritativos recalculados", 
		"player", pi.PlayerID, 
		"atk", currentStats.BaseAttack, 
		"weaponDmg", currentStats.WeaponDamage, 
		"def", currentStats.Defense, 
		"res", currentStats.Resistance, 
		"crit", currentStats.CritChance)
}

// EquipItem move um item da mochila (backpack) para um slot de equipamento, realizando todas as validações autoritativas
func (pi *PlayerInventory) EquipItem(fromSlot, toSlot int, currentStats *combat.EntityStats) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	// 1. Validação de limites de slots
	if fromSlot < SlotMinBackpack || fromSlot > SlotMaxBackpack {
		return fmt.Errorf("slot de origem inválido: %d", fromSlot)
	}
	if toSlot != SlotWeapon && toSlot != SlotArmor && toSlot != SlotAccessory {
		return fmt.Errorf("slot de equipamento de destino inválido: %d", toSlot)
	}

	// 2. Verifica se o item existe no slot de origem
	item, ok := pi.Items[fromSlot]
	if !ok || item == nil {
		return errors.New("não há nenhum item no slot de origem")
	}

	// 3. Verifica se a definição do item existe (Thread-Safe)
	itemDictMu.RLock()
	def, exists := ItemDictionary[item.ItemID]
	itemDictMu.RUnlock()
	if !exists {
		return fmt.Errorf("definição de item não encontrada: %s", item.ItemID)
	}

	// 4. Validação autoritativa de tipo de item correspondente ao slot
	switch toSlot {
	case SlotWeapon:
		if def.Type != "weapon" {
			return fmt.Errorf("item %s não é uma arma e não pode ser equipado no slot de arma", def.Name)
		}
	case SlotArmor:
		if def.Type != "armor" {
			return fmt.Errorf("item %s não é uma armadura e não pode ser equipado no slot de armadura", def.Name)
		}
	case SlotAccessory:
		if def.Type != "accessory" {
			return fmt.Errorf("item %s não é um acessório e não pode ser equipado no slot de acessórios", def.Name)
		}
	}

	// 4.5 Restrições de Equipamento da Classe Aprendiz (Novice Phase - Sprint 3 Task 5)
	if currentStats.Class == "Novice" || currentStats.Class == "" {
		if def.Tier > 1 {
			return fmt.Errorf("personagens Aprendizes (Novice) não podem equipar itens de Tier %d (limite de Tier 1)", def.Tier)
		}
	}

	// 5. Permuta itens (Equipa o item e envia o antigo de volta para o mesmo slot da mochila)
	oldEquip := pi.Items[toSlot]

	// Move para o equipamento
	item.SlotIndex = toSlot
	pi.Items[toSlot] = item

	// Se tinha algo equipado, move de volta para a mochila
	if oldEquip != nil {
		oldEquip.SlotIndex = fromSlot
		pi.Items[fromSlot] = oldEquip
	} else {
		delete(pi.Items, fromSlot)
	}

	// Marca o estado do inventário como alterado (PATCH 2)
	pi.isDirty = true

	// Recalcula atributos do personagem imediatamente
	pi.mu.Unlock()
	pi.RecalculateStats(currentStats)
	pi.mu.Lock()

	slog.Info("Item equipado com sucesso", "player", pi.PlayerID, "item", def.Name, "fromSlot", fromSlot, "toSlot", toSlot)
	return nil
}

// UnequipItem move um item equipado de volta para a mochila (backpack) no primeiro slot livre
func (pi *PlayerInventory) UnequipItem(fromSlot int, currentStats *combat.EntityStats) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	// 1. Validação de slot de equipamento
	if fromSlot != SlotWeapon && fromSlot != SlotArmor && fromSlot != SlotAccessory {
		return fmt.Errorf("slot de equipamento inválido para desequipar: %d", fromSlot)
	}

	// 2. Verifica se o item está equipado
	item, ok := pi.Items[fromSlot]
	if !ok || item == nil {
		return errors.New("não há nenhum item equipado neste slot")
	}

	// 3. Encontra o primeiro slot livre na mochila (0 a 29)
	freeSlot := -1
	for slot := SlotMinBackpack; slot <= SlotMaxBackpack; slot++ {
		if _, occupied := pi.Items[slot]; !occupied {
			freeSlot = slot
			break
		}
	}

	if freeSlot == -1 {
		return errors.New("mochila cheia! Libere espaço antes de desequipar")
	}

	// 4. Transfere o item de volta para o slot livre
	delete(pi.Items, fromSlot)
	item.SlotIndex = freeSlot
	pi.Items[freeSlot] = item

	// Marca o estado do inventário como alterado (PATCH 2)
	pi.isDirty = true

	// Recalcula atributos do personagem imediatamente
	pi.mu.Unlock()
	pi.RecalculateStats(currentStats)
	pi.mu.Lock()

	slog.Info("Item desequipado com sucesso", "player", pi.PlayerID, "slotEquip", fromSlot, "freeSlot", freeSlot)
	return nil
}

// SwapSlots move ou permuta itens entre dois slots genéricos (usado para reorganizar mochila ou mover de mochilas)
func (pi *PlayerInventory) SwapSlots(slotA, slotB int) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	if slotA < 0 || slotA >= TotalSlots || slotB < 0 || slotB >= TotalSlots {
		return fmt.Errorf("índices de slot inválidos: %d, %d", slotA, slotB)
	}

	// Não permite mover itens diretamente para slots de equipamentos sem passar pela validação de tipo adequada (use EquipItem)
	isEquipA := slotA >= SlotWeapon
	isEquipB := slotB >= SlotWeapon

	if isEquipA || isEquipB {
		return errors.New("use a função de Equipar para interagir com slots de equipamentos")
	}

	itemA := pi.Items[slotA]
	itemB := pi.Items[slotB]

	// Se ambos forem nulos, nada a fazer
	if itemA == nil && itemB == nil {
		return nil
	}

	// 1. Lógica de agrupamento de itens stackable iguais
	if itemA != nil && itemB != nil && itemA.ItemID == itemB.ItemID {
		itemDictMu.RLock()
		def, exists := ItemDictionary[itemA.ItemID]
		itemDictMu.RUnlock()
		if exists && def.Stackable {
			totalQty := itemA.Quantity + itemB.Quantity
			if totalQty <= def.MaxStack {
				itemB.Quantity = totalQty
				delete(pi.Items, slotA)
				pi.isDirty = true
				return nil
			} else {
				itemB.Quantity = def.MaxStack
				itemA.Quantity = totalQty - def.MaxStack
				pi.isDirty = true
				return nil
			}
		}
	}

	// 2. Permuta de itens simples
	pi.Items[slotA] = itemB
	if itemB != nil {
		itemB.SlotIndex = slotA
	}

	pi.Items[slotB] = itemA
	if itemA != nil {
		itemA.SlotIndex = slotB
	}

	// Marca o estado do inventário como alterado (PATCH 2)
	pi.isDirty = true

	return nil
}

// AddItem adiciona itens no inventário do jogador respeitando capacidade e empilhamento
func (pi *PlayerInventory) AddItem(itemID string, qty int) bool {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	itemDictMu.RLock()
	def, exists := ItemDictionary[itemID]
	itemDictMu.RUnlock()

	if !exists {
		return false
	}

	// 1. Tenta empilhar se o item for empilhável (Stackable)
	if def.Stackable {
		for _, item := range pi.Items {
			if item != nil && item.ItemID == itemID && item.Quantity < def.MaxStack {
				spaceLeft := def.MaxStack - item.Quantity
				if qty <= spaceLeft {
					item.Quantity += qty
					pi.isDirty = true
					return true
				} else {
					item.Quantity = def.MaxStack
					qty -= spaceLeft
				}
			}
		}
	}

	// 2. Se sobrar quantidade ou não for empilhável, coloca em slots livres na mochila (0 a 29)
	for slot := SlotMinBackpack; slot <= SlotMaxBackpack; slot++ {
		if _, occupied := pi.Items[slot]; !occupied {
			currentQty := qty
			if def.Stackable && currentQty > def.MaxStack {
				currentQty = def.MaxStack
			}

			pi.Items[slot] = &InventoryItem{
				ItemID:     itemID,
				Quantity:   currentQty,
				Durability: 100,
				SlotIndex:  slot,
			}
			qty -= currentQty
			pi.isDirty = true

			if qty <= 0 {
				return true
			}
		}
	}

	return qty <= 0
}

func init() {
	// Definições de fallback estáticas
	fallbackDict := map[string]ItemDef{
		"sword_basic": {
			ID:         "sword_basic",
			Name:       "Espada Básica",
			Type:       "weapon",
			Stackable:  false,
			MaxStack:   1,
			BaseDamage: 15.0,
		},
		"sword_excalibur": {
			ID:         "sword_excalibur",
			Name:       "Excalibur",
			Type:       "weapon",
			Stackable:  false,
			MaxStack:   1,
			BaseDamage: 60.0,
		},
		"bow_hunter": {
			ID:         "bow_hunter",
			Name:       "Arco de Caçador",
			Type:       "weapon",
			Stackable:  false,
			MaxStack:   1,
			BaseDamage: 25.0,
		},
		"armor_leather": {
			ID:        "armor_leather",
			Name:      "Armadura de Couro",
			Type:      "armor",
			Stackable: false,
			MaxStack:  1,
			BaseDef:   15.0,
		},
		"armor_plate": {
			ID:        "armor_plate",
			Name:      "Armadura de Placas",
			Type:      "armor",
			Stackable: false,
			MaxStack:  1,
			BaseDef:   45.0,
		},
		"ring_crit": {
			ID:        "ring_crit",
			Name:      "Anel do Destruidor",
			Type:      "accessory",
			Stackable: false,
			MaxStack:  1,
			BaseRes:   5.0,
			CritBonus: 0.12,
		},
		"potion_heal": {
			ID:        "potion_heal",
			Name:      "Poção de Cura",
			Type:      "consumable",
			Stackable: true,
			MaxStack:  99,
		},
		"iron_ore": {
			ID:        "iron_ore",
			Name:      "Minério de Ferro",
			Type:      "material",
			Stackable: true,
			MaxStack:  99,
		},
	}

	itemDictMu.Lock()
	ItemDictionary = fallbackDict
	itemDictMu.Unlock()

	paths := []string{"backend/config/items.json", "config/items.json", "../config/items.json"}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			if err := LoadItemDefinitions(p); err == nil {
				SetupItemHotReload(p)
				break
			}
		}
	}
}

// GetItemDef retorna de forma thread-safe a definição de um item (Sprint 4 Task 1)
func GetItemDef(itemID string) (ItemDef, bool) {
	itemDictMu.RLock()
	defer itemDictMu.RUnlock()
	def, exists := ItemDictionary[itemID]
	return def, exists
}

// IsAoLEquipped checks if the player has an Amulet of Loss equipped in slots 0-3 or accessory slot
func (pi *PlayerInventory) IsAoLEquipped() bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	for slot := 0; slot < TotalSlots; slot++ {
		if item, exists := pi.Items[slot]; exists && item != nil {
			if item.ItemID == "amulet_of_loss" {
				return true
			}
		}
	}
	return false
}

// ConsumeAoL removes one Amulet of Loss from the player's inventory or equipped slots
func (pi *PlayerInventory) ConsumeAoL() bool {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	for slot, item := range pi.Items {
		if item != nil && item.ItemID == "amulet_of_loss" {
			delete(pi.Items, slot)
			pi.isDirty = true
			return true
		}
	}
	return false
}

