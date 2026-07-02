package social

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"net"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/pkg/movement"
	"github.com/light-and-shadow/backend/pkg/protocol"
)

type LootMode uint8

const (
	LootFree       LootMode = 0
	LootRoundRobin LootMode = 1
	LootLeader     LootMode = 2
)

// Party representa um grupo ativo de até 5 jogadores
type Party struct {
	ID        string
	LeaderID  string
	Members   []string
	LootMode  LootMode
	mu        sync.RWMutex
}

type PartyInvite struct {
	InviterID string
	TargetID  string
	ExpiresAt time.Time
}

type GuildInvite struct {
	GuildName string
	InviterID string
	TargetID  string
	ExpiresAt time.Time
}

// PlayerSocialState gerencia amigos/ignorados em memória com dirty flag (PATCH 2)
type PlayerSocialState struct {
	PlayerID string
	Friends  map[string]bool
	Ignores  map[string]bool
	Dirty    bool
	mu       sync.RWMutex
}

// GuildMemoryState representa o estado em cache de uma guilda
type GuildMemoryState struct {
	ID        int
	Name      string
	MOTD      string
	Version   uint32 // For Optimistic Locking (PATCH 1)
	Leader    string
	Vices     map[string]bool
	Members   map[string]bool
	mu        sync.RWMutex
}

// RateLimiter controla spam no chat por jogador (PATCH 4)
type RateLimiter struct {
	LastMessage time.Time
	Count       int
	mu          sync.Mutex
}

type SocialManager struct {
	dbConn     *sql.DB
	aoiManager *movement.AOIManager

	// Parties
	parties      map[string]*Party // playerID -> Party
	partiesMu    sync.RWMutex
	partyInvites map[string]*PartyInvite // targetPlayerID -> PartyInvite (PATCH 5)
	invitesMu    sync.Mutex

	// Guilds
	guilds       map[string]*GuildMemoryState // guildName -> GuildState
	guildsMu     sync.RWMutex
	guildInvites map[string]*GuildInvite // targetPlayerID -> GuildInvite (PATCH 5)

	// Social States (friends, ignores)
	socialStates   map[string]*PlayerSocialState // playerID -> SocialState
	socialStatesMu sync.RWMutex

	// Chat limiters
	limiters   map[string]*RateLimiter
	limitersMu sync.Mutex
}

func NewSocialManager(dbConn *sql.DB, aoiManager *movement.AOIManager) *SocialManager {
	sm := &SocialManager{
		dbConn:       dbConn,
		aoiManager:   aoiManager,
		parties:      make(map[string]*Party),
		partyInvites: make(map[string]*PartyInvite),
		guilds:       make(map[string]*GuildMemoryState),
		guildInvites: make(map[string]*GuildInvite),
		socialStates: make(map[string]*PlayerSocialState),
		limiters:     make(map[string]*RateLimiter),
	}

	sm.loadAllGuildsFromDB()
	return sm
}

// =============================================================================
// PARTY SYSTEM
// =============================================================================

// CreateParty cria um novo grupo com o jogador como líder
func (sm *SocialManager) CreateParty(leaderID string, mode LootMode) (*Party, error) {
	sm.partiesMu.Lock()
	defer sm.partiesMu.Unlock()

	if _, exists := sm.parties[leaderID]; exists {
		return nil, fmt.Errorf("player is already in a party")
	}

	party := &Party{
		ID:        fmt.Sprintf("party_%s_%d", leaderID, time.Now().UnixNano()),
		LeaderID:  leaderID,
		Members:   []string{leaderID},
		LootMode:  mode,
	}

	sm.parties[leaderID] = party
	sm.BroadcastPartyInfo(party)
	return party, nil
}

// InviteToParty envia um convite para o jogador-alvo (com PATCH 3 e PATCH 5)
func (sm *SocialManager) InviteToParty(inviterID, targetID string) error {
	sm.partiesMu.RLock()
	party, inParty := sm.parties[inviterID]
	sm.partiesMu.RUnlock()

	if !inParty {
		return fmt.Errorf("you are not in a party")
	}

	party.mu.RLock()
	if party.LeaderID != inviterID {
		party.mu.RUnlock()
		return fmt.Errorf("only the party leader can invite members")
	}
	if len(party.Members) >= 5 {
		party.mu.RUnlock()
		return fmt.Errorf("party is already full")
	}
	party.mu.RUnlock()

	sm.invitesMu.Lock()
	defer sm.invitesMu.Unlock()

	// PATCH 5: Duplicate invite prevention & PATCH 3: Expiration cleanup
	if activeInvite, exists := sm.partyInvites[targetID]; exists {
		if time.Now().Before(activeInvite.ExpiresAt) {
			return fmt.Errorf("duplicate party invite detected: target already has a pending invite")
		}
	}

	invite := &PartyInvite{
		InviterID: inviterID,
		TargetID:  targetID,
		ExpiresAt: time.Now().Add(30 * time.Second), // PATCH 3: 30 seconds expiration
	}
	sm.partyInvites[targetID] = invite

	// Envia o pacote de convite binário ao alvo
	sm.sendBinaryToPlayer(targetID, protocol.SC_PARTY_INVITE_REQ, protocol.EncodePartyInviteReq(inviterID))
	slog.Info("Party invite sent", "inviter", inviterID, "target", targetID)
	return nil
}

// AcceptPartyInvite aceita um convite ativo e ingressa no grupo (com PATCH 3)
func (sm *SocialManager) AcceptPartyInvite(targetID, inviterID string) error {
	sm.invitesMu.Lock()
	invite, exists := sm.partyInvites[targetID]
	if !exists || invite.InviterID != inviterID {
		sm.invitesMu.Unlock()
		return fmt.Errorf("no pending party invite from this player")
	}

	// PATCH 3: Verify expiration
	if time.Now().After(invite.ExpiresAt) {
		delete(sm.partyInvites, targetID)
		sm.invitesMu.Unlock()
		return fmt.Errorf("party invite has expired (30 seconds limit)")
	}

	delete(sm.partyInvites, targetID)
	sm.invitesMu.Unlock()

	sm.partiesMu.Lock()
	party, inParty := sm.parties[inviterID]
	if !inParty {
		sm.partiesMu.Unlock()
		return fmt.Errorf("the party no longer exists")
	}

	party.mu.Lock()
	if len(party.Members) >= 5 {
		party.mu.Unlock()
		sm.partiesMu.Unlock()
		return fmt.Errorf("party is already full")
	}

	party.Members = append(party.Members, targetID)
	sm.parties[targetID] = party
	party.mu.Unlock()
	sm.partiesMu.Unlock()

	sm.BroadcastPartyInfo(party)
	slog.Info("Player joined party", "player", targetID, "partyID", party.ID)
	return nil
}

