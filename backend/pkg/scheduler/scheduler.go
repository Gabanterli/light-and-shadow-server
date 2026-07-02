package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type TickHandler func(deltaTime time.Duration, currentTick uint64)

type TickScheduler struct {
	hz           int
	interval     time.Duration
	handler      TickHandler
	stopChan     chan struct{}
	currentTick  uint64
}

func NewTickScheduler(hz int, handler TickHandler) *TickScheduler {
	return &TickScheduler{
		hz:       hz,
		interval: time.Second / time.Duration(hz),
		handler:  handler,
		stopChan: make(chan struct{}),
	}
}

func (ts *TickScheduler) Start(ctx context.Context) {
	slog.Info("Starting TickScheduler...", "Hz", ts.hz, "TargetInterval", ts.interval)

	// Timestep fixo usando time.Ticker e temporização baseada em relógio de alta precisão
	ticker := time.NewTicker(ts.interval)
	defer ticker.Stop()

	lastTime := time.Now()

	for {
		select {
		case <-ts.stopChan:
			slog.Info("TickScheduler stopped gracefully by request.")
			return
		case <-ctx.Done():
			slog.Info("TickScheduler stopped due to context cancellation.")
			return
		case current := <-ticker.C:
			ts.currentTick++
			
			// Calcular o delta de tempo real
			elapsed := current.Sub(lastTime)
			lastTime = current

			// Detecção de Overrun / Lag
			if elapsed > ts.interval + (5 * time.Millisecond) {
				lagAmount := elapsed - ts.interval
				slog.Warn("Tick Overrun Detected (Game Loop Lagging!)",
					"tick", ts.currentTick,
					"targetInterval", ts.interval,
					"actualElapsed", elapsed,
					"lag", lagAmount,
				)
			}

			// Executa a função de processamento físico de MMORPG
			ts.executeHandler(elapsed)
		}
	}
}

func (ts *TickScheduler) executeHandler(elapsed time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic recovered inside Game Loop Tick handler", "error", r, "tick", ts.currentTick)
		}
	}()
	
	ts.handler(elapsed, ts.currentTick)
}

func (ts *TickScheduler) Stop() {
	close(ts.stopChan)
}
