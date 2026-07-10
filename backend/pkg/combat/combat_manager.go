package combat

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/pkg/movement"
	"log/slog"
)

// WeaponConfig mapeia a velocidade, alcance e multiplicador de escalonamento de cada arma
type WeaponConfig struct {
	Name    string
	Speed   time.Duration
	Range   float64
	Scaling float64
}

// WeaponConfigs armazena as configurações oficiais de cada tipo de arma
var WeaponConfigs = map[string]WeaponConfig{
	"dagger":   {Name: "Adaga", Speed: 650 * time.Millisecond, Range: 1.0, Scaling: 0.9},
	"sword":    {Name: "Espada", Speed: 1000 * time.Millisecond, Range: 1.0, Scaling: 1.1},
	"axe":      {Name: "Machado", Speed: 1350 * time.Millisecond, Range: 1.0, Scaling: 1.3},
	"hammer":   {Name: "Martelo", Speed: 1800 * time.Millisecond, Range: 1.0, Scaling: 1.6},
	"spear":    {Name: "Lança", Speed: 1200 * time.Millisecond, Range: 2.0, Scaling: 1.2},
	"staff":    {Name: "Cajado", Speed: 1500 * time.Millisecond, Range: 4.0, Scaling: 1.0},
	"bow":      {Name: "Arco", Speed: 1250 * time.Millisecond, Range: 7.0, Scaling: 1.0},
	"crossbow": {Name: "Besta", Speed: 1400 * time.Millisecond, Range: 9.0, Scaling: 1.4},
}

// Projectile representa um projétil de ataque à distância ou magia em trânsito
type Projectile struct {
	ID         string
	AttackerID string
	TargetID   string
	LaunchTime time.Time
	HitTime    time.Time
	StartX     float64
	StartY     float64
	TargetX    float64
	TargetY    float64
	SkillID    uint32
	SkillName  string
	IsArea     bool
	AreaRadius float64
	Scaling    float64
}

// CombatEventHandler define callbacks para notificar o gateway síncrono sobre eventos assíncronos
type CombatEventHandler struct {
	OnDamage     func(attackerID, targetID string, damage float64, isCrit, isHit bool, skillName string)
	OnTargetDead func(targetID string)
}

// CombatManager coordena todo o ecossistema de combate do servidor
type CombatManager struct {
	mu             sync.RWMutex
	entities       map[string]*EntityStats
	entityPos      map[string]struct{ X, Y float64 }
	skills         map[uint32]Skill
	aggroTables    map[string]*AggroTable          // NPC ID -> AggroTable
	nextAttackTime map[string]time.Time            // ID -> Próximo momento de ataque permitido
	skillCooldowns map[string]map[uint32]time.Time // ID -> SkillID -> Próximo momento de uso
	scheduler      *CombatScheduler
	hpCounter      float64
	manaCounter    float64
	chunkManager   *movement.ChunkManager
	eventHandler   CombatEventHandler
	PvPValidator   func(attackerID, targetID string) error // PvP Validation Callback (PATCH PvP)
	lastAttacker   map[string]string                       // targetID -> lastPlayerAttackerID (PATCH PvP)
	OnPvPDealt     func(attackerID, targetID string)       // On PvP Damage callback (PATCH 6)
	OnPvPSpellCast func(attackerID string)                 // On PvP offensive skill cast (PATCH 6)
}

