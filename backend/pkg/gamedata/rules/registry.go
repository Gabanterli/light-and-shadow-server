package rules

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// RuleID é um identificador de texto único e estável para uma regra de jogo.
type RuleID string

// RuleCategory define a qual grupo uma regra pertence (ex: raça, classe).
type RuleCategory string

// Categorias de regras iniciais.
const (
	CategoryRace    RuleCategory = "race"
	CategoryClass   RuleCategory = "class"
	CategoryElement RuleCategory = "element"
	CategorySystem  RuleCategory = "system"
)

// RuleDefinition define uma única regra de jogo no registro.
type RuleDefinition struct {
	ID          RuleID
	Category    RuleCategory
	DisplayName string
	Description string
}

// ruleIDRegex valida o formato snake_case para IDs de regras.
var ruleIDRegex = regexp.MustCompile(`^[a-z][a-z0-9_]*[a-z0-9]$`)

// validateRuleID verifica se um RuleID segue as convenções obrigatórias.
func validateRuleID(id RuleID) error {
	if id == "" {
		return fmt.Errorf("rule ID cannot be empty")
	}
	if !ruleIDRegex.MatchString(string(id)) {
		return fmt.Errorf("rule ID '%s' has invalid format: must be lowercase snake_case, cannot start with a number or underscore, and cannot end with an underscore", id)
	}
	if strings.Contains(string(id), "__") {
		return fmt.Errorf("rule ID '%s' cannot contain consecutive underscores", id)
	}
	return nil
}

// Registry armazena e fornece acesso a todas as definições de regras do jogo.
// É a fonte da verdade server-authoritative para as regras.
type Registry struct {
	rules map[RuleID]RuleDefinition
}

// NewRegistry cria e inicializa um novo Rule Registry a partir de uma lista de definições.
// Ele valida cada definição e retorna um erro se alguma regra for inválida ou se houver IDs duplicados.
func NewRegistry(definitions []RuleDefinition) (*Registry, error) {
	rules := make(map[RuleID]RuleDefinition, len(definitions))

	for _, def := range definitions {
		if err := validateRuleID(def.ID); err != nil {
			return nil, fmt.Errorf("invalid rule definition: %w", err)
		}
		if def.Category == "" {
			return nil, fmt.Errorf("rule '%s' has an empty category", def.ID)
		}
		if def.DisplayName == "" {
			return nil, fmt.Errorf("rule '%s' has an empty display name", def.ID)
		}

		if _, exists := rules[def.ID]; exists {
			return nil, fmt.Errorf("duplicate rule ID found: '%s'", def.ID)
		}

		// Armazena uma cópia para garantir imutabilidade da entrada.
		rules[def.ID] = def
	}

	return &Registry{rules: rules}, nil
}

// Get busca uma RuleDefinition por seu ID. Retorna a definição e um booleano
// indicando se ela foi encontrada. A definição retornada é uma cópia.
func (r *Registry) Get(id RuleID) (RuleDefinition, bool) {
	def, ok := r.rules[id]
	return def, ok // A struct é retornada por valor (cópia).
}

// List retorna uma nova lista de todas as RuleDefinitions no registro,
// ordenada de forma determinística por categoria e depois por ID.
func (r *Registry) List() []RuleDefinition {
	list := make([]RuleDefinition, 0, len(r.rules))
	for _, def := range r.rules {
		list = append(list, def)
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].Category != list[j].Category {
			return list[i].Category < list[j].Category
		}
		return list[i].ID < list[j].ID
	})

	return list
}

// Count retorna o número total de regras carregadas no registro.
func (r *Registry) Count() int {
	return len(r.rules)
}
