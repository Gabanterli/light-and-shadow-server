package rules

import "errors"

// Constantes relacionadas à evolução para classes avançadas.
const (
	// AdvancedClassMinimumCharacterLevel é o nível de personagem mínimo para a evolução.
	AdvancedClassMinimumCharacterLevel uint32 = 100
	// AdvancedClassMinimumAffinityLevel é o nível de afinidade elemental mínimo para a evolução.
	AdvancedClassMinimumAffinityLevel uint32 = 100
)

// Erros exportados para validação da evolução de classe avançada.
var (
	ErrAdvancedClassInvalidBaseClass     = errors.New("base class is not a valid official base class for evolution")
	ErrAdvancedClassInvalidElement       = errors.New("element is not a valid official element for evolution")
	ErrAdvancedClassCharacterLevelTooLow = errors.New("character level is too low for advanced class evolution")
	ErrAdvancedClassAffinityLevelTooLow  = errors.New("elemental affinity level is too low for advanced class evolution")
	ErrAdvancedClassQuestNotCompleted    = errors.New("required quest for advanced class evolution is not completed")
)

// officialElements armazena a lista canônica de elementos que podem ser usados na evolução.
var officialElements = []RuleID{
	ElementFire,
	ElementEarth,
	ElementIce,
	ElementShadow,
	ElementSacred,
}

// officialElementMap é usado para uma verificação rápida e eficiente de elementos.
var officialElementMap = map[RuleID]struct{}{
	ElementFire:   {},
	ElementEarth:  {},
	ElementIce:    {},
	ElementShadow: {},
	ElementSacred: {},
}

// OfficialElementIDs retorna uma nova cópia da lista de IDs de elementos oficiais.
func OfficialElementIDs() []RuleID {
	ids := make([]RuleID, len(officialElements))
	copy(ids, officialElements)
	return ids
}

// IsOfficialElement verifica se um RuleID corresponde a um elemento oficial do jogo.
func IsOfficialElement(id RuleID) bool {
	_, ok := officialElementMap[id]
	return ok
}

// AdvancedClassEvolutionRequest agrupa todos os dados necessários para validar
// uma tentativa de evolução para uma classe avançada.
type AdvancedClassEvolutionRequest struct {
	CharacterLevel uint32
	AffinityLevel  uint32
	QuestCompleted bool
	BaseClass      RuleID
	Element        RuleID
}

// CanEvolveAdvancedClass valida se um personagem atende a todos os requisitos para evoluir.
// A ordem de validação é determinística para garantir erros consistentes.
func CanEvolveAdvancedClass(request AdvancedClassEvolutionRequest) error {
	// 1. Valida se a classe base é uma classe oficial selecionável.
	if !IsOfficialBaseClass(request.BaseClass) {
		return ErrAdvancedClassInvalidBaseClass
	}
	// 2. Valida se o elemento é um elemento oficial do jogo.
	if !IsOfficialElement(request.Element) {
		return ErrAdvancedClassInvalidElement
	}
	// 3. Valida se o personagem atingiu o nível mínimo.
	if request.CharacterLevel < AdvancedClassMinimumCharacterLevel {
		return ErrAdvancedClassCharacterLevelTooLow
	}
	// 4. Valida se a afinidade elemental atingiu o nível mínimo.
	if request.AffinityLevel < AdvancedClassMinimumAffinityLevel {
		return ErrAdvancedClassAffinityLevelTooLow
	}
	// 5. Valida se a quest de evolução foi concluída.
	if !request.QuestCompleted {
		return ErrAdvancedClassQuestNotCompleted
	}
	return nil
}

// MustReachCharacterLevelForAdvancedClass retorna o nível de personagem canônico para evolução.
func MustReachCharacterLevelForAdvancedClass() uint32 {
	return AdvancedClassMinimumCharacterLevel
}

// MustReachAffinityLevelForAdvancedClass retorna o nível de afinidade canônico para evolução.
func MustReachAffinityLevelForAdvancedClass() uint32 {
	return AdvancedClassMinimumAffinityLevel
}
