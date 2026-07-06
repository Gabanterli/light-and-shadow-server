package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/config"
	"github.com/light-and-shadow/backend/pkg/blessing"
	"github.com/light-and-shadow/backend/pkg/charactercreation"
	"github.com/light-and-shadow/backend/pkg/combat"
	"github.com/light-and-shadow/backend/pkg/db"
	"github.com/light-and-shadow/backend/pkg/dungeon"
	"github.com/light-and-shadow/backend/pkg/economy"
	"github.com/light-and-shadow/backend/pkg/gamedata/rules"
	"github.com/light-and-shadow/backend/pkg/housing"
	"github.com/light-and-shadow/backend/pkg/inventory"
	"github.com/light-and-shadow/backend/pkg/lifecycle"
	"github.com/light-and-shadow/backend/pkg/logger"
	"github.com/light-and-shadow/backend/pkg/messaging"
	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/npc"
	"github.com/light-and-shadow/backend/pkg/persistence"
	"github.com/light-and-shadow/backend/pkg/professions"
	"github.com/light-and-shadow/backend/pkg/progression"
	"github.com/light-and-shadow/backend/pkg/protocol"
	"github.com/light-and-shadow/backend/pkg/pve"
	"github.com/light-and-shadow/backend/pkg/pvp"
	"github.com/light-and-shadow/backend/pkg/quest"
	"github.com/light-and-shadow/backend/pkg/social"

	"database/sql"
)

type GatewayServer struct {
	config                   *config.Config
	tcpListener              net.Listener
	httpServer               *http.Server
	pgPool                   *db.PostgresPool
	redisClient              *db.RedisClient
	clientsMu                sync.Mutex
	clients                  map[net.Conn]bool
	wg                       sync.WaitGroup
	spatialIndex             *movement.SpatialIndex
	chunkManager             *movement.ChunkManager
	aoiManager               *movement.AOIManager
	movementSystem           *movement.MovementSystem
	combatManager            *combat.CombatManager
	persistenceMgr           *persistence.PersistenceManager
	characterCreationService *charactercreation.Service
	pveManager               *pve.PveManager
	questManager             *quest.QuestManager
	npcManager               *npc.NPCManager
	socialManager            *social.SocialManager
	economyManager           *economy.EconomyManager
	professionsManager       *professions.ProfessionsManager
	dungeonManager           *dungeon.DungeonManager
	progressionManager       *progression.ProgressionManager
	blessingManager          *blessing.BlessingManager
	respawnManager           *lifecycle.RespawnManager
	deathPenaltyManager      *lifecycle.DeathPenaltyManager
	housingManager           *housing.HousingManager
	pvpManager               *pvp.PvPManager
	activeGatheringsMu       sync.Mutex
	activeGatherings         map[string]string // playerID -> nodeID
	inventoriesMu            sync.RWMutex
	inventories              map[string]*inventory.PlayerInventory
	stopAutosave             chan struct{}
}

type gatewayAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type gatewayAuthResponse struct {
	Success   bool   `json:"success"`
	Token     string `json:"token,omitempty"`
	AccountID int    `json:"account_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (s *GatewayServer) authenticateWithAuthServer(ctx context.Context, username string, password string) (*gatewayAuthResponse, error) {
	requestBody, err := json.Marshal(gatewayAuthRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode auth request: %w", err)
	}

	authURL := fmt.Sprintf("http://auth-server:%d/api/v1/auth", s.config.AuthPort)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth server request failed: %w", err)
	}
	defer resp.Body.Close()

	var authResp gatewayAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || !authResp.Success || authResp.Token == "" {
		if authResp.Error == "" {
			authResp.Error = "auth_failed"
		}
		return nil, fmt.Errorf("auth rejected: %s", authResp.Error)
	}

	return &authResp, nil
}
func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.LogLevel)

	slog.Info("Starting Light and Shadow Gateway Server...")

	// InicializaÃ§Ã£o de bancos de dados (tolerante a fallbacks locais)
	pgPool, err := db.NewPostgresPool(cfg.PostgresDSN)
	if err != nil {
		slog.Error("PostgreSQL pool initialization failed; refusing to start Gateway", "error", err)
		os.Exit(1)
	}

	redisClient, err := db.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		slog.Error("Redis client initialization failed; refusing to start Gateway", "error", err)
		os.Exit(1)
	}

	// Inicializa e configura Sistemas de Movimento e AOI (Sprint 2 Task 4)
	spatialIndex := movement.NewSpatialIndex()
	chunkManager := movement.NewChunkManager()
	aoiManager := movement.NewAOIManager(spatialIndex)
	movementSystem := movement.NewMovementSystem(spatialIndex, chunkManager, aoiManager)

	// Inicializa Sistema de Combate Autorizativo (Sprint 2 Task 5)
	combatManager := combat.NewCombatManager(chunkManager)

	// Configura LevelProvider para checagens de regiÃ£o da Sprint 3 Task 5 (PATCH 1)
	movementSystem.LevelProvider = func(playerID string) int {
		if stats, exists := combatManager.GetEntityStats(playerID); exists {
			return stats.Level
		}
		return 1
	}

	// Inicializa Sistema de Combate Autorizativo (Sprint 2 Task 5)
	combatManager = combat.NewCombatManager(chunkManager)

	// Configura manipuladores de eventos de combate assÃ­ncronos (projÃ©teis)
	combatManager.SetEventHandler(combat.CombatEventHandler{
		OnDamage: func(attackerID, targetID string, damage float64, isCrit, isHit bool, skillName string) {
			dmgPayload := protocol.EncodeDamageEvent(attackerID, targetID, damage, isCrit, isHit, true, skillName)
			aoiManager.BroadcastCombat(attackerID, protocol.SC_DAMAGE_EVENT, dmgPayload)
		},
		OnTargetDead: func(targetID string) {
			deadPayload := protocol.EncodeTargetDeadEvent(targetID)
			aoiManager.BroadcastCombat(targetID, protocol.SC_TARGET_DEAD, deadPayload)
		},
	})

	// Inicializa e configura PersistenceManager e Esquemas do DB
	persistenceMgr := persistence.NewPersistenceManager(pgPool)
	if err := persistenceMgr.InitSchema(); err != nil {
		slog.Error("Failed to initialize database schema; refusing to start Gateway", "error", err)
		os.Exit(1)
	}

	// Initialize Character Creation Service (R1-M-E)
	ruleRegistry, err := rules.NewDefaultRegistry()
	if err != nil {
		slog.Error("Failed to create default rule registry; refusing to start Gateway", "error", err)
		os.Exit(1)
	}
	raceValidator := charactercreation.NewRuleRegistryRaceValidator(ruleRegistry)
	characterCreationService := charactercreation.NewService(persistenceMgr, raceValidator)

	// Inicializa e configura Gateway
	server := &GatewayServer{
		config:                   cfg,
		pgPool:                   pgPool,
		redisClient:              redisClient,
		clients:                  make(map[net.Conn]bool),
		spatialIndex:             spatialIndex,
		chunkManager:             chunkManager,
		aoiManager:               aoiManager,
		movementSystem:           movementSystem,
		combatManager:            combatManager,
		persistenceMgr:           persistenceMgr,
		characterCreationService: characterCreationService,
		inventories:              make(map[string]*inventory.PlayerInventory),
		activeGatherings:         make(map[string]string),
		stopAutosave:             make(chan struct{}),
	}

	// Inicializa e configura PveManager (Sprint 3 Task 2)
	pveMgr := pve.NewPveManager(spatialIndex, aoiManager, combatManager, server.inventories)
	pveMgr.RegisterLevelUpCallback(func(playerID string, level int, stats *combat.EntityStats) {
		slog.Info("Broadcasting level up event and syncing stats", "player", playerID, "level", level)
		if conn, exists := aoiManager.GetPlayerConn(playerID); exists {
			server.inventoriesMu.RLock()
			playerInv, existsInv := server.inventories[playerID]
			server.inventoriesMu.RUnlock()
			if existsInv {
				server.sendInventorySync(conn, playerID, stats, playerInv)
			}
		}
	})

	// Inicializa e configura NPC e Quest System (Sprint 3 Task 3)
	var sqlDB *sql.DB
	if pgPool != nil {
		sqlDB = pgPool.DB
	}
	questManager := quest.NewQuestManager(sqlDB, combatManager, server.inventories)
	npcManager := npc.NewNPCManager(combatManager)

	socialManager := social.NewSocialManager(sqlDB, aoiManager)
	server.socialManager = socialManager
	pveMgr.RegisterGetPartyMembersCallback(socialManager.GetSharedXpPlayers)

	server.questManager = questManager
	server.npcManager = npcManager

	// Inicializa e configura ProgressionManager (Sprint 3 Task 5)
	progressionManager := progression.NewProgressionManager(sqlDB, combatManager, server.inventories)
	server.progressionManager = progressionManager

	// Inicializa e configura EconomyManager (Sprint 3 Task 4)
	var rawItemsJSON []byte
	itemsPaths := []string{"backend/config/items.json", "config/items.json", "../config/items.json"}
	for _, p := range itemsPaths {
		if data, err := os.ReadFile(p); err == nil {
			rawItemsJSON = data
			break
		}
	}
	server.economyManager = economy.NewEconomyManager(sqlDB, movementSystem, rawItemsJSON)

	// Inicializa e configura ProfessionsManager (Sprint 4 Task 1)
	server.professionsManager = professions.NewProfessionsManager(sqlDB, combatManager)
	server.professionsManager.RegisterCallbacks(
		func(nodeID string, depleted bool) {
			depValue := uint8(0)
			if depleted {
				depValue = 1
			}
			packet := &protocol.Packet{
				Opcode:  protocol.SC_GATHER_COMPLETE,
				Payload: protocol.EncodeGatherComplete(nodeID, "", 0, depValue),
			}
			server.broadcastToAll(packet)
		},
		func(playerID string, prof string, level int, xp int) {
			if conn, ok := server.aoiManager.GetPlayerConn(playerID); ok {
				packet := &protocol.Packet{
					Opcode:  protocol.SC_PROFESSION_XP_UPDATE,
					Payload: protocol.EncodeProfessionXPUpdate(prof, uint32(level), uint32(xp)),
				}
				conn.Write(packet.Serialize())
			}
		},
	)

	// Registra callbacks de atualizaÃ§Ã£o de quests por rede
	questManager.RegisterQuestUpdateCallback(func(pID string, qID string, state *quest.CharacterQuestState) {
		if pConn, ok := aoiManager.GetPlayerConn(pID); ok {
			if state == nil {
				// Quest abandonada
				update := &protocol.QuestUpdateEvent{
					QuestID:    qID,
					Status:     "abandoned",
					Objectives: []protocol.ProtocolObjectiveState{},
				}
				packet := &protocol.Packet{
					Opcode:  protocol.SC_QUEST_UPDATE,
					Payload: protocol.EncodeQuestUpdate(update),
				}
				pConn.Write(packet.Serialize())
			} else {
				objectives := make([]protocol.ProtocolObjectiveState, 0, len(state.Objectives))
				for idx, obj := range state.Objectives {
					objectives = append(objectives, protocol.ProtocolObjectiveState{
						Index:      uint16(idx),
						CurrentQty: uint32(obj.CurrentQty),
					})
				}
				update := &protocol.QuestUpdateEvent{
					QuestID:    qID,
					Status:     state.Status,
					Objectives: objectives,
				}
				packet := &protocol.Packet{
					Opcode:  protocol.SC_QUEST_UPDATE,
					Payload: protocol.EncodeQuestUpdate(update),
				}
				pConn.Write(packet.Serialize())

				if state.Status == "completed" {
					completePacket := &protocol.Packet{
						Opcode:  protocol.SC_QUEST_COMPLETE,
						Payload: protocol.EncodeQuestComplete(qID),
					}
					pConn.Write(completePacket.Serialize())
				}
			}
		}
	})

	// Registra o hook de morte de monstros publicando no MessageBus (PATCH 5)
	pveMgr.RegisterMonsterKilledCallback(func(pID string, monsterTemplateID string) {
		messaging.GetInstance().Publish("monster_killed", messaging.MonsterKilledPayload{
			PlayerID:  pID,
			MonsterID: monsterTemplateID,
		})
	})

	// Registra o hook de saque de itens publicando no MessageBus (PATCH 5)
	pveMgr.RegisterItemLootedCallback(func(pID string, itemID string, qty int) {
		messaging.GetInstance().Publish("item_looted", messaging.ItemLootedPayload{
			PlayerID: pID,
			ItemID:   itemID,
			Qty:      qty,
		})
	})

	// Inicializa os gerenciadores de BÃªnÃ§Ã£os, Respawn e Penalidades de Morte (Sprint 5)
	blessingManager := blessing.NewBlessingManager(sqlDB)
	server.blessingManager = blessingManager

	respawnManager := lifecycle.NewRespawnManager()
	server.respawnManager = respawnManager

	housingManager := housing.NewHousingManager(sqlDB)
	server.housingManager = housingManager
	respawnManager.SetHousingManager(housingManager)

	pvpManager := pvp.NewPvPManager(sqlDB, housingManager, blessingManager, combatManager, server.inventories)
	server.pvpManager = pvpManager
	combatManager.PvPValidator = pvpManager.ValidatePvPPermission
	respawnManager.CanUseHouseRespawn = pvpManager.CanUseHouseRespawn

	deathPenaltyManager := lifecycle.NewDeathPenaltyManager(sqlDB, blessingManager, server.professionsManager, respawnManager, housingManager, server.inventories, combatManager)
	server.deathPenaltyManager = deathPenaltyManager
	deathPenaltyManager.SetPvPManager(pvpManager)

	pveMgr.Start()
	server.pveManager = pveMgr

	// Inicializa e configura DungeonManager (Sprint 4 Task 2)
	server.dungeonManager = dungeon.NewDungeonManager(sqlDB, combatManager, spatialIndex, aoiManager, server.inventories)
	server.dungeonManager.RegisterGetPartyMembersCallback(func(playerID string, x, y float64) []string {
		return socialManager.GetPartyMembers(playerID)
	})
	server.dungeonManager.Start()

	// Inicia Loop de Autosave a cada 30 segundos
	server.startAutosaveLoop()

	lifecycleMgr := lifecycle.NewManager()

	// Inicia HTTP Server para /health
	server.startHTTPServer()

	// Inicia TCP Server para clientes do MMORPG
	server.startTCPServer()

	// Registro de Encerramento Gracioso
	lifecycleMgr.Register(server.Shutdown)

	// Aguarda sinais de OS
	lifecycleMgr.Wait()
}

func (s *GatewayServer) startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "UP", "service": "gateway"}`))
	})

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.GatewayPort+1000), // Ex: 9080 se gateway estÃ¡ na 8080
		Handler: mux,
	}

	go func() {
		slog.Info("Gateway HTTP Health Server running", "port", s.config.GatewayPort+1000)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Gateway HTTP server failed", "error", err)
		}
	}()
}

