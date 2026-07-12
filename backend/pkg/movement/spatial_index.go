package movement

import (
	"math"
	"sync"
)

// Entity representa qualquer entidade móvel no mundo de jogo (jogador ou NPC)
type Entity struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	X              float64 `json:"x"`    // Coordenada X em Tiles
	Y              float64 `json:"y"`    // Coordenada Y em Tiles
	Z              int     `json:"z"`    // Floor (0 a 15)
	Type           string  `json:"type"` // "player" ou "npc"
	BlocksMovement bool    `json:"blocks_movement"`
}

// SpatialIndex gerencia o posicionamento de entidades em blocos 3D (X, Y, Z/Andar)
type SpatialIndex struct {
	mu       sync.RWMutex
	entities map[string]*Entity
	// Mapeia chunkKey -> lista de IDs de entidades
	// chunkKey é calculada como (chunkX << 32) | chunkY
	// Cada Z (andar) tem seu próprio mapa de chunks para isolamento total
	floors [16]map[uint64]map[string]*Entity
}

// NewSpatialIndex instancia o indexador espacial
func NewSpatialIndex() *SpatialIndex {
	si := &SpatialIndex{
		entities: make(map[string]*Entity),
	}
	for i := 0; i < 16; i++ {
		si.floors[i] = make(map[uint64]map[string]*Entity)
	}
	return si
}

// getChunkKey calcula a chave única para o chunk de tamanho 32x32 tiles
func getChunkKey(x, y float64) uint64 {
	cx := uint32(int(x) / 32)
	cy := uint32(int(y) / 32)
	return (uint64(cx) << 32) | uint64(cy)
}

// RegisterEntity adiciona uma nova entidade ao indexador espacial
func (si *SpatialIndex) RegisterEntity(entity *Entity) {
	if entity == nil || entity.ID == "" {
		return
	}

	si.mu.Lock()
	defer si.mu.Unlock()

	if entity.Z < 0 || entity.Z >= len(si.floors) {
		entity.Z = 0
	}

	// Registro idempotente: remove qualquer posição anterior do mesmo ID.
	if existing, exists := si.entities[entity.ID]; exists {
		oldKey := getChunkKey(existing.X, existing.Y)

		if existing.Z >= 0 && existing.Z < len(si.floors) {
			if chunkEntities := si.floors[existing.Z][oldKey]; chunkEntities != nil {
				delete(chunkEntities, entity.ID)

				if len(chunkEntities) == 0 {
					delete(si.floors[existing.Z], oldKey)
				}
			}
		}
	}

	si.entities[entity.ID] = entity
	key := getChunkKey(entity.X, entity.Y)

	if si.floors[entity.Z][key] == nil {
		si.floors[entity.Z][key] = make(map[string]*Entity)
	}

	si.floors[entity.Z][key][entity.ID] = entity
}

// UpdateEntityPosition atualiza de forma atômica e thread-safe a posição de uma entidade
func (si *SpatialIndex) UpdateEntityPosition(id string, newX, newY float64, newZ int) (bool, *Entity) {
	si.mu.Lock()
	defer si.mu.Unlock()

	entity, exists := si.entities[id]
	if !exists || entity == nil {
		return false, nil
	}

	if newZ < 0 || newZ >= len(si.floors) {
		return false, entity
	}

	// Entidades bloqueantes não podem ocupar atomicamente um tile que já
	// contenha outra entidade bloqueante.
	if entity.BlocksMovement && si.isTileOccupiedLocked(id, newX, newY, newZ) {
		return false, entity
	}

	oldX := entity.X
	oldY := entity.Y
	oldZ := entity.Z

	oldKey := getChunkKey(oldX, oldY)
	newKey := getChunkKey(newX, newY)

	positionChanged :=
		oldX != newX ||
			oldY != newY ||
			oldZ != newZ

	if !positionChanged {
		return true, entity
	}

	if oldZ >= 0 && oldZ < len(si.floors) {
		if oldChunk := si.floors[oldZ][oldKey]; oldChunk != nil {
			delete(oldChunk, id)

			if len(oldChunk) == 0 {
				delete(si.floors[oldZ], oldKey)
			}
		}
	}

	if si.floors[newZ][newKey] == nil {
		si.floors[newZ][newKey] = make(map[string]*Entity)
	}

	si.floors[newZ][newKey][id] = entity

	entity.X = newX
	entity.Y = newY
	entity.Z = newZ

	return true, entity
}

