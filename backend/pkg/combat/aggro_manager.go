package combat

import (
	"sync"
)

// AggroTable gerencia os níveis de ameaça (threat) de vários jogadores para um NPC
type AggroTable struct {
	mu     sync.RWMutex
	threat map[string]float64
}

// NewAggroTable cria uma nova tabela de aggro vazia
func NewAggroTable() *AggroTable {
	return &AggroTable{
		threat: make(map[string]float64),
	}
}

// AddThreat adiciona ou atualiza o nível de ameaça de um jogador
func (at *AggroTable) AddThreat(playerID string, amount float64) {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.threat[playerID] += amount
}

// SetThreat define diretamente um valor de ameaça de um jogador
func (at *AggroTable) SetThreat(playerID string, amount float64) {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.threat[playerID] = amount
}

// GetThreat retorna o nível de ameaça atual de um jogador
func (at *AggroTable) GetThreat(playerID string) float64 {
	at.mu.RLock()
	defer at.mu.RUnlock()
	return at.threat[playerID]
}

// ClearThreat limpa toda a tabela de aggro
func (at *AggroTable) ClearThreat() {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.threat = make(map[string]float64)
}

// RemovePlayer remove um jogador específico da tabela
func (at *AggroTable) RemovePlayer(playerID string) {
	at.mu.Lock()
	defer at.mu.Unlock()
	delete(at.threat, playerID)
}

// GetTopTarget retorna o ID do jogador com maior ameaça
func (at *AggroTable) GetTopTarget() (string, bool) {
	at.mu.RLock()
	defer at.mu.RUnlock()

	var topPlayer string
	maxThreat := -1.0

	for playerID, amount := range at.threat {
		if amount > maxThreat {
			maxThreat = amount
			topPlayer = playerID
		}
	}

	if maxThreat < 0 {
		return "", false
	}
	return topPlayer, true
}

// DecayThreats reduz a ameaça de todos os jogadores em uma taxa baseada no tempo (2% por segundo)
func (at *AggroTable) DecayThreats(elapsedSeconds float64) {
	at.mu.Lock()
	defer at.mu.Unlock()

	// 2% por segundo
	multiplier := 1.0 - (0.02 * elapsedSeconds)
	if multiplier < 0 {
		multiplier = 0
	}

	for playerID := range at.threat {
		at.threat[playerID] *= multiplier
		if at.threat[playerID] < 1.0 {
			delete(at.threat, playerID) // Remove ameaças insignificantes
		}
	}
}