func (s *GatewayServer) startTCPServer() {
	addr := fmt.Sprintf("0.0.0.0:%d", s.config.GatewayPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Failed to bind TCP listener", "addr", addr, "error", err)
		return
	}
	s.tcpListener = listener

	go func() {
		slog.Info("Gateway TCP Server listening for clients", "addr", addr)
		for {
			conn, err := s.tcpListener.Accept()
			if err != nil {
				// Listener closed
				break
			}
			s.wg.Add(1)
			go s.handleClient(conn)
		}
	}()
}

func (s *GatewayServer) handleClient(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	var sessionToken string
	var authenticatedAccountID int
	var playerID string
	var lastRefresh time.Time

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		// Sessão de autenticação preservada no Redis até TTL ou logout explícito.
		// O disconnect limpa estado ativo do player, mas não revoga a sessão.
		if sessionToken != "" {
			slog.Info("Client disconnected; auth session preserved until TTL", "account_id", authenticatedAccountID)
		}

		// Desregistra jogador do motor de movimentos e AOI para liberar recursos de rede
		if playerID != "" {
			// Salva o personagem antes de remover as referÃªncias in-memory (Save on logout/disconnect)
			slog.Info("Saving player state on disconnect / logout...", "player", playerID)
			s.saveCharacterState(playerID)

			s.inventoriesMu.Lock()
			delete(s.inventories, playerID)
			s.inventoriesMu.Unlock()

			s.movementSystem.RemovePlayerState(playerID)
			s.aoiManager.DeregisterPlayer(playerID)
			s.questManager.CleanPlayerState(playerID)
			s.socialManager.OnPlayerLogout(playerID)
			slog.Info("Cleaned up player states from systems on disconnect", "player", playerID)
		}
	}()

	slog.Info("Client connected to Gateway", "remote_addr", conn.RemoteAddr().String())

	for {
		packet, err := protocol.ReadPacket(conn)
		if err != nil {
			slog.Info("Client disconnected from Gateway", "remote_addr", conn.RemoteAddr().String(), "reason", err.Error())
			break
		}

		slog.Info("Received packet", "opcode", packet.Opcode, "size", packet.Size, "seq", packet.Sequence)

		// Refresh automÃ¡tico de sessÃ£o a cada 60s (Sliding Window) (PATCH 3)
		if sessionToken != "" && s.redisClient != nil && time.Since(lastRefresh) >= 60*time.Second {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err := s.redisClient.Client.Expire(ctx, sessionToken, 2*time.Hour).Err()
			cancel()
			if err != nil {
				slog.Error("Failed to automatically refresh session", "token", sessionToken, "error", err)
			} else {
				lastRefresh = time.Now()
				slog.Info("Sliding window session refreshed successfully", "token", sessionToken)
			}
		}

		// LÃ³gica do protocolo Gateway
		switch packet.Opcode {
		case protocol.CS_HEARTBEAT:
			// Responder Ack de Heartbeat imediatamente para manter conexÃ£o viva
			ack := &protocol.Packet{
				Opcode:   protocol.SC_HEARTBEAT_ACK,
				Sequence: packet.Sequence,
			}
			conn.Write(ack.Serialize())
			slog.Debug("Sent Heartbeat Ack", "seq", packet.Sequence)

		case protocol.CS_LOGIN_REQUEST:
			slog.Info("Routing login request to Auth Server")

			loginReq, err := protocol.DecodeLoginRequest(packet.Payload)
			if err != nil || loginReq.Username == "" || loginReq.Password == "" {
				slog.Warn("Invalid CS_LOGIN_REQUEST payload", "error", err)

				response := &protocol.Packet{
					Opcode:   protocol.SC_LOGIN_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeLoginResponse(false, 0, "", "invalid_login_payload"),
				}
				conn.Write(response.Serialize())
				break
			}

			authCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			authResp, err := s.authenticateWithAuthServer(authCtx, loginReq.Username, loginReq.Password)
			cancel()

			if err != nil {
				slog.Warn("Login rejected by Auth Server", "username", loginReq.Username, "error", err)

				response := &protocol.Packet{
					Opcode:   protocol.SC_LOGIN_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeLoginResponse(false, 0, "", "invalid_credentials"),
				}
				conn.Write(response.Serialize())
				break
			}

			sessionToken = authResp.Token
			authenticatedAccountID = authResp.AccountID
			lastRefresh = time.Now()

			slog.Info("Login accepted by Auth Server", "username", loginReq.Username, "account_id", authResp.AccountID)

			messaging.GetInstance().Publish("gateway.login", sessionToken)

			response := &protocol.Packet{
				Opcode:   protocol.SC_LOGIN_RESPONSE,
				Sequence: packet.Sequence,
				Payload:  protocol.EncodeLoginResponse(true, uint32(authenticatedAccountID), sessionToken, ""),
			}
			conn.Write(response.Serialize())

		case protocol.CS_CHAR_LIST_REQUEST:
			if authenticatedAccountID <= 0 {
				slog.Warn("Character list rejected: client is not authenticated")
				response := &protocol.Packet{
					Opcode:   protocol.SC_CHAR_LIST_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeCharacterListResponse(false, "not_authenticated", nil),
				}
				conn.Write(response.Serialize())
				break
			}

			slog.Info("Requesting character list from PostgreSQL")
			// FASE 3.3 Task 4D: usa account_id autenticado retornado pelo Auth Server.
			characters, err := s.persistenceMgr.ListCharactersByAccount(authenticatedAccountID)
			if err != nil {
				slog.Error("Failed to list characters from PostgreSQL", "error", err)
				response := &protocol.Packet{
					Opcode:   protocol.SC_CHAR_LIST_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeCharacterListResponse(false, "failed_to_list_characters", nil),
				}
				conn.Write(response.Serialize())
				break
			}

			entries := make([]protocol.CharacterListEntry, 0, len(characters))
			for _, ch := range characters {
				entries = append(entries, protocol.CharacterListEntry{
					Name:   ch.Name,
					Class:  ch.Class,
					Level:  uint32(ch.Level),
					RaceID: ch.RaceID, // (R1-I-B)
				})
			}

			response := &protocol.Packet{
				Opcode:   protocol.SC_CHAR_LIST_RESPONSE,
				Sequence: packet.Sequence,
				Payload:  protocol.EncodeCharacterListResponse(true, "", entries),
			}
			conn.Write(response.Serialize())

		case protocol.CS_CHAR_CREATE_REQUEST:
			if authenticatedAccountID <= 0 {
				slog.Warn("Character creation rejected: client is not authenticated")
				response := &protocol.Packet{
					Opcode:   protocol.SC_CHAR_CREATE_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeCharacterCreateResponse(false, "not_authenticated", protocol.CharacterListEntry{}),
				}
				conn.Write(response.Serialize())
				break
			}

			createReq, err := protocol.DecodeCharacterCreateRequest(packet.Payload)
			if err != nil {
				slog.Warn("Invalid CS_CHAR_CREATE_REQUEST payload", "error", err)
				response := &protocol.Packet{
					Opcode:   protocol.SC_CHAR_CREATE_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeCharacterCreateResponse(false, "invalid_character_create_payload", protocol.CharacterListEntry{}),
				}
				conn.Write(response.Serialize())
				break
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			summary, errorCode, err := s.characterCreationService.CreateCharacter(
				ctx,
				authenticatedAccountID,
				charactercreation.CreateRequest{
					DesiredName: createReq.DesiredName,
					RaceID:      createReq.RaceID,
				},
			)
			cancel()

			if err != nil {
				if errorCode == "" {
					errorCode = "internal_error"
				}
				slog.Error("Character creation failed", "account_id", authenticatedAccountID, "name", createReq.DesiredName, "error", err)
				response := &protocol.Packet{
					Opcode:   protocol.SC_CHAR_CREATE_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeCharacterCreateResponse(false, errorCode, protocol.CharacterListEntry{}),
				}
				conn.Write(response.Serialize())
				break
			}

			if summary == nil {
				slog.Error("Character creation returned nil summary without an error", "account_id", authenticatedAccountID)
				response := &protocol.Packet{
					Opcode:   protocol.SC_CHAR_CREATE_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeCharacterCreateResponse(false, "internal_error", protocol.CharacterListEntry{}),
				}
				conn.Write(response.Serialize())
				break
			}

			// Success
			characterEntry := protocol.CharacterListEntry{
				Name:   summary.Name,
				Class:  summary.Class,
				Level:  uint32(summary.Level),
				RaceID: summary.RaceID,
			}
			response := &protocol.Packet{
				Opcode:   protocol.SC_CHAR_CREATE_RESPONSE,
				Sequence: packet.Sequence,
				Payload:  protocol.EncodeCharacterCreateResponse(true, "", characterEntry),
			}
			conn.Write(response.Serialize())

		case protocol.CS_CHAR_SELECT_REQUEST:
			if authenticatedAccountID <= 0 {
				slog.Warn("Character selection rejected: client is not authenticated")
				response := &protocol.Packet{
					Opcode:   protocol.SC_CHAR_SELECT_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeCharacterSelectResponse(false, "", "not_authenticated"),
				}
				conn.Write(response.Serialize())
				break
			}
			slog.Info("Routing character selection to World Server")

			offset := 0
			selectedCharacterName, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil || selectedCharacterName == "" {
				slog.Warn("Invalid CS_CHAR_SELECT_REQUEST payload", "error", err)
				response := &protocol.Packet{
					Opcode:   protocol.SC_CHAR_SELECT_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeCharacterSelectResponse(false, "", "invalid_character_select_payload"), // Status: failed
				}
				conn.Write(response.Serialize())
				break
			}

			// FASE 3.3 Task 4D: valida ownership usando account_id autenticado.
			ownsCharacter, err := s.persistenceMgr.CharacterBelongsToAccount(authenticatedAccountID, selectedCharacterName)
			if err != nil {
				slog.Error("Failed to validate selected character ownership", "character", selectedCharacterName, "error", err)
				response := &protocol.Packet{
					Opcode:   protocol.SC_CHAR_SELECT_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeCharacterSelectResponse(false, "", "ownership_validation_failed"), // Status: failed
				}
				conn.Write(response.Serialize())
				break
			}

			if !ownsCharacter {
				slog.Warn("Character selection rejected: character does not belong to account", "character", selectedCharacterName, "account_id", authenticatedAccountID)
				response := &protocol.Packet{
					Opcode:   protocol.SC_CHAR_SELECT_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeCharacterSelectResponse(false, "", "character_not_owned"), // Status: failed
				}
				conn.Write(response.Serialize())
				break
			}

			characterID := selectedCharacterName // ID validado, mas ainda não ativo até LoadCharacter concluir

			// Carrega dados persistentes do banco PostgreSQL de forma atÃ´mica (PATCH 4)
			stats, items, savedX, savedY, savedZ, version, exp, gold, err := s.persistenceMgr.LoadCharacter(characterID)
			if err != nil {
				slog.Error("Failed to load character from PostgreSQL; rejecting character selection", "character", characterID, "account_id", authenticatedAccountID, "error", err)
				response := &protocol.Packet{
					Opcode:   protocol.SC_CHAR_SELECT_RESPONSE,
					Sequence: packet.Sequence,
					Payload:  protocol.EncodeCharacterSelectResponse(false, "", "character_load_failed"), // Status: failed
				}
				conn.Write(response.Serialize())
				break
			}

			playerID = characterID // Ativa o jogador somente após load persistente bem-sucedido

			// Inicializa inventÃ¡rio in-memory do jogador
			playerInv := inventory.NewPlayerInventory(playerID)
			playerInv.SetItems(items)
			playerInv.BaseStats = *stats   // Configura bÃ´nus de stats base antes do recÃ¡lculo
			playerInv.SetVersion(version)  // (PATCH 4)
			playerInv.SetDirty(false)      // (PATCH 2)
			playerInv.SetGold(gold)        // Define o gold do inventÃ¡rio
			pve.SetPlayerXp(playerID, exp) // Inicializa o XP do jogador para o PvE e ProgressÃ£o de NÃ­vel

			// Recalcula stats baseado nos equipamentos equipados de verdade no banco!
			playerInv.RecalculateStats(stats)

			s.inventoriesMu.Lock()
			s.inventories[playerID] = playerInv
			s.inventoriesMu.Unlock()

			s.aoiManager.RegisterPlayer(playerID, conn)
			s.movementSystem.InitPlayerState(playerID, savedX, savedY, int(savedZ))

			// Carrega e sincroniza estado de quests e diÃ¡logos de NPCs (Sprint 3 Task 3)
			_ = s.questManager.GetPlayerState(playerID)
			s.questManager.SyncAllActiveQuests(playerID)

			// Registra o jogador no CombatManager (com seus bÃ´nus de atributos)
			s.combatManager.RegisterEntity(stats, savedX, savedY)

			// Registra o NPC Orc Elite inimigo para simular combate PvE
			s.combatManager.RegisterEntity(&combat.EntityStats{
				ID:                 "Orc_Elite",
				Name:               "Orc Elite",
				IsPlayer:           false,
				Faction:            "Monsters",
				Level:              42,
				BaseAttack:         35.0,
				WeaponDamage:       25.0,
				Defense:            20.0,
				Resistance:         5.0,
				Accuracy:           85.0,
				Evasion:            12.0,
				CritChance:         0.05,
				CritMultiplier:     1.50,
				ArmorPenetration:   0.05,
				Element:            "Shadow",
				ElementAttackBonus: 0.05,
				ElementDefBonus:    0.10,
				Health:             500.0,
				MaxHealth:          500.0,
			}, savedX+2.0, savedY+2.0)

			response := &protocol.Packet{
				Opcode:   protocol.SC_CHAR_SELECT_RESPONSE,
				Sequence: packet.Sequence,
				Payload:  protocol.EncodeCharacterSelectResponse(true, characterID, ""),
			}
			conn.Write(response.Serialize())

			// Send initial position update to the client itself (R1-O-D1)
			initialPositionPayload, _ := json.Marshal(struct {
				PlayerID string  `json:"id"`
				X        float64 `json:"x"`
				Y        float64 `json:"y"`
				Z        int     `json:"z"`
			}{
				PlayerID: playerID,
				X:        savedX,
				Y:        savedY,
				Z:        int(savedZ),
			})
			initialPositionPacket := &protocol.Packet{
				Opcode:  protocol.SC_PLAYER_UPDATE,
				Payload: initialPositionPayload,
			}
			conn.Write(initialPositionPacket.Serialize())
			slog.Info("Initial player position update sent to client", "playerID", playerID, "x", savedX, "y", savedY, "z", savedZ)

			// Envia sincronizaÃ§Ã£o binÃ¡ria inicial de inventÃ¡rio e atributos recalculados
			s.sendInventorySync(conn, playerID, stats, playerInv)

			// Streaming inicial de chunks (janela deslizante 3x3 ao redor de sua Spawn Zone salva)
			chunks := s.chunkManager.GetSurroundingChunks(savedX, savedY)
			for _, ch := range chunks {
				chunkPacket := &protocol.Packet{
					Opcode:  protocol.SC_CHUNK_DATA,
					Payload: ch.Serialize(),
				}
				conn.Write(chunkPacket.Serialize())
			}
			slog.Info("Initial sliding chunks streamed to client", "playerID", playerID)

			// Trigger social login events
			s.socialManager.OnPlayerLogin(playerID)

		case protocol.CS_INVENTORY_REQUEST:
			if playerID == "" {
				slog.Warn("Inventory request received but player hasn't selected a character yet.")
				break
			}
			s.inventoriesMu.RLock()
			playerInv, ok := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if ok && playerInv != nil {
				stats, _ := s.combatManager.GetEntityStats(playerID)
				if stats != nil {
					s.sendInventorySync(conn, playerID, stats, playerInv)
				}
			}

		case protocol.CS_EQUIP_ITEM:
			if playerID == "" {
				slog.Warn("Equip item request received but player hasn't selected a character yet.")
				break
			}
			req, err := protocol.DecodeEquipItemRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode CS_EQUIP_ITEM packet", "error", err)
				break
			}

			slog.Info("Received CS_EQUIP_ITEM request", "player", playerID, "from", req.FromSlot, "to", req.ToSlot)

			s.inventoriesMu.RLock()
			playerInv, ok := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if !ok || playerInv == nil {
				slog.Warn("Inventory not found for player on equip", "player", playerID)
				respPayload := protocol.EncodeEquipItemResponse(false, "InventÃ¡rio nÃ£o encontrado")
				conn.Write((&protocol.Packet{Opcode: protocol.SC_EQUIP_RESPONSE, Sequence: packet.Sequence, Payload: respPayload}).Serialize())
				break
			}

			stats, exists := s.combatManager.GetEntityStats(playerID)
			if !exists || stats == nil {
				respPayload := protocol.EncodeEquipItemResponse(false, "Jogador nÃ£o registrado no motor de combate")
				conn.Write((&protocol.Packet{Opcode: protocol.SC_EQUIP_RESPONSE, Sequence: packet.Sequence, Payload: respPayload}).Serialize())
				break
			}

			err = playerInv.EquipItem(int(req.FromSlot), int(req.ToSlot), stats)
			if err != nil {
				slog.Warn("Equip item validation failed", "player", playerID, "error", err)
				respPayload := protocol.EncodeEquipItemResponse(false, err.Error())
				conn.Write((&protocol.Packet{Opcode: protocol.SC_EQUIP_RESPONSE, Sequence: packet.Sequence, Payload: respPayload}).Serialize())
			} else {
				respPayload := protocol.EncodeEquipItemResponse(true, "")
				conn.Write((&protocol.Packet{Opcode: protocol.SC_EQUIP_RESPONSE, Sequence: packet.Sequence, Payload: respPayload}).Serialize())
				s.sendInventorySync(conn, playerID, stats, playerInv)
			}

		case protocol.CS_UNEQUIP_ITEM:
			if playerID == "" {
				slog.Warn("Unequip item request received but player hasn't selected a character yet.")
				break
			}
			req, err := protocol.DecodeUnequipItemRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode CS_UNEQUIP_ITEM packet", "error", err)
				break
			}

			slog.Info("Received CS_UNEQUIP_ITEM request", "player", playerID, "from", req.FromSlot)

			s.inventoriesMu.RLock()
			playerInv, ok := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if !ok || playerInv == nil {
				respPayload := protocol.EncodeUnequipItemResponse(false, "InventÃ¡rio nÃ£o encontrado")
				conn.Write((&protocol.Packet{Opcode: protocol.SC_UNEQUIP_RESPONSE, Sequence: packet.Sequence, Payload: respPayload}).Serialize())
				break
			}

			stats, exists := s.combatManager.GetEntityStats(playerID)
			if !exists || stats == nil {
				respPayload := protocol.EncodeUnequipItemResponse(false, "Jogador nÃ£o registrado")
				conn.Write((&protocol.Packet{Opcode: protocol.SC_UNEQUIP_RESPONSE, Sequence: packet.Sequence, Payload: respPayload}).Serialize())
				break
			}

			err = playerInv.UnequipItem(int(req.FromSlot), stats)
			if err != nil {
				slog.Warn("Unequip item validation failed", "player", playerID, "error", err)
				respPayload := protocol.EncodeUnequipItemResponse(false, err.Error())
				conn.Write((&protocol.Packet{Opcode: protocol.SC_UNEQUIP_RESPONSE, Sequence: packet.Sequence, Payload: respPayload}).Serialize())
			} else {
				respPayload := protocol.EncodeUnequipItemResponse(true, "")
				conn.Write((&protocol.Packet{Opcode: protocol.SC_UNEQUIP_RESPONSE, Sequence: packet.Sequence, Payload: respPayload}).Serialize())
				s.sendInventorySync(conn, playerID, stats, playerInv)
			}

		case protocol.CS_SWAP_SLOTS:
			if playerID == "" {
				slog.Warn("Swap slots request received but player hasn't selected a character yet.")
				break
			}
			req, err := protocol.DecodeSwapSlotsRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode CS_SWAP_SLOTS packet", "error", err)
				break
			}

			slog.Info("Received CS_SWAP_SLOTS request", "player", playerID, "slotA", req.SlotA, "slotB", req.SlotB)

			s.inventoriesMu.RLock()
			playerInv, ok := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if !ok || playerInv == nil {
				respPayload := protocol.EncodeSwapSlotsResponse(false, "InventÃ¡rio nÃ£o encontrado")
				conn.Write((&protocol.Packet{Opcode: protocol.SC_SWAP_RESPONSE, Sequence: packet.Sequence, Payload: respPayload}).Serialize())
				break
			}

			err = playerInv.SwapSlots(int(req.SlotA), int(req.SlotB))
			if err != nil {
				slog.Warn("Swap slots failed", "player", playerID, "error", err)
				respPayload := protocol.EncodeSwapSlotsResponse(false, err.Error())
				conn.Write((&protocol.Packet{Opcode: protocol.SC_SWAP_RESPONSE, Sequence: packet.Sequence, Payload: respPayload}).Serialize())
			} else {
				respPayload := protocol.EncodeSwapSlotsResponse(true, "")
				conn.Write((&protocol.Packet{Opcode: protocol.SC_SWAP_RESPONSE, Sequence: packet.Sequence, Payload: respPayload}).Serialize())
				stats, _ := s.combatManager.GetEntityStats(playerID)
				if stats != nil {
					s.sendInventorySync(conn, playerID, stats, playerInv)
				}
			}

		case protocol.CS_NPC_INTERACT:
			if playerID == "" {
				slog.Warn("NPC interact request received but player hasn't selected a character yet.")
				break
			}
			req, err := protocol.DecodeNPCInteractRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode CS_NPC_INTERACT", "error", err)
				break
			}

			// Validar distÃ¢ncia
			if err := s.npcManager.ValidateInteractionDistance(playerID, req.NPCID, s.spatialIndex); err != nil {
				slog.Warn("NPC interaction rejected due to distance/floor check", "player", playerID, "npc", req.NPCID, "error", err)
				break
			}

			// Carrega nÃ³ inicial
			node, err := s.npcManager.GetVisibleNode(playerID, req.NPCID, "start", s.questManager)
			if err != nil {
				slog.Error("Failed to load initial dialogue node", "player", playerID, "npc", req.NPCID, "error", err)
				break
			}

			// Define estado de conversa atual no jogador
			s.questManager.SetDialogueFlag(playerID, req.NPCID, node.NodeID)

			// Envia diÃ¡logo aberto
			choices := make([]protocol.DialogueOpenChoice, 0, len(node.Responses))
			for _, r := range node.Responses {
				choices = append(choices, protocol.DialogueOpenChoice{
					NextNodeID: r.NextNodeID,
					Text:       r.Text,
				})
			}
			openEvent := &protocol.DialogueOpenEvent{
				NPCID:    req.NPCID,
				NodeID:   node.NodeID,
				NodeText: node.Text,
				Choices:  choices,
			}
			responsePacket := &protocol.Packet{
				Opcode:   protocol.SC_DIALOGUE_OPEN,
				Sequence: packet.Sequence,
				Payload:  protocol.EncodeDialogueOpen(openEvent),
			}
			conn.Write(responsePacket.Serialize())

			// Aciona hook de conversa para quests via MessageBus (PATCH 5)
			messaging.GetInstance().Publish("npc_interacted", messaging.NPCInteractedPayload{
				PlayerID: playerID,
				NPCID:    req.NPCID,
			})

		case protocol.CS_DIALOGUE_RESPONSE:
			if playerID == "" {
				slog.Warn("Dialogue response request received but player hasn't selected a character yet.")
				break
			}
			req, err := protocol.DecodeDialogueResponseRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode CS_DIALOGUE_RESPONSE", "error", err)
				break
			}

			// Validar distÃ¢ncia
			if err := s.npcManager.ValidateInteractionDistance(playerID, req.NPCID, s.spatialIndex); err != nil {
				slog.Warn("Dialogue response rejected due to distance/floor check", "player", playerID, "npc", req.NPCID, "error", err)
				break
			}

			// Se next node for "end", encerra o diÃ¡logo e limpa ou atualiza a flag
			if req.NextNodeID == "end" || req.NextNodeID == "" {
				s.questManager.SetDialogueFlag(playerID, req.NPCID, "completed_conversation")
				break
			}

			// ObtÃ©m nÃ³ de diÃ¡logo selecionado
			node, err := s.npcManager.GetVisibleNode(playerID, req.NPCID, req.NextNodeID, s.questManager)
			if err != nil {
				slog.Error("Failed to load dialogue node", "player", playerID, "npc", req.NPCID, "node", req.NextNodeID, "error", err)
				break
			}

			// Processa gatilhos de quest ANTES de avanÃ§ar para novo diÃ¡logo, garantindo transaÃ§Ãµes atÃ´micas (PATCH 2)
			if node.QuestTrigger != nil {
				if node.QuestTrigger.Action == "accept" {
					if err := s.questManager.AcceptQuest(playerID, node.QuestTrigger.QuestID); err != nil {
						slog.Warn("Failed to accept quest through dialogue", "player", playerID, "quest", node.QuestTrigger.QuestID, "error", err)
					}
				} else if node.QuestTrigger.Action == "complete" {
					if err := s.questManager.CompleteQuest(playerID, node.QuestTrigger.QuestID); err != nil {
						slog.Warn("Failed to complete quest through dialogue", "player", playerID, "quest", node.QuestTrigger.QuestID, "error", err)
					}
				}
			}

			// Define estado de conversa atual no jogador
			s.questManager.SetDialogueFlag(playerID, req.NPCID, node.NodeID)

			// Envia diÃ¡logo aberto
			choices := make([]protocol.DialogueOpenChoice, 0, len(node.Responses))
			for _, r := range node.Responses {
				choices = append(choices, protocol.DialogueOpenChoice{
					NextNodeID: r.NextNodeID,
					Text:       r.Text,
				})
			}
			openEvent := &protocol.DialogueOpenEvent{
				NPCID:    req.NPCID,
				NodeID:   node.NodeID,
				NodeText: node.Text,
				Choices:  choices,
			}
			responsePacket := &protocol.Packet{
				Opcode:   protocol.SC_DIALOGUE_OPEN,
				Sequence: packet.Sequence,
				Payload:  protocol.EncodeDialogueOpen(openEvent),
			}
			conn.Write(responsePacket.Serialize())

		case protocol.CS_ACCEPT_QUEST:
			if playerID == "" {
				slog.Warn("Accept quest request received but player hasn't selected a character yet.")
				break
			}
			req, err := protocol.DecodeAcceptQuestRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode CS_ACCEPT_QUEST", "error", err)
				break
			}
			if err := s.questManager.AcceptQuest(playerID, req.QuestID); err != nil {
				slog.Warn("Accept quest failed", "player", playerID, "quest", req.QuestID, "error", err)
			}

		case protocol.CS_COMPLETE_QUEST:
			if playerID == "" {
				slog.Warn("Complete quest request received but player hasn't selected a character yet.")
				break
			}
			req, err := protocol.DecodeCompleteQuestRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode CS_COMPLETE_QUEST", "error", err)
				break
			}
			if err := s.questManager.CompleteQuest(playerID, req.QuestID); err != nil {
				slog.Warn("Complete quest failed", "player", playerID, "quest", req.QuestID, "error", err)
			}

		// =====================================================================
		// SPRINT 3 TASK 3: PARTY, GUILD & SOCIAL HANDLERS
		// =====================================================================
		case protocol.CS_PARTY_CREATE:
			if playerID == "" {
				break
			}
			if len(packet.Payload) < 1 {
				break
			}
			mode := packet.Payload[0]
			_, err := s.socialManager.CreateParty(playerID, social.LootMode(mode))
			if err != nil {
				slog.Warn("Failed to create party", "player", playerID, "error", err)
			}

		case protocol.CS_PARTY_INVITE:
			if playerID == "" {
				break
			}
			offset := 0
			targetID, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				slog.Error("Failed to read party invite target", "error", err)
				break
			}
			if err := s.socialManager.InviteToParty(playerID, targetID); err != nil {
				slog.Warn("Failed to invite to party", "player", playerID, "target", targetID, "error", err)
			}

		case protocol.CS_PARTY_INVITE_RESP:
			if playerID == "" {
				break
			}
			resp, err := protocol.DecodePartyInviteResponse(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode party invite response", "error", err)
				break
			}
			if resp.Accept == 1 {
				if err := s.socialManager.AcceptPartyInvite(playerID, resp.InviterID); err != nil {
					slog.Warn("Failed to accept party invite", "player", playerID, "inviter", resp.InviterID, "error", err)
				}
			} else {
				if err := s.socialManager.RejectPartyInvite(playerID, resp.InviterID); err != nil {
					slog.Warn("Failed to reject party invite", "player", playerID, "inviter", resp.InviterID, "error", err)
				}
			}

		case protocol.CS_PARTY_LEAVE:
			if playerID == "" {
				break
			}
			if err := s.socialManager.LeaveParty(playerID); err != nil {
				slog.Warn("Failed to leave party", "player", playerID, "error", err)
			}

		case protocol.CS_PARTY_KICK:
			if playerID == "" {
				break
			}
			offset := 0
			targetID, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				break
			}
			if err := s.socialManager.KickMember(playerID, targetID); err != nil {
				slog.Warn("Failed to kick party member", "player", playerID, "target", targetID, "error", err)
			}

		case protocol.CS_PARTY_TRANSFER:
			if playerID == "" {
				break
			}
			offset := 0
			targetID, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				break
			}
			if err := s.socialManager.TransferLeader(playerID, targetID); err != nil {
				slog.Warn("Failed to transfer party leadership", "player", playerID, "target", targetID, "error", err)
			}

		case protocol.CS_PARTY_LOOT_MODE:
			if playerID == "" {
				break
			}
			if len(packet.Payload) < 1 {
				break
			}
			mode := packet.Payload[0]
			if err := s.socialManager.SetLootMode(playerID, social.LootMode(mode)); err != nil {
				slog.Warn("Failed to set party loot mode", "player", playerID, "mode", mode, "error", err)
			}

		case protocol.CS_GUILD_CREATE:
			if playerID == "" {
				break
			}
			offset := 0
			guildName, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				break
			}
			if err := s.socialManager.CreateGuild(playerID, guildName); err != nil {
				slog.Warn("Failed to create guild", "player", playerID, "name", guildName, "error", err)
			}

		case protocol.CS_GUILD_INVITE:
			if playerID == "" {
				break
			}
			offset := 0
			targetID, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				break
			}
			if err := s.socialManager.InviteToGuild(playerID, targetID); err != nil {
				slog.Warn("Failed to invite to guild", "player", playerID, "target", targetID, "error", err)
			}

		case protocol.CS_GUILD_INVITE_RESP:
			if playerID == "" {
				break
			}
			resp, err := protocol.DecodeGuildInviteResponse(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode guild invite response", "error", err)
				break
			}
			if resp.Accept == 1 {
				if err := s.socialManager.AcceptGuildInvite(playerID, resp.GuildName); err != nil {
					slog.Warn("Failed to accept guild invite", "player", playerID, "guild", resp.GuildName, "error", err)
				}
			}

		case protocol.CS_GUILD_LEAVE:
			if playerID == "" {
				break
			}
			if err := s.socialManager.LeaveGuild(playerID); err != nil {
				slog.Warn("Failed to leave guild", "player", playerID, "error", err)
			}

		case protocol.CS_GUILD_KICK:
			if playerID == "" {
				break
			}
			offset := 0
			targetID, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				break
			}
			if err := s.socialManager.KickFromGuild(playerID, targetID); err != nil {
				slog.Warn("Failed to kick guild member", "player", playerID, "target", targetID, "error", err)
			}

		case protocol.CS_GUILD_PROMOTE:
			if playerID == "" {
				break
			}
			offset := 0
			targetID, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				break
			}
			if err := s.socialManager.PromoteMember(playerID, targetID); err != nil {
				slog.Warn("Failed to promote guild member", "player", playerID, "target", targetID, "error", err)
			}

		case protocol.CS_GUILD_DEMOTE:
			if playerID == "" {
				break
			}
			offset := 0
			targetID, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				break
			}
			if err := s.socialManager.DemoteMember(playerID, targetID); err != nil {
				slog.Warn("Failed to demote guild member", "player", playerID, "target", targetID, "error", err)
			}

		case protocol.CS_GUILD_MOTD:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeGuildMOTDRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode guild motd request", "error", err)
				break
			}
			guildName := s.socialManager.FindGuildNameForPlayer(playerID)
			if guildName == "" {
				slog.Warn("Guild MOTD update rejected: player not in guild", "player", playerID)
				break
			}
			if err := s.socialManager.UpdateMOTD(playerID, guildName, req.MOTD, req.ExpectedVersion); err != nil {
				slog.Warn("Failed to update guild MOTD", "player", playerID, "error", err)
			}

		case protocol.CS_SOCIAL_ADD_FRIEND:
			if playerID == "" {
				break
			}
			offset := 0
			friendID, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				break
			}
			if err := s.socialManager.AddFriend(playerID, friendID); err != nil {
				slog.Warn("Failed to add friend", "player", playerID, "friend", friendID, "error", err)
			}

		case protocol.CS_SOCIAL_REMOVE_FRIEND:
			if playerID == "" {
				break
			}
			offset := 0
			friendID, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				break
			}
			if err := s.socialManager.RemoveFriend(playerID, friendID); err != nil {
				slog.Warn("Failed to remove friend", "player", playerID, "friend", friendID, "error", err)
			}

		case protocol.CS_SOCIAL_ADD_IGNORE:
			if playerID == "" {
				break
			}
			offset := 0
			targetID, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				break
			}
			if err := s.socialManager.AddIgnore(playerID, targetID); err != nil {
				slog.Warn("Failed to ignore player", "player", playerID, "target", targetID, "error", err)
			}

		case protocol.CS_SOCIAL_REMOVE_IGNORE:
			if playerID == "" {
				break
			}
			offset := 0
			targetID, err := protocol.ReadString(packet.Payload, &offset)
			if err != nil {
				break
			}
			if err := s.socialManager.RemoveIgnore(playerID, targetID); err != nil {
				slog.Warn("Failed to unignore player", "player", playerID, "target", targetID, "error", err)
			}

		case protocol.CS_CHAT_SEND:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeChatSendRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode chat send request", "error", err)
				break
			}
			s.socialManager.HandleChatMessage(playerID, req)

		case protocol.CS_MOVE_REQUEST:
			if playerID == "" {
				slog.Warn("Move request received but player hasn't selected a character yet.")
				break
			}

			// Cancela coleta ativa se o jogador se mover (Sprint 4 Task 1)
			s.cancelGatheringIfActive(playerID)

			if len(packet.Payload) < 18 {
				slog.Error("Failed to decode CS_MOVE_REQUEST payload: too short", "length", len(packet.Payload))
				break
			}

			targetX := int32(binary.LittleEndian.Uint32(packet.Payload[0:4]))
			targetY := int32(binary.LittleEndian.Uint32(packet.Payload[4:8]))
			targetZ := int8(packet.Payload[8])
			direction := uint8(packet.Payload[9])
			clientTimestamp := binary.LittleEndian.Uint64(packet.Payload[10:18])

			// For compiler / logging references
			_ = direction
			_ = clientTimestamp

			// ValidaÃ§Ã£o de movimento fÃ­sica e temporal autoritativa (Sprint 2 Task 4)
			success, confX, confY, confZ := s.movementSystem.ValidateAndMove(playerID, float64(targetX), float64(targetY), int(targetZ), packet.Sequence)

			// Envia confirmaÃ§Ã£o de volta ao cliente (SC_MOVE_CONFIRM)
			confirm := struct {
				X       float64 `json:"x"`
				Y       float64 `json:"y"`
				Z       int     `json:"z"`
				Seq     uint32  `json:"seq"`
				Success bool    `json:"success"`
			}{
				X:       confX,
				Y:       confY,
				Z:       confZ,
				Seq:     packet.Sequence,
				Success: success,
			}

			confirmPayload, _ := json.Marshal(confirm)
			confirmPacket := &protocol.Packet{
				Opcode:   protocol.SC_MOVE_CONFIRM,
				Sequence: packet.Sequence,
				Payload:  confirmPayload,
			}
			conn.Write(confirmPacket.Serialize())

			if success {
				// Atualiza a posiÃ§Ã£o no CombatManager para validaÃ§Ã£o autoritativa de combate (Sprint 2 Task 5)
				s.combatManager.UpdateEntityPosition(playerID, confX, confY)

				// Marca o estado do jogador como alterado na persistÃªncia (PATCH 2)
				s.inventoriesMu.RLock()
				if playerInv, ok := s.inventories[playerID]; ok && playerInv != nil {
					playerInv.SetDirty(true)
				}
				s.inventoriesMu.RUnlock()

				// Broadcast da nova posiÃ§Ã£o para jogadores vizinhos na AOI
				updatePayload, _ := json.Marshal(struct {
					PlayerID string  `json:"id"`
					X        float64 `json:"x"`
					Y        float64 `json:"y"`
					Z        int     `json:"z"`
				}{
					PlayerID: playerID,
					X:        confX,
					Y:        confY,
					Z:        confZ,
				})
				s.aoiManager.BroadcastMove(playerID, confX, confY, confZ, updatePayload)

				// Aciona hook de alcance de localizaÃ§Ãµes de quest via MessageBus (PATCH 5)
				messaging.GetInstance().Publish("location_reached", messaging.LocationReachedPayload{
					PlayerID: playerID,
					X:        confX,
					Y:        confY,
					Z:        float64(confZ),
				})
			} else {
				slog.Warn("Authoritative movement validation failed (Client out of sync/rubberbanded)", "player", playerID, "x", targetX, "y", targetY)
			}

		case protocol.CS_PLAYER_MOVE:
			if playerID != "" {
				s.cancelGatheringIfActive(playerID)
			}

			// Propagar via Message Bus interno (PATCH 1)
			messaging.GetInstance().Publish("player.move", packet.Payload)

			slog.Debug("Broadcasting move event to World Server")
			response := &protocol.Packet{
				Opcode:   protocol.SC_PLAYER_UPDATE,
				Sequence: packet.Sequence,
				Payload:  packet.Payload,
			}
			conn.Write(response.Serialize())

		case protocol.CS_ATTACK_REQUEST:
			if playerID == "" {
				slog.Warn("Attack request received but player hasn't selected a character yet.")
				break
			}
			req, err := protocol.DecodeAttackRequest(packet.Payload)
			if err != nil {
				slog.Warn("Failed to decode binary CS_ATTACK_REQUEST", "error", err)
				break
			}

			damage, isCrit, isProj, err := s.combatManager.ProcessAttackRequest(playerID, req.TargetID, req.WeaponType)
			if err != nil {
				slog.Warn("Failed to process basic attack", "error", err)
				errPayload := protocol.EncodeDamageEvent(playerID, req.TargetID, 0, false, false, false, err.Error())
				errPacket := &protocol.Packet{
					Opcode:   protocol.SC_DAMAGE_EVENT,
					Sequence: packet.Sequence,
					Payload:  errPayload,
				}
				conn.Write(errPacket.Serialize())
				break
			}

			// Marca estado como alterado por eventos de combate (PATCH 2)
			s.inventoriesMu.RLock()
			if playerInv, ok := s.inventories[playerID]; ok && playerInv != nil {
				playerInv.SetDirty(true)
			}
			s.inventoriesMu.RUnlock()

			if isProj {
				// ProjÃ©til agendado. Broadcast visual do efeito de tiro via BroadcastEffects
				spawnPayload := protocol.EncodeDamageEvent(playerID, req.TargetID, 0, false, false, true, req.WeaponType)
				s.aoiManager.BroadcastEffects(playerID, protocol.SC_DAMAGE_EVENT, spawnPayload)
				break
			}

			// Retorna evento de dano com sucesso para melee instantÃ¢neo
			dmgPayload := protocol.EncodeDamageEvent(playerID, req.TargetID, damage, isCrit, damage > 0, true, "")
			dmgPacket := &protocol.Packet{
				Opcode:   protocol.SC_DAMAGE_EVENT,
				Sequence: packet.Sequence,
				Payload:  dmgPayload,
			}
			conn.Write(dmgPacket.Serialize())

			// Broadcast do evento de dano para a Ã¡rea de interesse (AOI) via BroadcastCombat
			s.aoiManager.BroadcastCombat(playerID, protocol.SC_DAMAGE_EVENT, dmgPayload)

			// Verifica se o alvo morreu
			targetStats, exists := s.combatManager.GetEntityStats(req.TargetID)
			if exists && targetStats.Health <= 0 {
				deadPayload := protocol.EncodeTargetDeadEvent(req.TargetID)
				deadPacket := &protocol.Packet{
					Opcode:   protocol.SC_TARGET_DEAD,
					Sequence: packet.Sequence,
					Payload:  deadPayload,
				}
				conn.Write(deadPacket.Serialize())
				s.aoiManager.BroadcastCombat(playerID, protocol.SC_TARGET_DEAD, deadPayload)
			}

		case protocol.CS_CAST_SKILL:
			if playerID == "" {
				slog.Warn("Cast skill request received but player hasn't selected a character yet.")
				break
			}
			req, err := protocol.DecodeCastSkillRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode binary CS_CAST_SKILL", "error", err)
				break
			}

			res, err := s.combatManager.ProcessCastSkillRequest(playerID, req.SkillID, req.TargetID, req.TargetX, req.TargetY)
			if err != nil {
				slog.Warn("Failed to process cast skill", "error", err)
				errPayload := protocol.EncodeDamageEvent(playerID, req.TargetID, 0, false, false, false, err.Error())
				errPacket := &protocol.Packet{
					Opcode:   protocol.SC_DAMAGE_EVENT,
					Sequence: packet.Sequence,
					Payload:  errPayload,
				}
				conn.Write(errPacket.Serialize())
				break
			}

			// Dispara o evento de habilidade conjurada para fins de ganho de afinidade (Sprint 3 Task 5)
			messaging.GetInstance().Publish("skill_cast", map[string]interface{}{
				"player_id": playerID,
				"skill_id":  int(req.SkillID),
			})

			// Marca estado como alterado por eventos de combate (PATCH 2)
			s.inventoriesMu.RLock()
			if playerInv, ok := s.inventories[playerID]; ok && playerInv != nil {
				playerInv.SetDirty(true)
			}
			s.inventoriesMu.RUnlock()

			if res.IsProjectile {
				// Habilidade de projÃ©til agendada. Broadcast de efeito visual via BroadcastEffects
				spawnPayload := protocol.EncodeDamageEvent(playerID, req.TargetID, 0, false, false, true, res.Skill.Name)
				s.aoiManager.BroadcastEffects(playerID, protocol.SC_DAMAGE_EVENT, spawnPayload)
				break
			}

			// Retorna cada hit da habilidade instantÃ¢nea como evento de dano
			for _, hit := range res.TargetsHit {
				dmgPayload := protocol.EncodeDamageEvent(playerID, hit.TargetID, hit.Damage, hit.IsCrit, hit.IsHit, true, res.Skill.Name)
				dmgPacket := &protocol.Packet{
					Opcode:   protocol.SC_DAMAGE_EVENT,
					Sequence: packet.Sequence,
					Payload:  dmgPayload,
				}
				conn.Write(dmgPacket.Serialize())
				s.aoiManager.BroadcastCombat(playerID, protocol.SC_DAMAGE_EVENT, dmgPayload)

				// Se o alvo morreu, notifica morte
				targetStats, exists := s.combatManager.GetEntityStats(hit.TargetID)
				if exists && targetStats.Health <= 0 {
					deadPayload := protocol.EncodeTargetDeadEvent(hit.TargetID)
					deadPacket := &protocol.Packet{
						Opcode:   protocol.SC_TARGET_DEAD,
						Sequence: packet.Sequence,
						Payload:  deadPayload,
					}
					conn.Write(deadPacket.Serialize())
					s.aoiManager.BroadcastCombat(playerID, protocol.SC_TARGET_DEAD, deadPayload)
				}
			}

		// =========================================================================
		// ECONOMY, PLAYER TRADING, NPC SHOP & MARKETPLACE (Sprint 3 Task 4)
		// =========================================================================

		case protocol.CS_TRADE_REQUEST:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeTradeRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode trade request", "error", err)
				break
			}
			targetConn, ok := s.aoiManager.GetPlayerConn(req.TargetName)
			if !ok {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, "Target player is not online")}).Serialize())
				break
			}
			// Send proposal to target
			targetConn.Write((&protocol.Packet{Opcode: protocol.SC_TRADE_PROPOSAL, Payload: protocol.EncodeTradeProposal(playerID)}).Serialize())

		case protocol.CS_TRADE_RESPOND:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeTradeRespond(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode trade response", "error", err)
				break
			}
			if req.Accepted == 1 {
				// Procura o jogador mais prÃ³ximo para iniciar a troca (ou o primeiro online nas redondezas)
				var partner string
				px, py, _, _ := s.movementSystem.GetPlayerPos(playerID)
				s.inventoriesMu.RLock()
				for otherID := range s.inventories {
					if otherID != playerID {
						ox, oy, _, _ := s.movementSystem.GetPlayerPos(otherID)
						dist := math.Sqrt(math.Pow(px-ox, 2) + math.Pow(py-oy, 2))
						if dist <= 3.0 {
							partner = otherID
							break
						}
					}
				}
				s.inventoriesMu.RUnlock()

				if partner == "" {
					conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, "No player nearby to trade with")}).Serialize())
					break
				}

				_, err := s.economyManager.StartTradeSession(partner, playerID)
				if err != nil {
					conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
					break
				}

				s.broadcastTradeUpdate(playerID)
			} else {
				// Rejeitou a troca: envia uma resposta simples
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, "Trade request declined")}).Serialize())
			}

		case protocol.CS_TRADE_OFFER_GOLD:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeTradeOfferGold(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode trade gold offer", "error", err)
				break
			}
			s.inventoriesMu.RLock()
			playerInv := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if playerInv == nil {
				break
			}

			err = s.economyManager.OfferGold(playerID, int64(req.Gold), playerInv)
			if err != nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
				break
			}
			s.broadcastTradeUpdate(playerID)

		case protocol.CS_TRADE_OFFER_ITEM:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeTradeOfferItem(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode trade item offer", "error", err)
				break
			}
			s.inventoriesMu.RLock()
			playerInv := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if playerInv == nil {
				break
			}

			err = s.economyManager.OfferItem(playerID, int(req.SlotIndex), int(req.Quantity), playerInv)
			if err != nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
				break
			}
			s.broadcastTradeUpdate(playerID)

		case protocol.CS_TRADE_LOCK:
			if playerID == "" {
				break
			}
			_, err := s.economyManager.LockTrade(playerID)
			if err != nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
				break
			}
			s.broadcastTradeUpdate(playerID)

		case protocol.CS_TRADE_CONFIRM:
			if playerID == "" {
				break
			}
			// Busca parceiro
			session, exists := s.economyManager.GetTradeSession(playerID)
			if !exists {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, "No active trade session")}).Serialize())
				break
			}

			s.inventoriesMu.RLock()
			invA := s.inventories[session.PlayerA]
			invB := s.inventories[session.PlayerB]
			s.inventoriesMu.RUnlock()

			if invA == nil || invB == nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, "Trade partner offline")}).Serialize())
				break
			}

			committed, msg, err := s.economyManager.CompleteTrade(playerID, invA, invB)
			if err != nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
				break
			}

			if committed {
				// Envia confirmaÃ§Ã£o de trade finalizado para ambos
				closedPkt := &protocol.Packet{Opcode: protocol.SC_TRADE_CLOSED, Payload: protocol.EncodeTradeClosed("Trade committed successfully")}
				serializedClosed := closedPkt.Serialize()
				if connA, ok := s.aoiManager.GetPlayerConn(session.PlayerA); ok {
					connA.Write(serializedClosed)
				}
				if connB, ok := s.aoiManager.GetPlayerConn(session.PlayerB); ok {
					connB.Write(serializedClosed)
				}
				// Sincroniza inventÃ¡rios para ambos verem as mudanÃ§as imediatas
				s.inventoriesMu.RLock()
				statsA, existsA := s.combatManager.GetEntityStats(session.PlayerA)
				statsB, existsB := s.combatManager.GetEntityStats(session.PlayerB)
				s.inventoriesMu.RUnlock()
				if existsA {
					s.sendInventorySync(conn, session.PlayerA, statsA, invA)
				}
				if existsB {
					if connB, ok := s.aoiManager.GetPlayerConn(session.PlayerB); ok {
						s.sendInventorySync(connB, session.PlayerB, statsB, invB)
					}
				}
			} else {
				// Sincroniza estado intermediÃ¡rio
				s.broadcastTradeUpdate(playerID)
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(1, msg)}).Serialize())
			}

		case protocol.CS_TRADE_CANCEL:
			if playerID == "" {
				break
			}
			session, exists := s.economyManager.GetTradeSession(playerID)
			if exists {
				s.economyManager.CancelTrade(playerID)
				closedPkt := &protocol.Packet{Opcode: protocol.SC_TRADE_CLOSED, Payload: protocol.EncodeTradeClosed("Trade cancelled by player")}
				serializedClosed := closedPkt.Serialize()
				if connA, ok := s.aoiManager.GetPlayerConn(session.PlayerA); ok {
					connA.Write(serializedClosed)
				}
				if connB, ok := s.aoiManager.GetPlayerConn(session.PlayerB); ok {
					connB.Write(serializedClosed)
				}
			}

		case protocol.CS_NPC_SHOP_BUY:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeNPCShopBuy(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode NPCShopBuy", "error", err)
				break
			}
			s.inventoriesMu.RLock()
			playerInv := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if playerInv == nil {
				break
			}

			msg, err := s.economyManager.BuyNPCItem(playerID, req.ItemID, int(req.Quantity), playerInv)
			if err != nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
			} else {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(1, msg)}).Serialize())
				// Envia atualizaÃ§Ã£o de inventÃ¡rio imediata
				stats, _ := s.combatManager.GetEntityStats(playerID)
				s.sendInventorySync(conn, playerID, stats, playerInv)
			}

		case protocol.CS_NPC_SHOP_SELL:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeNPCShopSell(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode NPCShopSell", "error", err)
				break
			}
			s.inventoriesMu.RLock()
			playerInv := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if playerInv == nil {
				break
			}

			msg, err := s.economyManager.SellNPCItem(playerID, int(req.SlotIndex), int(req.Quantity), playerInv)
			if err != nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
			} else {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(1, msg)}).Serialize())
				stats, _ := s.combatManager.GetEntityStats(playerID)
				s.sendInventorySync(conn, playerID, stats, playerInv)
			}

		case protocol.CS_NPC_SHOP_REPAIR:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeNPCShopRepair(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode NPCShopRepair", "error", err)
				break
			}
			s.inventoriesMu.RLock()
			playerInv := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if playerInv == nil {
				break
			}

			msg, err := s.economyManager.RepairItem(playerID, int(req.SlotIndex), playerInv)
			if err != nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
			} else {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(1, msg)}).Serialize())
				stats, _ := s.combatManager.GetEntityStats(playerID)
				s.sendInventorySync(conn, playerID, stats, playerInv)
			}

		case protocol.CS_MARKET_CREATE_ORDER:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeMarketCreateOrder(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode MarketCreateOrder", "error", err)
				break
			}
			s.inventoriesMu.RLock()
			playerInv := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if playerInv == nil {
				break
			}

			msg, err := s.economyManager.CreateMarketOrder(playerID, int(req.SlotIndex), int(req.Quantity), int64(req.Price), playerInv)
			if err != nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
			} else {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(1, msg)}).Serialize())
				stats, _ := s.combatManager.GetEntityStats(playerID)
				s.sendInventorySync(conn, playerID, stats, playerInv)
			}

		case protocol.CS_MARKET_BUY_ITEM:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeMarketBuyItem(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode MarketBuyItem", "error", err)
				break
			}
			s.inventoriesMu.RLock()
			playerInv := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if playerInv == nil {
				break
			}

			msg, err := s.economyManager.BuyMarketItem(playerID, int64(req.OrderID), playerInv)
			if err != nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
			} else {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(1, msg)}).Serialize())
				stats, _ := s.combatManager.GetEntityStats(playerID)
				s.sendInventorySync(conn, playerID, stats, playerInv)
			}

		case protocol.CS_MARKET_CANCEL_ORDER:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeMarketCancelOrder(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode MarketCancelOrder", "error", err)
				break
			}
			s.inventoriesMu.RLock()
			playerInv := s.inventories[playerID]
			s.inventoriesMu.RUnlock()

			if playerInv == nil {
				break
			}

			msg, err := s.economyManager.CancelMarketOrder(playerID, int64(req.OrderID), playerInv)
			if err != nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
			} else {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(1, msg)}).Serialize())
				stats, _ := s.combatManager.GetEntityStats(playerID)
				s.sendInventorySync(conn, playerID, stats, playerInv)
			}

		case protocol.CS_MARKET_SEARCH:
			if playerID == "" {
				break
			}
			req, err := protocol.DecodeMarketSearch(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode MarketSearch", "error", err)
				break
			}

			orders, err := s.economyManager.SearchMarketOrders(req.FilterItemID)
			if err != nil {
				conn.Write((&protocol.Packet{Opcode: protocol.SC_NPC_SHOP_RESPONSE, Payload: protocol.EncodeNPCShopResponse(0, err.Error())}).Serialize())
				break
			}

			// Converte ordens para codec binÃ¡rio
			codecs := make([]protocol.MarketOrderCodec, 0, len(orders))
			for _, o := range orders {
				codecs = append(codecs, protocol.MarketOrderCodec{
					OrderID:      uint32(o.OrderID),
					SellerName:   o.SellerName,
					ItemID:       o.ItemID,
					Quantity:     uint32(o.Quantity),
					PriceGold:    uint32(o.PriceGold),
					ExpiresEpoch: uint32(o.ExpiresAt.Unix()),
				})
			}

			resultPayload := protocol.EncodeMarketSearchResult(codecs)
			conn.Write((&protocol.Packet{
				Opcode:   protocol.SC_MARKET_SEARCH_RESULT,
				Sequence: packet.Sequence,
				Payload:  resultPayload,
			}).Serialize())

		case protocol.CS_GATHER_START:
			if playerID == "" {
				break
			}
			nodeID, err := protocol.DecodeGatherStart(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode GatherStart", "error", err)
				break
			}

			s.inventoriesMu.RLock()
			playerInv, existsInv := s.inventories[playerID]
			s.inventoriesMu.RUnlock()
			if !existsInv {
				break
			}

			duration, err := s.professionsManager.StartGathering(playerID, nodeID, playerInv)
			if err != nil {
				conn.Write((&protocol.Packet{
					Opcode:  protocol.SC_GATHER_COMPLETE,
					Payload: protocol.EncodeGatherComplete(nodeID, "", 0, 0),
				}).Serialize())
				conn.Write((&protocol.Packet{
					Opcode:  protocol.SC_CHAT_MESSAGE,
					Payload: protocol.EncodeChatMessage(0, "System", "Coleta falhou: "+err.Error()),
				}).Serialize())
				break
			}

			s.activeGatheringsMu.Lock()
			s.activeGatherings[playerID] = nodeID
			s.activeGatheringsMu.Unlock()

			conn.Write((&protocol.Packet{
				Opcode:  protocol.SC_GATHER_PROGRESS,
				Payload: protocol.EncodeGatherProgress(nodeID, duration),
			}).Serialize())

			go func(pID string, nID string, dur float64, currentConn net.Conn, pi *inventory.PlayerInventory) {
				time.Sleep(time.Duration(dur * float64(time.Second)))

				s.activeGatheringsMu.Lock()
				currentActive, ok := s.activeGatherings[pID]
				if ok && currentActive == nID {
					delete(s.activeGatherings, pID)
				} else {
					s.activeGatheringsMu.Unlock()
					return
				}
				s.activeGatheringsMu.Unlock()

				itemID, xp, err := s.professionsManager.CompleteGathering(pID, nID, pi)
				if err != nil {
					currentConn.Write((&protocol.Packet{
						Opcode:  protocol.SC_GATHER_COMPLETE,
						Payload: protocol.EncodeGatherComplete(nID, "", 0, 0),
					}).Serialize())
					currentConn.Write((&protocol.Packet{
						Opcode:  protocol.SC_CHAT_MESSAGE,
						Payload: protocol.EncodeChatMessage(0, "System", "Coleta falhou: "+err.Error()),
					}).Serialize())
					return
				}

				currentConn.Write((&protocol.Packet{
					Opcode:  protocol.SC_GATHER_COMPLETE,
					Payload: protocol.EncodeGatherComplete(nID, itemID, uint32(xp), 1),
				}).Serialize())

				stats, _ := s.combatManager.GetEntityStats(pID)
				s.sendInventorySync(currentConn, pID, stats, pi)

				currentConn.Write((&protocol.Packet{
					Opcode:  protocol.SC_CHAT_MESSAGE,
					Payload: protocol.EncodeChatMessage(0, "System", fmt.Sprintf("VocÃª coletou 1x %s e ganhou %d XP de ProfissÃ£o!", itemID, xp)),
				}).Serialize())
			}(playerID, nodeID, duration, conn, playerInv)

		case protocol.CS_GATHER_CANCEL:
			if playerID == "" {
				break
			}
			s.cancelGatheringIfActive(playerID)

		case protocol.CS_CRAFT_START:
			if playerID == "" {
				break
			}
			recipeID, err := protocol.DecodeCraftStart(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode CraftStart", "error", err)
				break
			}

			s.inventoriesMu.RLock()
			playerInv, existsInv := s.inventories[playerID]
			s.inventoriesMu.RUnlock()
			if !existsInv {
				break
			}

			outputItemID, xp, success, err := s.professionsManager.PerformCraft(playerID, recipeID, playerInv)
			if err != nil {
				conn.Write((&protocol.Packet{
					Opcode:  protocol.SC_CRAFT_RESPONSE,
					Payload: protocol.EncodeCraftResponse(recipeID, "", 0, 0),
				}).Serialize())
				conn.Write((&protocol.Packet{
					Opcode:  protocol.SC_CHAT_MESSAGE,
					Payload: protocol.EncodeChatMessage(0, "System", "CriaÃ§Ã£o falhou: "+err.Error()),
				}).Serialize())
				break
			}

			successVal := uint8(0)
			if success {
				successVal = 1
			}

			conn.Write((&protocol.Packet{
				Opcode:   protocol.SC_CRAFT_RESPONSE,
				Sequence: packet.Sequence,
				Payload:  protocol.EncodeCraftResponse(recipeID, outputItemID, uint32(xp), successVal),
			}).Serialize())

			stats, _ := s.combatManager.GetEntityStats(playerID)
			s.sendInventorySync(conn, playerID, stats, playerInv)

			if success {
				conn.Write((&protocol.Packet{
					Opcode:  protocol.SC_CHAT_MESSAGE,
					Payload: protocol.EncodeChatMessage(0, "System", fmt.Sprintf("VocÃª criou com sucesso 1x %s e ganhou %d XP!", outputItemID, xp)),
				}).Serialize())
			} else {
				conn.Write((&protocol.Packet{
					Opcode:  protocol.SC_CHAT_MESSAGE,
					Payload: protocol.EncodeChatMessage(0, "System", fmt.Sprintf("CriaÃ§Ã£o falhou! Os materiais foram consumidos, mas ganhou %d XP de consolaÃ§Ã£o.", xp)),
				}).Serialize())
			}

		case protocol.CS_DUNGEON_ENTER:
			if playerID == "" {
				break
			}
			dungeonID, mode, err := protocol.DecodeDungeonEnter(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode DungeonEnter", "error", err)
				break
			}
			err = s.dungeonManager.EnterDungeon(playerID, dungeonID, mode)
			if err != nil {
				conn.Write((&protocol.Packet{
					Opcode:  protocol.SC_CHAT_MESSAGE,
					Payload: protocol.EncodeChatMessage(0, "System", "Entrada falhou: "+err.Error()),
				}).Serialize())
			}

		case protocol.CS_DUNGEON_LEAVE:
			if playerID == "" {
				break
			}
			s.dungeonManager.LeaveDungeon(playerID)

		case protocol.CS_DUNGEON_CLAIM_LOOT:
			if playerID == "" {
				break
			}
			_, itemID, err := protocol.DecodeDungeonClaimLoot(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode ClaimLoot", "error", err)
				break
			}
			err = s.dungeonManager.ClaimDungeonLoot(playerID, itemID)
			if err != nil {
				conn.Write((&protocol.Packet{
					Opcode:  protocol.SC_CHAT_MESSAGE,
					Payload: protocol.EncodeChatMessage(0, "System", "Resgate falhou: "+err.Error()),
				}).Serialize())
			}

		case protocol.CS_WORLD_BOSS_SPAWN_REQ:
			s.dungeonManager.ForceWorldBossSpawn()

		case protocol.CS_CHOOSE_VOCATION:
			if playerID == "" {
				break
			}
			vocation, err := protocol.DecodeChooseVocationRequest(packet.Payload)
			if err != nil {
				slog.Error("Failed to decode choose vocation request", "error", err)
				break
			}
			err = s.progressionManager.ChooseVocation(playerID, vocation)
			var respPayload []byte
			if err != nil {
				slog.Warn("Choose vocation rejected", "player", playerID, "vocation", vocation, "error", err)
				respPayload = protocol.EncodeChooseVocationResponse(false, err.Error(), "")
			} else {
				slog.Info("Choose vocation succeeded", "player", playerID, "vocation", vocation)
				respPayload = protocol.EncodeChooseVocationResponse(true, "", vocation)

				// Sync inventory immediately to client to reflect new vocation stats
				s.inventoriesMu.RLock()
				playerInv, existsInv := s.inventories[playerID]
				s.inventoriesMu.RUnlock()
				if stats, existsStats := s.combatManager.GetEntityStats(playerID); existsStats && existsInv {
					s.sendInventorySync(conn, playerID, stats, playerInv)
				}
			}
			conn.Write((&protocol.Packet{
				Opcode:   protocol.SC_CHOOSE_VOCATION_RESP,
				Sequence: packet.Sequence,
				Payload:  respPayload,
			}).Serialize())

		case protocol.CS_UNLOCK_SUBCLASS:
			if playerID == "" {
				break
			}
			subclass, err := s.progressionManager.TriggerSubclassUnlock(playerID)
			var respPayload []byte
			if err != nil {
				slog.Warn("Unlock subclass rejected", "player", playerID, "error", err)
				respPayload = protocol.EncodeUnlockSubclassResponse(false, err.Error(), "", "")
			} else {
				slog.Info("Unlock subclass succeeded", "player", playerID, "subclass", subclass)
				stats, _ := s.combatManager.GetEntityStats(playerID)
				respPayload = protocol.EncodeUnlockSubclassResponse(true, "", subclass, stats.Element)

				// Sync inventory immediately to client to reflect new subclass stats
				s.inventoriesMu.RLock()
				playerInv, existsInv := s.inventories[playerID]
				s.inventoriesMu.RUnlock()
				if existsInv {
					s.sendInventorySync(conn, playerID, stats, playerInv)
				}
			}
			conn.Write((&protocol.Packet{
				Opcode:   protocol.SC_UNLOCK_SUBCLASS_RESP,
				Sequence: packet.Sequence,
				Payload:  respPayload,
			}).Serialize())
		}
	}
}

