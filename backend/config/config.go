package config

import (
	"os"
	"strconv"
)

type Config struct {
	GatewayPort                  int
	AuthPort                     int
	WorldPort                    int
	PostgresDSN                  string
	RedisAddr                    string
	RedisPassword                string
	RedisDB                      int
	LogLevel                     string
	EnableDevGM                  bool
	WorldMapMode                 string
	WorldMapManifestPath         string
	WorldMapInitialResidencyPath string
}

func LoadConfig() *Config {
	gatewayPort, _ := strconv.Atoi(getEnv("GATEWAY_PORT", "8080"))
	authPort, _ := strconv.Atoi(getEnv("AUTH_PORT", "8081"))
	worldPort, _ := strconv.Atoi(getEnv("WORLD_PORT", "8082"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	enableDevGM, _ := strconv.ParseBool(getEnv("LS_ENABLE_DEV_GM", "false"))

	worldMapMode := getEnv("LS_WORLD_MAP_MODE", "debug")
	worldMapManifestPath := getEnv(
		"LS_WORLD_MAP_MANIFEST_PATH",
		"config/worldmap/world_manifest.json",
	)
	worldMapInitialResidencyPath := getEnv(
		"LS_WORLD_MAP_INITIAL_RESIDENCY_PATH",
		"config/worldmap/initial_residency.json",
	)
	return &Config{
		GatewayPort:                  gatewayPort,
		AuthPort:                     authPort,
		WorldPort:                    worldPort,
		PostgresDSN:                  getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/light_and_shadow?sslmode=disable"),
		RedisAddr:                    getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:                getEnv("REDIS_PASSWORD", ""),
		RedisDB:                      redisDB,
		LogLevel:                     getEnv("LOG_LEVEL", "info"),
		EnableDevGM:                  enableDevGM,
		WorldMapMode:                 worldMapMode,
		WorldMapManifestPath:         worldMapManifestPath,
		WorldMapInitialResidencyPath: worldMapInitialResidencyPath,
	}
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}