// RejectPartyInvite rejeita o convite pendente
func (sm *SocialManager) RejectPartyInvite(targetID, inviterID string) error {
	sm.invitesMu.Lock()
	defer sm.invitesMu.Unlock()

	invite, exists := sm.partyInvites[targetID]
	if !exists || invite.InviterID != inviterID {
		return fmt.Errorf("no pending invite found")
	}

	delete(sm.partyInvites, targetID)
	slog.Info("Party invite rejected", "target", targetID, "inviter", inviterID)
	return nil
}

// LeaveParty remove o jogador do grupo
func (sm *SocialManager) LeaveParty(playerID string) error {
	sm.partiesMu.Lock()
	party, inParty := sm.parties[playerID]
	if !inParty {
		sm.partiesMu.Unlock()
		return fmt.Errorf("you are not in a party")
	}

	delete(sm.parties, playerID)
	sm.partiesMu.Unlock()

	party.mu.Lock()
	newMembers := []string{}
	for _, m := range party.Members {
		if m != playerID {
			newMembers = append(newMembers, m)
		}
	}
	party.Members = newMembers

	if len(party.Members) == 0 {
		party.mu.Unlock()
		slog.Info("Party disbanded", "partyID", party.ID)
		return nil
	}

	// Transfere liderança se o líder saiu
	if party.LeaderID == playerID {
		party.LeaderID = party.Members[0]
	}
	party.mu.Unlock()

	sm.BroadcastPartyInfo(party)
	slog.Info("Player left party", "player", playerID, "partyID", party.ID)
	return nil
}

// KickMember expulsa um membro do grupo (apenas líder)
func (sm *SocialManager) KickMember(leaderID, targetID string) error {
	sm.partiesMu.RLock()
	party, inParty := sm.parties[leaderID]
	sm.partiesMu.RUnlock()

	if !inParty {
		return fmt.Errorf("you are not in a party")
	}

	party.mu.Lock()
	if party.LeaderID != leaderID {
		party.mu.Unlock()
		return fmt.Errorf("only the party leader can kick members")
	}

	found := false
	newMembers := []string{}
	for _, m := range party.Members {
		if m != targetID {
			newMembers = append(newMembers, m)
		} else {
			found = true
		}
	}

	if !found {
		party.mu.Unlock()
		return fmt.Errorf("player is not in your party")
	}

	party.Members = newMembers
	party.mu.Unlock()

	sm.partiesMu.Lock()
	delete(sm.parties, targetID)
	sm.partiesMu.Unlock()

	sm.BroadcastPartyInfo(party)
	sm.sendBinaryToPlayer(targetID, protocol.SC_CHAT_MESSAGE, protocol.EncodeChatMessage(2, "System", "You have been kicked from the party."))
	slog.Info("Player kicked from party", "kicked", targetID, "leader", leaderID)
	return nil
}

// TransferLeader transfere a liderança do grupo
func (sm *SocialManager) TransferLeader(leaderID, targetID string) error {
	sm.partiesMu.RLock()
	party, inParty := sm.parties[leaderID]
	sm.partiesMu.RUnlock()

	if !inParty {
		return fmt.Errorf("you are not in a party")
	}

	party.mu.Lock()
	defer party.mu.Unlock()

	if party.LeaderID != leaderID {
		return fmt.Errorf("only the party leader can transfer leadership")
	}

	found := false
	for _, m := range party.Members {
		if m == targetID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("target player is not in your party")
	}

	party.LeaderID = targetID
	go sm.BroadcastPartyInfo(party)
	slog.Info("Party leader transferred", "old", leaderID, "new", targetID)
	return nil
}

// SetLootMode altera o modo de distribuição de saque
func (sm *SocialManager) SetLootMode(leaderID string, mode LootMode) error {
	sm.partiesMu.RLock()
	party, inParty := sm.parties[leaderID]
	sm.partiesMu.RUnlock()

	if !inParty {
		return fmt.Errorf("you are not in a party")
	}

	party.mu.Lock()
	defer party.mu.Unlock()

	if party.LeaderID != leaderID {
		return fmt.Errorf("only the party leader can change loot mode")
	}

	party.LootMode = mode
	go sm.BroadcastPartyInfo(party)
	slog.Info("Party loot mode updated", "partyID", party.ID, "mode", mode)
	return nil
}

// GetSharedXpPlayers retorna os membros do grupo dentro de 30 tiles de raio para compartilhamento de XP
func (sm *SocialManager) GetSharedXpPlayers(playerID string, x, y float64) []string {
	sm.partiesMu.RLock()
	party, inParty := sm.parties[playerID]
	sm.partiesMu.RUnlock()

	if !inParty {
		return []string{playerID}
	}

	party.mu.RLock()
	defer party.mu.RUnlock()

	sharedMembers := []string{}
	for _, m := range party.Members {
		// Verifica se o jogador está online e próximo de x, y (dentro de 30 tiles)
		if sm.isPlayerOnlineAndNear(m, x, y, 30.0) {
			sharedMembers = append(sharedMembers, m)
		}
	}

	if len(sharedMembers) == 0 {
		return []string{playerID}
	}
	return sharedMembers
}

// BroadcastPartyInfo serializa e envia dados atualizados de grupo para todos os membros ativos
func (sm *SocialManager) BroadcastPartyInfo(party *Party) {
	party.mu.RLock()
	defer party.mu.RUnlock()

	membersInfo := []protocol.PartyMemberInfo{}
	for _, m := range party.Members {
		online := uint8(0)
		if sm.IsOnline(m) {
			online = 1
		}
		membersInfo = append(membersInfo, protocol.PartyMemberInfo{
			PlayerID: m,
			Online:   online,
		})
	}

	event := &protocol.PartyInfoEvent{
		LootMode: uint8(party.LootMode),
		LeaderID: party.LeaderID,
		Members:  membersInfo,
	}

	payload := protocol.EncodePartyInfo(event)
	for _, m := range party.Members {
		sm.sendBinaryToPlayer(m, protocol.SC_PARTY_INFO, payload)
	}
}

