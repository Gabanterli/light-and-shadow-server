package rules

import "errors"

// ZoneType define o tipo de uma área do jogo, determinando suas regras de combate.
type ZoneType string

// Constantes para os tipos de zona oficiais.
const (
	ZoneTypeSafe    ZoneType = "safe"
	ZoneTypeCombat  ZoneType = "combat"
	ZoneTypeNeutral ZoneType = "neutral"
)

// Erros exportados para validação de regras de zona.
var (
	ErrInvalidZoneType         = errors.New("zone type is not an official zone type")
	ErrCombatBlockedInSafeZone = errors.New("combat is blocked in a safe zone")
)

// officialZoneTypes armazena a lista canônica de tipos de zona.
var officialZoneTypes = []ZoneType{
	ZoneTypeSafe,
	ZoneTypeCombat,
	ZoneTypeNeutral,
}

// officialZoneTypeMap é usado para uma verificação rápida e eficiente.
var officialZoneTypeMap = map[ZoneType]struct{}{
	ZoneTypeSafe:    {},
	ZoneTypeCombat:  {},
	ZoneTypeNeutral: {},
}

// OfficialZoneTypes retorna uma nova cópia da lista de tipos de zona oficiais.
func OfficialZoneTypes() []ZoneType {
	types := make([]ZoneType, len(officialZoneTypes))
	copy(types, officialZoneTypes)
	return types
}

// IsOfficialZoneType verifica se um ZoneType é um tipo oficial reconhecido.
func IsOfficialZoneType(zoneType ZoneType) bool {
	_, ok := officialZoneTypeMap[zoneType]
	return ok
}

// IsSafeZone verifica se um ZoneType corresponde a uma zona segura.
func IsSafeZone(zoneType ZoneType) bool {
	return zoneType == ZoneTypeSafe
}

// ZoneCombatRuleRequest agrupa os dados necessários para validar as regras de combate de uma zona.
type ZoneCombatRuleRequest struct {
	ZoneType ZoneType
}

// CanCombatOccurInZone valida se uma ação de combate pode ocorrer em uma determinada zona.
func CanCombatOccurInZone(request ZoneCombatRuleRequest) error {
	if !IsOfficialZoneType(request.ZoneType) {
		return ErrInvalidZoneType
	}
	if IsSafeZone(request.ZoneType) {
		return ErrCombatBlockedInSafeZone
	}
	return nil
}

// MustBlockCombatInSafeZone retorna um booleano que confirma a regra de bloqueio de combate.
func MustBlockCombatInSafeZone() bool {
	return true
}