func (s *GatewayServer) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down Gateway Server gracefully...")

	// Parar PvE Manager (Sprint 3 Task 2)
	if s.pveManager != nil {
		s.pveManager.Stop()
	}

	// Parar loop de autosave
	if s.stopAutosave != nil {
		close(s.stopAutosave)
	}

	// Salva todos os personagens de forma sÃ­ncrona/atÃ´mica antes do desligamento do servidor (crash shutdown)
	slog.Info("Saving all active character states to PostgreSQL on shutdown...")
	s.saveAllActiveCharacters()

	// 1. Fecha o TCP Listener para novas conexÃµes
	if s.tcpListener != nil {
		s.tcpListener.Close()
	}

	// 2. Fecha conexÃµes HTTP do health check
	if s.httpServer != nil {
		s.httpServer.Shutdown(ctx)
	}

	// 3. Fecha conexÃµes de clientes ativos graciosamente
	s.clientsMu.Lock()
	for conn := range s.clients {
		conn.Close()
	}
	s.clientsMu.Unlock()

	// 4. Aguarda processamentos pendentes de pacotes
	waitChan := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		slog.Info("All active client routines finished.")
	case <-ctx.Done():
		slog.Warn("Shutdown timed out before client routines finished.")
	}

	// 5. Fecha pools de bancos de dados
	if s.pgPool != nil {
		s.pgPool.Close(ctx)
	}
	if s.redisClient != nil {
		s.redisClient.Close(ctx)
	}

	slog.Info("Gateway Server shutdown complete.")
	return nil
}

