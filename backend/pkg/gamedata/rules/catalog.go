package rules

// Constantes de RuleID para raças jogáveis oficiais.
const (
	RaceHuman     RuleID = "human"
	RaceForestElf RuleID = "forest_elf"
	RaceDwarf     RuleID = "dwarf"
	RaceIceElf    RuleID = "ice_elf"
	RaceGreenOrc  RuleID = "green_orc"
)

// Constantes de RuleID para classes base oficiais.
const (
	ClassKnight   RuleID = "knight"
	ClassMage     RuleID = "mage"
	ClassArcher   RuleID = "archer"
	ClassAssassin RuleID = "assassin"
	ClassCleric   RuleID = "cleric"
)

// Constantes de RuleID para elementos oficiais.
const (
	ElementFire   RuleID = "fire"
	ElementEarth  RuleID = "earth"
	ElementIce    RuleID = "ice"
	ElementShadow RuleID = "shadow"
	ElementSacred RuleID = "sacred"
)

// defaultRuleDefinitions armazena a lista canônica de todas as regras do jogo.
// Não deve ser exposta ou modificada diretamente.
var defaultRuleDefinitions = []RuleDefinition{
	// Raças Jogáveis
	{
		ID:          RaceHuman,
		Category:    CategoryRace,
		DisplayName: "Humano",
		Description: "Uma raça versátil e adaptável, conhecida por sua determinação.",
	},
	{
		ID:          RaceForestElf,
		Category:    CategoryRace,
		DisplayName: "Elfo da Floresta",
		Description: "Uma raça ágil e sábia, com profunda conexão com a natureza.",
	},
	{
		ID:          RaceDwarf,
		Category:    CategoryRace,
		DisplayName: "Anão",
		Description: "Uma raça robusta e engenhosa, mestres da forja e da mineração.",
	},
	{
		ID:          RaceIceElf,
		Category:    CategoryRace,
		DisplayName: "Elfo de Gelo",
		Description: "Uma raça resiliente e reclusa, adaptada aos climas mais frios.",
	},
	{
		ID:          RaceGreenOrc,
		Category:    CategoryRace,
		DisplayName: "Orc Verde",
		Description: "Uma raça forte e orgulhosa, com grande afinidade com o ambiente selvagem.",
	},
	// Classes Base
	{
		ID:          ClassKnight,
		Category:    CategoryClass,
		DisplayName: "Cavaleiro",
		Description: "Um combatente nobre e disciplinado, mestre na defesa e no combate corpo a corpo.",
	},
	{
		ID:          ClassMage,
		Category:    CategoryClass,
		DisplayName: "Mago",
		Description: "Um estudioso das artes arcanas, capaz de manipular energias elementais.",
	},
	{
		ID:          ClassArcher,
		Category:    CategoryClass,
		DisplayName: "Arqueiro",
		Description: "Um atirador preciso e ágil, especialista em combate à distância.",
	},
	{
		ID:          ClassAssassin,
		Category:    CategoryClass,
		DisplayName: "Assassino",
		Description: "Um mestre da furtividade e dos ataques rápidos e letais.",
	},
	{
		ID:          ClassCleric,
		Category:    CategoryClass,
		DisplayName: "Clérigo",
		Description: "Um devoto que canaliza poder divino para curar aliados e punir inimigos.",
	},
	// Elementos
	{ID: ElementFire, Category: CategoryElement, DisplayName: "Fogo", Description: "O elemento da destruição e da paixão, causa dano massivo e contínuo."},
	{ID: ElementEarth, Category: CategoryElement, DisplayName: "Terra", Description: "O elemento da estabilidade e da resistência, focado em defesa e controle."},
	{ID: ElementIce, Category: CategoryElement, DisplayName: "Gelo", Description: "O elemento do controle e da precisão, especializado em retardar e imobilizar inimigos."},
	{ID: ElementShadow, Category: CategoryElement, DisplayName: "Sombrio", Description: "O elemento da furtividade e da corrupção, usa truques e debilitações."},
	{ID: ElementSacred, Category: CategoryElement, DisplayName: "Sagrado", Description: "O elemento da purificação e da cura, focado em suporte e dano contra mortos-vivos."},
}

// DefaultDefinitions retorna uma nova cópia da lista de todas as definições de regras oficiais.
func DefaultDefinitions() []RuleDefinition {
	// Retorna uma cópia para garantir que a lista original não seja mutável.
	defs := make([]RuleDefinition, len(defaultRuleDefinitions))
	copy(defs, defaultRuleDefinitions)
	return defs
}

// NewDefaultRegistry cria um novo Registry pré-populado com todas as regras oficiais do jogo.
func NewDefaultRegistry() (*Registry, error) {
	return NewRegistry(DefaultDefinitions())
}
