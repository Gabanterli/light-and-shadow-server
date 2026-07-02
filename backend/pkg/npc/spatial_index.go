package npc

import (
	"sync"
)

// GridCell define a chave espacial para uma célula 3x3 ou correspondente a chunks (32x32 tiles)
type GridCell struct {
	X int
	Y int
	Z int
}

// NPCSpatialIndex gerencia uma grade espacial para rápido lookup e validação de proximidade dos NPCs (PATCH 3)
type NPCSpatialIndex struct {
	mu    sync.RWMutex
	cells map[GridCell][]string // GridCell -> Lista de NPCIDs
}

// NewNPCSpatialIndex cria um novo índice espacial de NPCs
func NewNPCSpatialIndex() *NPCSpatialIndex {
	return &NPCSpatialIndex{
		cells: make(map[GridCell][]string),
	}
}

// getCellKey calcula a GridCell correspondente a coordenadas (X, Y) do mundo (chunk 32x32)
func (nsi *NPCSpatialIndex) getCellKey(x, y float64, z int) GridCell {
	return GridCell{
		X: int(x) / 32,
		Y: int(y) / 32,
		Z: z,
	}
}

// Clear limpa todas as entradas do índice
func (nsi *NPCSpatialIndex) Clear() {
	nsi.mu.Lock()
	defer nsi.mu.Unlock()
	nsi.cells = make(map[GridCell][]string)
}

// Insert insere um NPC no índice espacial baseado em sua localização
func (nsi *NPCSpatialIndex) Insert(npcID string, x, y float64, z int) {
	nsi.mu.Lock()
	defer nsi.mu.Unlock()

	cell := nsi.getCellKey(x, y, z)
	nsi.cells[cell] = append(nsi.cells[cell], npcID)
}

// GetNearbyNPCs busca todos os NPCs localizados nas células adjacentes (incluindo a célula central)
func (nsi *NPCSpatialIndex) GetNearbyNPCs(x, y float64, z int) []string {
	nsi.mu.RLock()
	defer nsi.mu.RUnlock()

	centerCell := nsi.getCellKey(x, y, z)
	var nearby []string

	// Consulta em matriz 3x3 ao redor da célula do jogador
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			cell := GridCell{
				X: centerCell.X + dx,
				Y: centerCell.Y + dy,
				Z: centerCell.Z,
			}
			if list, exists := nsi.cells[cell]; exists {
				nearby = append(nearby, list...)
			}
		}
	}
	return nearby
}