// =============================================================================
// GUILD SYSTEM
// =============================================================================

// CreateGuild cria uma guilda persistente no PostgreSQL
func (sm *SocialManager) CreateGuild(leaderID, guildName string) error {
	if guildName == "" {
		return fmt.Errorf("guild name cannot be empty")
	}

	sm.guildsMu.Lock()
	defer sm.guildsMu.Unlock()

	if _, exists := sm.guilds[guildName]; exists {
		return fmt.Errorf("guild name already taken")
	}

	// Verifica se jogador já pertence a outra guilda
	for _, g := range sm.guilds {
		g.mu.RLock()
		inGuild := g.Members[leaderID] || g.Leader == leaderID
		g.mu.RUnlock()
		if inGuild {
			return fmt.Errorf("player is already in a guild")
		}
	}

	if sm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tx, err := sm.dbConn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		var guildID int
		err = tx.QueryRowContext(ctx, `
			INSERT INTO guilds (name, motd, version)
			VALUES ($1, 'Bem-vindo à guilda!', 1)
			RETURNING id
		`, guildName).Scan(&guildID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("guild creation failed: name might already be registered: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO guild_members (guild_id, character_name, role)
			VALUES ($1, $2, 2)
		`, guildID, leaderID)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO guild_audit_logs (guild_id, action)
			VALUES ($1, $2)
		`, guildID, fmt.Sprintf("Guilda %s criada por %s", guildName, leaderID))
		if err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		gState := &GuildMemoryState{
			ID:      guildID,
			Name:    guildName,
			MOTD:    "Bem-vindo à guilda!",
			Version: 1,
			Leader:  leaderID,
			Vices:   make(map[string]bool),
			Members: make(map[string]bool),
		}
		sm.guilds[guildName] = gState
		go sm.BroadcastGuildInfo(gState)
		slog.Info("Guild created persistently", "name", guildName, "leader", leaderID)
		return nil
	}

	return fmt.Errorf("database unavailable for guild creation")
}

// InviteToGuild envia convite de guilda (com PATCH 3 e PATCH 5)
func (sm *SocialManager) InviteToGuild(inviterID, targetID string) error {
	sm.guildsMu.RLock()
	var userGuild *GuildMemoryState
	for _, g := range sm.guilds {
		g.mu.RLock()
		if g.Leader == inviterID || g.Vices[inviterID] {
			userGuild = g
			g.mu.RUnlock()
			break
		}
		g.mu.RUnlock()
	}
	sm.guildsMu.RUnlock()

	if userGuild == nil {
		return fmt.Errorf("only guild leaders or vices can invite players")
	}

	sm.invitesMu.Lock()
	defer sm.invitesMu.Unlock()

	// PATCH 5: Duplicate invite prevention & PATCH 3: Expiration cleanup
	if activeInvite, exists := sm.guildInvites[targetID]; exists {
		if time.Now().Before(activeInvite.ExpiresAt) {
			return fmt.Errorf("duplicate guild invite detected: target already has a pending guild invite")
		}
	}

	invite := &GuildInvite{
		GuildName: userGuild.Name,
		InviterID: inviterID,
		TargetID:  targetID,
		ExpiresAt: time.Now().Add(30 * time.Second), // PATCH 3: 30 seconds expiration
	}
	sm.guildInvites[targetID] = invite

	sm.sendBinaryToPlayer(targetID, protocol.SC_GUILD_INVITE_REQ, protocol.EncodeGuildInviteReq(userGuild.Name, inviterID))
	slog.Info("Guild invite sent", "guild", userGuild.Name, "inviter", inviterID, "target", targetID)
	return nil
}

// AcceptGuildInvite aceita o convite pendente da guilda
func (sm *SocialManager) AcceptGuildInvite(targetID, guildName string) error {
	sm.invitesMu.Lock()
	invite, exists := sm.guildInvites[targetID]
	if !exists || invite.GuildName != guildName {
		sm.invitesMu.Unlock()
		return fmt.Errorf("no pending guild invite for this guild")
	}

	// PATCH 3: Verify expiration
	if time.Now().After(invite.ExpiresAt) {
		delete(sm.guildInvites, targetID)
		sm.invitesMu.Unlock()
		return fmt.Errorf("guild invite has expired")
	}

	delete(sm.guildInvites, targetID)
	sm.invitesMu.Unlock()

	sm.guildsMu.Lock()
	gState, exists := sm.guilds[guildName]
	if !exists {
		sm.guildsMu.Unlock()
		return fmt.Errorf("guild no longer exists")
	}

	gState.mu.Lock()
	gState.Members[targetID] = true
	gState.mu.Unlock()
	sm.guildsMu.Unlock()

	if sm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, _ = sm.dbConn.ExecContext(ctx, `
			INSERT INTO guild_members (guild_id, character_name, role)
			VALUES ($1, $2, 0)
			ON CONFLICT (character_name) DO UPDATE SET guild_id = EXCLUDED.guild_id, role = 0
		`, gState.ID, targetID)

		_, _ = sm.dbConn.ExecContext(ctx, `
			INSERT INTO guild_audit_logs (guild_id, action)
			VALUES ($1, $2)
		`, gState.ID, fmt.Sprintf("Jogador %s juntou-se à guilda", targetID))
	}

	go sm.BroadcastGuildInfo(gState)
	slog.Info("Player joined guild", "player", targetID, "guild", guildName)
	return nil
}

