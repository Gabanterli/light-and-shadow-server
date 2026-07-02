package movement

import (
	"encoding/binary"
	"sync"
)

// Chunk representa um bloco espacial de 32x32 tiles
type Chunk struct {
	ChunkX int    `json:"chunk_x"`
	ChunkY int    `json:"chunk_y"`
	// Grid bidimensional de 32x32 representando IDs de tiles. 
	// 0 = Walkable (Grama/Chão), 1 = Obstáculo (Parede/Pedra)
	Tiles  [32][32]byte `json:"tiles"`
}

// ChunkManager gerencia o carregamento, cache e geração procedural dos chunks
type ChunkManager struct {
	mu     sync.RWMutex
	chunks map[uint64]*Chunk
}

// NewChunkManager inicializa o gerenciador de chunks do servidor
func NewChunkManager() *ChunkManager {
	return &ChunkManager{
		chunks: make(map[uint64]*Chunk),
	}
}

// getChunkKey calcula chave binária do chunk
func (cm *ChunkManager) getChunkKey(cx, cy int) uint64 {
	return (uint64(uint32(cx)) << 32) | uint64(uint32(cy))
}

// GetChunk recupera um chunk do cache ou gera de forma procedural (Lazy Loading)
func (cm *ChunkManager) GetChunk(cx, cy int) *Chunk {
	// Limites do mundo de 16384x16384 tiles (512x512 chunks de 32x32)
	if cx < 0 || cx >= 512 || cy < 0 || cy >= 512 {
		return nil
	}

	key := cm.getChunkKey(cx, cy)

	cm.mu.RLock()
	chunk, exists := cm.chunks[key]
	cm.mu.RUnlock()

	if exists {
		return chunk
	}

	// Geração procedural do chunk se não existir
	cm.mu.Lock()
	// Duplo check sob escrita segura
	if chunk, exists = cm.chunks[key]; exists {
		cm.mu.Unlock()
		return chunk
	}

	chunk = &Chunk{
		ChunkX: cx,
		ChunkY: cy,
	}

	// Popula tiles do chunk de forma inteligente e jogável
	for y := 0; y < 32; y++ {
		globalY := cy*32 + y
		for x := 0; x < 32; x++ {
			globalX := cx*32 + x

			// Garante uma zona segura (Spawn Zone) livre de obstáculos (ex: em torno de coord 100, 100)
			if globalX >= 80 && globalX <= 120 && globalY >= 80 && globalY <= 120 {
				chunk.Tiles[y][x] = 0 // Totalmente livre de colisões
			} else {
				// Adiciona obstáculos em padrão geométrico para simular paredes, árvores e pedras
				if (globalX%11 == 0 && globalY%7 == 0) || (globalX%13 == 0 && globalY%13 == 0) || (globalX%19 == 0 && globalY%5 == 0) {
					chunk.Tiles[y][x] = 1 // Colisão / Bloqueado
				} else {
					chunk.Tiles[y][x] = 0 // Caminhável
				}
			}
		}
	}

	cm.chunks[key] = chunk
	cm.mu.Unlock()

	return chunk
}

// IsBlocked verifica se um tile global específico no mundo de jogo é intransponível
func (cm *ChunkManager) IsBlocked(tileX, tileY int) bool {
	if tileX < 0 || tileX >= 16384 || tileY < 0 || tileY >= 16384 {
		return true // Fora do mapa é considerado obstáculo
	}

	cx := tileX / 32
	cy := tileY / 32
	rx := tileX % 32
	ry := tileY % 32

	chunk := cm.GetChunk(cx, cy)
	if chunk == nil {
		return true
	}

	return chunk.Tiles[ry][rx] == 1
}

// SerializeChunk compacta o chunk em um array binário otimizado para a rede
// Formato: 4 bytes ChunkX (LE), 4 bytes ChunkY (LE), 1024 bytes (tiles 32x32)
func (chunk *Chunk) Serialize() []byte {
	buf := make([]byte, 8+1024)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(chunk.ChunkX))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(chunk.ChunkY))

	offset := 8
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			buf[offset] = chunk.Tiles[y][x]
			offset++
		}
	}
	return buf
}

// GetSurroundingChunks recupera a lista de chunks em uma matriz de 3x3 ao redor de uma coordenada
func (cm *ChunkManager) GetSurroundingChunks(playerX, playerY float64) []*Chunk {
	cx := int(playerX) / 32
	cy := int(playerY) / 32

	var surrounding []*Chunk
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			chunk := cm.GetChunk(cx+dx, cy+dy)
			if chunk != nil {
				surrounding = append(surrounding, chunk)
			}
		}
	}
	return surrounding
}
