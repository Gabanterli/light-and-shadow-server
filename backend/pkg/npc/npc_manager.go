package npc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"sync"

	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/dialogue"
	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/quest"
)

// Position representa a coordenada do NPC no mapa
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z int     `json:"z"`
}

// NPC representa a definição estática de um NPC no JSON
type NPC struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Position          Position `json:"position"`
	SpriteID          string   `json:"sprite_id"`
	InteractionRadius float64  `json:"interaction_radius"`
	QuestIDs          []string `json:"quest_ids"`
}

// NPCManager gerencia NPCs ativos no servidor e roteamento de diálogos
type NPCManager struct {
	mu            sync.RWMutex
	npcs          map[string]NPC
	dialogueCache *dialogue.DialogueCache
	spatialIndex  *NPCSpatialIndex
	combatManager *combat.CombatManager
}

// NewNPCManager instancia o NPCManager
func NewNPCManager(cm *combat.CombatManager) *NPCManager {
	nm := &NPCManager{
		npcs:          make(map[string]NPC),
		dialogueCache: dialogue.NewDialogueCache(),
		spatialIndex:  NewNPCSpatialIndex(),
		combatManager: cm,
	}

	nm.loadConfigs()
	return nm
}

// loadConfigs carrega os NPCs dos JSONs correspondentes e inicializa o índice espacial
func (nm *NPCManager) loadConfigs() {
	paths := []string{"backend/config/", "config/", "../config/", "/backend/config/"}

	// 1. Carrega NPCs
	npcsLoaded := false
	for _, p := range paths {
		filePath := p + "npcs.json"
		if _, err := os.Stat(filePath); err == nil {
			data, err := os.ReadFile(filePath)
			if err == nil {
				var list []NPC
				if err := json.Unmarshal(data, &list); err == nil {
					nm.mu.Lock()
					nm.spatialIndex.Clear()
					for _, npc := range list {
						nm.npcs[npc.ID] = npc
						nm.spatialIndex.Insert(npc.ID, npc.Position.X, npc.Position.Y, npc.Position.Z)
					}
					nm.mu.Unlock()
					slog.Info("Successfully loaded npcs.json and updated NPCSpatialIndex", "count", len(nm.npcs), "path", filePath)
					npcsLoaded = true
					break
				}
			}
		}
	}
	if !npcsLoaded {
		slog.Error("Failed to find or load npcs.json files")
	}
}

// GetNPC retorna as propriedades de um NPC
func (nm *NPCManager) GetNPC(npcID string) (NPC, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	npc, ok := nm.npcs[npcID]
	return npc, ok
}

// GetNPCs retorna todos os NPCs
func (nm *NPCManager) GetNPCs() []NPC {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	list := make([]NPC, 0, len(nm.npcs))
	for _, npc := range nm.npcs {
		list = append(list, npc)
	}
	return list
}

// ValidateInteractionDistance valida se o jogador está próximo o suficiente do NPC (Requisitos de Segurança + PATCH 3 Grid Lookup)
func (nm *NPCManager) ValidateInteractionDistance(playerID string, npcID string, si *movement.SpatialIndex) error {
	npc, ok := nm.GetNPC(npcID)
	if !ok {
		return fmt.Errorf("NPC %s does not exist", npcID)
	}

	pEnt, ok := si.GetEntity(playerID)
	if !ok {
		return fmt.Errorf("player %s not found in spatial index", playerID)
	}

	// PATCH 3: Spatial Index lookup (O(k) instead of global scan or direct lookup)
	nearby := nm.spatialIndex.GetNearbyNPCs(pEnt.X, pEnt.Y, pEnt.Z)
	found := false
	for _, id := range nearby {
		if id == npcID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("interaction failed: NPC %s is not in adjacent spatial cells for player %s", npcID, playerID)
	}

	// Verifica se estão no mesmo andar (andar/Z)
	if pEnt.Z != npc.Position.Z {
		return fmt.Errorf("player and NPC are on different floors (player floor %d, NPC floor %d)", pEnt.Z, npc.Position.Z)
	}

	// Distância Euclidiana 2D
	dx := pEnt.X - npc.Position.X
	dy := pEnt.Y - npc.Position.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Validação estrita do raio de interação (Requirement 6)
	if distance > npc.InteractionRadius {
		return fmt.Errorf("interaction failed: player is too far from NPC %s (distance: %.2f, allowed: %.2f)", npcID, distance, npc.InteractionRadius)
	}

	return nil
}