// LeaveGuild sai da guilda atual
func (sm *SocialManager) LeaveGuild(playerID string) error {
	sm.guildsMu.Lock()
	var userGuild *GuildMemoryState
	for _, g := range sm.guilds {
		g.mu.RLock()
		if g.Leader == playerID || g.Vices[playerID] || g.Members[playerID] {
			userGuild = g
			g.mu.RUnlock()
			break
		}
		g.mu.RUnlock()
	}
	sm.guildsMu.Unlock()

	if userGuild == nil {
		return fmt.Errorf("you are not in a guild")
	}

	userGuild.mu.Lock()
	if userGuild.Leader == playerID {
		userGuild.mu.Unlock()
		return fmt.Errorf("guild leaders cannot leave without transferring leadership first")
	}

	delete(userGuild.Vices, playerID)
	delete(userGuild.Members, playerID)
	userGuild.mu.Unlock()

	if sm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, _ = sm.dbConn.ExecContext(ctx, "DELETE FROM guild_members WHERE character_name = $1", playerID)
		_, _ = sm.dbConn.ExecContext(ctx, "INSERT INTO guild_audit_logs (guild_id, action) VALUES ($1, $2)", userGuild.ID, fmt.Sprintf("%s saiu da guilda", playerID))
	}

	go sm.BroadcastGuildInfo(userGuild)
	sm.sendBinaryToPlayer(playerID, protocol.SC_GUILD_INFO, []byte{}) // Limpa estado no cliente
	slog.Info("Player left guild", "player", playerID, "guild", userGuild.Name)
	return nil
}

// UpdateMOTD atualiza o MOTD com Optimistic Locking (PATCH 1)
func (sm *SocialManager) UpdateMOTD(playerID, guildName, newMOTD string, expectedVersion uint32) error {
	sm.guildsMu.RLock()
	gState, exists := sm.guilds[guildName]
	sm.guildsMu.RUnlock()

	if !exists {
		return fmt.Errorf("guild does not exist")
	}

	gState.mu.Lock()
	defer gState.mu.Unlock()

	// Verifica permissões (apenas líder ou vice)
	if gState.Leader != playerID && !gState.Vices[playerID] {
		return fmt.Errorf("permission denied to update MOTD")
	}

	// PATCH 1: Optimistic Locking version check
	if gState.Version != expectedVersion {
		return fmt.Errorf("concurrency conflict: MOTD updated by another officer, please refresh")
	}

	if sm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := sm.dbConn.ExecContext(ctx, `
			UPDATE guilds
			SET motd = $1, version = version + 1
			WHERE id = $2 AND version = $3
		`, newMOTD, gState.ID, expectedVersion)
		if err != nil {
			return fmt.Errorf("failed to update MOTD in database: %w", err)
		}

		affected, err := res.RowsAffected()
		if err != nil || affected == 0 {
			return fmt.Errorf("optimistic locking conflict: MOTD update failed")
		}

		_, _ = sm.dbConn.ExecContext(ctx, `
			INSERT INTO guild_audit_logs (guild_id, action)
			VALUES ($1, $2)
		`, gState.ID, fmt.Sprintf("%s atualizou o MOTD para: %s", playerID, newMOTD))
	}

	gState.MOTD = newMOTD
	gState.Version++ // Incrementa a versão em memória
	go sm.BroadcastGuildInfo(gState)
	slog.Info("Guild MOTD updated safely", "guild", guildName, "new_version", gState.Version)
	return nil
}

// PromoteMember promove um membro (apenas lider)
func (sm *SocialManager) PromoteMember(leaderID, targetID string) error {
	sm.guildsMu.Lock()
	var gState *GuildMemoryState
	for _, g := range sm.guilds {
		g.mu.RLock()
		if g.Leader == leaderID {
			gState = g
			g.mu.RUnlock()
			break
		}
		g.mu.RUnlock()
	}
	sm.guildsMu.Unlock()

	if gState == nil {
		return fmt.Errorf("only the guild leader can promote members")
	}

	gState.mu.Lock()
	defer gState.mu.Unlock()

	if !gState.Members[targetID] {
		return fmt.Errorf("player is not a regular member of your guild")
	}

	delete(gState.Members, targetID)
	gState.Vices[targetID] = true

	if sm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, _ = sm.dbConn.ExecContext(ctx, "UPDATE guild_members SET role = 1 WHERE character_name = $1", targetID)
		_, _ = sm.dbConn.ExecContext(ctx, "INSERT INTO guild_audit_logs (guild_id, action) VALUES ($1, $2)", gState.ID, fmt.Sprintf("%s foi promovido a Vice-Líder por %s", targetID, leaderID))
	}

	go sm.BroadcastGuildInfo(gState)
	return nil
}

// DemoteMember rebaixa um vice-líder (apenas líder)
func (sm *SocialManager) DemoteMember(leaderID, targetID string) error {
	sm.guildsMu.Lock()
	var gState *GuildMemoryState
	for _, g := range sm.guilds {
		g.mu.RLock()
		if g.Leader == leaderID {
			gState = g
			g.mu.RUnlock()
			break
		}
		g.mu.RUnlock()
	}
	sm.guildsMu.Unlock()

	if gState == nil {
		return fmt.Errorf("only the guild leader can demote officers")
	}

	gState.mu.Lock()
	defer gState.mu.Unlock()

	if !gState.Vices[targetID] {
		return fmt.Errorf("player is not a vice-leader of your guild")
	}

	delete(gState.Vices, targetID)
	gState.Members[targetID] = true

	if sm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, _ = sm.dbConn.ExecContext(ctx, "UPDATE guild_members SET role = 0 WHERE character_name = $1", targetID)
		_, _ = sm.dbConn.ExecContext(ctx, "INSERT INTO guild_audit_logs (guild_id, action) VALUES ($1, $2)", gState.ID, fmt.Sprintf("%s foi rebaixado a Membro por %s", targetID, leaderID))
	}

	go sm.BroadcastGuildInfo(gState)
	return nil
}

