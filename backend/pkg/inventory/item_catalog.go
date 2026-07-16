package inventory

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync/atomic"
)

const MaxItemCatalogBytes int64 = 1 << 20

type ItemCatalog struct {
	definitions map[string]ItemDefinition
	orderedIDs  []string
}

var defaultItemCatalog atomic.Pointer[ItemCatalog]

func NewItemCatalog(definitions []ItemDefinition) (*ItemCatalog, error) {
	if len(definitions) == 0 {
		return nil, errors.New("item catalog must contain at least one definition")
	}

	catalog := &ItemCatalog{
		definitions: make(map[string]ItemDefinition, len(definitions)),
		orderedIDs:  make([]string, 0, len(definitions)),
	}

	for index, definition := range definitions {
		if err := definition.validate(); err != nil {
			return nil, fmt.Errorf(
				"invalid item definition at index %d (ID %q): %w",
				index,
				definition.ID,
				err,
			)
		}
		if _, duplicate := catalog.definitions[definition.ID]; duplicate {
			return nil, fmt.Errorf(
				"duplicate item definition ID %q at index %d",
				definition.ID,
				index,
			)
		}

		catalog.definitions[definition.ID] = cloneItemDefinition(definition)
		catalog.orderedIDs = append(catalog.orderedIDs, definition.ID)
	}

	sort.Strings(catalog.orderedIDs)
	return catalog, nil
}

func LoadItemCatalog(path string) (*ItemCatalog, error) {
	catalog, _, err := LoadItemCatalogSource(path)
	return catalog, err
}

// LoadItemCatalogSource loads and validates a catalog while returning a copy of
// the exact source bytes for legacy consumers that still require the JSON.
func LoadItemCatalogSource(path string) (*ItemCatalog, []byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil, errors.New("item catalog path is required")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open item catalog %q: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("stat item catalog %q: %w", path, err)
	}
	if info.Size() > MaxItemCatalogBytes {
		return nil, nil, fmt.Errorf(
			"item catalog %q exceeds %d bytes",
			path,
			MaxItemCatalogBytes,
		)
	}

	data, err := io.ReadAll(io.LimitReader(file, MaxItemCatalogBytes+1))
	if err != nil {
		return nil, nil, fmt.Errorf("read item catalog %q: %w", path, err)
	}
	if int64(len(data)) > MaxItemCatalogBytes {
		return nil, nil, fmt.Errorf(
			"item catalog %q exceeds %d bytes",
			path,
			MaxItemCatalogBytes,
		)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil, fmt.Errorf("item catalog %q is empty", path)
	}

	catalog, err := DecodeItemCatalog(data)
	if err != nil {
		return nil, nil, fmt.Errorf("decode item catalog %q: %w", path, err)
	}
	return catalog, append([]byte(nil), data...), nil
}

func DecodeItemCatalog(data []byte) (*ItemCatalog, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, errors.New("item catalog JSON is empty")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var definitions []ItemDefinition
	if err := decoder.Decode(&definitions); err != nil {
		return nil, fmt.Errorf("decode item definitions: %w", err)
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("unexpected trailing JSON after item catalog")
		}
		return nil, fmt.Errorf("decode trailing item catalog data: %w", err)
	}

	return NewItemCatalog(definitions)
}

func (c *ItemCatalog) Get(id string) (ItemDefinition, bool) {
	if c == nil {
		return ItemDefinition{}, false
	}
	definition, exists := c.definitions[id]
	if !exists {
		return ItemDefinition{}, false
	}
	return cloneItemDefinition(definition), true
}

func (c *ItemCatalog) Count() int {
	if c == nil {
		return 0
	}
	return len(c.definitions)
}

func (c *ItemCatalog) IDs() []string {
	if c == nil {
		return nil
	}
	return append([]string(nil), c.orderedIDs...)
}

func InstallDefaultItemCatalog(catalog *ItemCatalog) error {
	if catalog == nil {
		return errors.New("default item catalog cannot be nil")
	}
	if !defaultItemCatalog.CompareAndSwap(nil, catalog) {
		return errors.New("default item catalog is already installed")
	}
	return nil
}

func GetItemDef(itemID string) (ItemDefinition, bool) {
	catalog := defaultItemCatalog.Load()
	if catalog == nil {
		return ItemDefinition{}, false
	}
	return catalog.Get(itemID)
}
