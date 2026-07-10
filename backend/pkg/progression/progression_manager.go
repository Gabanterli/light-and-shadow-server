package progression

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/gamedata/rules"
	"github.com/light-and-shadow/backend/pkg/inventory"
	"github.com/light-and-shadow/backend/pkg/messaging"
)

// ProgressionManager coordena a evolução, afinidade elemental, vocações e subclasses dos jogadores de forma thread-safe
type ProgressionManager struct {
	mu            sync.RWMutex
	dbConn        *sql.DB
	combatManager *combat.CombatManager
	inventories   map[string]*inventory.PlayerInventory

	// Anti-spam para acúmulo de afinidade (PATCH 4)
	lastAffinityGain map[string]time.Time // playerID -> last gain timestamp
	spamGuardMu      sync.Mutex
}

// NewProgressionManager instancia o ProgressionManager
func NewProgressionManager(db *sql.DB, cm *combat.CombatManager, invs map[string]*inventory.PlayerInventory) *ProgressionManager {
	pm := &ProgressionManager{
		dbConn:           db,
		combatManager:    cm,
		inventories:      invs,
		lastAffinityGain: make(map[string]time.Time),
	}
	pm.startEventSubscription()
	return pm
}

// ChooseVocation realiza a seleção irreversível de classe a partir do Level 10 (PATCH 2)
func (pm *ProgressionManager) ChooseVocation(playerID string, baseClass string) error {
	pStats, existsStats := pm.combatManager.GetEntityStats(playerID)
	playerInv, existsInv := pm.inventories[playerID]
	if !existsStats || !existsInv {
		return fmt.Errorf("jogador %s não está totalmente carregado no servidor", playerID)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	targetClass := rules.RuleID(strings.ToLower(strings.TrimSpace(baseClass)))

	// A41-B3: Persist class change atomically before mutating memory.
	if pm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tx, err := pm.dbConn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			return fmt.Errorf("failed to begin vocation transaction: %w", err)
		}
		defer tx.Rollback()

		var dbClass sql.NullString
		var dbLevel int
		err = tx.QueryRowContext(ctx, `SELECT class, level FROM characters WHERE name = $1 FOR UPDATE`, playerID).Scan(&dbClass, &dbLevel)
		if err != nil {
			return fmt.Errorf("failed to fetch and lock character for vocation change: %w", err)
		}

		var dbCurrentClass rules.RuleID
		switch strings.ToLower(strings.TrimSpace(dbClass.String)) {
		case "", "novice":
			dbCurrentClass = rules.StartingClassNovice
		case "knight":
			dbCurrentClass = rules.ClassKnight
		case "mage":
			dbCurrentClass = rules.ClassMage
		case "archer":
			dbCurrentClass = rules.ClassArcher
		case "assassin":
			dbCurrentClass = rules.ClassAssassin
		case "cleric":
			dbCurrentClass = rules.ClassCleric
		default:
			dbCurrentClass = rules.RuleID(strings.ToLower(strings.TrimSpace(dbClass.String)))
		}

		if err := rules.CanSelectBaseClass(uint32(dbLevel), dbCurrentClass, targetClass); err != nil {
			return fmt.Errorf("seleção de vocação rejeitada: %w", err)
		}

		res, err := tx.ExecContext(ctx, `
			UPDATE characters
			SET class = $1, updated_at = CURRENT_TIMESTAMP
			WHERE name = $2 AND (class = 'Novice' OR class = 'novice' OR class = '')
		`, string(targetClass), playerID)
		if err != nil {
			return fmt.Errorf("failed to persist vocation change: %w", err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to verify vocation update result: %w", err)
		}
		if rowsAffected != 1 {
			return fmt.Errorf("vocation already chosen or race condition detected (rows affected: %d)", rowsAffected)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit vocation transaction: %w", err)
		}

		slog.Info("Canonical vocation persisted atomically", "player", playerID, "class", string(targetClass), "source_class", dbCurrentClass, "level", dbLevel)
	} else {
		// Fallback for development/testing without a database.
		var currentClass rules.RuleID
		switch strings.ToLower(strings.TrimSpace(pStats.Class)) {
		case "", "novice":
			currentClass = rules.StartingClassNovice
		case "knight":
			currentClass = rules.ClassKnight
		case "mage":
			currentClass = rules.ClassMage
		case "archer":
			currentClass = rules.ClassArcher
		case "assassin":
			currentClass = rules.ClassAssassin
		case "cleric":
			currentClass = rules.ClassCleric
		default:
			currentClass = rules.RuleID(strings.ToLower(strings.TrimSpace(pStats.Class)))
		}

		if err := rules.CanSelectBaseClass(uint32(pStats.Level), currentClass, targetClass); err != nil {
			return fmt.Errorf("seleção de vocação rejeitada: %w", err)
		}
	}

	// Apply changes to in-memory state only after successful validation (and DB commit if applicable).
	pStats.Class = string(targetClass)
	playerInv.BaseStats.Class = string(targetClass)

	// Aplica bônus de vocação passivos de forma permanente na BaseStats (recalculados depois)
	switch targetClass {
	case "knight":
		playerInv.BaseStats.MaxHealth += 100.0 // +100 MaxHP
		playerInv.BaseStats.Defense += 15.0    // +15 Defesa
	case "mage":
		playerInv.BaseStats.MaxMana += 80.0            // +80 MaxMP
		playerInv.BaseStats.ElementAttackBonus += 0.10 // +10% Bônus de Ataque Elemental
	case "archer":
		playerInv.BaseStats.CritChance += 0.05 // +5% CritChance
		playerInv.BaseStats.Accuracy += 10.0   // +10 Precisão
	case "assassin":
		playerInv.BaseStats.Evasion += 10.0    // +10 Evasão
		playerInv.BaseStats.CritChance += 0.05 // +5% CritChance
	case "cleric":
		playerInv.BaseStats.MaxHealth += 50.0       // +50 MaxHP
		playerInv.BaseStats.Resistance += 10.0      // +10% Resistência
		playerInv.BaseStats.ElementDefBonus += 0.05 // +5% Mitigação Elemental
	}

	// Sincroniza vida/mana atuais com os novos limites
	pStats.MaxHealth = playerInv.BaseStats.MaxHealth
	pStats.Health = pStats.MaxHealth
	pStats.MaxMana = playerInv.BaseStats.MaxMana
	pStats.Mana = pStats.MaxMana
	pStats.Defense = playerInv.BaseStats.Defense
	pStats.Resistance = playerInv.BaseStats.Resistance
	pStats.CritChance = playerInv.BaseStats.CritChance
	pStats.Accuracy = playerInv.BaseStats.Accuracy
	pStats.Evasion = playerInv.BaseStats.Evasion
	pStats.ElementAttackBonus = playerInv.BaseStats.ElementAttackBonus
	pStats.ElementDefBonus = playerInv.BaseStats.ElementDefBonus

	// Sincroniza estado para o banco marcar como modificado (PATCH 5)
	pStats.ProgressionDirty = true
	playerInv.SetDirty(true)

	slog.Info("Vocação escolhida e atributos de bônus aplicados com sucesso", "player", playerID, "vocation", string(targetClass))

	// Dispara evento de vocação desbloqueada
	messaging.GetInstance().Publish("vocation_unlocked", map[string]interface{}{
		"player_id": playerID,
		"class":     string(targetClass),
	})

	return nil
}