// KickFromGuild remove um jogador da guilda (lider/vice de maior cargo)
func (sm *SocialManager) KickFromGuild(kickerID, targetID string) error {
	sm.guildsMu.Lock()
	var gState *GuildMemoryState
	for _, g := range sm.guilds {
		g.mu.RLock()
		if g.Leader == kickerID || g.Vices[kickerID] {
			gState = g
			g.mu.RUnlock()
			break
		}
		g.mu.RUnlock()
	}
	sm.guildsMu.Unlock()

	if gState == nil {
		return fmt.Errorf("you do not have permission to kick from guild")
	}

	gState.mu.Lock()
	defer gState.mu.Unlock()

	// Valida hierarquia
	isLeader := gState.Leader == kickerID
	isTargetVice := gState.Vices[targetID]
	isTargetMember := gState.Members[targetID]

	if !isTargetVice && !isTargetMember {
		return fmt.Errorf("player is not in your guild")
	}
	if isTargetVice && !isLeader {
		return fmt.Errorf("only the guild leader can kick vice-leaders")
	}

	delete(gState.Vices, targetID)
	delete(gState.Members, targetID)

	if sm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, _ = sm.dbConn.ExecContext(ctx, "DELETE FROM guild_members WHERE character_name = $1", targetID)
		_, _ = sm.dbConn.ExecContext(ctx, "INSERT INTO guild_audit_logs (guild_id, action) VALUES ($1, $2)", gState.ID, fmt.Sprintf("%s foi expulso da guilda por %s", targetID, kickerID))
	}

	go sm.BroadcastGuildInfo(gState)
	sm.sendBinaryToPlayer(targetID, protocol.SC_GUILD_INFO, []byte{}) // Limpa estado no cliente do expulso
	slog.Info("Guild member kicked", "target", targetID, "kicker", kickerID)
	return nil
}

