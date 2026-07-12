package movement

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"sync"
	"time"
)

// BoundingBox define intervalos de coordenadas para um continente/regiÃ£o
type BoundingBox struct {
	MinX float64 `json:"min_x"`
	MaxX float64 `json:"max_x"`
	MinY float64 `json:"min_y"`
	MaxY float64 `json:"max_y"`
}

// SafeRespawn define coordenadas para renascimento seguro de jogadores
type SafeRespawn struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z int     `json:"z"`
}

// BossHook representa uma ancoragem/definiÃ§Ã£o para chefes mundiais por continente
type BossHook struct {
	BossID   string  `json:"boss_id"`
	BossName string  `json:"boss_name"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        int     `json:"z"`
}

// WorldRegion representa as configuraÃ§Ãµes de um continente carregado do JSON
type WorldRegion struct {
	RegionID      string        `json:"region_id"`
	ContinentName string        `json:"continent_name"`
	MinLevel      int           `json:"min_level"`
	BoundingBoxes []BoundingBox `json:"bounding_boxes"`
	SafeRespawn   SafeRespawn   `json:"safe_respawn"`
	BossHooks     []BossHook    `json:"boss_hooks"` // Suporte para mÃºltiplos chefes mundiais futuros
}

// PlayerMoveState armazena o histÃ³rico posicional e temporal para validaÃ§Ã£o autoritativa
type PlayerMoveState struct {
	LastX               float64
	LastY               float64
	LastZ               int
	LastTime            time.Time
	Sequence            uint32
	IsInit              bool
	NextAllowedMoveTime time.Time
	CurrentContinent    string // Metadados de continente (Sprint 3 Task 5 Patch 6)
	CurrentRegionID     string // Metadados de ID de regiÃ£o (Sprint 3 Task 5 Patch 6)
}

// BossSpawnCallback define uma assinatura para invocaÃ§Ã£o de chefe mundial
type BossSpawnCallback func(bossID string, x, y float64, z int) error

// MovementSystem coordena a fÃ­sica autoritativa de movimentaÃ§Ã£o e colisÃµes no servidor
type MovementSystem struct {
	mu                 sync.RWMutex
	spatialIndex       *SpatialIndex
	chunkManager       *ChunkManager
	aoiManager         *AOIManager
	playerStates       map[string]*PlayerMoveState
	LevelProvider      func(string) int // Callback para obter level do jogador (Sprint 3 Task 5)
	regions            []WorldRegion
	bossSpawnCallbacks map[string][]BossSpawnCallback // continent_name -> callbacks de spawn
}

// RegisterEntity registra uma entidade no SpatialIndex autoritativo.
func (ms *MovementSystem) RegisterEntity(entity *Entity) {
	if entity == nil {
		return
	}

	ms.spatialIndex.RegisterEntity(entity)
}

// DeregisterEntity remove uma entidade do SpatialIndex autoritativo.
func (ms *MovementSystem) DeregisterEntity(entityID string) {
	ms.spatialIndex.RemoveEntity(entityID)
}

// UpdateEntityPosition atualiza uma entidade no SpatialIndex autoritativo.
func (ms *MovementSystem) UpdateEntityPosition(entityID string, x, y float64, z int) bool {
	updated, _ := ms.spatialIndex.UpdateEntityPosition(entityID, x, y, z)
	return updated
}

// NewMovementSystem inicializa o sistema de movimentaÃ§Ã£o autoritativo e carrega regiÃµes
func NewMovementSystem(si *SpatialIndex, cm *ChunkManager, aoi *AOIManager) *MovementSystem {
	ms := &MovementSystem{
		spatialIndex:       si,
		chunkManager:       cm,
		aoiManager:         aoi,
		playerStates:       make(map[string]*PlayerMoveState),
		bossSpawnCallbacks: make(map[string][]BossSpawnCallback),
	}
	ms.loadWorldRegions()
	return ms
}

// loadWorldRegions busca e carrega as regiÃµes definidas no JSON de forma resiliente
func (ms *MovementSystem) loadWorldRegions() {
	paths := []string{"backend/config/", "config/", "../config/", "../../config/"}
	for _, p := range paths {
		filePath := p + "world_regions.json"
		if _, err := os.Stat(filePath); err == nil {
			data, err := os.ReadFile(filePath)
			if err == nil {
				var list []WorldRegion
				if err := json.Unmarshal(data, &list); err == nil {
					ms.regions = list
					slog.Info("Successfully loaded world_regions.json", "count", len(ms.regions), "path", filePath)
					return
				} else {
					slog.Error("Failed to parse world_regions.json", "error", err)
				}
			}
		}
	}

	slog.Warn("Could not find or load world_regions.json, initializing default world regions fallback")
	// Fallback padrÃ£o se arquivo nÃ£o for encontrado
	ms.regions = []WorldRegion{
		{
			RegionID:      "main_continent",
			ContinentName: "Main Continent",
			MinLevel:      1,
			BoundingBoxes: []BoundingBox{{MinX: 0, MaxX: 1999, MinY: 0, MaxY: 1999}},
			SafeRespawn:   SafeRespawn{X: 100, Y: 100, Z: 0},
		},
		{
			RegionID:      "fire_continent",
			ContinentName: "Fire Continent",
			MinLevel:      50,
			BoundingBoxes: []BoundingBox{{MinX: 2000, MaxX: 2199, MinY: 2000, MaxY: 2199}},
			SafeRespawn:   SafeRespawn{X: 2100, Y: 2100, Z: 0},
		},
		{
			RegionID:      "ice_continent",
			ContinentName: "Ice Continent",
			MinLevel:      50,
			BoundingBoxes: []BoundingBox{{MinX: 2200, MaxX: 2399, MinY: 2200, MaxY: 2399}},
			SafeRespawn:   SafeRespawn{X: 2300, Y: 2300, Z: 0},
		},
		{
			RegionID:      "holy_continent",
			ContinentName: "Holy Continent",
			MinLevel:      50,
			BoundingBoxes: []BoundingBox{
				{MinX: 2400, MaxX: 2599, MinY: 2400, MaxY: 2599},
				{MinX: 4800, MaxX: 5000, MinY: 4800, MaxY: 5000},
			},
			SafeRespawn: SafeRespawn{X: 4900, Y: 4950, Z: 0},
		},
		{
			RegionID:      "shadow_continent",
			ContinentName: "Shadow Continent",
			MinLevel:      50,
			BoundingBoxes: []BoundingBox{
				{MinX: 2600, MaxX: 2799, MinY: 2600, MaxY: 2799},
				{MinX: 3800, MaxX: 4000, MinY: 3800, MaxY: 4000},
			},
			SafeRespawn: SafeRespawn{X: 3900, Y: 3950, Z: 0},
		},
		{
			RegionID:      "nature_continent",
			ContinentName: "Nature Continent",
			MinLevel:      50,
			BoundingBoxes: []BoundingBox{{MinX: 2800, MaxX: 2999, MinY: 2800, MaxY: 2999}},
			SafeRespawn:   SafeRespawn{X: 2900, Y: 2900, Z: 0},
		},
		{
			RegionID:      "abyssia_continent",
			ContinentName: "Abyssia",
			MinLevel:      150,
			BoundingBoxes: []BoundingBox{{MinX: 3000, MaxX: 3799, MinY: 3000, MaxY: 3799}},
			SafeRespawn:   SafeRespawn{X: 3400, Y: 3100, Z: 0},
		},
	}
}

// GetRegionByCoords busca qual regiÃ£o contÃ©m as coordenadas dadas
func (ms *MovementSystem) GetRegionByCoords(x, y float64) *WorldRegion {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	for _, reg := range ms.regions {
		for _, box := range reg.BoundingBoxes {
			if x >= box.MinX && x <= box.MaxX && y >= box.MinY && y <= box.MaxY {
				return &reg
			}
		}
	}
	return nil
}

// GetPlayerContinent retorna o continente atual do jogador
func (ms *MovementSystem) GetPlayerContinent(playerID string) string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if state, exists := ms.playerStates[playerID]; exists {
		if state.CurrentContinent != "" {
			return state.CurrentContinent
		}
	}
	return "Main Continent"
}

// GetPlayerRegionID retorna o ID da regiÃ£o do jogador
func (ms *MovementSystem) GetPlayerRegionID(playerID string) string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if state, exists := ms.playerStates[playerID]; exists {
		if state.CurrentRegionID != "" {
			return state.CurrentRegionID
		}
	}
	return "main_continent"
}

// GetPlayerSafeRespawn retorna as coordenadas de seguranÃ§a baseadas no continente atual do jogador
func (ms *MovementSystem) GetPlayerSafeRespawn(playerID string) (float64, float64, int) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if state, exists := ms.playerStates[playerID]; exists {
		for _, r := range ms.regions {
			if r.RegionID == state.CurrentRegionID {
				return r.SafeRespawn.X, r.SafeRespawn.Y, r.SafeRespawn.Z
			}
		}
	}
	// Fallback padrÃ£o se nÃ£o encontrado
	return 100.0, 100.0, 0
}

// RegisterBossSpawnHook adiciona uma funÃ§Ã£o para processamento de spawn de world boss futuros por continente
func (ms *MovementSystem) RegisterBossSpawnHook(continentName string, cb BossSpawnCallback) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.bossSpawnCallbacks[continentName] = append(ms.bossSpawnCallbacks[continentName], cb)
	slog.Info("Registered boss spawn hook for continent", "continent", continentName)
}

// TriggerWorldBosses dispara mÃºltiplos chefes mundiais para um continente especÃ­fico
func (ms *MovementSystem) TriggerWorldBosses(continentName string, bossID string) error {
	ms.mu.RLock()
	callbacks, ok := ms.bossSpawnCallbacks[continentName]
	var hooks []BossHook
	for _, reg := range ms.regions {
		if reg.ContinentName == continentName {
			hooks = reg.BossHooks
			break
		}
	}
	ms.mu.RUnlock()

	if !ok || len(callbacks) == 0 {
		return fmt.Errorf("no boss spawn hooks registered for continent %s", continentName)
	}

	// Executa os ganchos injetados
	for _, cb := range callbacks {
		if len(hooks) > 0 {
			for _, hk := range hooks {
				if hk.BossID == bossID || bossID == "" {
					if err := cb(hk.BossID, hk.X, hk.Y, hk.Z); err != nil {
						slog.Error("Failed to trigger boss spawn hook", "boss", hk.BossID, "error", err)
					}
				}
			}
		} else {
			// Sem hooks estÃ¡ticos, usa spawn genÃ©rico ou o default passado
			if err := cb(bossID, 150.0, 150.0, 0); err != nil {
				return err
			}
		}
	}
	return nil
}

// InitPlayerState define a posiÃ§Ã£o inicial confiÃ¡vel do jogador no servidor e resolve o continente
func (ms *MovementSystem) InitPlayerState(playerID string, x, y float64, z int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Identifica a regiÃ£o inicial pelas coordenadas
	regionID := "main_continent"
	continent := "Main Continent"
	for _, reg := range ms.regions {
		for _, box := range reg.BoundingBoxes {
			if x >= box.MinX && x <= box.MaxX && y >= box.MinY && y <= box.MaxY {
				regionID = reg.RegionID
				continent = reg.ContinentName
				break
			}
		}
	}

	ms.playerStates[playerID] = &PlayerMoveState{
		LastX:               x,
		LastY:               y,
		LastZ:               z,
		LastTime:            time.Now(),
		IsInit:              true,
		NextAllowedMoveTime: time.Now(),
		CurrentContinent:    continent,
		CurrentRegionID:     regionID,
	}

	// Insere no indexador espacial se nÃ£o existir
	ms.spatialIndex.RegisterEntity(&Entity{
		ID:   playerID,
		Name: "Player_" + playerID,
		X:    x,
		Y:    y,
		Z:    z,
		Type: "player",
	})
}

// RemovePlayerState limpa recursos de movimentaÃ§Ã£o do jogador ao desconectar
func (ms *MovementSystem) RemovePlayerState(playerID string) {
	ms.mu.Lock()
	delete(ms.playerStates, playerID)
	ms.mu.Unlock()

	ms.spatialIndex.RemoveEntity(playerID)
}

// ValidateAndMove realiza as checagens autoritativas (velocidade, colisÃµes de mapa, andares)
func (ms *MovementSystem) ValidateAndMove(playerID string, targetX, targetY float64, targetZ int, seq uint32) (bool, float64, float64, int) {
	ms.mu.Lock()
	state, exists := ms.playerStates[playerID]
	if !exists {
		ms.mu.Unlock()
		ms.InitPlayerState(playerID, targetX, targetY, targetZ)
		return true, targetX, targetY, targetZ
	}
	ms.mu.Unlock()

	// DetecÃ§Ã£o de Teleporte Externo (Compatibilidade de regressÃ£o com instÃ¢ncias de masmorras/morte/teleportes)
	if ent, ok := ms.spatialIndex.GetEntity(playerID); ok {
		const TeleportThreshold = 2.0
		if math.Abs(ent.X-state.LastX) > TeleportThreshold || math.Abs(ent.Y-state.LastY) > TeleportThreshold || ent.Z != state.LastZ {
			ms.mu.Lock()
			state.LastX = ent.X
			state.LastY = ent.Y
			state.LastZ = ent.Z
			ms.mu.Unlock()
			slog.Info("Detected external teleport, resetting cached movement system coordinate state", "player", playerID, "new_x", ent.X, "new_y", ent.Y)
		}
	}

	now := time.Now()

	// 0. Cooldown Check (PATCH 3 â€” Movement Cooldown)
	// B3-D: Centralized validation continues to use existing cooldown logic.
	if now.Before(state.NextAllowedMoveTime) {
		return false, state.LastX, state.LastY, state.LastZ
	}

	// 1. Verificar ColisÃ£o de ObstÃ¡culo EstÃ¡tico no Chunk
	if ms.chunkManager.IsBlocked(int(targetX), int(targetY)) {
		return false, state.LastX, state.LastY, state.LastZ
	}

	// B3-D: Centralized authoritative spatial occupancy check for dynamic entities.
	if ms.spatialIndex.IsTileOccupied(playerID, targetX, targetY, targetZ) {
		return false, state.LastX, state.LastY, state.LastZ
	}

	// 1.5 ValidaÃ§Ã£o Autoritativa de RegiÃ£o baseada em world_regions.json (Sprint 3 Task 5 Patch 6)
	targetRegion := ms.GetRegionByCoords(targetX, targetY)
	if targetRegion != nil {
		level := 1
		if ms.LevelProvider != nil {
			level = ms.LevelProvider(playerID)
		}
		if level < targetRegion.MinLevel {
			slog.Warn("Level requirement not met for region access", "player", playerID, "level", level, "required", targetRegion.MinLevel, "region", targetRegion.ContinentName)
			return false, state.LastX, state.LastY, state.LastZ
		}
	}

	// 2. ValidaÃ§Ã£o Autoritativa de Velocidade (Anti-Speedhack com tolerÃ¢ncia a Jitter)
	dt := now.Sub(state.LastTime).Seconds()
	dx := targetX - state.LastX
	dy := targetY - state.LastY

	distance := math.Sqrt(dx*dx + dy*dy)
	const BaseSpeed = 4.0
	const Tolerance = 1.15
	maxAllowedDistance := (BaseSpeed * dt) * Tolerance

	if distance > 0.01 && dt > 0.0 {
		const MaxLagBuffer = 1.5
		if distance > maxAllowedDistance+MaxLagBuffer {
			slog.Warn("Speedhack check rejected movement", "player", playerID, "distance", distance, "max_allowed", maxAllowedDistance+MaxLagBuffer)
			return false, state.LastX, state.LastY, state.LastZ
		}
	}

	// 3. Atualizar Estado VÃ¡lido, Metadados de Continente e Ãndices Espaciais
	ms.mu.Lock()
	state.LastX = targetX
	state.LastY = targetY
	state.LastZ = targetZ
	state.LastTime = now
	state.Sequence = seq
	state.NextAllowedMoveTime = now.Add(250 * time.Millisecond)

	// Atualiza metadados do continente no player state
	if targetRegion != nil {
		state.CurrentContinent = targetRegion.ContinentName
		state.CurrentRegionID = targetRegion.RegionID
	} else {
		state.CurrentContinent = "Main Continent"
		state.CurrentRegionID = "main_continent"
	}
	ms.mu.Unlock()

	// Sincroniza a nova posiÃ§Ã£o no SpatialIndex
	ms.spatialIndex.UpdateEntityPosition(playerID, targetX, targetY, targetZ)

	// Dispara o recÃ¡lculo e transmissÃ£o de AOI de visibilidade (Spawn / Despawn de vizinhos)
	ms.aoiManager.UpdatePlayerAOI(playerID, targetX, targetY, targetZ)

	return true, targetX, targetY, targetZ
}

// GetPlayerPos retorna a Ãºltima posiÃ§Ã£o vÃ¡lida conhecida do jogador no servidor
func (ms *MovementSystem) GetPlayerPos(playerID string) (float64, float64, int, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	state, exists := ms.playerStates[playerID]
	if !exists {
		return 0, 0, 0, false
	}
	return state.LastX, state.LastY, state.LastZ, true
}
