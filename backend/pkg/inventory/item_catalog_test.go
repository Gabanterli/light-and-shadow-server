package inventory

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
)

func validWeaponDefinition() ItemDefinition {
	return ItemDefinition{
		ID:                "sword_test",
		Name:              "Test Sword",
		Type:              ItemTypeWeapon,
		Category:          "Sword",
		Tier:              1,
		BaseDamage:        10,
		CompatibleClasses: []string{"Knight"},
		Stackable:         false,
		MaxStack:          1,
		Description:       "Test definition.",
	}
}

func validMaterialDefinition() ItemDefinition {
	return ItemDefinition{
		ID:        "material_test",
		Name:      "Test Material",
		Type:      ItemTypeMaterial,
		Tier:      0,
		Stackable: true,
		MaxStack:  99,
	}
}

func TestItemCatalogValidAndImmutable(t *testing.T) {
	definitions := []ItemDefinition{
		validWeaponDefinition(),
		validMaterialDefinition(),
	}
	catalog, err := NewItemCatalog(definitions)
	if err != nil {
		t.Fatalf("NewItemCatalog() error = %v", err)
	}

	definitions[0].Name = "mutated"
	definitions[0].CompatibleClasses[0] = "Mage"

	got, ok := catalog.Get("sword_test")
	if !ok {
		t.Fatal("expected sword_test")
	}
	if got.Name != "Test Sword" {
		t.Fatalf("catalog mutated through input slice: %q", got.Name)
	}
	if !reflect.DeepEqual(got.CompatibleClasses, []string{"Knight"}) {
		t.Fatalf("catalog classes mutated through input: %v", got.CompatibleClasses)
	}

	got.CompatibleClasses[0] = "Mage"
	again, ok := catalog.Get("sword_test")
	if !ok {
		t.Fatal("expected sword_test on second read")
	}
	if !reflect.DeepEqual(again.CompatibleClasses, []string{"Knight"}) {
		t.Fatalf("catalog classes mutated through Get result: %v", again.CompatibleClasses)
	}

	ids := catalog.IDs()
	if !sort.StringsAreSorted(ids) {
		t.Fatalf("IDs are not sorted: %v", ids)
	}
	ids[0] = "mutated"
	if catalog.IDs()[0] == "mutated" {
		t.Fatal("IDs exposed internal storage")
	}
	if catalog.Count() != 2 {
		t.Fatalf("Count() = %d, want 2", catalog.Count())
	}
	if _, ok := catalog.Get("missing"); ok {
		t.Fatal("unexpected missing definition")
	}
}

func TestLoadItemCatalogRealSeed(t *testing.T) {
	path := filepath.Join("..", "..", "config", "items.json")
	catalog, source, err := LoadItemCatalogSource(path)
	if err != nil {
		t.Fatalf("LoadItemCatalogSource() error = %v", err)
	}
	if catalog.Count() != 23 {
		t.Fatalf("Count() = %d, want 23", catalog.Count())
	}
	if len(source) == 0 {
		t.Fatal("expected source JSON")
	}
	if _, ok := catalog.Get("sword_t1_rusty"); !ok {
		t.Fatal("expected sword_t1_rusty")
	}
}

func TestItemDefinitionValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ItemDefinition)
	}{
		{"empty ID", func(d *ItemDefinition) { d.ID = "" }},
		{"uppercase ID", func(d *ItemDefinition) { d.ID = "Sword_Test" }},
		{"spaced ID", func(d *ItemDefinition) { d.ID = " sword_test" }},
		{"empty Name", func(d *ItemDefinition) { d.Name = " " }},
		{"invalid Type", func(d *ItemDefinition) { d.Type = "quest" }},
		{"missing Category", func(d *ItemDefinition) { d.Category = "" }},
		{"invalid equipment Tier", func(d *ItemDefinition) { d.Tier = 0 }},
		{"invalid MaxStack", func(d *ItemDefinition) { d.MaxStack = 0 }},
		{"non-stackable MaxStack", func(d *ItemDefinition) { d.MaxStack = 2 }},
		{"equipment stackable", func(d *ItemDefinition) { d.Stackable = true; d.MaxStack = 2 }},
		{"negative BaseDamage", func(d *ItemDefinition) { d.BaseDamage = -1 }},
		{"CritBonus above one", func(d *ItemDefinition) { d.CritBonus = 1.01 }},
		{"NaN", func(d *ItemDefinition) { d.BaseDamage = math.NaN() }},
		{"infinity", func(d *ItemDefinition) { d.BaseDamage = math.Inf(1) }},
		{"invalid class", func(d *ItemDefinition) { d.CompatibleClasses = []string{"Paladin"} }},
		{"duplicate class", func(d *ItemDefinition) { d.CompatibleClasses = []string{"Knight", "Knight"} }},
		{"whitespace Description", func(d *ItemDefinition) { d.Description = " " }},
		{"negative elemental stat", func(d *ItemDefinition) { d.ElementalStats.FireAttack = -1 }},
		{"negative Price", func(d *ItemDefinition) { d.Price = -1 }},
		{"negative MarketPrice", func(d *ItemDefinition) { d.MarketPrice = -1 }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validWeaponDefinition()
			test.mutate(&definition)
			if _, err := NewItemCatalog([]ItemDefinition{definition}); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestItemCatalogRejectsDuplicateIDAndEmptyCatalog(t *testing.T) {
	definition := validWeaponDefinition()
	if _, err := NewItemCatalog(nil); err == nil {
		t.Fatal("expected empty catalog error")
	}
	if _, err := NewItemCatalog([]ItemDefinition{definition, definition}); err == nil {
		t.Fatal("expected duplicate ID error")
	}
}

func TestItemCatalogMaterialStackRules(t *testing.T) {
	definition := validMaterialDefinition()
	definition.MaxStack = 1
	if _, err := NewItemCatalog([]ItemDefinition{definition}); err == nil {
		t.Fatal("expected stackable MaxStack validation error")
	}

	definition = validMaterialDefinition()
	definition.Stackable = false
	definition.MaxStack = 2
	if _, err := NewItemCatalog([]ItemDefinition{definition}); err == nil {
		t.Fatal("expected non-stackable MaxStack validation error")
	}

	definition = validMaterialDefinition()
	definition.Tier = 6
	if _, err := NewItemCatalog([]ItemDefinition{definition}); err == nil {
		t.Fatal("expected material tier validation error")
	}
}

func TestDecodeItemCatalogStrictJSON(t *testing.T) {
	valid := `[
		{
			"ID":"material_test",
			"Name":"Test Material",
			"Type":"material",
			"Tier":0,
			"Stackable":true,
			"MaxStack":99
		}
	]`

	tests := []struct {
		name string
		data string
	}{
		{"empty", ""},
		{"empty array", "[]"},
		{"unknown item field", strings.Replace(valid, `"Name"`, `"Unknown"`, 1)},
		{"unknown elemental field", `[
			{
				"ID":"material_test",
				"Name":"Test Material",
				"Type":"material",
				"Tier":0,
				"Stackable":true,
				"MaxStack":99,
				"ElementalStats":{"LightningAttack":1}
			}
		]`},
		{"trailing JSON", valid + `{}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := DecodeItemCatalog([]byte(test.data)); err == nil {
				t.Fatal("expected decode error")
			}
		})
	}
}

func TestLoadItemCatalogFileErrors(t *testing.T) {
	if _, err := LoadItemCatalog(""); err == nil {
		t.Fatal("expected empty path error")
	}
	if _, err := LoadItemCatalog(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected missing file error")
	}

	dir := t.TempDir()
	emptyPath := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(emptyPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadItemCatalog(emptyPath); err == nil {
		t.Fatal("expected empty file error")
	}

	largePath := filepath.Join(dir, "large.json")
	large := bytes.Repeat([]byte("x"), int(MaxItemCatalogBytes)+1)
	if err := os.WriteFile(largePath, large, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadItemCatalog(largePath); err == nil {
		t.Fatal("expected oversized file error")
	}
}

func TestItemCatalogConcurrentReads(t *testing.T) {
	catalog, err := NewItemCatalog([]ItemDefinition{
		validWeaponDefinition(),
		validMaterialDefinition(),
	})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for worker := 0; worker < 64; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for iteration := 0; iteration < 1000; iteration++ {
				if _, ok := catalog.Get("sword_test"); !ok {
					t.Errorf("missing definition")
					return
				}
				if got := catalog.Count(); got != 2 {
					t.Errorf("Count() = %d", got)
					return
				}
				if got := len(catalog.IDs()); got != 2 {
					t.Errorf("len(IDs()) = %d", got)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestDefaultItemCatalogSubprocess(t *testing.T) {
	mode := os.Getenv("LS_ITEM_CATALOG_SUBPROCESS")
	if mode == "" {
		return
	}

	if _, ok := GetItemDef("sword_test"); ok {
		t.Fatal("GetItemDef unexpectedly succeeded before installation")
	}
	if err := InstallDefaultItemCatalog(nil); err == nil {
		t.Fatal("expected nil install error")
	}

	catalog, err := NewItemCatalog([]ItemDefinition{validWeaponDefinition()})
	if err != nil {
		t.Fatal(err)
	}
	if err := InstallDefaultItemCatalog(catalog); err != nil {
		t.Fatalf("first install error = %v", err)
	}
	if err := InstallDefaultItemCatalog(catalog); err == nil {
		t.Fatal("expected second install error")
	}
	if _, ok := GetItemDef("sword_test"); !ok {
		t.Fatal("GetItemDef failed after installation")
	}
}

func TestDefaultItemCatalogInstallation(t *testing.T) {
	command := exec.Command(os.Args[0], "-test.run=^TestDefaultItemCatalogSubprocess$")
	command.Env = append(os.Environ(), "LS_ITEM_CATALOG_SUBPROCESS=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v\n%s", err, output)
	}
}