// NewCombatManager instancia um novo CombatManager completo
func NewCombatManager(chunkManager *movement.ChunkManager, skillOverrides ...map[uint32]Skill) *CombatManager {
	cm := &CombatManager{
		entities:       make(map[string]*EntityStats),
		entityPos:      make(map[string]struct{ X, Y float64 }),
		skills:         PredefinedSkills, // Começa com o hardcode
		aggroTables:    make(map[string]*AggroTable),
		nextAttackTime: make(map[string]time.Time),
		skillCooldowns: make(map[string]map[uint32]time.Time),
		scheduler:      NewCombatScheduler(500 * time.Millisecond), // 500ms Tick Rate
		hpCounter:      0,
		manaCounter:    0,
		chunkManager:   chunkManager,
		lastAttacker:   make(map[string]string),
	}

	// Se um mapa de skills foi fornecido, ele sobrescreve ou adiciona ao padrão.
	if len(skillOverrides) > 0 && skillOverrides[0] != nil && len(skillOverrides[0]) > 0 {
		for id, skill := range skillOverrides[0] {
			cm.skills[id] = skill
		}
	}

	// Registra tarefas periódicas no Scheduler
	cm.scheduler.RegisterTask("HpRegenAndAggroDecay", cm.tickHpRegenAndAggro)
	cm.scheduler.Start()

	return cm
}

// SetEventHandler registra callbacks de notificação de combate
func (cm *CombatManager) SetEventHandler(handler CombatEventHandler) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.eventHandler = handler
}

// HasLineOfSight verifica se há linha de visão desobstruída entre duas coordenadas (Bresenham's Grid Raycast)
func (cm *CombatManager) HasLineOfSight(x1, y1, x2, y2 float64) bool {
	if cm.chunkManager == nil {
		return true // Fallback se chunkManager não estiver configurado
	}

	x := int(math.Floor(x1))
	y := int(math.Floor(y1))
	targetX := int(math.Floor(x2))
	targetY := int(math.Floor(y2))

	dx := int(math.Abs(float64(targetX - x)))
	dy := int(math.Abs(float64(targetY - y)))

	var sx, sy int
	if x < targetX {
		sx = 1
	} else {
		sx = -1
	}
	if y < targetY {
		sy = 1
	} else {
		sy = -1
	}

	err := dx - dy

	for {
		if x == targetX && y == targetY {
			break
		}

		// Ignora verificação no tile inicial (do atacante)
		if x != int(math.Floor(x1)) || y != int(math.Floor(y1)) {
			if cm.chunkManager.IsBlocked(x, y) {
				return false
			}
		}

		err2 := 2 * err
		if err2 > -dy {
			err -= dy
			x += sx
		}
		if err2 < dx {
			err += dx
			y += sy
		}
	}

	return true
}

// LaunchProjectile agenda a resolução de um projétil após o tempo de viagem
func (cm *CombatManager) LaunchProjectile(p *Projectile, travelTime time.Duration) {
	slog.Info("Projectile launched", "id", p.ID, "attacker", p.AttackerID, "target", p.TargetID, "travelTime", travelTime)
	time.AfterFunc(travelTime, func() {
		cm.ResolveProjectile(p)
	})
}

