package pve

import (
    "fmt"
    "sync"
    "time"
)

// CreatureSpawnState represents the authoritative runtime state of one spawned creature.
// It is intentionally in-memory for the first R2 PvE foundation step.
type CreatureSpawnState struct {
    SpawnID         string
    CreatureID      string
    RuntimeEntityID string

    X float64
    Y float64
    Z int

    Alive       bool
    CurrentHP   float64
    MaxHP       float64
    SpawnedAt   time.Time
    DiedAt      time.Time
    NextRespawn time.Time

    KillerPlayerID      string
    LastDamagerPlayerID string
    LootGenerated       bool
    Version             uint64
}

// CreatureSpawnManager owns creature spawn runtime state server-side.
// The manager is concurrency-safe and does not trust client state.
type CreatureSpawnManager struct {
    mu      sync.RWMutex
    spawns  map[string]*CreatureSpawnState
    counter map[string]uint64
}

// NewCreatureSpawnManager creates an empty in-memory creature spawn manager.
func NewCreatureSpawnManager() *CreatureSpawnManager {
    return &CreatureSpawnManager{
        spawns:  make(map[string]*CreatureSpawnState),
        counter: make(map[string]uint64),
    }
}

// RegisterSpawn creates or replaces one authoritative spawn state.
func (m *CreatureSpawnManager) RegisterSpawn(spawnID, creatureID string, x, y float64, z int, maxHP float64) (*CreatureSpawnState, error) {
    if spawnID == "" {
        return nil, fmt.Errorf("spawn id is required")
    }
    if creatureID == "" {
        return nil, fmt.Errorf("creature id is required")
    }
    if maxHP <= 0 {
        return nil, fmt.Errorf("max hp must be greater than zero")
    }

    m.mu.Lock()
    defer m.mu.Unlock()

    m.counter[spawnID]++
    state := &CreatureSpawnState{
        SpawnID:         spawnID,
        CreatureID:      creatureID,
        RuntimeEntityID: buildRuntimeEntityID(spawnID, m.counter[spawnID]),
        X:               x,
        Y:               y,
        Z:               z,
        Alive:           true,
        CurrentHP:       maxHP,
        MaxHP:           maxHP,
        SpawnedAt:       time.Now().UTC(),
        Version:         1,
    }

    m.spawns[spawnID] = state
    return cloneCreatureSpawnState(state), nil
}

// GetSpawn returns a safe copy of a spawn state.
func (m *CreatureSpawnManager) GetSpawn(spawnID string) (*CreatureSpawnState, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    state, exists := m.spawns[spawnID]
    if !exists || state == nil {
        return nil, false
    }

    return cloneCreatureSpawnState(state), true
}

// GetSpawnByRuntimeEntityID returns a safe copy of a spawn state by runtime entity id.
func (m *CreatureSpawnManager) GetSpawnByRuntimeEntityID(runtimeEntityID string) (*CreatureSpawnState, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    for _, state := range m.spawns {
        if state != nil && state.RuntimeEntityID == runtimeEntityID {
            return cloneCreatureSpawnState(state), true
        }
    }

    return nil, false
}

// MarkDamaged records basic damage contributor metadata.
func (m *CreatureSpawnManager) MarkDamaged(spawnID, playerID string) bool {
    m.mu.Lock()
    defer m.mu.Unlock()

    state, exists := m.spawns[spawnID]
    if !exists || state == nil || !state.Alive {
        return false
    }

    state.LastDamagerPlayerID = playerID
    state.Version++
    return true
}

// MarkDead performs an atomic alive -> dead transition.
// It returns false if the spawn is missing or was already dead.
func (m *CreatureSpawnManager) MarkDead(spawnID, killerPlayerID string, respawnDelay time.Duration) (*CreatureSpawnState, bool) {
    m.mu.Lock()
    defer m.mu.Unlock()

    state, exists := m.spawns[spawnID]
    if !exists || state == nil || !state.Alive {
        return nil, false
    }

    now := time.Now().UTC()
    state.Alive = false
    state.CurrentHP = 0
    state.DiedAt = now
    state.NextRespawn = now.Add(respawnDelay)
    state.KillerPlayerID = killerPlayerID
    state.Version++

    return cloneCreatureSpawnState(state), true
}

// MarkLootGenerated performs a one-way loot generation guard.
// It returns false if loot was already generated or the spawn is not dead.
func (m *CreatureSpawnManager) MarkLootGenerated(spawnID string) bool {
    m.mu.Lock()
    defer m.mu.Unlock()

    state, exists := m.spawns[spawnID]
    if !exists || state == nil || state.Alive || state.LootGenerated {
        return false
    }

    state.LootGenerated = true
    state.Version++
    return true
}

// ReviveRespawn resets a dead spawn to a new alive runtime entity.
// It is intended for future server-side respawn timers, not client-triggered revive.
func (m *CreatureSpawnManager) ReviveRespawn(spawnID string) (*CreatureSpawnState, bool) {
    m.mu.Lock()
    defer m.mu.Unlock()

    state, exists := m.spawns[spawnID]
    if !exists || state == nil || state.Alive {
        return nil, false
    }

    m.counter[spawnID]++
    state.RuntimeEntityID = buildRuntimeEntityID(spawnID, m.counter[spawnID])
    state.Alive = true
    state.CurrentHP = state.MaxHP
    state.SpawnedAt = time.Now().UTC()
    state.DiedAt = time.Time{}
    state.NextRespawn = time.Time{}
    state.KillerPlayerID = ""
    state.LastDamagerPlayerID = ""
    state.LootGenerated = false
    state.Version++

    return cloneCreatureSpawnState(state), true
}

// ListSpawns returns safe copies of all known spawn states.
func (m *CreatureSpawnManager) ListSpawns() []*CreatureSpawnState {
    m.mu.RLock()
    defer m.mu.RUnlock()

    result := make([]*CreatureSpawnState, 0, len(m.spawns))
    for _, state := range m.spawns {
        if state != nil {
            result = append(result, cloneCreatureSpawnState(state))
        }
    }

    return result
}

func buildRuntimeEntityID(spawnID string, version uint64) string {
    return fmt.Sprintf("creature:%s:%d", spawnID, version)
}

func cloneCreatureSpawnState(state *CreatureSpawnState) *CreatureSpawnState {
    if state == nil {
        return nil
    }

    copyState := *state
    return &copyState
}