// Inicia o loop de autosave a cada 30 segundos usando ticker em goroutine dedicada
func (s *GatewayServer) startAutosaveLoop() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				slog.Info("Autosave ticker fired. Persisting all active character states...")
				s.saveAllActiveCharacters()
			case <-s.stopAutosave:
				slog.Info("Autosave loop stopped.")
				return
			}
		}
	}()
}

// Varre todos os inventÃ¡rios ativos cadastrados e os persiste no PostgreSQL
func (s *GatewayServer) saveAllActiveCharacters() {
	s.inventoriesMu.RLock()
	playerIDs := make([]string, 0, len(s.inventories))
	for id := range s.inventories {
		playerIDs = append(playerIDs, id)
	}
	s.inventoriesMu.RUnlock()

	for _, pid := range playerIDs {
		s.saveCharacterState(pid)
	}
}

// Salva de forma transacional e atÃ´mica o estado de combate, inventÃ¡rio e posiÃ§Ãµes fÃ­sicas do jogador (PATCH 1, 2, 3, 4)
func (s *GatewayServer) saveCharacterState(playerID string) {
	s.inventoriesMu.RLock()
	playerInv, ok := s.inventories[playerID]
	s.inventoriesMu.RUnlock()

	if !ok || playerInv == nil {
		return
	}

	// 1. Obter cÃ³pia thread-safe dos atributos de combate (evita race conditions)
	stats, exists := s.combatManager.GetEntityStatsCopy(playerID)
	if !exists {
		slog.Warn("Attempted to save player stats, but they are not registered in combatManager", "playerID", playerID)
		return
	}

	posX, posY, _ := s.combatManager.GetEntityPosition(playerID)
	posZ := 0.0

	// 2. Criar snapshot atÃ´mico e imutÃ¡vel (PATCH 1)
	snapshot := playerInv.CreateSnapshot(stats, posX, posY, posZ)

	// 3. PolÃ­tica de dirty-flag: Se nÃ£o estiver marcado como modificado, ignorar gravaÃ§Ã£o (PATCH 2)
	if !snapshot.IsDirty {
		slog.Debug("Skipping character state save: state is not dirty", "player", playerID)
		return
	}

	// 4. Salvar usando o snapshot imutÃ¡vel, o isolamento transacional estrito e o optimistic locking (PATCH 3 & 4)
	err := s.persistenceMgr.SaveCharacter(
		snapshot.PlayerID,
		&snapshot.Stats,
		snapshot.Items,
		snapshot.PosX,
		snapshot.PosY,
		snapshot.PosZ,
		snapshot.Version,
		pve.GetPlayerXp(snapshot.PlayerID),
		snapshot.Gold,
	)

	if err != nil {
		slog.Error("Failed to save character state", "player", playerID, "error", err)
	} else {
		// 5. ApÃ³s sucesso no banco de dados, limpa a flag dirty e incrementa a versÃ£o local (PATCH 2 & 4)
		playerInv.SetDirty(false)
		playerInv.SetVersion(snapshot.Version + 1)
		slog.Info("Successfully saved character state on database", "player", playerID, "new_version", snapshot.Version+1)
	}

	// Salva estado de quests e diÃ¡logos do jogador (PATCH 1 - Dirty Flag Writes)
	if err := s.questManager.SaveDirtyQuests(playerID); err != nil {
		slog.Error("Failed to save dirty quest and dialogue states on autosave", "player", playerID, "error", err)
	}
}