// ResolveProjectile processa a colisão espacial e o dano do projétil ao atingir o destino
func (cm *CombatManager) ResolveProjectile(p *Projectile) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	attacker, existsAtt := cm.entities[p.AttackerID]
	if !existsAtt {
		return
	}

	if p.IsArea {
		// Habilidade em área (ex: Fireball, Arrow Rain)
		// Aplica dano a todas as entidades na área de impacto
		var hitEntities []*EntityStats
		for id, ent := range cm.entities {
			if ent.Health <= 0 {
				continue
			}
			entPos := cm.entityPos[id]
			dist := math.Hypot(entPos.X-p.TargetX, entPos.Y-p.TargetY)
			if dist <= p.AreaRadius {
				hitEntities = append(hitEntities, ent)
			}
		}

		for _, ent := range hitEntities {
			// Friendly fire check
			if attacker.IsPlayer && ent.IsPlayer && attacker.Faction == ent.Faction {
				continue
			}

			isHit := RollHit(attacker, ent)
			if !isHit {
				if cm.eventHandler.OnDamage != nil {
					cm.eventHandler.OnDamage(p.AttackerID, ent.ID, 0, false, false, p.SkillName)
				}
				continue
			}

			isPvP := attacker.IsPlayer && ent.IsPlayer
			damage, isCrit := CalculateDamage(attacker, ent, 1.0, p.Scaling, isPvP)

			ent.Health -= damage
			if ent.Health < 0 {
				ent.Health = 0
			}
			if damage > 0 && attacker.IsPlayer {
				cm.lastAttacker[ent.ID] = p.AttackerID
				if ent.IsPlayer && cm.OnPvPDealt != nil {
					cm.OnPvPDealt(p.AttackerID, ent.ID)
				}
			}

			if !ent.IsPlayer {
				if atTable, ok := cm.aggroTables[ent.ID]; ok {
					atTable.AddThreat(p.AttackerID, damage)
				}
			}

			if cm.eventHandler.OnDamage != nil {
				cm.eventHandler.OnDamage(p.AttackerID, ent.ID, damage, isCrit, true, p.SkillName)
			}

			if ent.Health <= 0 {
				go cm.handleEntityDeath(ent.ID)
				if cm.eventHandler.OnTargetDead != nil {
					cm.eventHandler.OnTargetDead(ent.ID)
				}
			}
		}
	} else {
		// Alvo único (ex: Arco, Besta, Cajado)
		target, existsTar := cm.entities[p.TargetID]
		if !existsTar || target.Health <= 0 {
			slog.Info("Projectile missed: target dead or logged out", "id", p.ID)
			return
		}

		// Validação de colisão espacial: o alvo não pode ter se movido mais do que 3 tiles do ponto esperado
		tarPos := cm.entityPos[p.TargetID]
		dist := math.Hypot(tarPos.X-p.TargetX, tarPos.Y-p.TargetY)
		if dist > 3.0 {
			slog.Info("Projectile missed: target evaded out of collision box", "id", p.ID, "dist", dist)
			if cm.eventHandler.OnDamage != nil {
				cm.eventHandler.OnDamage(p.AttackerID, p.TargetID, 0, false, false, p.SkillName)
			}
			return
		}

		isHit := RollHit(attacker, target)
		if !isHit {
			if cm.eventHandler.OnDamage != nil {
				cm.eventHandler.OnDamage(p.AttackerID, p.TargetID, 0, false, false, p.SkillName)
			}
			return
		}

		isPvP := attacker.IsPlayer && target.IsPlayer
		damage, isCrit := CalculateDamage(attacker, target, p.Scaling, 1.0, isPvP)

		target.Health -= damage
		if target.Health < 0 {
			target.Health = 0
		}
		if damage > 0 && attacker.IsPlayer {
			cm.lastAttacker[p.TargetID] = p.AttackerID
			if target.IsPlayer && cm.OnPvPDealt != nil {
				cm.OnPvPDealt(p.AttackerID, p.TargetID)
			}
		}

		if !target.IsPlayer {
			if atTable, ok := cm.aggroTables[p.TargetID]; ok {
				atTable.AddThreat(p.AttackerID, damage)
			}
		}

		if cm.eventHandler.OnDamage != nil {
			cm.eventHandler.OnDamage(p.AttackerID, p.TargetID, damage, isCrit, true, p.SkillName)
		}

		if target.Health <= 0 {
			go cm.handleEntityDeath(p.TargetID)
			if cm.eventHandler.OnTargetDead != nil {
				cm.eventHandler.OnTargetDead(p.TargetID)
			}
		}
	}
}

// RegisterEntity registra um jogador ou NPC no sistema de combate
func (cm *CombatManager) RegisterEntity(entity *EntityStats, x, y float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.entities[entity.ID] = entity
	cm.entityPos[entity.ID] = struct{ X, Y float64 }{X: x, Y: y}
	cm.nextAttackTime[entity.ID] = time.Now()
	cm.skillCooldowns[entity.ID] = make(map[uint32]time.Time)

	if !entity.IsPlayer {
		cm.aggroTables[entity.ID] = NewAggroTable()
	}
}

// UpdateEntityPosition atualiza as coordenadas para validação autoritativa de distância
func (cm *CombatManager) UpdateEntityPosition(id string, x, y float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, exists := cm.entityPos[id]; exists {
		cm.entityPos[id] = struct{ X, Y float64 }{X: x, Y: y}
	}
}

