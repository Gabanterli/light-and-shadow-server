# Backend Rule Registry Implementation Status

## 1. Objetivo

Este documento consolida o status atual da implementação do **Rule Registry** no backend do projeto Light and Shadow. Ele serve como um guia técnico para desenvolvedores, detalhando as regras de negócio que foram codificadas, testadas e validadas, e estabelecendo a base para as futuras integrações com os sistemas de jogo em tempo real (runtime).

O objetivo é garantir que toda a equipe tenha uma visão clara da arquitetura server-authoritative e das regras que já são consideradas canônicas no servidor.

## 2. Escopo da Fase 6 até agora

As seguintes tasks, representando a construção da fundação do Rule Registry, foram concluídas e integradas:

- `3adca56`: Document game design rule registry plan
- `c99326f`: Add backend rule registry skeleton
- `4e398f5`: Add official game rule catalog
- `5f08e2e`: Add class selection rule definitions
- `84c1e39`: Add advanced class evolution rule definitions
- `e0dc4db`: Add open pvp level gate rules
- `22f70c0`: Add safe zone combat rules
- `bd3361b`: Add skill progression eligibility rules
- `460ced9`: Add economy loot authority rules

## 3. Package principal

Toda a lógica de regras de negócio puras está centralizada no package:

`backend/pkg/gamedata/rules`

Este package contém as definições, constantes, erros e funções de validação que formam a "constituição" do jogo. Ele é projetado para ser a única fonte da verdade para as regras, sem dependências de sistemas runtime como combate, inventário ou protocolo de rede.

## 4. Regras de design bloqueadas

As seguintes entidades de jogo foram oficialmente definidas e implementadas no catálogo de regras. Qualquer entidade fora desta lista não é considerada oficial para mecânicas de jogador.

**Raças Jogáveis:**
- `human` — Humano
- `forest_elf` — Elfo da Floresta
- `dwarf` — Anão
- `ice_elf` — Elfo de Gelo
- `green_orc` — Orc Verde

> **Importante:** `Ogre` **NÃO** é uma raça jogável. Ele pode existir futuramente como criatura ou NPC, mas não pode ser selecionado por jogadores.

**Classes Base:**
- `knight` — Cavaleiro
- `mage` — Mago
- `archer` — Arqueiro
- `assassin` — Assassino
- `cleric` — Clérigo

**Elementos:**
- `fire` — Fogo
- `earth` — Terra
- `ice` — Gelo
- `shadow` — Sombrio
- `sacred` — Sagrado

> **Importante:** Os elementos `air` e `water` **NÃO** existem no jogo. O elemento correto para mecânicas de frio é `ice` (Gelo).

## 5. Componentes implementados

As seguintes regras foram implementadas como módulos puros e testáveis dentro do package `rules`.

### Registry skeleton
- Define os tipos base `RuleID`, `RuleCategory` e `RuleDefinition`.
- Implementa o `Registry` para armazenar e gerenciar todas as regras.
- Garante validação de formato de ID (`snake_case`), rejeita IDs duplicados e fornece listagem determinística.

### Official catalog
- Cadastra as 15 definições oficiais (5 raças, 5 classes, 5 elementos).
- Garante que entidades não oficiais como `ogre`, `air` e `water` não façam parte do catálogo.

### Class selection
- Define que todo personagem começa como `novice`.
- Bloqueia a escolha de classe base para antes do **level 10**.
- Garante que a classe base só pode ser escolhida uma vez (partindo de `novice`).
- Valida que a classe alvo é uma das 5 classes base oficiais.

### Advanced class evolution
- Define os requisitos para a evolução avançada:
  - Nível do personagem: **100+**
  - Nível de afinidade: **100+**
  - Quest de evolução: **concluída**
- Valida que a evolução parte de uma classe base e um elemento oficiais.

### Open PvP level gate
- Define o nível mínimo para engajamento em PvP aberto como **level 10**.
- Valida que tanto o atacante quanto o alvo devem ter level 10 ou superior.
- A validação ocorre antes de qualquer outra lógica de combate.

### Safe zone
- Define os tipos de zona: `safe`, `combat`, `neutral`.
- Implementa a regra de que combate é **bloqueado** em `ZoneTypeSafe`.
- Permite combate em zonas `combat` e `neutral` (por enquanto).

