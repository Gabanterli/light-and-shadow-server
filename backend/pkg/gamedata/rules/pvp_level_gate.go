package rules

import "errors"

// Constantes relacionadas ao level gate de PvP.
const (
	// OpenPvPMinimumLevel é o nível mínimo para um personagem poder participar de PvP aberto.
	OpenPvPMinimumLevel uint32 = 10
)

// Erros exportados para validação de engajamento em PvP.
var (
	ErrPvPAttackerLevelTooLow = errors.New("attacker level is too low for open pvp")
	ErrPvPTargetLevelTooLow   = errors.New("target level is too low for open pvp")
)

// PvPLevelGateRequest agrupa os dados necessários para validar uma tentativa de engajamento em PvP.
type PvPLevelGateRequest struct {
	AttackerLevel uint32
	TargetLevel   uint32
}

// IsLevelEligibleForOpenPvP verifica se um nível de personagem atende ao requisito mínimo para PvP aberto.
func IsLevelEligibleForOpenPvP(level uint32) bool {
	return level >= OpenPvPMinimumLevel
}

// CanEngageOpenPvP valida se um atacante pode engajar um alvo em PvP aberto, baseado no nível de ambos.
// A ordem de validação é determinística, verificando primeiro o atacante.
func CanEngageOpenPvP(request PvPLevelGateRequest) error {
	if !IsLevelEligibleForOpenPvP(request.AttackerLevel) {
		return ErrPvPAttackerLevelTooLow
	}
	if !IsLevelEligibleForOpenPvP(request.TargetLevel) {
		return ErrPvPTargetLevelTooLow
	}
	return nil
}

// MustReachLevelForOpenPvP retorna o nível mínimo canônico para PvP aberto.
// Esta função serve para que outros módulos obtenham esta regra de forma centralizada.
func MustReachLevelForOpenPvP() uint32 {
	return OpenPvPMinimumLevel
}