// DeregisterEntity remove a entidade do sistema
func (cm *CombatManager) DeregisterEntity(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.entities, id)
	delete(cm.entityPos, id)
	delete(cm.aggroTables, id)
	delete(cm.nextAttackTime, id)
	delete(cm.skillCooldowns, id)
}

// GetEntityStats retorna as estatísticas de uma entidade
func (cm *CombatManager) GetEntityStats(id string) (*EntityStats, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	stats, exists := cm.entities[id]
	return stats, exists
}

// GetEntityStatsCopy retorna uma cópia das estatísticas de uma entidade de forma thread-safe
func (cm *CombatManager) GetEntityStatsCopy(id string) (EntityStats, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	stats, exists := cm.entities[id]
	if !exists || stats == nil {
		return EntityStats{}, false
	}
	return *stats, true
}

// GetEntityPosition retorna as coordenadas X e Y de uma entidade de forma thread-safe
func (cm *CombatManager) GetEntityPosition(id string) (float64, float64, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	pos, exists := cm.entityPos[id]
	return pos.X, pos.Y, exists
}

// GetLastAttacker retorna o ID do último jogador que atacou a entidade (PATCH PvP)
func (cm *CombatManager) GetLastAttacker(id string) (string, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	atk, exists := cm.lastAttacker[id]
	return atk, exists
}

// ClearLastAttacker limpa o histórico de último atacante para uma entidade (PATCH PvP)
func (cm *CombatManager) ClearLastAttacker(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.lastAttacker, id)
}

