package rules

import "errors"

// Constantes relacionadas à seleção de classe.
const (
	// StartingClassNovice é a classe inicial de todo personagem recém-criado.
	StartingClassNovice RuleID = "novice"
	// ClassSelectionMinimumLevel é o nível mínimo necessário para escolher uma classe base.
	ClassSelectionMinimumLevel uint32 = 10
)

// Erros exportados para validação de seleção de classe.
var (
	ErrClassSelectionLevelTooLow   = errors.New("character level is too low to select a base class")
	ErrClassSelectionAlreadyChosen = errors.New("a base class has already been chosen")
	ErrInvalidBaseClass            = errors.New("target class is not a valid official base class")
)

// officialBaseClasses armazena a lista canônica de classes base que podem ser escolhidas.
var officialBaseClasses = []RuleID{
	ClassKnight,
	ClassMage,
	ClassArcher,
	ClassAssassin,
	ClassCleric,
}

// officialBaseClassMap é usado para uma verificação rápida e eficiente.
var officialBaseClassMap = map[RuleID]struct{}{
	ClassKnight:   {},
	ClassMage:     {},
	ClassArcher:   {},
	ClassAssassin: {},
	ClassCleric:   {},
}

// OfficialBaseClassIDs retorna uma nova cópia da lista de IDs de classes base oficiais.
func OfficialBaseClassIDs() []RuleID {
	ids := make([]RuleID, len(officialBaseClasses))
	copy(ids, officialBaseClasses)
	return ids
}

// IsOfficialBaseClass verifica se um RuleID corresponde a uma classe base oficial.
func IsOfficialBaseClass(id RuleID) bool {
	_, ok := officialBaseClassMap[id]
	return ok
}

// CanSelectBaseClass valida se um personagem pode mudar para uma classe base alvo.
// A ordem de validação é determinística: classe alvo, nível e classe atual.
func CanSelectBaseClass(characterLevel uint32, currentClass RuleID, targetClass RuleID) error {
	// 1. Valida se a classe alvo é uma classe base oficial.
	if !IsOfficialBaseClass(targetClass) {
		return ErrInvalidBaseClass
	}

	// 2. Valida se o personagem atingiu o nível mínimo.
	if characterLevel < ClassSelectionMinimumLevel {
		return ErrClassSelectionLevelTooLow
	}

	// 3. Valida se o personagem ainda é um 'novice'.
	if currentClass != StartingClassNovice {
		return ErrClassSelectionAlreadyChosen
	}

	return nil
}

// MustSelectClassAtOrAfterLevel retorna o nível mínimo de seleção de classe.
// Esta função serve para que outros módulos obtenham esta regra de forma canônica.
func MustSelectClassAtOrAfterLevel() uint32 {
	return ClassSelectionMinimumLevel
}
