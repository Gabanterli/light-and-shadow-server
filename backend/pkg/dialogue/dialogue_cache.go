package dialogue

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"
)

// Condition representa as condições de visibilidade de uma resposta de diálogo
type Condition struct {
	Level   int    `json:"level,omitempty"`
	QuestID string `json:"quest_id,omitempty"`
	Status  string `json:"status,omitempty"` // "not_started", "active", "ready_to_complete", "completed"
}

// QuestTrigger define um gatilho de aceitação ou entrega de quest associado a um nó de diálogo
type QuestTrigger struct {
	Action  string `json:"action"`   // "accept", "complete"
	QuestID string `json:"quest_id"`
}

// Response representa uma resposta selecionável do jogador a um diálogo
type Response struct {
	Text       string    `json:"text"`
	NextNodeID string    `json:"next_node_id"`
	Condition  Condition `json:"condition"`
}

// DialogueNode representa um único slide ou balão de fala do NPC
type DialogueNode struct {
	NodeID       string        `json:"node_id"`
	Text         string        `json:"text"`
	QuestTrigger *QuestTrigger `json:"quest_trigger,omitempty"`
	Responses    []Response    `json:"responses"`
}

// DialogueTree representa a árvore de diálogos completa associada a um NPC
type DialogueTree struct {
	NPCID string         `json:"npc_id"`
	Nodes []DialogueNode `json:"nodes"`
}

// DialogueCache gerencia o carregamento de diálogos em cache para evitar JSON parsing em tempo de execução
type DialogueCache struct {
	trees       map[string]*DialogueTree
	mu          sync.RWMutex
	filePath    string
	lastModTime time.Time
}

// NewDialogueCache inicializa o DialogueCache
func NewDialogueCache() *DialogueCache {
	dc := &DialogueCache{
		trees: make(map[string]*DialogueTree),
	}
	// Encontra o caminho correto para dialogues.json
	paths := []string{"backend/config/dialogues.json", "config/dialogues.json", "../config/dialogues.json", "/backend/config/dialogues.json"}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			dc.filePath = p
			break
		}
	}
	if dc.filePath == "" {
		slog.Error("Failed to find dialogues.json for DialogueCache")
	} else {
		dc.LoadAll()
		// Inicia hot-reload periódico
		go dc.startWatcher()
	}
	return dc
}

// LoadAll carrega todos os diálogos de dialogues.json para a memória
func (dc *DialogueCache) LoadAll() {
	if dc.filePath == "" {
		return
	}
	info, err := os.Stat(dc.filePath)
	if err != nil {
		slog.Error("Failed to stat dialogues.json", "error", err)
		return
	}

	data, err := os.ReadFile(dc.filePath)
	if err != nil {
		slog.Error("Failed to read dialogues.json", "error", err)
		return
	}

	var list []DialogueTree
	if err := json.Unmarshal(data, &list); err != nil {
		slog.Error("Failed to unmarshal dialogues.json", "error", err)
		return
	}

	dc.mu.Lock()
	dc.trees = make(map[string]*DialogueTree)
	for i := range list {
		tree := list[i]
		dc.trees[tree.NPCID] = &tree
	}
	dc.lastModTime = info.ModTime()
	dc.mu.Unlock()

	slog.Info("DialogueCache loaded/reloaded", "count", len(list), "path", dc.filePath)
}

// GetDialogueTree recupera a árvore de diálogos de um NPC direto da memória
func (dc *DialogueCache) GetDialogueTree(npcID string) (*DialogueTree, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	tree, ok := dc.trees[npcID]
	return tree, ok
}

// startWatcher implementa Hot Reload de dialogues.json
func (dc *DialogueCache) startWatcher() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if dc.filePath == "" {
			continue
		}
		info, err := os.Stat(dc.filePath)
		if err != nil {
			continue
		}
		dc.mu.RLock()
		stale := info.ModTime().After(dc.lastModTime)
		dc.mu.RUnlock()

		if stale {
			slog.Info("dialogues.json changed, reloading DialogueCache...", "path", dc.filePath)
			dc.LoadAll()
		}
	}
}