// AddAffinity incrementa os scores ocultos de afinidade elemental com proteção anti-spam (PATCH 4, 5)
func (pm *ProgressionManager) AddAffinity(playerID string, element string, points int, source string) (bool, error) {
	pStats, exists := pm.combatManager.GetEntityStats(playerID)
	if !exists {
		return false, errors.New("jogador não encontrado nos registros do CombatManager")
	}

	// 1. Proteção de spam de afinidade baseada em tempo (PATCH 4)
	pm.spamGuardMu.Lock()
	lastGain, ok := pm.lastAffinityGain[playerID]
	now := time.Now()
	if ok && now.Sub(lastGain) < 1500*time.Millisecond {
		pm.spamGuardMu.Unlock()
		slog.Warn("Acúmulo de afinidade elemental bloqueado por proteção anti-spam (PATCH 4)", "player", playerID, "source", source)
		return false, nil // Bloqueado silenciosamente ou com flag
	}
	pm.lastAffinityGain[playerID] = now
	pm.spamGuardMu.Unlock()

	// 2. Atualiza scores ocultos em memória com thread safety
	pm.mu.Lock()
	switch strings.ToLower(element) {
	case "fire":
		pStats.AffinityFire += points
	case "ice":
		pStats.AffinityIce += points
	case "holy":
		pStats.AffinityHoly += points
	case "shadow":
		pStats.AffinityShadow += points
	case "nature":
		pStats.AffinityNature += points
	default:
		pm.mu.Unlock()
		return false, fmt.Errorf("elemento desconhecido para afinidade: %s", element)
	}

	// Ativa a flag dirty de progresso (PATCH 5)
	pStats.ProgressionDirty = true
	if playerInv, existsInv := pm.inventories[playerID]; existsInv {
		playerInv.SetDirty(true)
	}
	pm.mu.Unlock()

	slog.Info("Afinidade elemental acumulada com sucesso", "player", playerID, "element", element, "points", points, "source", source)
	return true, nil
}

