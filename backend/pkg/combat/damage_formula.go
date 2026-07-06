package combat

import (
	"math"
	"math/rand"
	"time"
)

// EntityStats armazena os atributos relevantes de combate de uma entidade (Player ou NPC)
type EntityStats struct {
	ID                 string
	Name               string
	IsPlayer           bool
	Faction            string // Para friendly fire
	RaceID             string // (R1-H) Raça do personagem
	Level              int    // Added for level scaling
	BaseAttack         float64
	WeaponDamage       float64 // Added for weapon damage
	Defense            float64
	Resistance         float64 // Resistencia percentual (0.0 a 100.0)
	Accuracy           float64
	Evasion            float64
	CritChance         float64 // Ex: 0.05 (5%)
	CritMultiplier     float64 // Ex: 1.50 (150%)
	ArmorPenetration   float64 // Added: Percent of defense ignored (0.0 to 1.0)
	Element            string  // Added: "Fire", "Ice", "Light", "Shadow", "None"
	ElementAttackBonus float64 // Added: Extra elemental damage percent (0.0 to 1.0)
	ElementDefBonus    float64 // Added: Elemental defense/mitigation percent (0.0 to 1.0)
	Health             float64
	MaxHealth          float64
	Mana               float64
	MaxMana            float64
	LastCombatTime     time.Time

	// Sprint 3 Task 5 - Progression & Vocation system
	Class            string // "Novice", "Knight", "Mage", "Archer", "Assassin", "Cleric"
	Subclass         string // e.g. "Holy Knight", "Fire Mage", etc.
	AffinityFire     int
	AffinityIce      int
	AffinityHoly     int
	AffinityShadow   int
	AffinityNature   int
	ProgressionDirty bool
}

// CalculateHitChance calcula a chance de acerto entre duas entidades, limitada entre 10% e 95%
func CalculateHitChance(attacker, defender *EntityStats) float64 {
	acc := attacker.Accuracy
	if acc <= 0 {
		acc = 1.0
	}
	eva := defender.Evasion
	if eva < 0 {
		eva = 0
	}
	hitChance := acc / (acc + eva)
	if hitChance < 0.10 {
		return 0.10
	}
	if hitChance > 0.95 {
		return 0.95
	}
	return hitChance
}

// RollHit determina se um ataque acerta o alvo
func RollHit(attacker, defender *EntityStats) bool {
	chance := CalculateHitChance(attacker, defender)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rng.Float64() <= chance
}

// CalculateDamage calcula o dano final baseado na fórmula oficial MMO completa
func CalculateDamage(attacker, defender *EntityStats, weaponScale float64, skillScale float64, isPvP bool) (float64, bool) {
	// Pipeline: Weapon Damage -> Skill Scaling -> Level Scaling -> Element Scaling -> Critical -> Armor Penetration & Defense Mitigation -> Resistance Reduction -> PvP Modifier -> Random Variance -> Final Damage

	// 1. Weapon Damage & Base Attack
	weaponDmg := attacker.WeaponDamage
	if weaponDmg <= 0 {
		weaponDmg = 15.0 // Default weapon damage fallback
	}
	rawAttack := attacker.BaseAttack + (weaponDmg * weaponScale)

	// 2. Skill Scaling
	baseDamage := rawAttack * skillScale

	// 3. Level Scaling
	attLvl := attacker.Level
	if attLvl <= 0 {
		attLvl = 1
	}
	defLvl := defender.Level
	if defLvl <= 0 {
		defLvl = 1
	}
	levelDiff := float64(attLvl - defLvl)
	levelScale := 1.0 + (levelDiff * 0.02)
	if levelScale < 0.2 {
		levelScale = 0.2
	}
	if levelScale > 2.0 {
		levelScale = 2.0
	}
	baseDamage *= levelScale

	// 3.5 Subclass Damage Boost (Combat Integration)
	if attacker.Subclass != "" {
		// Subclasses gain a persistent 15% passive damage multiplier (Sprint 3 Task 5)
		baseDamage *= 1.15
	}

	// 4. Element Scaling
	elementScale := 1.0
	attElement := attacker.Element
	defElement := defender.Element

	if attElement != "" && defElement != "" {
		if attElement == "Fire" && defElement == "Ice" {
			elementScale = 1.5
		} else if attElement == "Ice" && defElement == "Fire" {
			elementScale = 0.5
		} else if (attElement == "Light" && defElement == "Shadow") || (attElement == "Shadow" && defElement == "Light") {
			elementScale = 1.5
		}
	}

	elementScale *= (1.0 + attacker.ElementAttackBonus)
	elementScale *= (1.0 - defender.ElementDefBonus)
	if elementScale < 0.1 {
		elementScale = 0.1
	}
	baseDamage *= elementScale

	// 5. Critical Roll
	isCrit := false
	critChance := attacker.CritChance
	if critChance < 0.05 {
		critChance = 0.05 // Base crit chance 5%
	}
	if critChance > 0.80 {
		critChance = 0.80 // Cap at 80%
	}
	critMult := attacker.CritMultiplier
	if critMult < 1.50 {
		critMult = 1.50 // Base crit multiplier 150%
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	if rng.Float64() <= critChance {
		isCrit = true
		baseDamage *= critMult
	}

	// 6. Armor Penetration & Defense Mitigation
	armorPen := attacker.ArmorPenetration
	if armorPen < 0 {
		armorPen = 0
	}
	if armorPen > 1.0 {
		armorPen = 1.0
	}
	effectiveDefense := defender.Defense * (1.0 - armorPen)
	if effectiveDefense < 0 {
		effectiveDefense = 0
	}

	defReduction := effectiveDefense / (effectiveDefense + 100.0)
	damageAfterDef := baseDamage * (1.0 - defReduction)

	// 7. Resistance Reduction
	res := defender.Resistance
	if res < 0 {
		res = 0
	}
	if res > 100.0 {
		res = 100.0
	}
	resReduction := res / 100.0
	damageAfterRes := damageAfterDef * (1.0 - resReduction)

	// 8. PvP Modifier (70% do dano)
	finalDamage := damageAfterRes
	if isPvP {
		finalDamage *= 0.70
	}

	// 9. Random Variance (±10% variance)
	variance := 0.90 + (rng.Float64() * 0.20) // 0.90 to 1.10
	finalDamage *= variance

	if finalDamage < 1.0 {
		finalDamage = 1.0 // Garante dano mínimo
	}

	return math.Round(finalDamage*100.0) / 100.0, isCrit
}