// B3-C: RemoveEntity is part of the blocking creature lifecycle.
// RemoveEntity desregistra uma entidade do indexador espacial
func (si *SpatialIndex) RemoveEntity(id string) {
	si.mu.Lock()
	defer si.mu.Unlock()

	entity, exists := si.entities[id]
	if !exists {
		return
	}

	key := getChunkKey(entity.X, entity.Y)
	if si.floors[entity.Z][key] != nil {
		delete(si.floors[entity.Z][key], id)
		if len(si.floors[entity.Z][key]) == 0 {
			delete(si.floors[entity.Z], key)
		}
	}

	delete(si.entities, id)
}

// GetEntity retorna uma entidade pelo ID de forma thread-safe
func (si *SpatialIndex) GetEntity(id string) (*Entity, bool) {
	si.mu.RLock()
	defer si.mu.RUnlock()
	entity, exists := si.entities[id]
	return entity, exists
}

// B3-D: IsTileOccupied is the centralized authority for spatial occupancy queries.
// IsTileOccupied verifica se outra entidade bloqueante ocupa o tile informado.
func (si *SpatialIndex) IsTileOccupied(excludeEntityID string, x, y float64, z int) bool {
	si.mu.RLock()
	defer si.mu.RUnlock()

	return si.isTileOccupiedLocked(excludeEntityID, x, y, z)
}

// isTileOccupiedLocked consulta a ocupação sem adquirir um novo lock.
// O chamador deve possuir si.mu em modo de leitura ou escrita.
func (si *SpatialIndex) isTileOccupiedLocked(excludeEntityID string, x, y float64, z int) bool {
	if z < 0 || z >= len(si.floors) {
		return false
	}

	key := getChunkKey(x, y)
	chunkEntities, exists := si.floors[z][key]
	if !exists {
		return false
	}

	targetTileX := int(math.Floor(x))
	targetTileY := int(math.Floor(y))

	for _, entity := range chunkEntities {
		if entity == nil {
			continue
		}

		if entity.ID == excludeEntityID {
			continue
		}

		if !entity.BlocksMovement {
			continue
		}

		entityTileX := int(math.Floor(entity.X))
		entityTileY := int(math.Floor(entity.Y))

		if entityTileX == targetTileX && entityTileY == targetTileY {
			return true
		}
	}

	return false
}

// GetEntitiesInRegion retorna entidades presentes nos chunks vizinhos que cobrem a região pesquisada
func (si *SpatialIndex) GetEntitiesInRegion(x, y float64, radius float64, z int) []*Entity {
	if z < 0 || z >= 16 {
		return nil
	}

	si.mu.RLock()
	defer si.mu.RUnlock()

	var result []*Entity

	// Determinar a faixa de chunks a varrer
	minX := math.Max(0, x-radius)
	maxX := math.Min(16384, x+radius)
	minY := math.Max(0, y-radius)
	maxY := math.Min(16384, y+radius)

	minChunkX := int(minX) / 32
	maxChunkX := int(maxX) / 32
	minChunkY := int(minY) / 32
	maxChunkY := int(maxY) / 32

	for cx := minChunkX; cx <= maxChunkX; cx++ {
		for cy := minChunkY; cy <= maxChunkY; cy++ {
			key := (uint64(cx) << 32) | uint64(cy)
			if chunkEntities, exists := si.floors[z][key]; exists {
				for _, ent := range chunkEntities {
					// Verifica distância euclidiana fina
					dx := ent.X - x
					dy := ent.Y - y
					dist := math.Sqrt(dx*dx + dy*dy)
					if dist <= radius {
						result = append(result, ent)
					}
				}
			}
		}
	}

	return result
}