// TriggerSubclassUnlock executa a transação atômica de desbloqueio de subclasse (PATCH 3)
func (pm *ProgressionManager) TriggerSubclassUnlock(playerID string) (string, error) {
	pStats, existsStats := pm.combatManager.GetEntityStats(playerID)
	playerInv, existsInv := pm.inventories[playerID]
	if !existsStats || !existsInv {
		return "", fmt.Errorf("jogador %s não carregado completamente", playerID)
	}

	// 1. Validação de Nível Mínimo para Subclasse
	if pStats.Level < 100 {
		return "", fmt.Errorf("nível insuficiente para desbloquear subclasse (atual: %d, necessário: 100)", pStats.Level)
	}

	// 2. Executa Transação Atômica PostgreSQL com Repeatable Read (PATCH 3)
	if pm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tx, err := pm.dbConn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			return "", fmt.Errorf("failed to begin subclass transaction: %w", err)
		}
		defer tx.Rollback()

		var class, subclass string
		var level int
		var fire, ice, holy, shadow, nature int

		// Lock FOR UPDATE para evitar race condition
		err = tx.QueryRowContext(ctx, `
			SELECT class, subclass, level, affinity_fire, affinity_ice, affinity_holy, affinity_shadow, affinity_nature
			FROM characters WHERE name = $1 FOR UPDATE
		`, playerID).Scan(&class, &subclass, &level, &fire, &ice, &holy, &shadow, &nature)

		if err != nil {
			return "", fmt.Errorf("failed to fetch and lock character state: %w", err)
		}

		if class == "Novice" || class == "" {
			return "", errors.New("é necessário possuir uma vocação base antes de obter uma subclasse")
		}

		if subclass != "" {
			return "", fmt.Errorf("subclasse já desbloqueada anteriormente: %s", subclass)
		}

		// 3. Cálculo de Dominante Elemental
		dominantElement := "Holy"
		maxScore := holy

		if fire > maxScore {
			dominantElement = "Fire"
			maxScore = fire
		}
		if ice > maxScore {
			dominantElement = "Ice"
			maxScore = ice
		}
		if shadow > maxScore {
			dominantElement = "Shadow"
			maxScore = shadow
		}
		if nature > maxScore {
			dominantElement = "Nature"
			maxScore = nature
		}

		// Constrói o nome da subclasse (Subclass = Elemento + Vocação)
		newSubclass := dominantElement + " " + class

		// Persiste atômicamente as colunas no banco
		_, err = tx.ExecContext(ctx, `
			UPDATE characters 
			SET subclass = $1, element = $2, element_attack_bonus = element_attack_bonus + 0.15, element_def_bonus = element_def_bonus + 0.10
			WHERE name = $3
		`, newSubclass, dominantElement, playerID)
		if err != nil {
			return "", fmt.Errorf("failed to persist subclass: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return "", fmt.Errorf("failed to commit subclass transaction: %w", err)
		}

		// 4. Aplica os bônus e atualiza o estado em memória local de forma thread-safe
		pm.mu.Lock()
		pStats.Subclass = newSubclass
		pStats.Element = dominantElement
		pStats.ElementAttackBonus += 0.15
		pStats.ElementDefBonus += 0.10
		pStats.ProgressionDirty = false // Limpa dirty flag pois já foi gravado

		playerInv.BaseStats.Subclass = newSubclass
		playerInv.BaseStats.Element = dominantElement
		playerInv.BaseStats.ElementAttackBonus += 0.15
		playerInv.BaseStats.ElementDefBonus += 0.10
		pm.mu.Unlock()

		slog.Info("Subclasse desbloqueada transacionalmente", "player", playerID, "subclass", newSubclass, "element", dominantElement)
		return newSubclass, nil
	}

	// Fallback em memória caso banco esteja em fallback
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pStats.Class == "Novice" || pStats.Class == "" {
		return "", errors.New("é necessário possuir uma vocação base antes de obter uma subclasse")
	}
	if pStats.Subclass != "" {
		return "", fmt.Errorf("subclasse já desbloqueada anteriormente: %s", pStats.Subclass)
	}

	dominantElement := "Holy"
	maxScore := pStats.AffinityHoly

	if pStats.AffinityFire > maxScore {
		dominantElement = "Fire"
		maxScore = pStats.AffinityFire
	}
	if pStats.AffinityIce > maxScore {
		dominantElement = "Ice"
		maxScore = pStats.AffinityIce
	}
	if pStats.AffinityShadow > maxScore {
		dominantElement = "Shadow"
		maxScore = pStats.AffinityShadow
	}
	if pStats.AffinityNature > maxScore {
		dominantElement = "Nature"
		maxScore = pStats.AffinityNature
	}

	newSubclass := dominantElement + " " + pStats.Class
	pStats.Subclass = newSubclass
	pStats.Element = dominantElement
	pStats.ElementAttackBonus += 0.15
	pStats.ElementDefBonus += 0.10

	playerInv.BaseStats.Subclass = newSubclass
	playerInv.BaseStats.Element = dominantElement
	playerInv.BaseStats.ElementAttackBonus += 0.15
	playerInv.BaseStats.ElementDefBonus += 0.10

	slog.Info("Subclasse desbloqueada (Fallback Modo Memória)", "player", playerID, "subclass", newSubclass)
	return newSubclass, nil
}

