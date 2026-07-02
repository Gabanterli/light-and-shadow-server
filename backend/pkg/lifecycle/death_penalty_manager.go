package lifecycle

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/light-and-shadow/backend/pkg/blessing"
	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/inventory"
	"github.com/light-and-shadow/backend/pkg/professions"
	"github.com/light-and-shadow/backend/pkg/pve"
	"github.com/light-and-shadow/backend/pkg/housing"
	"github.com/light-and-shadow/backend/pkg/pvp"
)

// DeathPenaltyConfig armazena as taxas e percentuais de perdas canônicas sob morte não abençoada
type DeathPenaltyConfig struct {
	XpLossPercentage     float64 // Ex: 0.10 (10% de perda de XP)
	SkillLossPercentage  float64 // Ex: 0.05 (5% de perda de XP da profissão/skill)
	GoldLossPercentage   float64 // Ex: 0.10 (10% de perda do ouro do inventário)
	EquippedDropChance   float64 // Ex: 0.05 (5% de chance de dropar um item equipado)
	InventoryDropChance  float64 // Ex: 0.20 (20% de chance de dropar materiais ou consumíveis)
}

// DeathPenaltyManager centraliza as punições de morte e coordena ressurreições seguras
type DeathPenaltyManager struct {
	dbConn             *sql.DB
	blessingMgr        *blessing.BlessingManager
	professionsMgr     *professions.ProfessionsManager
	respawnMgr         *RespawnManager
	housingMgr         *housing.HousingManager
	inventories        map[string]*inventory.PlayerInventory
	combatManager      *combat.CombatManager
	config             DeathPenaltyConfig
	pvpMgr             *pvp.PvPManager
}

// NewDeathPenaltyManager inicializa o DeathPenaltyManager de forma integrada
func NewDeathPenaltyManager(
	db *sql.DB,
	bm *blessing.BlessingManager,
	pm *professions.ProfessionsManager,
	rm *RespawnManager,
	hm *housing.HousingManager,
	invs map[string]*inventory.PlayerInventory,
	cm *combat.CombatManager,
) *DeathPenaltyManager {
	return &DeathPenaltyManager{
		dbConn:         db,
		blessingMgr:    bm,
		professionsMgr: pm,
		respawnMgr:     rm,
		housingMgr:     hm,
		inventories:    invs,
		combatManager:  cm,
		config: DeathPenaltyConfig{
			XpLossPercentage:    0.10, // 10% de perda canônica de XP
			SkillLossPercentage: 0.05, // 5% de perda canônica de skill
			GoldLossPercentage:  0.10, // 10% de perda de gold
			EquippedDropChance:  0.05, // 5% chance para equipados (Slots 0-3)
			InventoryDropChance: 0.20, // 20% chance para consumíveis/recursos (Slots >= 4)
		},
	}
}

// SetPvPManager associa o PvPManager ao gerenciador de penalidades de morte (PATCH PvP)
func (dpm *DeathPenaltyManager) SetPvPManager(pm *pvp.PvPManager) {
	dpm.pvpMgr = pm
}

