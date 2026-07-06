# Backend Rule Registry Integration Audit

## 1. Objetivo

Este documento audita e mapeia onde as regras de negócio puras, já implementadas no **Rule Registry** (`backend/pkg/gamedata/rules`), deverão ser integradas futuramente nos sistemas de jogo em tempo real (runtime) do backend.

O objetivo é fornecer um guia claro para os desenvolvedores sobre como e onde invocar as funções de validação de regras, garantindo que a arquitetura server-authoritative seja mantida de forma consistente em todo o projeto.

## 2. Estado atual

- As regras de negócio puras já existem no package `backend/pkg/gamedata/rules` e possuem cobertura de testes unitários.
- Todos os testes para este package e para o backend como um todo estão passando.
- **Nenhuma integração runtime foi feita ainda.** As funções de validação de regras não estão sendo chamadas por nenhum sistema de jogo (combate, inventário, etc.).
- Esta task é exclusivamente de documentação e auditoria. Nenhuma alteração no comportamento do jogo foi realizada.

## 3. Princípio de integração

Toda integração futura deve seguir rigorosamente os seguintes princípios:

- **Fonte da Verdade:** O backend é a única fonte da verdade. As funções do Rule Registry são a manifestação canônica dessa verdade.
- **Intenção do Cliente:** O cliente apenas envia intenções (ex: "quero atacar o alvo X", "quero escolher a classe Y"). O backend valida essa intenção usando as regras antes de permitir a ação.
- **Validação Pré-Mutação:** Toda regra deve ser validada **antes** de qualquer mutação de estado (ex: antes de salvar no banco de dados, antes de calcular dano, antes de mover um item).
- **Falha Bloqueia Mutação:** Uma falha na validação de uma regra deve impedir a mutação de estado e, idealmente, retornar um erro claro para o sistema chamador.
- **Auditoria:** Ações críticas, especialmente as relacionadas à economia, progressão e PvP, devem gerar logs de auditoria para rastreabilidade e detecção de anomalias.

## 4. Auditoria por regra

### 4.1. Official Catalog Integration

- **Pontos de Integração Futuros:**
  - **Character Creation:** O serviço de criação de personagem deve usar o `NewDefaultRegistry()` para validar se a raça e a classe inicial escolhidas pelo jogador são válidas.
  - **Client UI:** A UI de criação de personagem no cliente deve, idealmente, ser populada com dados de uma cópia read-only do catálogo para evitar divergências.
  - **Sistemas Futuros:** Qualquer sistema que lide com raças, classes ou elementos (ex: bônus raciais, restrições de equipamento) deve validar os IDs contra o catálogo.
- **Riscos de Falha na Integração:**
  - Permitir a criação de um personagem com uma raça não jogável (`ogre`).
  - Permitir o uso de elementos inválidos (`air`, `water`) em sistemas de afinidade ou encantamento.
  - Divergência entre as opções exibidas no cliente e as regras reais do backend, causando frustração no jogador.

### 4.2. Class Selection Integration

- **Pontos de Integração Futuros:**
  - **Character Progression System:** O sistema que gerencia a evolução do personagem.
  - **Endpoint/Handler de Seleção de Classe:** A rota da API (ou handler de pacote de rede) que recebe a intenção do jogador de escolher uma classe.
  - **Banco de Dados:** O campo `class_id` na tabela de personagens.
  - **Client UI:** A interface que o jogador usa para escolher a classe no level 10.
- **Regras a serem validadas por `CanSelectBaseClass`:**
  - O personagem deve estar na classe `novice`.
  - O personagem deve ter `level >= 10`.
  - A classe alvo deve ser uma das 5 classes base oficiais.
- **Riscos de Falha na Integração:**
  - Permitir que um jogador troque de classe após já ter escolhido uma.
  - Permitir que um cliente malicioso force a escolha de uma classe inválida.
  - Permitir a escolha de classe antes do level 10.

### 4.3. Advanced Class Evolution Integration

- **Pontos de Integração Futuros:**
  - **Quest Runtime System:** O sistema que verifica a conclusão de quests.
  - **Progression/Affinity System:** O sistema que rastreia o nível do personagem e os níveis de afinidade.
  - **Character State:** O estado atual do personagem (classe, etc.).
  - **Client UI:** A interface de evolução de classe.
- **Regras a serem validadas por `CanEvolveAdvancedClass`:**
  - `characterLevel >= 100`.
  - `affinityLevel >= 100`.
  - `questCompleted == true`.
  - A `BaseClass` e o `Element` devem ser oficiais.
- **Riscos de Falha na Integração:**
  - Evolução sem ter completado a quest necessária.
  - Evolução antes do nível 100.
  - Evolução com uma combinação de classe/elemento inválida.
  - Cliente forçar uma evolução que não atende aos requisitos.

### 4.4. Open PvP Level Gate Integration

- **Pontos de Integração Futuros:**
  - `backend/pkg/pvp` e `backend/pkg/combat`: Os pacotes que gerenciam a lógica de combate.
  - **Damage Calculation Logic:** A função que calcula o dano final.
  - **Attack/Cast Validation:** A validação inicial de qualquer ação hostil.
- **Regras a serem validadas por `CanEngageOpenPvP`:**
  - `AttackerLevel >= 10`.
  - `TargetLevel >= 10`.
  - A validação deve ocorrer **antes** de qualquer cálculo de dano ou aplicação de efeitos.
- **Riscos de Falha na Integração:**
  - Permitir que um jogador de alto nível ataque um personagem de nível baixo (ex: level 9).
  - Permitir que um personagem de nível baixo inicie um combate PvP.
  - Gastar recursos do servidor calculando dano para uma ação que deveria ser bloqueada.

