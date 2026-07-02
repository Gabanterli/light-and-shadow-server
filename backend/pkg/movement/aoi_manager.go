package movement

import (
	"encoding/json"
	"log/slog"
	"net"
	"sync"

	"github.com/light-and-shadow/backend/pkg/protocol"
)

// AOIManager coordena a visibilidade mútua e sincronização de rede entre jogadores próximos
type AOIManager struct {
	mu           sync.RWMutex
	spatialIndex *SpatialIndex
	// Conexões de rede ativas indexadas pelo ID do jogador
	connections  map[string]net.Conn
	// Rastreamento de quais entidades cada jogador atualmente "enxerga"
	// playerID -> set de entityIDs visíveis
	visibility   map[string]map[string]bool
}

// NewAOIManager instancia o gerenciador de visibilidade (AOI)
func NewAOIManager(si *SpatialIndex) *AOIManager {
	return &AOIManager{
		spatialIndex: si,
		connections:  make(map[string]net.Conn),
		visibility:   make(map[string]map[string]bool),
	}
}

// RegisterPlayer registra uma conexão ativa para receber atualizações espaciais
func (aoi *AOIManager) RegisterPlayer(id string, conn net.Conn) {
	aoi.mu.Lock()
	defer aoi.mu.Unlock()

	aoi.connections[id] = conn
	aoi.visibility[id] = make(map[string]bool)
	slog.Info("Player connection registered in AOIManager", "id", id)
}

// GetPlayerConn retorna a conexão de rede ativa de um jogador de forma thread-safe
func (aoi *AOIManager) GetPlayerConn(id string) (net.Conn, bool) {
	aoi.mu.RLock()
	defer aoi.mu.RUnlock()
	conn, exists := aoi.connections[id]
	return conn, exists
}

// DeregisterPlayer remove a conexão e envia pacotes de despawn para todos os que viam este jogador
func (aoi *AOIManager) DeregisterPlayer(id string) {
	aoi.mu.Lock()
	// Remove a conexão e o conjunto de visibilidade do jogador
	delete(aoi.connections, id)
	delete(aoi.visibility, id)
	aoi.mu.Unlock()

	// Notifica outros jogadores sobre o despawn deste jogador desregistrado
	aoi.broadcastDespawn(id)
	slog.Info("Player connection deregistered from AOIManager", "id", id)
}

// UpdatePlayerAOI recalcula a visibilidade ao redor do jogador e dispara eventos de Spawn/Despawn (Deltas)
func (aoi *AOIManager) UpdatePlayerAOI(playerID string, x, y float64, z int) {
	aoi.mu.Lock()
	conn, hasConn := aoi.connections[playerID]
	observed, hasObs := aoi.visibility[playerID]
	aoi.mu.Unlock()

	if !hasConn || !hasObs {
		return
	}

	// Viewport é 24x18. Definimos o raio do AOI como 20 tiles (Sprint 2 Patch 4)
	const AOIRadius = 20.0

	// Consulta o SpatialIndex por todas as entidades próximas
	nearby := aoi.spatialIndex.GetEntitiesInRegion(x, y, AOIRadius, z)

	newVisible := make(map[string]*Entity)
	newVisibleIDs := make(map[string]bool)

	for _, ent := range nearby {
		// Não precisamos registrar spawn de nós mesmos
		if ent.ID == playerID {
			continue
		}
		newVisible[ent.ID] = ent
		newVisibleIDs[ent.ID] = true
	}

	// 1. Identificar entidades que SAÍRAM da área de visão (Despawn)
	var despawnIDs []string
	for oldID := range observed {
		if !newVisibleIDs[oldID] {
			despawnIDs = append(despawnIDs, oldID)
		}
	}

	// 2. Identificar entidades que ENTRARAM na área de visão (Spawn)
	var spawnEntities []*Entity
	for newID, ent := range newVisible {
		if !observed[newID] {
			spawnEntities = append(spawnEntities, ent)
		}
	}

	// Atualizar o set de visibilidade sob proteção
	aoi.mu.Lock()
	for _, id := range despawnIDs {
		delete(aoi.visibility[playerID], id)
	}
	for _, ent := range spawnEntities {
		aoi.visibility[playerID][ent.ID] = true
	}
	aoi.mu.Unlock()

	// 3. Enviar pacotes de Despawn para o cliente do jogador
	for _, id := range despawnIDs {
		payload := []byte(id)
		packet := &protocol.Packet{
			Opcode:  protocol.SC_DESPAWN_ENTITY,
			Payload: payload,
		}
		conn.Write(packet.Serialize())

		// Se a outra entidade for um jogador ativo, notifica ela também sobre o despawn recíproco
		aoi.sendDespawnToPlayer(id, playerID)
	}

	// 4. Enviar pacotes de Spawn para o cliente do jogador
	for _, ent := range spawnEntities {
		payload, err := json.Marshal(ent)
		if err != nil {
			slog.Error("Failed to marshal spawn entity JSON", "error", err)
			continue
		}
		packet := &protocol.Packet{
			Opcode:  protocol.SC_SPAWN_ENTITY,
			Payload: payload,
		}
		conn.Write(packet.Serialize())

		// Se a outra entidade for um jogador ativo, envia spawn recíproco dela nos enxergando
		aoi.sendSpawnToPlayer(ent.ID, playerID, x, y, z)
	}
}