// GetVisibleNode avalia as condições do nó e filtra as respostas elegíveis para o jogador (PATCH 4, PATCH 2 - Cache)
func (nm *NPCManager) GetVisibleNode(playerID string, npcID string, nodeID string, qm *quest.QuestManager) (*dialogue.DialogueNode, error) {
	dialogueTree, hasDialogue := nm.dialogueCache.GetDialogueTree(npcID)
	if !hasDialogue {
		return nil, fmt.Errorf("no dialogue configuration found for NPC %s", npcID)
	}

	// Encontra o nó solicitado
	var targetNode *dialogue.DialogueNode
	for i := range dialogueTree.Nodes {
		if dialogueTree.Nodes[i].NodeID == nodeID {
			targetNode = &dialogueTree.Nodes[i]
			break
		}
	}

	if targetNode == nil {
		return nil, fmt.Errorf("dialogue node %s not found for NPC %s", nodeID, npcID)
	}

	pStats, hasStats := nm.combatManager.GetEntityStats(playerID)
	if !hasStats {
		return nil, fmt.Errorf("failed to retrieve player combat stats")
	}

	pQuestState := qm.GetPlayerState(playerID)

	// Filtra as respostas com base em condicionais de quest e nível de jogador
	visibleResponses := make([]dialogue.Response, 0, len(targetNode.Responses))
	for _, resp := range targetNode.Responses {
		visible := true

		// 1. Valida nível
		if resp.Condition.Level > 0 && pStats.Level < resp.Condition.Level {
			visible = false
		}

		// 2. Valida estado da quest
		if resp.Condition.QuestID != "" && visible {
			qState, exists := pQuestState.Quests[resp.Condition.QuestID]
			expectedStatus := resp.Condition.Status

			switch expectedStatus {
			case "not_started":
				if exists && qState.Status != "not_started" {
					visible = false
				}
			case "active":
				if !exists || qState.Status != "active" {
					visible = false
				} else {
					// Verifica se NÃO está pronto para concluir (pelo menos um objetivo pendente)
					qDef, hasDef := qm.GetDefinition(resp.Condition.QuestID)
					if hasDef {
						ready := true
						for idx, obj := range qDef.Objectives {
							prog, existsProg := qState.Objectives[idx]
							if !existsProg || prog.CurrentQty < obj.RequiredQty {
								ready = false
								break
							}
						}
						if ready {
							visible = false // Se estiver pronto para concluir, não mostra o nó "active" (mostra o "ready_to_complete")
						}
					}
				}
			case "ready_to_complete":
				if !exists || qState.Status != "active" {
					visible = false
				} else {
					// Verifica se de fato todos os objetivos estão cumpridos
					qDef, hasDef := qm.GetDefinition(resp.Condition.QuestID)
					if !hasDef {
						visible = false
					} else {
						for idx, obj := range qDef.Objectives {
							prog, existsProg := qState.Objectives[idx]
							if !existsProg || prog.CurrentQty < obj.RequiredQty {
								visible = false
								break
							}
						}
					}
				}
			case "completed":
				if !exists || qState.Status != "completed" {
					visible = false
				}
			}
		}

		if visible {
			visibleResponses = append(visibleResponses, resp)
		}
	}

	// Retorna um nó filtrado com apenas as respostas válidas
	return &dialogue.DialogueNode{
		NodeID:       targetNode.NodeID,
		Text:         targetNode.Text,
		QuestTrigger: targetNode.QuestTrigger,
		Responses:    visibleResponses,
	}, nil
}
