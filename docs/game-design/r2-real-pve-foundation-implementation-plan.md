# R2 Real PvE Foundation Implementation Plan

Data: 2026-07-07

## Contexto

Este documento define o plano de implementação da R2 — Real PvE Foundation.

A R2 deve transformar o loop técnico validado na R1 em uma base real de PvE MMO.

A R1 validou:

    login -> world entry -> movement -> attack -> damage -> death -> visual dead -> debug loot -> inventory sync -> retry

A R2 deve migrar esse fluxo para sistemas reais server-authoritative.

## Objetivo da R2

Criar a fundação real de PvE com:

- spawn state real de criatura;
- runtime entity por spawn;
- morte segura;
- respawn real por timer;
- loot table inicial server-side;
- separação entre debug-only e gameplay final;
- documentação runtime de fechamento.

## Fases internas da R2

A R2 será dividida em 5 blocos:

    R2-S — Real Creature Spawn System
    R2-C — Creature Combat State
    R2-L — Real Loot Foundation
    R2-R — Real Respawn Flow
    R2-D — Documentation and Runtime Closure

## R2-S — Real Creature Spawn System

Objetivo:

Criar o primeiro sistema real de spawn de criaturas no servidor.

Tasks recomendadas:

### R2-S-C — Add in-memory creature spawn state manager

Escopo provável:

- `backend/pkg/pve`
- `backend/cmd/gateway/main.go`

Objetivo:

- criar manager de spawn runtime;
- controlar `SpawnID`;
- controlar `CreatureID`;
- controlar alive/dead;
- expor lookup por runtime entity.

### R2-S-D — Register Orc_Elite through spawn manager

Objetivo:

- remover dependência gradual do `Orc_Elite` hardcoded como entidade solta;
- registrar `Orc_Elite` como spawn real debug-compatible;
- manter compatibilidade com botão debug temporariamente.

### R2-S-E — Add spawn lifecycle logs

Objetivo:

- logs de spawn;
- logs de death transition;
- logs de scheduled respawn;
- logs de runtime entity ID.

## R2-C — Creature Combat State

Objetivo:

Conectar combate com runtime creature state.

Tasks recomendadas:

### R2-C-A — Route creature damage through runtime entity state

Objetivo:

- garantir que dano mira runtime entity;
- validar alive/dead antes do ataque;
- proteger contra alvo inexistente.

### R2-C-B — Add atomic death transition guard

Objetivo:

- impedir double-death;
- impedir múltiplos `SC_TARGET_DEAD`;
- impedir múltiplas gerações de loot para a mesma morte.

### R2-C-C — Track basic damage contributors

Objetivo:

- registrar último atacante;
- registrar contribuição básica;
- preparar elegibilidade futura de loot.

## R2-L — Real Loot Foundation

Objetivo:

Substituir loot fixo debug por fundação de loot table server-side.

Tasks recomendadas:

### R2-L-A — Define loot table runtime model

Objetivo:

- item ID;
- chance;
- quantidade;
- raridade;
- source creature;
- audit metadata.

### R2-L-B — Generate real loot from Orc_Elite table

Objetivo:

- usar tabela real inicial;
- remover concessão fixa direta;
- manter `SC_INVENTORY_SYNC`.

### R2-L-C — Add loot audit log

Objetivo:

- player;
- creature;
- spawn;
- item;
- quantity;
- timestamp;
- death event ID.

## R2-R — Real Respawn Flow

Objetivo:

Remover retry técnico como mecanismo principal e criar respawn real.

Tasks recomendadas:

### R2-R-A — Add respawn timer to spawn state

Objetivo:

- `DiedAt`;
- `NextRespawnAt`;
- timer server-side;
- estado waiting_respawn.

### R2-R-B — Respawn Orc_Elite through server timer

Objetivo:

- após morte, aguardar delay;
- recriar entidade;
- notificar AOI;
- permitir novo kill.

### R2-R-C — Remove debug retry dependency from real flow

Objetivo:

- botão debug pode continuar para teste;
- sistema final não depende dele.

## R2-D — Documentation and Runtime Closure

Objetivo:

Documentar fechamento runtime da R2.

Tasks recomendadas:

### R2-D-A — Document real spawn runtime validation

### R2-D-B — Document real loot runtime validation

### R2-D-C — Document real respawn runtime validation

### R2-D-D — Document R2 phase closure

## Ordem recomendada de execução

Ordem segura:

    R2-S-C
    R2-S-D
    R2-S-E
    R2-C-A
    R2-C-B
    R2-C-C
    R2-L-A
    R2-L-B
    R2-L-C
    R2-R-A
    R2-R-B
    R2-R-C
    R2-D-A
    R2-D-B
    R2-D-C
    R2-D-D

## Critério de saída da R2

A R2 estará fechada quando for possível validar:

    real spawn -> attack -> damage -> death -> guarded death transition -> real loot table -> inventory sync -> respawn timer -> spawn alive again -> repeat kill

Sem depender de:

- loot fixo debug;
- revive no botão de ataque;
- target fixo sem spawn ID;
- death state apenas visual;
- estado de criatura sem runtime lifecycle.

## Regras de segurança para implementação

Cada task deve respeitar:

- uma task por commit;
- não usar `git add .`;
- validar build antes de commit;
- validar runtime quando houver alteração de código;
- sempre push;
- sempre verificar branch/status após push;
- não misturar loot, respawn e combat guard no mesmo commit;
- não transformar debug helper em sistema final sem documentação.

## Próxima task recomendada após este plano

A próxima task deve ser:

    R2-S-C — Add in-memory creature spawn state manager

Objetivo:

Criar o primeiro manager server-side em memória para estado real de spawn de criaturas.
