package rules

import "errors"

// ItemMutationSource define a origem de uma mutação de item (criação, transferência, etc.).
type ItemMutationSource string

// Constantes para as fontes oficiais de mutação de itens.
const (
	ItemMutationSourceLootDrop      ItemMutationSource = "loot_drop"
	ItemMutationSourcePlayerTrade   ItemMutationSource = "player_trade"
	ItemMutationSourceQuestReward   ItemMutationSource = "quest_reward"
	ItemMutationSourceInventoryMove ItemMutationSource = "inventory_move"
)

// Erros exportados para validação de autoridade sobre itens.
var (
	ErrInvalidItemMutationSource       = errors.New("source is not an official item mutation source")
	ErrItemMutationNotBackendValidated = errors.New("item mutation was not validated by the backend")
	ErrClientCreatedItemRejected       = errors.New("client-created items are strictly rejected")
	ErrItemMutationNotBackendGenerated = errors.New("item creation source requires a backend-generated item")
	ErrItemMutationTransactionRequired = errors.New("item mutation source requires a transaction")
	ErrItemMutationAuditRequired       = errors.New("item mutation must be audited")
)

// officialItemMutationSources armazena a lista canônica de fontes de mutação.
var officialItemMutationSources = []ItemMutationSource{
	ItemMutationSourceLootDrop,
	ItemMutationSourcePlayerTrade,
	ItemMutationSourceQuestReward,
	ItemMutationSourceInventoryMove,
}

// officialItemMutationSourceMap é usado para uma verificação rápida e eficiente.
var officialItemMutationSourceMap = map[ItemMutationSource]struct{}{
	ItemMutationSourceLootDrop:      {},
	ItemMutationSourcePlayerTrade:   {},
	ItemMutationSourceQuestReward:   {},
	ItemMutationSourceInventoryMove: {},
}

// OfficialItemMutationSources retorna uma nova cópia da lista de fontes oficiais.
func OfficialItemMutationSources() []ItemMutationSource {
	sources := make([]ItemMutationSource, len(officialItemMutationSources))
	copy(sources, officialItemMutationSources)
	return sources
}

// IsOfficialItemMutationSource verifica se uma fonte de mutação é oficialmente reconhecida.
func IsOfficialItemMutationSource(source ItemMutationSource) bool {
	_, ok := officialItemMutationSourceMap[source]
	return ok
}

// ItemMutationSourceCreatesItem verifica se uma fonte de mutação pode criar novos itens.
func ItemMutationSourceCreatesItem(source ItemMutationSource) bool {
	return source == ItemMutationSourceLootDrop || source == ItemMutationSourceQuestReward
}

// ItemMutationSourceRequiresTransaction verifica se uma fonte de mutação exige uma transação atômica.
func ItemMutationSourceRequiresTransaction(source ItemMutationSource) bool {
	return source == ItemMutationSourcePlayerTrade
}

// ItemMutationAuthorityRequest agrupa os dados para validar a autoridade de uma mutação de item.
type ItemMutationAuthorityRequest struct {
	Source               ItemMutationSource
	BackendValidated     bool
	ClientCreatedItem    bool
	BackendGeneratedItem bool
	Transactional        bool
	AuditLogged          bool
}

// CanApplyItemMutation valida se uma mutação de item pode ser aplicada.
func CanApplyItemMutation(request ItemMutationAuthorityRequest) error {
	if !IsOfficialItemMutationSource(request.Source) {
		return ErrInvalidItemMutationSource
	}
	if !request.BackendValidated {
		return ErrItemMutationNotBackendValidated
	}
	if request.ClientCreatedItem {
		return ErrClientCreatedItemRejected
	}
	if ItemMutationSourceCreatesItem(request.Source) && !request.BackendGeneratedItem {
		return ErrItemMutationNotBackendGenerated
	}
	if ItemMutationSourceRequiresTransaction(request.Source) && !request.Transactional {
		return ErrItemMutationTransactionRequired
	}
	if !request.AuditLogged {
		return ErrItemMutationAuditRequired
	}
	return nil
}

// MustValidateItemMutationsOnBackend confirma que a validação é server-authoritative.
func MustValidateItemMutationsOnBackend() bool {
	return true
}

// MustRejectClientCreatedItems confirma que itens criados pelo cliente são proibidos.
func MustRejectClientCreatedItems() bool {
	return true
}

// MustAuditItemMutations confirma que mutações de itens devem ser registradas.
func MustAuditItemMutations() bool {
	return true
}