// Sincroniza de forma segura o inventÃ¡rio e bÃ´nus de atributos por rede via protocolo binÃ¡rio compacto
func (s *GatewayServer) sendInventorySync(conn net.Conn, playerID string, stats *combat.EntityStats, playerInv *inventory.PlayerInventory) {
	items := playerInv.GetItems()
	syncItems := make([]protocol.SyncItem, 0, len(items))
	for _, it := range items {
		syncItems = append(syncItems, protocol.SyncItem{
			ItemID:     it.ItemID,
			Quantity:   uint32(it.Quantity),
			Durability: uint32(it.Durability),
			SlotIndex:  uint16(it.SlotIndex),
		})
	}

	event := &protocol.InventorySyncEvent{
		Items:        syncItems,
		Level:        uint32(stats.Level),
		MaxHealth:    stats.MaxHealth,
		Health:       stats.Health,
		MaxMana:      stats.MaxMana,
		Mana:         stats.Mana,
		BaseAttack:   stats.BaseAttack,
		WeaponDamage: stats.WeaponDamage,
		Defense:      stats.Defense,
		Resistance:   stats.Resistance,
		CritChance:   stats.CritChance,
	}

	payload := protocol.EncodeInventorySync(event)
	pkt := &protocol.Packet{
		Opcode:   protocol.SC_INVENTORY_SYNC,
		Sequence: 0,
		Payload:  payload,
	}
	conn.Write(pkt.Serialize())
	slog.Info("Sent inventory sync packet to client", "player", playerID, "items_count", len(syncItems))
}

