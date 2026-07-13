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
	WorldSpaceID        string // Authoritative identity used by static terrain
	CurrentContinent    string // Metadados de continente (Sprint 3 Task 5 Patch 6)
	CurrentRegionID     string // Metadados de ID de regiÃ£o (Sprint 3 Task 5 Patch 6)
}

// StaticStep describes one authoritative movement attempt without coupling the
// movement package to a concrete world-map implementation.
type StaticStep struct {
	WorldSpaceID string
	FromX        float64
	FromY        float64
	FromZ        int
	ToX          float64
	ToY          float64
	ToZ          int
}

// StaticStepValidator validates immutable terrain for one movement step.
// Implementations must return a non-nil error for blocked, unavailable,
// malformed or otherwise invalid terrain.
type StaticStepValidator interface {
	ValidateStaticStep(step StaticStep) error
}

// playerMovementLockEntry serializes one player's movement requests. The
// reference count includes both the current owner and queued waiters.
type playerMovementLockEntry struct {
	mutex      sync.Mutex
	references int
}

// BossSpawnCallback define uma assinatura para invocaÃ§Ã£o de chefe mundial
type BossSpawnCallback func(bossID string, x, y float64, z int) error

// MovementSystem coordena a fÃ­sica autoritativa de movimentaÃ§Ã£o e colisÃµes no servidor
type MovementSystem struct {
	mu                    sync.RWMutex
	spatialIndex          *SpatialIndex
	chunkManager          *ChunkManager
	aoiManager            *AOIManager
	staticStepValidator   StaticStepValidator
	playerStates          map[string]*PlayerMoveState
	playerMovementLocksMu sync.Mutex
	playerMovementLocks   map[string]*playerMovementLockEntry
	LevelProvider         func(string) int // Callback para obter level do jogador (Sprint 3 Task 5)
	regions               []WorldRegion
	bossSpawnCallbacks    map[string][]BossSpawnCallback // continent_name -> callbacks de spawn
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
// NewMovementSystem preserves the legacy/debug terrain behavior.
func NewMovementSystem(
	si *SpatialIndex,
	cm *ChunkManager,
	aoi *AOIManager,
) *MovementSystem {
	return NewMovementSystemWithStaticStepValidator(
		si,
		cm,
		aoi,
		nil,
	)
}

// NewMovementSystemWithStaticStepValidator creates an authoritative movement
// system with an optional immutable production-terrain validator.
func NewMovementSystemWithStaticStepValidator(
	si *SpatialIndex,
	cm *ChunkManager,
	aoi *AOIManager,
	staticStepValidator StaticStepValidator,
) *MovementSystem {
	ms := &MovementSystem{
		spatialIndex:        si,
		chunkManager:        cm,
		aoiManager:          aoi,
		staticStepValidator: staticStepValidator,
		playerStates:        make(map[string]*PlayerMoveState),
		playerMovementLocks: make(map[string]*playerMovementLockEntry),
		bossSpawnCallbacks:  make(map[string][]BossSpawnCallback),
	}
	ms.loadWorldRegions()
	return ms
}

// lockPlayerMovement acquires one reference-counted player-specific lock and
// returns an idempotence-by-contract release function for a single defer.
func (ms *MovementSystem) lockPlayerMovement(
	playerID string,
) func() {
	ms.playerMovementLocksMu.Lock()

	entry, exists := ms.playerMovementLocks[playerID]
	if !exists {
		entry = &playerMovementLockEntry{}
		ms.playerMovementLocks[playerID] = entry
	}

	entry.references++
	ms.playerMovementLocksMu.Unlock()

	entry.mutex.Lock()

	return func() {
		entry.mutex.Unlock()

		ms.playerMovementLocksMu.Lock()
		defer ms.playerMovementLocksMu.Unlock()

		entry.references--
		if entry.references == 0 {
			delete(ms.playerMovementLocks, playerID)
		}
	}
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
func (ms *MovementSystem) GetRegionByCoords(
	x, y float64,
) *WorldRegion {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.getRegionByCoordsLocked(x, y)
}

// getRegionByCoordsLocked resolves one region while the caller already owns
// ms.mu for reading or writing.
func (ms *MovementSystem) getRegionByCoordsLocked(
	x, y float64,
) *WorldRegion {
	for regionIndex := range ms.regions {
		region := &ms.regions[regionIndex]

		for _, box := range region.BoundingBoxes {
			if x >= box.MinX &&
				x <= box.MaxX &&
				y >= box.MinY &&
				y <= box.MaxY {
				return region
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

// InitPlayerState preserves the legacy initialization contract by assigning
// the canonical Alpha world-space identity.
func (ms *MovementSystem) InitPlayerState(
	playerID string,
	x, y float64,
	z int,
) {
	ms.InitPlayerStateInWorldSpace(
		playerID,
		"main_continent",
		x,
		y,
		z,
	)
}

// InitPlayerStateInWorldSpace defines the trusted initial position and explicit
// world-space identity used by authoritative static terrain validation.
func (ms *MovementSystem) InitPlayerStateInWorldSpace(
	playerID string,
	worldSpaceID string,
	x, y float64,
	z int,
) {
	unlockPlayer := ms.lockPlayerMovement(playerID)
	defer unlockPlayer()

	if worldSpaceID == "" {
		worldSpaceID = "main_continent"
	}

	ms.mu.Lock()

	regionID := "main_continent"
	continent := "Main Continent"
	for _, region := range ms.regions {
		for _, box := range region.BoundingBoxes {
			if x >= box.MinX &&
				x <= box.MaxX &&
				y >= box.MinY &&
				y <= box.MaxY {
				regionID = region.RegionID
				continent = region.ContinentName
				break
			}
		}
	}

	now := time.Now()

	ms.playerStates[playerID] = &PlayerMoveState{
		LastX:               x,
		LastY:               y,
		LastZ:               z,
		LastTime:            now,
		IsInit:              true,
		NextAllowedMoveTime: now,
		WorldSpaceID:        worldSpaceID,
		CurrentContinent:    continent,
		CurrentRegionID:     regionID,
	}

	ms.mu.Unlock()

	ms.spatialIndex.RegisterEntity(&Entity{
		ID:             playerID,
		Name:           "Player_" + playerID,
		X:              x,
		Y:              y,
		Z:              z,
		Type:           "player",
		BlocksMovement: true,
	})
}

// RemovePlayerState limpa recursos de movimentaÃ§Ã£o do jogador ao desconectar
func (ms *MovementSystem) RemovePlayerState(
	playerID string,
) {
	unlockPlayer := ms.lockPlayerMovement(playerID)
	defer unlockPlayer()

	ms.mu.Lock()
	delete(ms.playerStates, playerID)
	ms.mu.Unlock()

	ms.spatialIndex.RemoveEntity(playerID)
}

// ValidateAndMove realiza as checagens autoritativas (velocidade, colisÃµes de mapa, andares)
func (ms *MovementSystem) ValidateAndMove(
	playerID string,
	targetX, targetY float64,
	targetZ int,
	seq uint32,
) (bool, float64, float64, int) {
	unlockPlayer := ms.lockPlayerMovement(playerID)
	defer unlockPlayer()

	entity, entityFound :=
		ms.spatialIndex.GetEntity(playerID)

	ms.mu.Lock()

	state, exists := ms.playerStates[playerID]
	if !exists {
		ms.mu.Unlock()
		return false, 0, 0, 0
	}

	if entityFound {
		const teleportThreshold = 2.0

		if math.Abs(entity.X-state.LastX) > teleportThreshold ||
			math.Abs(entity.Y-state.LastY) > teleportThreshold ||
			entity.Z != state.LastZ {
			state.LastX = entity.X
			state.LastY = entity.Y
			state.LastZ = entity.Z

			slog.Info(
				"Detected external teleport, resetting cached movement system coordinate state",
				"player",
				playerID,
				"new_x",
				entity.X,
				"new_y",
				entity.Y,
				"new_z",
				entity.Z,
			)
		}
	}

	lastX := state.LastX
	lastY := state.LastY
	lastZ := state.LastZ
	lastTime := state.LastTime
	nextAllowedMoveTime := state.NextAllowedMoveTime
	worldSpaceID := state.WorldSpaceID

	ms.mu.Unlock()

	reject := func() (bool, float64, float64, int) {
		return false, lastX, lastY, lastZ
	}

	now := time.Now()

	if now.Before(nextAllowedMoveTime) {
		return reject()
	}

	if ms.staticStepValidator != nil {
		err := ms.staticStepValidator.ValidateStaticStep(
			StaticStep{
				WorldSpaceID: worldSpaceID,
				FromX:        lastX,
				FromY:        lastY,
				FromZ:        lastZ,
				ToX:          targetX,
				ToY:          targetY,
				ToZ:          targetZ,
			},
		)
		if err != nil {
			return reject()
		}
	} else if ms.chunkManager.IsBlocked(
		int(targetX),
		int(targetY),
	) {
		return reject()
	}

	if ms.spatialIndex.IsTileOccupied(
		playerID,
		targetX,
		targetY,
		targetZ,
	) {
		return reject()
	}

	targetRegion := ms.GetRegionByCoords(
		targetX,
		targetY,
	)
	if targetRegion != nil {
		level := 1
		if ms.LevelProvider != nil {
			level = ms.LevelProvider(playerID)
		}

		if level < targetRegion.MinLevel {
			slog.Warn(
				"Level requirement not met for region access",
				"player",
				playerID,
				"level",
				level,
				"required",
				targetRegion.MinLevel,
				"region",
				targetRegion.ContinentName,
			)

			return reject()
		}
	}

	deltaTime := now.Sub(lastTime).Seconds()
	deltaX := targetX - lastX
	deltaY := targetY - lastY
	distance := math.Sqrt(
		deltaX*deltaX +
			deltaY*deltaY,
	)

	const (
		baseSpeed    = 4.0
		tolerance    = 1.15
		maxLagBuffer = 1.5
	)

	maxAllowedDistance :=
		baseSpeed * deltaTime * tolerance

	if distance > 0.01 &&
		deltaTime > 0 &&
		distance > maxAllowedDistance+maxLagBuffer {
		slog.Warn(
			"Speedhack check rejected movement",
			"player",
			playerID,
			"distance",
			distance,
			"max_allowed",
			maxAllowedDistance+maxLagBuffer,
		)

		return reject()
	}

	ms.mu.Lock()

	state, exists = ms.playerStates[playerID]
	if !exists {
		ms.mu.Unlock()
		return false, 0, 0, 0
	}

	state.LastX = targetX
	state.LastY = targetY
	state.LastZ = targetZ
	state.LastTime = now
	state.Sequence = seq
	state.NextAllowedMoveTime = now.Add(
		250 * time.Millisecond,
	)

	if targetRegion != nil {
		state.CurrentContinent =
			targetRegion.ContinentName
		state.CurrentRegionID =
			targetRegion.RegionID
	} else {
		state.CurrentContinent = "Main Continent"
		state.CurrentRegionID = "main_continent"
	}

	ms.mu.Unlock()

	ms.spatialIndex.UpdateEntityPosition(
		playerID,
		targetX,
		targetY,
		targetZ,
	)

	ms.aoiManager.UpdatePlayerAOI(
		playerID,
		targetX,
		targetY,
		targetZ,
	)

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