// ProcessAttackRequest valida e processa um ataque básico (auto-attack) autoritativo
func (cm *CombatManager) ProcessAttackRequest(attackerID, targetID, weaponType string) (float64, bool, bool, error) {
	cm.mu.Lock()
	attacker, existsAtt := cm.entities[attackerID]
	target, existsTar := cm.entities[targetID]
	attPos, hasAttPos := cm.entityPos[attackerID]
	tarPos, hasTarPos := cm.entityPos[targetID]
	cm.mu.Unlock()

	if !existsAtt || !existsTar || !hasAttPos || !hasTarPos {
		return 0, false, false, errors.New("atacante ou alvo não encontrado no mundo")
	}

	if attacker.Health <= 0 {
		return 0, false, false, errors.New("atacante está morto")
	}
	if target.Health <= 0 {
		return 0, false, false, errors.New("alvo já está morto")
	}

	// 1. Cooldown / Anti-Spam Check
	cm.mu.Lock()
	nextAllowed, existsCooldown := cm.nextAttackTime[attackerID]
	cm.mu.Unlock()

	now := time.Now()
	if existsCooldown && now.Before(nextAllowed) {
		timeLeft := nextAllowed.Sub(now)
		return 0, false, false, fmt.Errorf("ataque em recarga. Espere %.2fs", timeLeft.Seconds())
	}

	// 2. Weapon Config & Range Check
	wConfig, existsWeapon := WeaponConfigs[weaponType]
	if !existsWeapon {
		wConfig = WeaponConfigs["sword"] // Fallback padrão
	}

	dist := math.Max(math.Abs(tarPos.X-attPos.X), math.Abs(tarPos.Y-attPos.Y))
	if dist > wConfig.Range {
		return 0, false, false, fmt.Errorf("alvo fora de alcance para %s. Distância: %.2fm, Alcance: %.2fm", wConfig.Name, dist, wConfig.Range)
	}

	// 3. Line of Sight (LOS) check para ataques à distância (> 2.0m)
	if wConfig.Range > 2.0 {
		if !cm.HasLineOfSight(attPos.X, attPos.Y, tarPos.X, tarPos.Y) {
			return 0, false, false, errors.New("linha de visão bloqueada por um obstáculo")
		}
	}

	// 4. Registrar novo momento de ataque permitido com base no Attack Speed da arma
	cm.mu.Lock()
	cm.nextAttackTime[attackerID] = now.Add(wConfig.Speed)
	cm.mu.Unlock()

	// Friendly Fire check
	if attacker.IsPlayer && target.IsPlayer {
		if cm.PvPValidator != nil {
			if err := cm.PvPValidator(attackerID, targetID); err != nil {
				return 0, false, false, err
			}
		}
		if attacker.Faction == target.Faction {
			return 0, false, false, errors.New("fogo amigo desabilitado")
		}
	}

	// Update combat timers
	cm.mu.Lock()
	attacker.LastCombatTime = now
	target.LastCombatTime = now
	cm.mu.Unlock()

	// 5. Determinar se é Projétil (Ataque à distância com arco, besta, cajado)
	isProjectile := weaponType == "bow" || weaponType == "crossbow" || weaponType == "staff"
	if isProjectile {
		speed := 15.0 // Bow
		if weaponType == "crossbow" {
			speed = 12.0
		} else if weaponType == "staff" {
			speed = 10.0
		}

		travelTime := time.Duration((dist / speed) * float64(time.Second))
		if travelTime < 50*time.Millisecond {
			travelTime = 50 * time.Millisecond
		}

		p := &Projectile{
			ID:         fmt.Sprintf("proj_att_%d", time.Now().UnixNano()),
			AttackerID: attackerID,
			TargetID:   targetID,
			LaunchTime: now,
			HitTime:    now.Add(travelTime),
			StartX:     attPos.X,
			StartY:     attPos.Y,
			TargetX:    tarPos.X,
			TargetY:    tarPos.Y,
			SkillName:  wConfig.Name,
			IsArea:     false,
			Scaling:    wConfig.Scaling,
		}

		cm.LaunchProjectile(p, travelTime)
		return 0, false, true, nil
	}

	// 6. Roll Hit / Accuracy vs Evasion (Para melee instantâneo)
	if !RollHit(attacker, target) {
		return 0, false, false, nil
	}

	// 7. Calculate Damage
	isPvP := attacker.IsPlayer && target.IsPlayer
	damage, isCrit := CalculateDamage(attacker, target, wConfig.Scaling, 1.0, isPvP)

	// 8. Apply Health reduction
	cm.mu.Lock()
	target.Health -= damage
	if target.Health < 0 {
		target.Health = 0
	}
	isDead := target.Health <= 0
	if damage > 0 && attacker.IsPlayer {
		cm.lastAttacker[targetID] = attackerID
		if target.IsPlayer && cm.OnPvPDealt != nil {
			cm.OnPvPDealt(attackerID, targetID)
		}
	}
	cm.mu.Unlock()

	// 9. PvE Aggro
	if attacker.IsPlayer && !target.IsPlayer {
		cm.mu.Lock()
		atTable, hasAt := cm.aggroTables[target.ID]
		cm.mu.Unlock()
		if hasAt {
			atTable.AddThreat(attackerID, damage)
		}
	}

	if isDead {
		cm.handleEntityDeath(target.ID)
	}

	return damage, isCrit, false, nil
}