// startEventSubscription assina tópicos para ganhar pontos de afinidade elemental através de múltiplas ações (Sprint 3 Task 5)
func (pm *ProgressionManager) startEventSubscription() {
	mb := messaging.GetInstance()
	chKilled := mb.Subscribe("monster_killed")
	chSkill := mb.Subscribe("skill_cast")
	chQuest := mb.Subscribe("quest_completed")

	go func() {
		for {
			select {
			case msg, ok := <-chKilled:
				if !ok {
					return
				}
				// Ganhos de afinidade matando monstros
				if payload, ok := msg.(messaging.MonsterKilledPayload); ok {
					// Ex: Matar goblin dá afinidade de Fogo, matar monstro de sombra dá afinidade Sombria, etc.
					element := "nature"
					if strings.Contains(strings.ToLower(payload.MonsterID), "goblin") {
						element = "fire"
					} else if strings.Contains(strings.ToLower(payload.MonsterID), "shadow") {
						element = "shadow"
					} else if strings.Contains(strings.ToLower(payload.MonsterID), "dragon") {
						element = "fire"
					} else if strings.Contains(strings.ToLower(payload.MonsterID), "ice") {
						element = "ice"
					}
					pm.AddAffinity(payload.PlayerID, element, 1, "Monster Kill ("+payload.MonsterID+")")
				}

			case msg, ok := <-chSkill:
				if !ok {
					return
				}
				// Ganhos de afinidade usando habilidades
				if payload, ok := msg.(map[string]interface{}); ok {
					playerID, _ := payload["player_id"].(string)
					skillID, _ := payload["skill_id"].(int)
					element := "nature"
					switch skillID {
					case 1: // Slash
						element = "holy"
					case 2: // Fireball
						element = "fire"
					case 3: // Spear Thrust
						element = "ice"
					case 4: // Arrow Rain
						element = "nature"
					}
					pm.AddAffinity(playerID, element, 1, "Skill Used")
				}

			case msg, ok := <-chQuest:
				if !ok {
					return
				}
				// Ganhos de afinidade ao completar quests
				if payload, ok := msg.(map[string]interface{}); ok {
					playerID, _ := payload["player_id"].(string)
					questID, _ := payload["quest_id"].(string)
					element := "holy"
					if strings.Contains(strings.ToLower(questID), "goblins") {
						element = "fire"
					} else if strings.Contains(strings.ToLower(questID), "iron") {
						element = "nature"
					}
					pm.AddAffinity(playerID, element, 3, "Quest Completed")
				}
			}
		}
	}()
}
