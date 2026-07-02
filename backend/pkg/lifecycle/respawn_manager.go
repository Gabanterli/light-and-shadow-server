package lifecycle

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/light-and-shadow/backend/pkg/housing"
)

// HospitalDefinition representa a localização de um hospital ou enfermaria militar
type HospitalDefinition struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ContinentName string  `json:"continent_name"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	Z             int     `json:"z"`
}

// FallbackCoords representa as coordenadas de segurança globais
type FallbackCoords struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z int     `json:"z"`
}

// RespawnConfig engloba a configuração de respawn externa
type RespawnConfig struct {
	Hospitals    []HospitalDefinition `json:"hospitals"`
	SafeFallback FallbackCoords       `json:"safe_fallback"`
}

// RespawnManager coordena o renascimento de jogadores e busca de pontos seguros/hospitalares
type RespawnManager struct {
	mu                 sync.RWMutex
	hospitals          map[string]HospitalDefinition // continent_name -> hospital
	safeFallback       FallbackCoords
	loaded             bool
	housingMgr         *housing.HousingManager
	CanUseHouseRespawn func(playerID string) bool // Callback to check if player can respawn in house (PATCH 6)
}

// NewRespawnManager inicializa o RespawnManager carregando as configurações externas
func NewRespawnManager() *RespawnManager {
	rm := &RespawnManager{
		hospitals: make(map[string]HospitalDefinition),
		safeFallback: FallbackCoords{
			X: 100.0,
			Y: 100.0,
			Z: 0,
		},
		CanUseHouseRespawn: func(playerID string) bool { return true },
	}
	rm.LoadConfig()
	return rm
}

// SetHousingManager injeta a dependência do HousingManager para validação de ativação do respawn (PATCH 6)
func (rm *RespawnManager) SetHousingManager(hm *housing.HousingManager) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.housingMgr = hm
}

// LoadConfig carrega as configurações do arquivo JSON de forma resiliente
func (rm *RespawnManager) LoadConfig() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	paths := []string{"backend/config/", "config/", "../config/", "../../config/"}
	fileName := "hospitals_config.json"
	var data []byte
	var err error
	var finalPath string

	for _, p := range paths {
		filePath := p + fileName
		if _, statErr := os.Stat(filePath); statErr == nil {
			data, err = os.ReadFile(filePath)
			if err == nil {
				finalPath = filePath
				break
			}
		}
	}

	if err != nil || len(data) == 0 {
		slog.Warn("Could not find or read hospitals_config.json, using default safe fallback")
		rm.loadFallbackHospitals()
		return
	}

	var config RespawnConfig
	if err := json.Unmarshal(data, &config); err != nil {
		slog.Error("Failed to parse hospitals_config.json, using fallback", "error", err)
		rm.loadFallbackHospitals()
		return
	}

	rm.hospitals = make(map[string]HospitalDefinition)
	for _, h := range config.Hospitals {
		rm.hospitals[h.ContinentName] = h
	}
	rm.safeFallback = config.SafeFallback
	rm.loaded = true

	slog.Info("Successfully loaded hospitals and fallback coordinates from JSON config", "hospitals_count", len(rm.hospitals), "path", finalPath)
}

func (rm *RespawnManager) loadFallbackHospitals() {
	fallbackList := []HospitalDefinition{
		{"beginner_infirmary", "Beginner Camp Military Infirmary", "Main Continent", 120.0, 120.0, 0},
		{"fire_temple", "Hearth Healing Temple", "Fire Continent", 2110.0, 2110.0, 0},
		{"ice_hospital", "Frostbite Infirmary", "Ice Continent", 2310.0, 2310.0, 0},
		{"holy_cathedral", "Sanctum Cathedral Hospital", "Holy Continent", 2510.0, 2510.0, 0},
		{"shadow_infirmary", "Eclipse Military Infirmary", "Shadow Continent", 2710.0, 2710.0, 0},
		{"nature_shrine", "Wildwood Healing Shrine", "Nature Continent", 2910.0, 2910.0, 0},
		{"abyssia_temple", "Last Bastion Recovery Clinic", "Abyssia", 3410.0, 3110.0, 0},
	}

	for _, h := range fallbackList {
		rm.hospitals[h.ContinentName] = h
	}
	rm.safeFallback = FallbackCoords{X: 100.0, Y: 100.0, Z: 0}
	rm.loaded = true
}

// LookupHospitalByContinent busca o hospital ou enfermaria configurado para um continente
func (rm *RespawnManager) LookupHospitalByContinent(continentName string) (HospitalDefinition, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	hospital, ok := rm.hospitals[continentName]
	return hospital, ok
}

// GetRespawnLocation determina as coordenadas de renascimento do jogador com suporte a casas e continentes
func (rm *RespawnManager) GetRespawnLocation(playerID string, currentContinent string, customHouseID string) (float64, float64, int, string) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// 1. Suporte a ressurreição em casas compradas/alugadas (Feature de Housing) (PATCH 6)
	if rm.housingMgr != nil && (rm.CanUseHouseRespawn == nil || rm.CanUseHouseRespawn(playerID)) {
		if rx, ry, rz, rName, ok := rm.housingMgr.GetPlayerActiveHouseLocation(playerID); ok {
			slog.Info("Resolving validated housing-based respawn", "player", playerID, "house", rName)
			return rx, ry, rz, rName
		}
	}

	if customHouseID != "" {
		// Mock representativo para housing futuro (ex: Carrega as coordenadas da casa do jogador)
		slog.Info("Resolving future housing-based respawn", "player", playerID, "house_id", customHouseID)
		// Vamos assumir coordenadas simuladas para a casa do jogador
		return 150.0, 150.0, 0, fmt.Sprintf("Player House (%s)", customHouseID)
	}

	// 2. Busca hospital/enfermaria local baseada no continente atual do jogador
	if hospital, exists := rm.hospitals[currentContinent]; exists {
		slog.Info("Determining local continent-based hospital respawn location", "player", playerID, "continent", currentContinent, "hospital", hospital.Name)
		return hospital.X, hospital.Y, hospital.Z, hospital.Name
	}

	// 3. Fallback de segurança absoluto se o continente ou hospital não estiver configurado
	slog.Warn("No hospital configuration found for continent, using global safe fallback coordinates", "player", playerID, "continent", currentContinent)
	return rm.safeFallback.X, rm.safeFallback.Y, rm.safeFallback.Z, "Global Safe Fallback Location"
}
