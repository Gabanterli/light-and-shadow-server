package combat

import (
	"context"
	"sync"
	"time"
)

// CombatScheduler gerencia tarefas periódicas de combate em segundo plano, como regeneração de HP e ticks de IA
type CombatScheduler struct {
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	tasks     map[string]func()
	interval  time.Duration
	running   bool
}

// NewCombatScheduler cria uma nova instância do agendador de combate
func NewCombatScheduler(interval time.Duration) *CombatScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &CombatScheduler{
		ctx:      ctx,
		cancel:   cancel,
		tasks:    make(map[string]func()),
		interval: interval,
	}
}

// RegisterTask adiciona uma tarefa repetitiva ao agendador
func (cs *CombatScheduler) RegisterTask(name string, task func()) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.tasks[name] = task
}

// UnregisterTask remove uma tarefa
func (cs *CombatScheduler) UnregisterTask(name string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.tasks, name)
}

// Start inicia o loop de ticks em background
func (cs *CombatScheduler) Start() {
	cs.mu.Lock()
	if cs.running {
		cs.mu.Unlock()
		return
	}
	cs.running = true
	cs.mu.Unlock()

	go cs.runLoop()
}

// Stop finaliza o agendador de combate
func (cs *CombatScheduler) Stop() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if !cs.running {
		return
	}
	cs.cancel()
	cs.running = false
}

func (cs *CombatScheduler) runLoop() {
	ticker := time.NewTicker(cs.interval)
	defer ticker.Stop()

	for {
		select {
		case <-cs.ctx.Done():
			return
		case <-ticker.C:
			cs.executeTasks()
		}
	}
}

func (cs *CombatScheduler) executeTasks() {
	cs.mu.Lock()
	// Copia as funções para evitar manter o lock durante a execução de tarefas potencialmente lentas
	tasksToRun := make([]func(), 0, len(cs.tasks))
	for _, task := range cs.tasks {
		tasksToRun = append(tasksToRun, task)
	}
	cs.mu.Unlock()

	for _, task := range tasksToRun {
		// Executa cada tarefa de maneira segura de panics para não derrubar o servidor
		safelyExecute(task)
	}
}

func safelyExecute(task func()) {
	defer func() {
		if r := recover(); r != nil {
			// Captura falhas para manter a resiliência autoritativa do servidor
		}
	}()
	task()
}

// PerformRegeneration aplica as regras oficiais de regeneração baseada em tempo para HP e Mana das entidades registradas (Sprint 2 Task 5 Patch 2)
func PerformRegeneration(entities map[string]*EntityStats, elapsedSeconds float64, hpCounter, manaCounter *float64) {
	now := time.Now()
	
	*hpCounter += elapsedSeconds
	*manaCounter += elapsedSeconds

	triggerHP := false
	if *hpCounter >= 5.0 {
		triggerHP = true
		*hpCounter -= 5.0
	}

	triggerMana := false
	if *manaCounter >= 5.0 {
		triggerMana = true
		*manaCounter -= 5.0
	}

	if !triggerHP && !triggerMana {
		return
	}

	for _, ent := range entities {
		if ent.Health <= 0 {
			continue
		}

		// Verifica se está em combate (último combate há menos de 5 segundos)
		inCombat := !ent.LastCombatTime.IsZero() && now.Sub(ent.LastCombatTime) < 5*time.Second

		// 1. HP Regen: Fora de combate, 1% max HP a cada 5 segundos
		if triggerHP {
			if !inCombat {
				if ent.Health < ent.MaxHealth {
					ent.Health += ent.MaxHealth * 0.01
					if ent.Health > ent.MaxHealth {
						ent.Health = ent.MaxHealth
					}
				}
			}
		}

		// 2. Mana Regen: 3% max mana a cada 5 segundos
		if triggerMana {
			if ent.Mana < ent.MaxMana {
				ent.Mana += ent.MaxMana * 0.03
				if ent.Mana > ent.MaxMana {
					ent.Mana = ent.MaxMana
				}
			}
		}
	}
}