// ProcessCastSkillRequest valida e processa a conjuração de uma habilidade ativa
func (cm *CombatManager) ProcessCastSkillRequest(attackerID string, skillID uint32, targetID string, targetX, targetY float64) (*SkillCastResult, error) {
	cm.mu.Lock()
	attacker, existsAtt := cm.entities[attackerID]
	attPos, hasAttPos := cm.entityPos[attackerID]
	cm.mu.Unlock()

	if !existsAtt || !hasAttPos {
		return nil, errors.New("atacante não encontrado no mundo")
	}

	if attacker.Health <= 0 {
		return nil, errors.New("atacante está morto")
	}

	skill, existsSkill := cm.skills[skillID]
	if !existsSkill {
		return nil, fmt.Errorf("habilidade ID %d não existe", skillID)
	}

	// 1. Validar Cooldown da Habilidade (Anti-spam)
	cm.mu.Lock()
	nextAllowed, existsCD := cm.skillCooldowns[attackerID][skillID]
	cm.mu.Unlock()

	now := time.Now()
	if existsCD && now.Before(nextAllowed) {
		return nil, fmt.Errorf("habilidade %s está em recarga", skill.Name)
	}

	// 2. Se for alvo único (target lock), encontrar o alvo e validar distância
	var target *EntityStats
	if !skill.IsArea {
		cm.mu.Lock()
		targetStats, hasTarget := cm.entities[targetID]
		tarPos, hasTarPos := cm.entityPos[targetID]
		cm.mu.Unlock()

		if !hasTarget || !hasTarPos {
			return nil, errors.New("alvo único da habilidade não encontrado")
		}
		if targetStats.Health <= 0 {
			return nil, errors.New("alvo já está morto")
		}
		target = targetStats
		targetX, targetY = tarPos.X, tarPos.Y

		if attacker.IsPlayer && target.IsPlayer {
			if cm.PvPValidator != nil {
				if err := cm.PvPValidator(attackerID, targetID); err != nil {
					return nil, err
				}
			}
			if cm.OnPvPSpellCast != nil {
				cm.OnPvPSpellCast(attackerID)
			}
		}
	}

	// 3. Line of Sight (LOS) check para habilidades com alcance > 2.0m
	if skill.Range > 2.0 {
		if !cm.HasLineOfSight(attPos.X, attPos.Y, targetX, targetY) {
			return nil, errors.New("linha de visão bloqueada por um obstáculo")
		}
	}

	// 4. Coleta entidades elegíveis próximas (AOI) no caso de habilidades em área instantâneas
	var nearbyCandidates []*EntityStats
	if skill.IsArea && skillID != 2 { // Fireball (ID 2) é projétil, resolve na colisão
		cm.mu.RLock()
		for id, ent := range cm.entities {
			if id == attackerID || ent.Health <= 0 {
				continue
			}
			entPos := cm.entityPos[id]
			dist := math.Hypot(entPos.X-targetX, entPos.Y-targetY)
			if dist <= skill.AreaRadius {
				nearbyCandidates = append(nearbyCandidates, ent)
			}
		}
		cm.mu.RUnlock()
	}

	// 5. Determinar se a habilidade usa projétil (ex: Fireball ID 2, Arrow Rain ID 4)
	isProjectileSkill := skillID == 2 || skillID == 4
	if isProjectileSkill {
		dist := math.Hypot(targetX-attPos.X, targetY-attPos.Y)
		if dist > skill.Range {
			return nil, fmt.Errorf("alvo fora de alcance para %s", skill.Name)
		}

		// Consumir recarga e atualizar timers de combate
		cm.mu.Lock()
		cm.skillCooldowns[attackerID][skillID] = now.Add(skill.Cooldown)
		attacker.LastCombatTime = now
		cm.mu.Unlock()

		speed := 14.0 // Fireball speed
		if skillID == 4 {
			speed = 10.0 // Arrow Rain speed (sky-drop)
		}

		travelTime := time.Duration((dist / speed) * float64(time.Second))
		if travelTime < 50*time.Millisecond {
			travelTime = 50 * time.Millisecond
		}

		p := &Projectile{
			ID:         fmt.Sprintf("proj_skill_%d_%d", skillID, time.Now().UnixNano()),
			AttackerID: attackerID,
			TargetID:   targetID,
			LaunchTime: now,
			HitTime:    now.Add(travelTime),
			StartX:     attPos.X,
			StartY:     attPos.Y,
			TargetX:    targetX,
			TargetY:    targetY,
			SkillID:    skillID,
			SkillName:  skill.Name,
			IsArea:     skill.IsArea,
			AreaRadius: skill.AreaRadius,
			Scaling:    skill.SkillScale,
		}

		cm.LaunchProjectile(p, travelTime)

		// Retorna resultado indicando que foi disparado como projétil assíncrono
		return &SkillCastResult{
			Skill:        skill,
			AttackerID:   attackerID,
			Success:      true,
			IsProjectile: true,
		}, nil
	}

	// 6. Resolver habilidade instantânea síncrona
	result := ResolveSkill(skill, attacker, attPos.X, attPos.Y, target, targetX, targetY, nearbyCandidates)

	if !result.Success {
		return result, errors.New(result.ErrorMessage)
	}

	// Consumir recarga e aplicar danos reais no backend
	cm.mu.Lock()
	cm.skillCooldowns[attackerID][skillID] = now.Add(skill.Cooldown)
	attacker.LastCombatTime = now

	for _, dmg := range result.TargetsHit {
		tEnt, ok := cm.entities[dmg.TargetID]
		if ok {
			tEnt.LastCombatTime = now
			if dmg.IsHit && dmg.Damage > 0 {
				tEnt.Health -= dmg.Damage
				if tEnt.Health < 0 {
					tEnt.Health = 0
				}
				if dmg.Damage > 0 && attacker.IsPlayer {
					cm.lastAttacker[dmg.TargetID] = attackerID
					if tEnt.IsPlayer && cm.OnPvPDealt != nil {
						cm.OnPvPDealt(attackerID, dmg.TargetID)
					}
				}
				// Aggro PvE
				if attacker.IsPlayer && !tEnt.IsPlayer {
					atTable, hasAt := cm.aggroTables[tEnt.ID]
					if hasAt {
						atTable.AddThreat(attackerID, dmg.Damage)
					}
				}

				if tEnt.Health <= 0 {
					go cm.handleEntityDeath(tEnt.ID)
				}
			}
		}
	}
	cm.mu.Unlock()

	return result, nil
}