// BroadcastMove envia atualizações de posição para todos os jogadores vizinhos
func (aoi *AOIManager) BroadcastMove(sourceID string, x, y float64, z int, payload []byte) {
	aoi.BroadcastMovement(sourceID, protocol.SC_PLAYER_UPDATE, payload)
}

// BroadcastMovement envia atualizações de movimento com opcode específico para jogadores na AOI que observam a fonte
func (aoi *AOIManager) BroadcastMovement(sourceID string, opcode uint16, payload []byte) {
	aoi.mu.RLock()
	defer aoi.mu.RUnlock()

	for playerID, observed := range aoi.visibility {
		if observed[sourceID] {
			if conn, exists := aoi.connections[playerID]; exists {
				packet := &protocol.Packet{
					Opcode:  opcode,
					Payload: payload,
				}
				conn.Write(packet.Serialize())
			}
		}
	}
}

// BroadcastCombat envia atualizações de combate com opcode específico para jogadores na AOI que observam a fonte
func (aoi *AOIManager) BroadcastCombat(sourceID string, opcode uint16, payload []byte) {
	aoi.mu.RLock()
	defer aoi.mu.RUnlock()

	for playerID, observed := range aoi.visibility {
		if observed[sourceID] {
			if conn, exists := aoi.connections[playerID]; exists {
				packet := &protocol.Packet{
					Opcode:  opcode,
					Payload: payload,
				}
				conn.Write(packet.Serialize())
			}
		}
	}
}

// BroadcastEffects envia atualizações visuais/gráficas com opcode específico para jogadores na AOI que observam a fonte
func (aoi *AOIManager) BroadcastEffects(sourceID string, opcode uint16, payload []byte) {
	aoi.mu.RLock()
	defer aoi.mu.RUnlock()

	for playerID, observed := range aoi.visibility {
		if observed[sourceID] {
			if conn, exists := aoi.connections[playerID]; exists {
				packet := &protocol.Packet{
					Opcode:  opcode,
					Payload: payload,
				}
				conn.Write(packet.Serialize())
			}
		}
	}
}

// sendDespawnToPlayer envia um despawn unilateral de targetID para o playerID
func (aoi *AOIManager) sendDespawnToPlayer(playerID, targetID string) {
	aoi.mu.Lock()
	conn, existsConn := aoi.connections[playerID]
	observed, existsObs := aoi.visibility[playerID]
	if existsConn && existsObs {
		delete(observed, targetID)
	}
	aoi.mu.Unlock()

	if existsConn {
		packet := &protocol.Packet{
			Opcode:  protocol.SC_DESPAWN_ENTITY,
			Payload: []byte(targetID),
		}
		conn.Write(packet.Serialize())
	}
}

// sendSpawnToPlayer envia o spawn unilateral do playerID (com coordenadas atuais) para o targetID
func (aoi *AOIManager) sendSpawnToPlayer(playerID, targetID string, x, y float64, z int) {
	aoi.mu.Lock()
	conn, existsConn := aoi.connections[playerID]
	observed, existsObs := aoi.visibility[playerID]
	if existsConn && existsObs {
		observed[targetID] = true
	}
	aoi.mu.Unlock()

	if existsConn {
		// Obtém informações da entidade a ser spawned
		aoi.spatialIndex.mu.RLock()
		targetEnt, ok := aoi.spatialIndex.entities[targetID]
		aoi.spatialIndex.mu.RUnlock()

		if !ok {
			// Se não achar entidade na memória, usa dados padrão
			targetEnt = &Entity{
				ID:   targetID,
				Name: "Player_" + targetID,
				X:    x,
				Y:    y,
				Z:    z,
				Type: "player",
			}
		}

		payload, err := json.Marshal(targetEnt)
		if err != nil {
			return
		}

		packet := &protocol.Packet{
			Opcode:  protocol.SC_SPAWN_ENTITY,
			Payload: payload,
		}
		conn.Write(packet.Serialize())
	}
}

// broadcastDespawn limpa e despawna um jogador deslogado de todas as telas vizinhas
func (aoi *AOIManager) broadcastDespawn(despawnedID string) {
	aoi.mu.Lock()
	defer aoi.mu.Unlock()

	packet := &protocol.Packet{
		Opcode:  protocol.SC_DESPAWN_ENTITY,
		Payload: []byte(despawnedID),
	}

	serialized := packet.Serialize()

	for playerID, observed := range aoi.visibility {
		if observed[despawnedID] {
			delete(observed, despawnedID)
			if conn, exists := aoi.connections[playerID]; exists {
				conn.Write(serialized)
			}
		}
	}
}
