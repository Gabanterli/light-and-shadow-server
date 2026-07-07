# R2 Real Creature Spawn State Architecture

Data: 2026-07-07

## Contexto

Este documento abre a R2 — Real PvE Foundation do projeto Light and Shadow.

A R1 fechou o loop técnico debug:

    login -> character select -> world entry -> movement -> attack -> damage -> death -> visual dead state -> debug loot -> inventory sync -> retry

A R2 inicia a migração do debug-only para uma base real de PvE MMO.

O objetivo desta task é definir a arquitetura server-authoritative para estado real de criaturas, substituindo gradualmente o `Orc_Elite` fixo/debug por spawn state real.

## Problema atual

Durante a R1, o alvo `Orc_Elite` foi usado como entidade técnica fixa para validar combate, morte, loot debug, inventário e retry.

Isso foi correto para bootstrap, mas não serve como sistema final porque:

- não existe `spawn_id`;
- não existe estado por spawn;
- não existe respawn timer real;
- não existe controle de múltiplos players;
- não existe ownership de morte;
- não existe elegibilidade de loot;
- não existe proteção final contra double-death/double-loot;
- não existe lifecycle real da criatura.

## Objetivo da arquitetura

Criar uma fundação server-side onde cada criatura existente no mundo seja controlada por um estado autoritativo de spawn.

Cada spawn deve ter identidade própria, lifecycle próprio e integração clara com combate, AOI, loot e persistência futura.

## Entidades conceituais

### CreatureDefinition

Representa o tipo/base da criatura.

Campos recomendados:

- `CreatureID`
- `Name`
- `Level`
- `Faction`
- `BaseStats`
- `CombatProfile`
- `LootTableID`
- `RespawnPolicyID`
- `AIProfileID`
- `Element`
- `RaceOrFamily`

Exemplo:

    creature_id: orc_elite
    name: Orc Elite
    level: 42
    faction: Monsters
    loot_table_id: orc_elite_t1
    respawn_policy_id: debug_fast_respawn

### CreatureSpawn

Representa um ponto real de spawn no mundo.

Campos recomendados:

- `SpawnID`
- `CreatureID`
- `MapID`
- `X`
- `Y`
- `Z`
- `SpawnRadius`
- `RespawnDelaySeconds`
- `MaxAlive`
- `RegionID`
- `ChunkID`
- `Enabled`

Exemplo:

    spawn_id: debug_orc_elite_001
    creature_id: orc_elite
    x: 136
    y: 127
    z: 0
    respawn_delay_seconds: 30

### CreatureRuntimeState

Representa o estado vivo/morto da criatura em runtime.

Campos recomendados:

- `RuntimeEntityID`
- `SpawnID`
- `CreatureID`
- `CurrentHP`
- `MaxHP`
- `Alive`
- `SpawnedAt`
- `DiedAt`
- `NextRespawnAt`
- `KillerPlayerID`
- `LastDamagerPlayerID`
- `DamageContributors`
- `LootGenerated`
- `CorpseID`
- `Version`

Exemplo:

    runtime_entity_id: creature:debug_orc_elite_001:1
    spawn_id: debug_orc_elite_001
    alive: true
    current_hp: 80
    max_hp: 80

## Lifecycle esperado

### Spawn inicial

Fluxo:

    load spawn definitions -> create runtime creature state -> register combat entity -> register spatial/AOI entity -> notify visible players

### Dano

Fluxo:

    player attack -> validate target runtime entity -> apply damage -> update HP -> track contributor -> emit damage event

### Morte

Fluxo:

    HP <= 0 -> transition alive=false -> lock death state -> emit target dead -> generate loot eligibility -> create corpse or direct loot event -> schedule respawn

### Respawn

Fluxo:

    now >= next_respawn_at -> recreate runtime combat entity -> set alive=true -> reset HP -> register AOI -> notify visible players

## Regras server-authoritative

O servidor deve ser autoridade absoluta sobre:

- spawn existente;
- estado alive/dead;
- HP;
- morte;
- geração de loot;
- respawn;
- posição;
- visibilidade AOI;
- elegibilidade de loot.

O client nunca deve decidir:

- que uma criatura existe;
- que uma criatura reviveu;
- que uma criatura dropou item;
- que uma criatura morreu;
- quem tem direito ao loot.

## Proteções contra exploits

### Double-death

A transição para morto deve ser atômica.

Regra:

    alive=true + hp<=0 -> alive=false

Somente uma execução deve gerar morte, loot e respawn timer.

### Double-loot

Loot deve ser gerado uma vez por `RuntimeEntityID` ou por `DeathEventID`.

Regra:

    loot_generated=false -> generate loot -> loot_generated=true

### Spawn abuse

Respawn não deve depender de botão client-side.

Regra:

    respawn somente por timer server-side ou admin/debug command autorizado

### AOI spoof

Client não pode atacar entidade fora da autoridade do servidor.

Regra:

    ataque deve validar existência, alive state, range, região, linha de visão futura e combat eligibility

## Integrações necessárias

### CombatManager

Deve registrar criatura por `RuntimeEntityID`, não apenas por nome fixo.

Exemplo futuro:

    creature:debug_orc_elite_001:1

### Movement/AOI

Spawn deve ser registrado no SpatialIndex/AOI como entidade observável.

### LootSystem

A morte deve chamar geração de loot baseada em tabela real, não item fixo.

### Persistence

Na primeira versão, runtime pode ser em memória.

Mais tarde, persistência pode registrar:

- mortes recentes;
- respawn state;
- auditoria de loot;
- estatísticas de kill.

## Fase inicial recomendada

A primeira implementação real da R2 deve ser mínima:

- criar `CreatureSpawnState` em memória;
- criar manager server-side;
- registrar um spawn real para `Orc_Elite`;
- matar por combat flow existente;
- impedir double-death;
- preparar respawn timer;
- manter loot final fora do primeiro patch.

## Não escopo desta arquitetura

Esta task não implementa:

- loot table real;
- corpse;
- party loot;
- AI completa;
- pathfinding de criatura;
- persistência final de spawn;
- UI final;
- economia player-driven.

## Critério de aceite futuro

A R2-S será considerada fechada quando o projeto tiver:

- spawn real com `SpawnID`;
- runtime state por criatura;
- morte por runtime entity;
- respawn server-side por timer;
- AOI notificando spawn/death/respawn;
- nenhum fluxo final dependendo do `Orc_Elite` fixo/debug.