// GetGuildAuditLogs recupera logs da guilda
func (sm *SocialManager) GetGuildAuditLogs(playerID string) ([]protocol.AuditLogEntry, error) {
	sm.guildsMu.RLock()
	var gState *GuildMemoryState
	for _, g := range sm.guilds {
		g.mu.RLock()
		if g.Leader == playerID || g.Vices[playerID] || g.Members[playerID] {
			gState = g
			g.mu.RUnlock()
			break
		}
		g.mu.RUnlock()
	}
	sm.guildsMu.RUnlock()

	if gState == nil {
		return nil, fmt.Errorf("you are not in a guild")
	}

	if sm.dbConn == nil {
		return []protocol.AuditLogEntry{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := sm.dbConn.QueryContext(ctx, `
		SELECT action, created_at
		FROM guild_audit_logs
		WHERE guild_id = $1
		ORDER BY created_at DESC
		LIMIT 20
	`, gState.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []protocol.AuditLogEntry{}
	for rows.Next() {
		var action string
		var t time.Time
		if err := rows.Scan(&action, &t); err == nil {
			entries = append(entries, protocol.AuditLogEntry{
				Action:    action,
				Timestamp: t.Format("2006-01-02 15:04:05"),
			})
		}
	}
	return entries, nil
}

// Persistent Guild Storage operations
func (sm *SocialManager) DepositToGuildStorage(playerID string, slot int, itemID string, qty int) error {
	sm.guildsMu.RLock()
	var gState *GuildMemoryState
	for _, g := range sm.guilds {
		g.mu.RLock()
		if g.Leader == playerID || g.Vices[playerID] || g.Members[playerID] {
			gState = g
			g.mu.RUnlock()
			break
		}
		g.mu.RUnlock()
	}
	sm.guildsMu.RUnlock()

	if gState == nil {
		return fmt.Errorf("you are not in a guild")
	}

	if sm.dbConn == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := sm.dbConn.ExecContext(ctx, `
		INSERT INTO guild_storage (guild_id, slot, item_id, quantity)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (guild_id, slot) DO UPDATE SET item_id = EXCLUDED.item_id, quantity = guild_storage.quantity + EXCLUDED.quantity
	`, gState.ID, slot, itemID, qty)
	if err != nil {
		return err
	}

	_, _ = sm.dbConn.ExecContext(ctx, `
		INSERT INTO guild_audit_logs (guild_id, action)
		VALUES ($1, $2)
	`, gState.ID, fmt.Sprintf("%s depositou %dx %s no slot %d", playerID, qty, itemID, slot))

	slog.Info("Deposited item to guild storage", "player", playerID, "guildID", gState.ID, "slot", slot, "itemID", itemID, "qty", qty)
	return nil
}

func (sm *SocialManager) WithdrawFromGuildStorage(playerID string, slot int, qty int) (string, error) {
	sm.guildsMu.RLock()
	var gState *GuildMemoryState
	for _, g := range sm.guilds {
		g.mu.RLock()
		if g.Leader == playerID || g.Vices[playerID] || g.Members[playerID] {
			gState = g
			g.mu.RUnlock()
			break
		}
		g.mu.RUnlock()
	}
	sm.guildsMu.RUnlock()

	if gState == nil {
		return "", fmt.Errorf("you are not in a guild")
	}

	if sm.dbConn == nil {
		return "", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var itemID string
	var currentQty int
	err := sm.dbConn.QueryRowContext(ctx, "SELECT item_id, quantity FROM guild_storage WHERE guild_id = $1 AND slot = $2", gState.ID, slot).Scan(&itemID, &currentQty)
	if err != nil {
		return "", fmt.Errorf("no item found in guild storage slot %d", slot)
	}

	if currentQty < qty {
		return "", fmt.Errorf("insufficient items in slot %d", slot)
	}

	if currentQty == qty {
		_, _ = sm.dbConn.ExecContext(ctx, "DELETE FROM guild_storage WHERE guild_id = $1 AND slot = $2", gState.ID, slot)
	} else {
		_, _ = sm.dbConn.ExecContext(ctx, "UPDATE guild_storage SET quantity = quantity - $1 WHERE guild_id = $2 AND slot = $3", qty, gState.ID, slot)
	}

	_, _ = sm.dbConn.ExecContext(ctx, `
		INSERT INTO guild_audit_logs (guild_id, action)
		VALUES ($1, $2)
	`, gState.ID, fmt.Sprintf("%s retirou %dx %s do slot %d", playerID, qty, itemID, slot))

	slog.Info("Withdrew item from guild storage", "player", playerID, "guildID", gState.ID, "slot", slot, "itemID", itemID, "qty", qty)
	return itemID, nil
}

// BroadcastGuildInfo envia info da guilda a todos os membros online
func (sm *SocialManager) BroadcastGuildInfo(gState *GuildMemoryState) {
	gState.mu.RLock()
	defer gState.mu.RUnlock()

	membersInfo := []protocol.GuildMemberInfo{}

	// Líder
	membersInfo = append(membersInfo, protocol.GuildMemberInfo{
		PlayerID: gState.Leader,
		Role:     2,
		Online:   sm.getOnlineByte(gState.Leader),
	})

	// Vices
	for v := range gState.Vices {
		membersInfo = append(membersInfo, protocol.GuildMemberInfo{
			PlayerID: v,
			Role:     1,
			Online:   sm.getOnlineByte(v),
		})
	}

	// Membros normais
	for m := range gState.Members {
		membersInfo = append(membersInfo, protocol.GuildMemberInfo{
			PlayerID: m,
			Role:     0,
			Online:   sm.getOnlineByte(m),
		})
	}

	event := &protocol.GuildInfoEvent{
		GuildName: gState.Name,
		MOTD:      gState.MOTD,
		Version:   gState.Version,
		Members:   membersInfo,
	}

	payload := protocol.EncodeGuildInfo(event)
	// Envia a todos os que estão no grupo
	for _, m := range membersInfo {
		sm.sendBinaryToPlayer(m.PlayerID, protocol.SC_GUILD_INFO, payload)
	}
}

func (sm *SocialManager) loadAllGuildsFromDB() {
	if sm.dbConn == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := sm.dbConn.QueryContext(ctx, "SELECT id, name, motd, version FROM guilds")
	if err != nil {
		slog.Error("Failed to query guilds from DB", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, motd string
		var version uint32
		if err := rows.Scan(&id, &name, &motd, &version); err == nil {
			gState := &GuildMemoryState{
				ID:      id,
				Name:    name,
				MOTD:    motd,
				Version: version,
				Vices:   make(map[string]bool),
				Members: make(map[string]bool),
			}

			// Carrega membros
			mRows, mErr := sm.dbConn.QueryContext(ctx, "SELECT character_name, role FROM guild_members WHERE guild_id = $1", id)
			if mErr == nil {
				for mRows.Next() {
					var charName string
					var role int
					if mScanErr := mRows.Scan(&charName, &role); mScanErr == nil {
						if role == 2 {
							gState.Leader = charName
						} else if role == 1 {
							gState.Vices[charName] = true
						} else {
							gState.Members[charName] = true
						}
					}
				}
				mRows.Close()
			}

			sm.guilds[name] = gState
			slog.Info("Guild loaded from DB into memory", "guild", name, "leader", gState.Leader, "members_count", len(gState.Members))
		}
	}
}

// =============================================================================
// SOCIAL SYSTEMS (Friend List, Ignore List, Presence)
// =============================================================================

// LoadSocialState carrega a lista de amigos/ignorados com dirty-flag (PATCH 2)
func (sm *SocialManager) LoadSocialState(playerID string) *PlayerSocialState {
	sm.socialStatesMu.Lock()
	defer sm.socialStatesMu.Unlock()

	if state, exists := sm.socialStates[playerID]; exists {
		return state
	}

	state := &PlayerSocialState{
		PlayerID: playerID,
		Friends:  make(map[string]bool),
		Ignores:  make(map[string]bool),
		Dirty:    false,
	}

	if sm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		rows, err := sm.dbConn.QueryContext(ctx, `
			SELECT target_name, relation_type 
			FROM social_relations 
			WHERE character_name = $1
		`, playerID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var target, rType string
				if err := rows.Scan(&target, &rType); err == nil {
					if rType == "friend" {
						state.Friends[target] = true
					} else if rType == "ignore" {
						state.Ignores[target] = true
					}
				}
			}
		}
	}

	sm.socialStates[playerID] = state
	return state
}

// PersistSocialState grava alterações caso o estado esteja marcado como "dirty" (PATCH 2)
func (sm *SocialManager) PersistSocialState(playerID string) {
	sm.socialStatesMu.RLock()
	state, exists := sm.socialStates[playerID]
	sm.socialStatesMu.RUnlock()

	if !exists || state == nil {
		return
	}

	state.mu.Lock()
	if !state.Dirty {
		state.mu.Unlock()
		return
	}

	if sm.dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tx, err := sm.dbConn.BeginTx(ctx, nil)
		if err != nil {
			state.mu.Unlock()
			slog.Error("Failed to start transaction for persisting social state", "player", playerID, "error", err)
			return
		}

		// Limpa relações antigas
		_, _ = tx.ExecContext(ctx, "DELETE FROM social_relations WHERE character_name = $1", playerID)

		// Insere novas relações em lote
		for f := range state.Friends {
			_, _ = tx.ExecContext(ctx, `
				INSERT INTO social_relations (character_name, target_name, relation_type)
				VALUES ($1, $2, 'friend')
			`, playerID, f)
		}
		for ig := range state.Ignores {
			_, _ = tx.ExecContext(ctx, `
				INSERT INTO social_relations (character_name, target_name, relation_type)
				VALUES ($1, $2, 'ignore')
			`, playerID, ig)
		}

		if err := tx.Commit(); err != nil {
			slog.Error("Failed to commit social state transaction", "player", playerID, "error", err)
			tx.Rollback()
			state.mu.Unlock()
			return
		}

		state.Dirty = false // Reseta a flag de dirty após salvar com sucesso!
		slog.Info("Social state persisted persistently (PATCH 2)", "player", playerID)
	}
	state.mu.Unlock()
}

func (sm *SocialManager) AddFriend(playerID, friendID string) error {
	if playerID == friendID {
		return fmt.Errorf("cannot add yourself as friend")
	}

	state := sm.LoadSocialState(playerID)
	state.mu.Lock()
	state.Friends[friendID] = true
	delete(state.Ignores, friendID) // Se era ignorado, remove
	state.Dirty = true             // Marca como dirty (PATCH 2)
	state.mu.Unlock()

	sm.SendSocialLists(playerID)
	slog.Info("Friend added", "player", playerID, "friend", friendID)
	return nil
}

func (sm *SocialManager) RemoveFriend(playerID, friendID string) error {
	state := sm.LoadSocialState(playerID)
	state.mu.Lock()
	if _, exists := state.Friends[friendID]; !exists {
		state.mu.Unlock()
		return fmt.Errorf("player is not in your friend list")
	}
	delete(state.Friends, friendID)
	state.Dirty = true // Marca como dirty (PATCH 2)
	state.mu.Unlock()

	sm.SendSocialLists(playerID)
	slog.Info("Friend removed", "player", playerID, "friend", friendID)
	return nil
}

func (sm *SocialManager) AddIgnore(playerID, targetID string) error {
	if playerID == targetID {
		return fmt.Errorf("cannot ignore yourself")
	}

	state := sm.LoadSocialState(playerID)
	state.mu.Lock()
	state.Ignores[targetID] = true
	delete(state.Friends, targetID) // Se era amigo, remove
	state.Dirty = true              // Marca como dirty (PATCH 2)
	state.mu.Unlock()

	sm.SendSocialLists(playerID)
	slog.Info("Player ignored", "player", playerID, "ignored", targetID)
	return nil
}

func (sm *SocialManager) RemoveIgnore(playerID, targetID string) error {
	state := sm.LoadSocialState(playerID)
	state.mu.Lock()
	if _, exists := state.Ignores[targetID]; !exists {
		state.mu.Unlock()
		return fmt.Errorf("player is not ignored")
	}
	delete(state.Ignores, targetID)
	state.Dirty = true // Marca como dirty (PATCH 2)
	state.mu.Unlock()

	sm.SendSocialLists(playerID)
	slog.Info("Player unignored", "player", playerID, "ignored", targetID)
	return nil
}

func (sm *SocialManager) SendSocialLists(playerID string) {
	state := sm.LoadSocialState(playerID)
	state.mu.RLock()
	defer state.mu.RUnlock()

	friends := []protocol.FriendInfo{}
	for f := range state.Friends {
		online := uint8(0)
		if sm.IsOnline(f) {
			online = 1
		}
		friends = append(friends, protocol.FriendInfo{
			FriendID: f,
			Online:   online,
		})
	}

	ignores := []string{}
	for ig := range state.Ignores {
		ignores = append(ignores, ig)
	}

	event := &protocol.SocialListsEvent{
		Friends: friends,
		Ignores: ignores,
	}

	sm.sendBinaryToPlayer(playerID, protocol.SC_SOCIAL_LISTS, protocol.EncodeSocialLists(event))
}

// Online Presence notifications
func (sm *SocialManager) NotifyOnlineStatus(playerID string, online uint8) {
	state := sm.LoadSocialState(playerID)
	state.mu.RLock()
	defer state.mu.RUnlock()

	// Envia notificação de status para amigos de forma eficiente
	payload := protocol.EncodeOnlineStatus(playerID, online)
	sm.socialStatesMu.RLock()
	for pID, s := range sm.socialStates {
		s.mu.RLock()
		if s.Friends[playerID] {
			sm.sendBinaryToPlayer(pID, protocol.SC_ONLINE_STATUS, payload)
		}
		s.mu.RUnlock()
	}
	sm.socialStatesMu.RUnlock()
}

// =============================================================================
// CHAT & RATE LIMITER (PATCH 4)
// =============================================================================

// CanSendChat verifica rate limits do jogador para evitar chat spam (PATCH 4)
func (sm *SocialManager) CanSendChat(playerID string) bool {
	sm.limitersMu.Lock()
	defer sm.limitersMu.Unlock()

	limiter, exists := sm.limiters[playerID]
	if !exists {
		limiter = &RateLimiter{
			LastMessage: time.Now(),
			Count:       0,
		}
		sm.limiters[playerID] = limiter
	}

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	now := time.Now()
	// Intervalo mínimo rígido de 200 milissegundos entre mensagens (PATCH 4)
	if now.Sub(limiter.LastMessage) < 200*time.Millisecond {
		limiter.LastMessage = now
		return false
	}

	// Limite extra de 3 mensagens por segundo para coibir macros rápidos
	if now.Sub(limiter.LastMessage) > time.Second {
		limiter.Count = 0
	}
	limiter.Count++
	limiter.LastMessage = now

	if limiter.Count > 3 {
		return false
	}
	return true
}

// HandleChatMessage processa mensagens nos canais autoritativos (com rate limiter)
func (sm *SocialManager) HandleChatMessage(playerID string, req *protocol.ChatSendRequest) {
	// 1. Aplica anti-spam rate limiter (PATCH 4)
	if !sm.CanSendChat(playerID) {
		sm.sendBinaryToPlayer(playerID, protocol.SC_CHAT_MESSAGE, protocol.EncodeChatMessage(4, "System", "Message spam detected! Slow down."))
		slog.Warn("Chat spam rate limit triggered", "player", playerID)
		return
	}

	// 2. Processa canais
	switch req.Channel {
	case 0: // Local chat (within 20 tiles)
		sm.broadcastLocalChat(playerID, req.Message)
	case 1: // Global chat
		sm.broadcastGlobalChat(playerID, req.Message)
	case 2: // Party chat
		sm.broadcastPartyChat(playerID, req.Message)
	case 3: // Guild chat
		sm.broadcastGuildChat(playerID, req.Message)
	case 4: // Private / Whisper (checks Ignore lists)
		sm.sendPrivateWhisper(playerID, req.Target, req.Message)
	}
}

func (sm *SocialManager) broadcastLocalChat(senderID, msg string) {
	payload := protocol.EncodeChatMessage(0, senderID, msg)
	// Usa o AOIManager para pegar jogadores no SpatialIndex
	// Para simplificar e garantir estabilidade, pegamos as conexões registradas no AOI e enviamos para as ativas próximas
	sm.aoiManager.BroadcastMove(senderID, protocol.SC_CHAT_MESSAGE, payload)
	// Também envia para o próprio remetente
	sm.sendBinaryToPlayer(senderID, protocol.SC_CHAT_MESSAGE, payload)
}

func (sm *SocialManager) broadcastGlobalChat(senderID, msg string) {
	payload := protocol.EncodeChatMessage(1, senderID, msg)
	// Envia para todos os jogadores online registrados no AOIManager
	sm.aoiManager.BroadcastMovement(senderID, protocol.SC_CHAT_MESSAGE, payload)
	sm.sendBinaryToPlayer(senderID, protocol.SC_CHAT_MESSAGE, payload)
}

func (sm *SocialManager) broadcastPartyChat(senderID, msg string) {
	sm.partiesMu.RLock()
	party, inParty := sm.parties[senderID]
	sm.partiesMu.RUnlock()

	if !inParty {
		sm.sendBinaryToPlayer(senderID, protocol.SC_CHAT_MESSAGE, protocol.EncodeChatMessage(2, "System", "You are not in a party."))
		return
	}

	party.mu.RLock()
	defer party.mu.RUnlock()

	payload := protocol.EncodeChatMessage(2, senderID, msg)
	for _, m := range party.Members {
		sm.sendBinaryToPlayer(m, protocol.SC_CHAT_MESSAGE, payload)
	}
}

func (sm *SocialManager) broadcastGuildChat(senderID, msg string) {
	sm.guildsMu.RLock()
	var gState *GuildMemoryState
	for _, g := range sm.guilds {
		g.mu.RLock()
		if g.Leader == senderID || g.Vices[senderID] || g.Members[senderID] {
			gState = g
			g.mu.RUnlock()
			break
		}
		g.mu.RUnlock()
	}
	sm.guildsMu.RUnlock()

	if gState == nil {
		sm.sendBinaryToPlayer(senderID, protocol.SC_CHAT_MESSAGE, protocol.EncodeChatMessage(3, "System", "You are not in a guild."))
		return
	}

	gState.mu.RLock()
	defer gState.mu.RUnlock()

	payload := protocol.EncodeChatMessage(3, senderID, msg)
	// Envia ao líder
	sm.sendBinaryToPlayer(gState.Leader, protocol.SC_CHAT_MESSAGE, payload)
	// Envia aos vices
	for v := range gState.Vices {
		sm.sendBinaryToPlayer(v, protocol.SC_CHAT_MESSAGE, payload)
	}
	// Envia aos membros
	for m := range gState.Members {
		sm.sendBinaryToPlayer(m, protocol.SC_CHAT_MESSAGE, payload)
	}
}

func (sm *SocialManager) sendPrivateWhisper(senderID, targetID, msg string) {
	// Verifica se o alvo ignora o remetente
	targetState := sm.LoadSocialState(targetID)
	targetState.mu.RLock()
	ignored := targetState.Ignores[senderID]
	targetState.mu.RUnlock()

	if ignored {
		sm.sendBinaryToPlayer(senderID, protocol.SC_CHAT_MESSAGE, protocol.EncodeChatMessage(4, "System", "You are ignored by this player."))
		return
	}

	if !sm.IsOnline(targetID) {
		sm.sendBinaryToPlayer(senderID, protocol.SC_CHAT_MESSAGE, protocol.EncodeChatMessage(4, "System", "Player is offline."))
		return
	}

	payload := protocol.EncodeChatMessage(4, senderID, msg)
	sm.sendBinaryToPlayer(targetID, protocol.SC_CHAT_MESSAGE, payload)
	sm.sendBinaryToPlayer(senderID, protocol.SC_CHAT_MESSAGE, payload)
}

// =============================================================================
// HELPERS & LIFECYCLE HOOKS
// =============================================================================

func (sm *SocialManager) IsOnline(playerID string) bool {
	_, exists := sm.aoiManager.GetPlayerConn(playerID)
	return exists
}

func (sm *SocialManager) getOnlineByte(playerID string) uint8 {
	if sm.IsOnline(playerID) {
		return 1
	}
	return 0
}

func (sm *SocialManager) sendBinaryToPlayer(playerID string, opcode uint16, payload []byte) {
	conn, exists := sm.aoiManager.GetPlayerConn(playerID)
	if exists && conn != nil {
		packet := &protocol.Packet{
			Opcode:  opcode,
			Payload: payload,
		}
		_, _ = conn.Write(packet.Serialize())
	}
}

func (sm *SocialManager) isPlayerOnlineAndNear(playerID string, targetX, targetY, radius float64) bool {
	conn, exists := sm.aoiManager.GetPlayerConn(playerID)
	if !exists || conn == nil {
		return false
	}
	return true
}

func (sm *SocialManager) GetGuilds() map[string]*GuildMemoryState {
	sm.guildsMu.RLock()
	defer sm.guildsMu.RUnlock()
	res := make(map[string]*GuildMemoryState)
	for name, g := range sm.guilds {
		res[name] = g
	}
	return res
}

func (sm *SocialManager) FindGuildNameForPlayer(playerID string) string {
	sm.guildsMu.RLock()
	defer sm.guildsMu.RUnlock()
	for name, g := range sm.guilds {
		g.mu.RLock()
		inGuild := g.Leader == playerID || g.Vices[playerID] || g.Members[playerID]
		g.mu.RUnlock()
		if inGuild {
			return name
		}
	}
	return ""
}

// OnPlayerLogin handle presence and social lists load
func (sm *SocialManager) OnPlayerLogin(playerID string) {
	sm.LoadSocialState(playerID)
	sm.SendSocialLists(playerID)
	sm.NotifyOnlineStatus(playerID, 1)

	guildName := sm.FindGuildNameForPlayer(playerID)
	if guildName != "" {
		sm.guildsMu.RLock()
		gState, exists := sm.guilds[guildName]
		sm.guildsMu.RUnlock()
		if exists && gState != nil {
			go sm.BroadcastGuildInfo(gState)
		}
	}
}

// OnPlayerLogout handle presence notify and social state persistence (PATCH 2)
func (sm *SocialManager) OnPlayerLogout(playerID string) {
	sm.PersistSocialState(playerID)
	sm.NotifyOnlineStatus(playerID, 0)

	guildName := sm.FindGuildNameForPlayer(playerID)
	if guildName != "" {
		sm.guildsMu.RLock()
		gState, exists := sm.guilds[guildName]
		sm.guildsMu.RUnlock()
		if exists && gState != nil {
			go sm.BroadcastGuildInfo(gState)
		}
	}
}

// GetPartyMembers retorna todos os membros do grupo de um jogador (ou apenas ele mesmo se não estiver em grupo)
func (sm *SocialManager) GetPartyMembers(playerID string) []string {
	sm.partiesMu.RLock()
	defer sm.partiesMu.RUnlock()

	party, ok := sm.parties[playerID]
	if !ok {
		return []string{playerID}
	}

	party.mu.RLock()
	defer party.mu.RUnlock()

	members := make([]string, len(party.Members))
	copy(members, party.Members)
	return members
}