// ApplyDeathPenalties aplica punições de morte autoritativas dependendo do status de bênção do jogador
func (dpm *DeathPenaltyManager) ApplyDeathPenalties(playerID string, currentContinent string) (
	xpLost int64,
	goldLost int64,
	skillsLost map[string]int,
	drops []string,
	respawnX float64,
	respawnY float64,
	respawnZ int,
	respawnLocationName string,
	wasProtected bool,
	err error,
) {
	skillsLost = make(map[string]int)
	isPvPDeath := false

	// 0. Se morto por jogador, processar estatísticas de PvP, crânio e recompensas de recompensa (PATCH PvP)
	if dpm.combatManager != nil {
		if killerID, exists := dpm.combatManager.GetLastAttacker(playerID); exists {
			isPvPDeath = true
			if dpm.pvpMgr != nil {
				// Registra o assassinato injusto e atualiza os crânios
				_, _ = dpm.pvpMgr.HandleKillRecord(killerID, playerID)

				// Claim do prêmio da Wanted Board
				_, _ = dpm.pvpMgr.HandleBountyClaim(killerID, playerID)
			}
			dpm.combatManager.ClearLastAttacker(playerID)
		}
	}

	// Retrieve inventory
	playerInv, invExists := dpm.inventories[playerID]

	// Check if Amulet of Loss is equipped/present
	hasAoL := false
	if invExists && playerInv != nil {
		hasAoL = playerInv.IsAoLEquipped()
	}

	// Check how many blessings the player has
	blessingsCount := 0
	isFullyBlessed := false
	if dpm.blessingMgr != nil {
		blessingsCount = dpm.blessingMgr.GetActiveBlessingsCount(playerID)
		isFullyBlessed = dpm.blessingMgr.IsFullyBlessed(playerID)
	}

	// Blessings are consumed on death
	if blessingsCount > 0 && dpm.blessingMgr != nil {
		dpm.blessingMgr.ConsumeAllBlessings(playerID)
	}

	// Amulet of Loss is consumed on death
	if hasAoL && playerInv != nil {
		playerInv.ConsumeAoL()
	}

	wasProtected = hasAoL || blessingsCount > 0

	slog.Warn("Applying formal death penalties.", "player", playerID, "is_pvp", isPvPDeath, "blessings_count", blessingsCount, "has_aol", hasAoL)

	// 1. Aplica perda de XP
	pStats, statsExists := dpm.combatManager.GetEntityStats(playerID)
	if statsExists {
		currentXp := pve.GetPlayerXp(playerID)
		xpNeeded := int64(pStats.Level * pStats.Level * 100)

		var lossPercentage float64
		if isPvPDeath {
			// Base PvP death XP loss: 10%
			baseLossPercent := 0.10

			if blessingsCount > 0 {
				// Each blessing reduces XP loss by 12%
				blessingReduction := 0.12 * float64(blessingsCount)
				if blessingReduction > 0.84 {
					blessingReduction = 0.84
				}
				xpLossMultiplier := 1.0 - blessingReduction
				// AoL does NOT further reduce XP loss if there are active blessings
				lossPercentage = baseLossPercent * xpLossMultiplier
			} else {
				// No active blessings: check if AoL is equipped
				if hasAoL {
					// AoL only: XP loss reduced to 5% (50% reduction)
					lossPercentage = 0.05
				} else {
					// No protection: 10%
					lossPercentage = baseLossPercent
				}
			}
		} else {
			// PvE death: 10%, completely protected if fully blessed
			if isFullyBlessed {
				lossPercentage = 0.0
			} else {
				lossPercentage = dpm.config.XpLossPercentage
			}
		}

		loss := int64(float64(xpNeeded) * lossPercentage)
		if loss > currentXp {
			loss = currentXp
		}

		newXp := currentXp - loss
		pve.SetPlayerXp(playerID, newXp)
		xpLost = loss

		slog.Info("Deducted XP penalty on death", "player", playerID, "xp_lost", loss, "new_xp", newXp, "loss_percentage", lossPercentage)
	}

	// 2. Aplica perda de Gold (Economia)
	// Fully blessed or AoL protects gold loss
	if invExists && !wasProtected {
		currentGold := playerInv.GetGold()
		loss := int64(float64(currentGold) * dpm.config.GoldLossPercentage)

		if loss > 0 {
			playerInv.RemoveGold(loss)
			goldLost = loss

			// Persiste a alteração no banco de dados se conectado
			if dpm.dbConn != nil {
				go func() {
					_, dbErr := dpm.dbConn.Exec("UPDATE characters SET gold = gold - $1 WHERE name = $2 AND gold >= $1", loss, playerID)
					if dbErr != nil {
						slog.Error("Failed to persist gold loss on death", "player", playerID, "error", dbErr)
					}
				}()
			}
			slog.Info("Deducted gold penalty on death", "player", playerID, "gold_lost", loss, "new_gold", playerInv.GetGold())
		}
	}

	// 3. Aplica perda de Skill / Profissões (Progresso)
	// Fully blessed or AoL protects skill loss
	if dpm.professionsMgr != nil && !wasProtected {
		profs, pErr := dpm.professionsMgr.LoadProfessions(playerID)
		if pErr == nil {
			for profName, pState := range profs {
				if pState != nil && pState.Experience > 0 {
					// Perda calculada em 5% do XP atual acumulado na profissão
					loss := int(float64(pState.Experience) * dpm.config.SkillLossPercentage)
					if loss <= 0 {
						loss = 1 // Perda mínima de 1 XP se o jogador tiver XP
					}
					if loss > pState.Experience {
						loss = pState.Experience
					}

					pState.Experience -= loss
					skillsLost[profName] = loss

					// Salva de forma persistente no banco de dados
					saveErr := dpm.professionsMgr.SaveProfession(playerID, profName, pState.Level, pState.Experience)
					if saveErr != nil {
						slog.Error("Failed to persist profession XP loss on death", "player", playerID, "profession", profName, "error", saveErr)
					}
				}
			}
			slog.Info("Applied skill/profession penalty on death", "player", playerID, "skills_penalized", len(skillsLost))
		}
	}

	// 4. Aplica regras de perda/durabilidade de itens de inventário
	// Fully blessed or AoL protects all items from drop
	if invExists {
		playerInv.Lock()
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		// Check if player has Black Skull penalty (PATCH PvP)
		isBlackSkull := false
		if dpm.pvpMgr != nil {
			if skull, exists := dpm.pvpMgr.GetSkullState(playerID); exists {
				if skull.SkullTier == "black" {
					isBlackSkull = true
				}
			}
		}

		for slot, item := range playerInv.Items {
			if item == nil {
				continue
			}

			isEquipped := slot >= 0 && slot <= 3

			if wasProtected {
				// Protected! If equipped and not dropped, suffers 10% durability wear if equipped
				if isEquipped {
					item.Durability -= 10
					if item.Durability < 0 {
						item.Durability = 0
					}
				}
			} else {
				var dropChance float64
				if isBlackSkull {
					if isEquipped {
						dropChance = 0.50 // 50% chance each under Black Skull penalty (PATCH 9)
					} else {
						dropChance = 0.60 // 60% chance each under Black Skull penalty (PATCH 9)
					}
				} else if isEquipped {
					dropChance = dpm.config.EquippedDropChance
				} else {
					dropChance = dpm.config.InventoryDropChance
				}

				if r.Float64() < dropChance {
					// Item dropado! Adiciona na lista e remove do inventário
					drops = append(drops, fmt.Sprintf("%s (x%d)", item.ItemID, item.Quantity))
					delete(playerInv.Items, slot)
				} else if isEquipped {
					// Se equipado e não dropado, sofre desgaste na durabilidade (10% de desgaste)
					item.Durability -= 10
					if item.Durability < 0 {
						item.Durability = 0
					}
				}
			}
		}
		playerInv.SetDirty(true) // Marca inventário como modificado para persistência
		playerInv.Unlock()

		slog.Info("Applied inventory item drop and durability penalty on death", "player", playerID, "dropped_items_count", len(drops))
	}

	// 5. Determina ponto seguro de renascimento (com suporte a custom house)
	var rx, ry float64
	var rz int
	var rName string
	var hasHouse bool

	if dpm.housingMgr != nil {
		rx, ry, rz, rName, hasHouse = dpm.housingMgr.GetPlayerActiveHouseLocation(playerID)
	}

	if !hasHouse {
		rx, ry, rz, rName = dpm.respawnMgr.GetRespawnLocation(playerID, currentContinent, "")
	}

	return xpLost, goldLost, skillsLost, drops, rx, ry, rz, rName, wasProtected, nil
}