// handleEntityDeath processa a morte e limpa tabelas de aggro
func (cm *CombatManager) handleEntityDeath(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Limpa da aggro table de outros NPCs
	for npcID, atTable := range cm.aggroTables {
		atTable.RemovePlayer(id)
		if npcID == id {
			atTable.ClearThreat()
		}
	}
}

// tickHpRegenAndAggro gerencia a cura passiva lenta e a decadência de ameaça a cada 500ms
func (cm *CombatManager) tickHpRegenAndAggro() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	const dt = 0.5 // Tick rate de 500ms é exatamente 0.5s

	// Decadência de ameaça baseada no tempo real (2% por segundo)
	for _, atTable := range cm.aggroTables {
		atTable.DecayThreats(dt)
	}

	// Executa a regeneração baseada em tempo (1% de HP e 3% de Mana a cada 5 segundos fora de combate)
	PerformRegeneration(cm.entities, dt, &cm.hpCounter, &cm.manaCounter)
}

// Close finaliza o CombatManager e o Scheduler
func (cm *CombatManager) Close() {
	cm.scheduler.Stop()
}

// GetAggroTable retorna a tabela de ameaça de forma thread-safe
func (cm *CombatManager) GetAggroTable(id string) (*AggroTable, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	table, ok := cm.aggroTables[id]
	return table, ok
}

// ReviveEntity restores an existing entity to MaxHealth in-place.
// This is intentionally generic server-authoritative combat state handling;
// callers decide whether a revive is allowed for a given gameplay/debug flow.
func (cm *CombatManager) ReviveEntity(id string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	stats, exists := cm.entities[id]
	if !exists || stats == nil || stats.MaxHealth <= 0 {
		return false
	}

	stats.Health = stats.MaxHealth
	stats.LastCombatTime = time.Time{}

	if cm.nextAttackTime != nil {
		cm.nextAttackTime[id] = time.Now()
	}
	if cm.skillCooldowns != nil {
		cm.skillCooldowns[id] = make(map[uint32]time.Time)
	}
	if cm.aggroTables != nil {
		delete(cm.aggroTables, id)
	}
	if cm.lastAttacker != nil {
		delete(cm.lastAttacker, id)
	}

	return true
}