### 4.5. Safe Zone Integration

- **Pontos de Integração Futuros:**
  - **Map/Chunk Metadata:** O sistema que define as propriedades de cada área do mapa.
  - **Region/Zone System:** O sistema que identifica em qual tipo de zona um jogador se encontra.
  - **Combat/PvP Validation:** A mesma validação de ação hostil do PvP Level Gate.
- **Regras a serem validadas por `CanCombatOccurInZone`:**
  - A `ZoneType` deve ser oficial.
  - A `ZoneType` não pode ser `safe`.
- **Riscos de Falha na Integração:**
  - Permitir combate dentro de cidades ou outras áreas seguras.
  - Implementar a lógica de safe zone apenas no cliente, permitindo que um cliente modificado a ignore.
  - Falha em associar uma `ZoneType` a todas as áreas do mapa, deixando brechas.

### 4.6. Skill Progression Integration

- **Pontos de Integração Futuros:**
  - `backend/pkg/progression`: O pacote que gerencia o ganho de XP.
  - `backend/pkg/combat`, `backend/pkg/professions`: Os sistemas que geram eventos de "uso de skill".
  - **Anti-Macro System:** Um futuro sistema para detectar comportamento repetitivo.
  - **Persistence Layer:** Onde o novo valor da skill é salvo.
- **Regras a serem validadas por `CanGrantSkillProgression`:**
  - A `Source` deve ser oficial (ex: `combat_use`).
  - A ação e o contexto devem ser validados pelo backend (ex: o ataque acertou um alvo válido?).
  - O `MillisecondsSinceLastGain` deve ser `>= 1000`.
  - `DiminishingReturnsBlocked` deve ser `false`.
- **Riscos de Falha na Integração:**
  - Permitir que jogadores ganhem skill XP com macros (ex: atacando o ar).
  - Permitir que um cliente envie pacotes para conceder XP de skill a si mesmo.
  - Ganhos de skill muito rápidos devido à falta de validação de intervalo.

### 4.7. Economy/Loot Authority Integration

- **Pontos de Integração Futuros:**
  - `backend/pkg/economy`, `backend/pkg/inventory`, `backend/pkg/persistence`.
  - **Loot Generation System:** O sistema que decide os drops de monstros.
  - **Trade System:** A lógica de troca entre jogadores.
  - **Quest Reward System:** O sistema que entrega recompensas de quests.
  - **Audit Logs:** O sistema de logging para ações críticas.
- **Regras a serem validadas por `CanApplyItemMutation`:**
  - O cliente **nunca** pode criar um item (`ClientCreatedItem` deve ser `false`).
  - `loot_drop` e `quest_reward` devem vir de um item gerado pelo backend.
  - `player_trade` deve ser transacional.
  - Toda mutação deve ser auditada.
- **Riscos de Falha na Integração:**
  - **Duplicação de itens (dupe):** O risco mais crítico, geralmente causado por trades não transacionais.
  - Itens criados do nada por um cliente malicioso.
  - Manipulação de loot ou recompensas de quests.
  - Ausência de logs para rastrear a origem e o destino de itens valiosos.

## 5. Ordem de integração recomendada

A integração deve seguir uma ordem lógica para construir sobre fundações seguras:

1.  **Character Creation Rule Contract:** Integrar o catálogo com a criação de personagem.
2.  **Class Selection Runtime Integration Plan:** Integrar a regra de seleção de classe.
3.  **PvP/Safe Zone Integration Plan:** Integrar as regras de combate e zona.
4.  **Skill Progression Runtime Integration Plan:** Integrar as regras de ganho de skill.
5.  **Economy/Loot Runtime Integration Plan:** Integrar as regras de autoridade sobre itens.
6.  **Client Read-Only Rule Catalog Plan:** Planejar como o cliente receberá uma cópia segura das regras.
7.  **Protocol Change Review:** Revisar o protocolo de rede somente se as integrações exigirem.

## 6. Critérios gerais para qualquer integração futura

Toda task de integração deve obrigatoriamente atender aos seguintes critérios:

- Compilar sem erros.
- Incluir testes unitários e de integração no backend para validar a integração.
- Não quebrar o funcionamento do cliente de debug atual.
- Não alterar o protocolo de rede sem um documento de planejamento e aprovação.
- Não expor tokens, senhas ou dados sensíveis em logs.
- Manter o princípio de que o cliente não pode mutar estado autoritativo.
- Rejeitar qualquer tentativa do cliente de criar item, XP, skill, classe, evolução ou dano.
- **Toda regra deve ser validada antes da mutação de estado.**

## 7. Arquivos runtime que provavelmente serão auditados futuramente

Os seguintes pacotes e diretórios do backend são os principais candidatos a serem modificados e auditados durante as tasks de integração:

- `backend/pkg/persistence/`
- `backend/pkg/progression/`
- `backend/pkg/combat/`
- `backend/pkg/pvp/`
- `backend/pkg/inventory/`
- `backend/pkg/economy/`
- `backend/pkg/quest/`
- `backend/pkg/protocol/`
- `backend/cmd/` (especificamente os handlers que recebem as intenções do jogador)

## 8. O que esta task NÃO faz

- **Não cria código Go.**
- **Não muda o comportamento runtime do jogo.**
- **Não integra nenhum sistema.**
- **Não altera o banco de dados, o protocolo de rede ou o cliente.**

## 9. Status

Este documento é uma auditoria técnica preparatória. Ele serve como um mapa para a integração segura e consistente do Rule Registry nos sistemas de jogo do backend, garantindo a robustez da arquitetura server-authoritative.