// broadcastToAll transmite um pacote para todos os clientes ativos (Sprint 4 Task 1)
func (s *GatewayServer) broadcastToAll(packet *protocol.Packet) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	serialized := packet.Serialize()
	for conn := range s.clients {
		conn.Write(serialized)
	}
}

// cancelGatheringIfActive cancela a coleta ativa do jogador caso ele se mova ou cancele (Sprint 4 Task 1)
func (s *GatewayServer) cancelGatheringIfActive(playerID string) {
	s.activeGatheringsMu.Lock()
	nodeID, ok := s.activeGatherings[playerID]
	if ok {
		delete(s.activeGatherings, playerID)
		s.activeGatheringsMu.Unlock()
		s.professionsManager.CancelGathering(playerID, nodeID)

		// Notifica o cliente do cancelamento
		if conn, ok := s.aoiManager.GetPlayerConn(playerID); ok {
			conn.Write((&protocol.Packet{
				Opcode:  protocol.SC_GATHER_COMPLETE,
				Payload: protocol.EncodeGatherComplete(nodeID, "", 0, 0),
			}).Serialize())
			conn.Write((&protocol.Packet{
				Opcode:  protocol.SC_CHAT_MESSAGE,
				Payload: protocol.EncodeChatMessage(0, "System", "Coleta cancelada devido ao movimento."),
			}).Serialize())
		}
	} else {
		s.activeGatheringsMu.Unlock()
	}
}