### Skill progression eligibility
- Define as fontes válidas para ganho de XP de skill (`combat_use`, `training_use`, `profession_use`).
- Exige que a ação e o contexto sejam validados pelo backend.
- Impõe um intervalo mínimo de **1000ms** entre ganhos para prevenir spam.
- Inclui um bloqueio por `DiminishingReturnsBlocked` para futuro sistema anti-macro.

### Economy/loot authority
- Define as fontes válidas de mutação de item (`loot_drop`, `player_trade`, etc.).
- Impõe a regra de que o **cliente nunca pode criar um item** (`ErrClientCreatedItemRejected`).
- Garante que `loot_drop` e `quest_reward` usem um item gerado pelo backend.
- Exige que `player_trade` seja transacional para prevenir duplicação.
- Exige que toda mutação de item seja auditada (`AuditLogged`).

## 6. Garantias server-authoritative

A arquitetura implementada reforça os seguintes princípios de segurança para um MMO:

- **O backend valida tudo:** Toda regra de negócio é validada no servidor antes de qualquer mutação de estado.
- **O cliente é um terminal "burro":** O cliente apenas exibe o estado recebido do servidor e envia as intenções do jogador. Ele não tem autoridade para tomar decisões.
- O cliente **não pode** criar itens, conceder XP/níveis/skills, escolher uma classe ou evolução inválida, ou decidir se uma ação de combate é permitida. Qualquer tentativa nesse sentido é rejeitada pelas regras no backend.

## 7. O que ainda NÃO foi integrado

É crucial entender que os módulos de regras implementados são **puros** e **desacoplados**. Eles ainda **NÃO** foram integrados com os sistemas de jogo em tempo real. As seguintes integrações são necessárias em fases futuras:

- **Character Creation:** A criação de personagem ainda não consome as regras para validar a raça/classe inicial.
- **Banco de Dados:** Nenhuma regra está sendo lida ou escrita no banco de dados.
- **Protocolo de Rede:** As mensagens de rede ainda não invocam as funções de validação de regras.
- **Cliente:** A UI do cliente não reflete dinamicamente o resultado das validações de regras.
- **Combat/PvP Runtime:** O sistema de combate real ainda não chama `CanEngageOpenPvP` ou `CanCombatOccurInZone` antes de calcular dano.
- **Mapa/Chunks:** O sistema de mapa ainda não associa uma `ZoneType` a cada área.
- **Inventory/Economy Runtime:** O inventário e o sistema de trade ainda não invocam `CanApplyItemMutation`.
- **Quest/Progression Runtime:** Os sistemas de quests e progressão ainda não invocam `CanEvolveAdvancedClass` ou `CanGrantSkillProgression`.

## 8. Validação atual

Todos os módulos de regras implementados no package `backend/pkg/gamedata/rules` possuem cobertura de testes unitários. A suíte de testes foi executada em um ambiente Docker isolado, garantindo que o código é válido e se comporta conforme o esperado.

Os seguintes comandos foram executados com sucesso:

```sh
# Validar testes específicos do package de regras
docker run --rm -v "${PWD}\backend:/app" -w /app golang:1.23-alpine go test ./pkg/gamedata/...

# Validar todos os testes do backend
docker run --rm -v "${PWD}\backend:/app" -w /app golang:1.23-alpine go test ./pkg/...
```

Ambos os comandos passaram, confirmando a integridade do código implementado.

## 9. Próximas tasks recomendadas

Com a fundação de regras puras estabelecida, as próximas tasks devem focar em planejar e executar a integração dessas regras com os sistemas runtime.

- **Task 6K-A** — Backend Rule Registry Integration Audit
- **Task 6L-A** — Character Creation Rule Contract
- **Task 6M-A** — Class Selection Runtime Integration Plan
- **Task 6N-A** — PvP/Safe Zone Integration Plan
- **Task 6O-A** — Skill Progression Runtime Integration Plan
- **Task 6P-A** — Economy/Loot Runtime Integration Plan

## 10. Status

Esta task é exclusivamente de documentação. Nenhuma alteração no comportamento do jogo, código ou assets foi realizada. Este documento reflete o estado do projeto após o commit `460ced9`.