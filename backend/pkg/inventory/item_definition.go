package inventory

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

type ItemType string

const (
	ItemTypeWeapon     ItemType = "weapon"
	ItemTypeArmor      ItemType = "armor"
	ItemTypeAccessory  ItemType = "accessory"
	ItemTypeConsumable ItemType = "consumable"
	ItemTypeMaterial   ItemType = "material"
)

var itemIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type ElementalStats struct {
	FireAttack       float64 `json:"FireAttack,omitempty"`
	IceAttack        float64 `json:"IceAttack,omitempty"`
	HolyAttack       float64 `json:"HolyAttack,omitempty"`
	ShadowAttack     float64 `json:"ShadowAttack,omitempty"`
	NatureAttack     float64 `json:"NatureAttack,omitempty"`
	FireResistance   float64 `json:"FireResistance,omitempty"`
	IceResistance    float64 `json:"IceResistance,omitempty"`
	HolyResistance   float64 `json:"HolyResistance,omitempty"`
	ShadowResistance float64 `json:"ShadowResistance,omitempty"`
	NatureResistance float64 `json:"NatureResistance,omitempty"`
}

type ItemDefinition struct {
	ID                string         `json:"ID"`
	Name              string         `json:"Name"`
	Type              ItemType       `json:"Type"`
	Category          string         `json:"Category,omitempty"`
	Tier              int            `json:"Tier"`
	BaseDamage        float64        `json:"BaseDamage,omitempty"`
	BaseDef           float64        `json:"BaseDef,omitempty"`
	BaseRes           float64        `json:"BaseRes,omitempty"`
	CritBonus         float64        `json:"CritBonus,omitempty"`
	CompatibleClasses []string       `json:"CompatibleClasses,omitempty"`
	Stackable         bool           `json:"Stackable"`
	MaxStack          int            `json:"MaxStack"`
	Craftable         bool           `json:"Craftable,omitempty"`
	ElementalStats    ElementalStats `json:"ElementalStats,omitempty"`
	Description       string         `json:"Description,omitempty"`
	Price             int64          `json:"Price,omitempty"`
	MarketPrice       int64          `json:"MarketPrice,omitempty"`
}

// ItemDef preserves the legacy public name while keeping one canonical model.
type ItemDef = ItemDefinition

func (d ItemDefinition) validate() error {
	if d.ID == "" {
		return fmt.Errorf("ID is required")
	}
	if strings.TrimSpace(d.ID) != d.ID {
		return fmt.Errorf("ID %q must not contain surrounding whitespace", d.ID)
	}
	if !itemIDPattern.MatchString(d.ID) {
		return fmt.Errorf("ID %q must match %s", d.ID, itemIDPattern.String())
	}
	if strings.TrimSpace(d.Name) == "" {
		return fmt.Errorf("Name is required")
	}

	switch d.Type {
	case ItemTypeWeapon, ItemTypeArmor, ItemTypeAccessory:
		if strings.TrimSpace(d.Category) == "" {
			return fmt.Errorf("Category is required for item type %q", d.Type)
		}
		if d.Tier < 1 || d.Tier > 5 {
			return fmt.Errorf("Tier for item type %q must be between 1 and 5", d.Type)
		}
		if d.Stackable {
			return fmt.Errorf("item type %q cannot be stackable", d.Type)
		}
	case ItemTypeConsumable, ItemTypeMaterial:
		if d.Tier < 0 || d.Tier > 5 {
			return fmt.Errorf("Tier for item type %q must be between 0 and 5", d.Type)
		}
	default:
		return fmt.Errorf("unsupported Type %q", d.Type)
	}

	if d.Category != "" && strings.TrimSpace(d.Category) == "" {
		return fmt.Errorf("Category cannot contain only whitespace")
	}
	if d.Description != "" && strings.TrimSpace(d.Description) == "" {
		return fmt.Errorf("Description cannot contain only whitespace")
	}
	if d.MaxStack < 1 {
		return fmt.Errorf("MaxStack must be at least 1")
	}
	if d.Stackable && d.MaxStack < 2 {
		return fmt.Errorf("stackable items require MaxStack of at least 2")
	}
	if !d.Stackable && d.MaxStack != 1 {
		return fmt.Errorf("non-stackable items require MaxStack equal to 1")
	}

	for _, field := range []struct {
		name  string
		value float64
	}{
		{"BaseDamage", d.BaseDamage},
		{"BaseDef", d.BaseDef},
		{"BaseRes", d.BaseRes},
		{"CritBonus", d.CritBonus},
	} {
		if err := validateFiniteNonNegative(field.name, field.value); err != nil {
			return err
		}
	}
	if d.CritBonus > 1 {
		return fmt.Errorf("CritBonus must be between 0 and 1")
	}
	if d.Price < 0 {
		return fmt.Errorf("Price cannot be negative")
	}
	if d.MarketPrice < 0 {
		return fmt.Errorf("MarketPrice cannot be negative")
	}
	if err := d.ElementalStats.validate(); err != nil {
		return err
	}

	allowedClasses := map[string]struct{}{
		"Novice":   {},
		"Knight":   {},
		"Mage":     {},
		"Archer":   {},
		"Assassin": {},
		"Cleric":   {},
	}
	seenClasses := make(map[string]struct{}, len(d.CompatibleClasses))
	for index, classID := range d.CompatibleClasses {
		if _, ok := allowedClasses[classID]; !ok {
			return fmt.Errorf("CompatibleClasses[%d] contains unsupported class %q", index, classID)
		}
		if _, duplicate := seenClasses[classID]; duplicate {
			return fmt.Errorf("CompatibleClasses contains duplicate class %q", classID)
		}
		seenClasses[classID] = struct{}{}
	}

	return nil
}

func (s ElementalStats) validate() error {
	values := []struct {
		name  string
		value float64
	}{
		{"ElementalStats.FireAttack", s.FireAttack},
		{"ElementalStats.IceAttack", s.IceAttack},
		{"ElementalStats.HolyAttack", s.HolyAttack},
		{"ElementalStats.ShadowAttack", s.ShadowAttack},
		{"ElementalStats.NatureAttack", s.NatureAttack},
		{"ElementalStats.FireResistance", s.FireResistance},
		{"ElementalStats.IceResistance", s.IceResistance},
		{"ElementalStats.HolyResistance", s.HolyResistance},
		{"ElementalStats.ShadowResistance", s.ShadowResistance},
		{"ElementalStats.NatureResistance", s.NatureResistance},
	}
	for _, field := range values {
		if err := validateFiniteNonNegative(field.name, field.value); err != nil {
			return err
		}
	}
	return nil
}

func validateFiniteNonNegative(field string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("%s must be finite", field)
	}
	if value < 0 {
		return fmt.Errorf("%s cannot be negative", field)
	}
	return nil
}

func cloneItemDefinition(definition ItemDefinition) ItemDefinition {
	copyDefinition := definition
	if definition.CompatibleClasses != nil {
		copyDefinition.CompatibleClasses = append(
			[]string(nil),
			definition.CompatibleClasses...,
		)
	}
	return copyDefinition
}