// broadcastTradeUpdate transmite o estado atual da proposta comercial em binÃ¡rio para ambos participantes (Dual Confirm)
func (s *GatewayServer) broadcastTradeUpdate(playerID string) {
	session, exists := s.economyManager.GetTradeSession(playerID)
	if !exists {
		return
	}

	itemsA := make([]protocol.TradeItemCodec, 0, len(session.ItemsA))
	for _, it := range session.ItemsA {
		if it != nil {
			itemsA = append(itemsA, protocol.TradeItemCodec{
				SlotIndex: uint32(it.SlotIndex),
				ItemID:    it.ItemID,
				Quantity:  uint32(it.Quantity),
				ItemUUID:  it.ItemUUID,
			})
		}
	}

	itemsB := make([]protocol.TradeItemCodec, 0, len(session.ItemsB))
	for _, it := range session.ItemsB {
		if it != nil {
			itemsB = append(itemsB, protocol.TradeItemCodec{
				SlotIndex: uint32(it.SlotIndex),
				ItemID:    it.ItemID,
				Quantity:  uint32(it.Quantity),
				ItemUUID:  it.ItemUUID,
			})
		}
	}

	lockedAVal := uint8(0)
	if session.LockedA {
		lockedAVal = 1
	}
	lockedBVal := uint8(0)
	if session.LockedB {
		lockedBVal = 1
	}
	acceptedAVal := uint8(0)
	if session.AcceptedA {
		acceptedAVal = 1
	}
	acceptedBVal := uint8(0)
	if session.AcceptedB {
		acceptedBVal = 1
	}

	event := &protocol.TradeUpdateEvent{
		GoldA:     uint32(session.GoldA),
		GoldB:     uint32(session.GoldB),
		LockedA:   lockedAVal,
		LockedB:   lockedBVal,
		AcceptedA: acceptedAVal,
		AcceptedB: acceptedBVal,
		ItemsA:    itemsA,
		ItemsB:    itemsB,
	}

	payload := protocol.EncodeTradeUpdate(event)
	pkt := &protocol.Packet{
		Opcode:  protocol.SC_TRADE_UPDATE,
		Payload: payload,
	}
	serialized := pkt.Serialize()

	if connA, ok := s.aoiManager.GetPlayerConn(session.PlayerA); ok {
		connA.Write(serialized)
	}
	if connB, ok := s.aoiManager.GetPlayerConn(session.PlayerB); ok {
		connB.Write(serialized)
	}
}
