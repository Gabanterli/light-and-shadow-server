package rules

import "errors"

// SkillProgressionSource defines the origin of a skill progression attempt.
type SkillProgressionSource string

// Constantes para as fontes oficiais de progressão de skill.
const (
	SkillProgressionSourceCombatUse     SkillProgressionSource = "combat_use"
	SkillProgressionSourceTrainingUse   SkillProgressionSource = "training_use"
	SkillProgressionSourceProfessionUse SkillProgressionSource = "profession_use"
)

// Constante para o intervalo mínimo entre ganhos de skill para evitar spam.
const (
	SkillProgressionMinimumIntervalMilliseconds uint32 = 1000
)

// Erros exportados para validação de elegibilidade de progressão de skill.
var (
	ErrInvalidSkillProgressionSource             = errors.New("source is not an official skill progression source")
	ErrSkillProgressionActionNotValidated        = errors.New("skill progression action was not validated by the backend")
	ErrSkillProgressionContextNotValidated       = errors.New("skill progression context was not validated by the backend")
	ErrSkillProgressionIntervalTooShort          = errors.New("time since last skill gain is too short")
	ErrSkillProgressionDiminishingReturnsBlocked = errors.New("skill progression is currently blocked by diminishing returns")
)

// officialSkillProgressionSources armazena a lista canônica de fontes de progressão.
var officialSkillProgressionSources = []SkillProgressionSource{
	SkillProgressionSourceCombatUse,
	SkillProgressionSourceTrainingUse,
	SkillProgressionSourceProfessionUse,
}

// officialSkillProgressionSourceMap é usado para uma verificação rápida e eficiente.
var officialSkillProgressionSourceMap = map[SkillProgressionSource]struct{}{
	SkillProgressionSourceCombatUse:     {},
	SkillProgressionSourceTrainingUse:   {},
	SkillProgressionSourceProfessionUse: {},
}

// OfficialSkillProgressionSources retorna uma nova cópia da lista de fontes oficiais.
func OfficialSkillProgressionSources() []SkillProgressionSource {
	sources := make([]SkillProgressionSource, len(officialSkillProgressionSources))
	copy(sources, officialSkillProgressionSources)
	return sources
}

// IsOfficialSkillProgressionSource verifica se uma fonte é oficialmente reconhecida.
func IsOfficialSkillProgressionSource(source SkillProgressionSource) bool {
	_, ok := officialSkillProgressionSourceMap[source]
	return ok
}

// SkillProgressionEligibilityRequest agrupa os dados para validar a elegibilidade de ganho de skill.
type SkillProgressionEligibilityRequest struct {
	Source                    SkillProgressionSource
	ActionValidatedByBackend  bool
	ContextValidatedByBackend bool
	MillisecondsSinceLastGain uint32
	DiminishingReturnsBlocked bool
}

// CanGrantSkillProgression valida se um personagem pode receber progresso de skill.
func CanGrantSkillProgression(request SkillProgressionEligibilityRequest) error {
	if !IsOfficialSkillProgressionSource(request.Source) {
		return ErrInvalidSkillProgressionSource
	}
	if !request.ActionValidatedByBackend {
		return ErrSkillProgressionActionNotValidated
	}
	if !request.ContextValidatedByBackend {
		return ErrSkillProgressionContextNotValidated
	}
	if request.MillisecondsSinceLastGain < SkillProgressionMinimumIntervalMilliseconds {
		return ErrSkillProgressionIntervalTooShort
	}
	if request.DiminishingReturnsBlocked {
		return ErrSkillProgressionDiminishingReturnsBlocked
	}
	return nil
}

// MustWaitMillisecondsBetweenSkillProgressionGains retorna o intervalo mínimo canônico.
func MustWaitMillisecondsBetweenSkillProgressionGains() uint32 {
	return SkillProgressionMinimumIntervalMilliseconds
}

// MustValidateSkillProgressionOnBackend confirma que a validação é server-authoritative.
func MustValidateSkillProgressionOnBackend() bool {
	return true
}
