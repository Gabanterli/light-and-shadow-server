# Game Design Rule Registry Plan

## 1. Objetivo

Este documento planeja como transformar as diretrizes de design da Fase 1 do projeto em um conjunto de regras de jogo versionadas, centralizadas e auditáveis, conhecido como "Rule Registry". O objetivo é criar uma fundação sólida para o balanceamento e a evolução do jogo, garantindo que a lógica de negócio seja segura e consistente.

## 2. O que é um Rule Registry

O Rule Registry é uma camada de configuração e dados que define as regras fundamentais do jogo. Ele não é código de lógica, mas sim os dados que a lógica utiliza. Cada entrada no registro possui um ID estável, um nome oficial, limites numéricos (como level mínimo) e outras regras de validação que governam o comportamento do jogo.

## 3. Por que não hardcodar regras espalhadas

Deixar regras de jogo "hardcodadas" (fixas diretamente no código) em múltiplos locais é uma prática de alto risco em um projeto MMO. Os principais problemas são:

- **Balanceamento Difícil:** Mudar uma regra, como o dano de uma skill, exige encontrar e alterar múltiplos trechos de código, com alto risco de erro.
- **Exploits:** Inconsistências entre a validação do cliente e do servidor criam brechas para exploits.
- **Inconsistência Cliente/Backend:** O cliente pode exibir uma informação (ex: custo de mana) que não corresponde à regra real no servidor.
- **Retrabalho:** Qualquer mudança de design se torna uma tarefa de programação complexa e demorada.
- **Dificuldade de Migração:** Atualizar o jogo ou migrar dados de jogadores se torna um pesadelo se as regras não forem versionadas.
- **Dificuldade de Auditoria:** É quase impossível auditar o estado das regras do jogo em um determinado momento.

## 4. Fonte da verdade

A arquitetura de um MMO seguro depende de uma clara separação de responsabilidades:

- O **Backend** é a única fonte da verdade para todas as regras de jogo. Ele valida 100% das ações do jogador.
- O **Cliente** apenas exibe o estado do jogo e envia as intenções do jogador para o backend. Ele nunca toma decisões finais.
- O **Banco de Dados** persiste o estado do jogador (level, itens, skills) conforme validado e modificado pelo backend.
- O **Protocolo** transporta o estado do jogo (do backend para o cliente) e as intenções do jogador (do cliente para o backend).
- A **Documentação** (como este Rule Registry) registra as decisões de design de forma que possam ser implementadas e auditadas.

## 5. Escopo inicial do Rule Registry

Para a Fase 1, o Rule Registry deve cobrir as seguintes áreas:

- Raças jogáveis.
- Classes base.
- Elementos.
- Regra de level mínimo para PvP (PvP level gate).
- Regras de zonas seguras (safe zone rules).
- Regras de progressão de skills por uso.
- Regras de afinidade de classe/elemento.
- Regras de evolução avançada de classe por quest.
- Restrições de economia e comércio (economy/trade constraints).
- Regras de geração de loot.
- Regras de progressão de personagem (character progression).

## 6. IDs internos propostos

Cada entidade no Rule Registry deve ter um ID de texto estável, único e padronizado (inglês, minúsculas, `snake_case`). Isso desacopla a lógica do nome de exibição (que pode mudar ou ser traduzido).

**Raças Jogáveis (Races):**
- `human`
- `forest_elf`
- `dwarf`
- `ice_elf`
- `green_orc`

**Classes Base (Classes):**
- `knight`
- `mage`
- `archer`
- `assassin`
- `cleric`

**Elementos (Elements):**
- `fire`
- `earth`
- `ice`
- `shadow`
- `sacred`

## 7. Regras de classe base

- Um personagem recém-criado começa como uma classe base (ex: `novice`).
- A escolha de uma classe base especializada (Knight, Mage, etc.) ocorre ao atingir o **level 10**.
- Após a escolha, a classe base do personagem fica fixa.
- O backend deve validar o level mínimo para a seleção de classe.
- O cliente apenas exibe as opções de classe disponíveis para o jogador, sem poder forçar uma escolha inválida.
- Uma futura mecânica de troca de classe base só poderá ser implementada com uma decisão explícita de design e uma task específica.

## 8. Regras de raça

- A raça jogável do personagem deve ser escolhida durante a criação.
- A raça não pode ser trocada livremente após a criação do personagem.
- Quaisquer atributos ou bônus raciais devem ser definidos no Rule Registry e validados exclusivamente pelo backend.
- Não implementar bônus raciais agora sem uma task de design e balanceamento dedicada.
- Ogres não são uma raça jogável nesta definição.

## 9. Regras de elementos

- Os elementos disponíveis são: `fire`, `earth`, `ice`, `shadow`, `sacred`.
- Os elementos `air` e `water` não existem no sistema de afinidade do jogo.
- A "Afinidade Elemental/de Classe" de um personagem deve aumentar com o uso de skills e ações associadas a ela.
- A evolução avançada de classe é decidida pelo backend com base na afinidade e outros critérios.
- O cliente não pode escolher ou forçar uma evolução final diretamente; ele apenas reflete o estado decidido pelo servidor.

## 10. Evolução avançada de classe

- A evolução para uma classe avançada ocorre somente quando o personagem atinge o **level 100**.
- Exige que a afinidade relevante (classe/elemento) também esteja no **level 100**.
- Exige a conclusão de uma **quest específica** para aquela evolução.
- O backend deve validar os três pré-requisitos: level do personagem, level da afinidade e status da quest.
- O cliente apenas exibe os requisitos e envia a intenção de evoluir.
- **Exemplos:**
  - `knight` + afinidade `fire` + quest = Cavaleiro do Fogo
  - `assassin` + afinidade `shadow` + quest = Assassino das Sombras
  - `mage` + afinidade `ice` + quest = Mago de Gelo
  - `cleric` + afinidade `sacred` + quest = Clérigo Sagrado
  - `archer` + afinidade `earth` + quest = Arqueiro da Terra

## 11. PvP e safe zones

- O PvP aberto (Open PvP) só é permitido para personagens de **level 10 ou superior**.
- O backend deve bloquear qualquer tentativa de dano PvP se um dos envolvidos estiver abaixo do level mínimo.
- Zonas seguras (Safe Zones), como cidades, devem bloquear completamente qualquer forma de combate.
- O "fogo amigo" (friendly fire) entre membros da mesma facção/guilda deve ser bloqueado por padrão.
- Modificadores de dano em PvP são calculados e aplicados exclusivamente pelo backend.
- O cliente pode exibir avisos visuais (ex: ícone de caveira, cor do nome), mas nunca decide se uma ação de combate é permitida.

## 12. Skill progression por uso

- As skills evoluem com o uso e treino contínuo.
- O backend é o único responsável por calcular o ganho de experiência da skill e persistir o novo valor.
- O sistema deve ser projetado para prevenir abuso por macros e exploits (ex: atacar um monstro imortal por horas).
- Regras de cooldown, "diminishing returns" (ganhos decrescentes) e validação de contexto (ex: a skill foi usada em um alvo válido?) devem ser consideradas na implementação do backend.

## 13. Economia e loot

- A economia do jogo é "player-driven" (movida pelos jogadores).
- Qualquer troca (trade) entre jogadores deve ser uma operação transacional e atômica no backend para evitar duplicação de itens (dupes).
- A geração de loot raro (drops) é calculada e concedida exclusivamente pelo backend.
- O cliente **nunca** cria um item. Ele apenas recebe a informação de que um novo item foi adicionado ao seu inventário.
- Logs detalhados e auditoria são necessários para todas as operações de trade, drop e movimentação de itens valiosos.

## 14. Estrutura futura sugerida

A implementação do Rule Registry será dividida entre o backend (fonte da verdade) e o cliente (cópia read-only para exibição). A estrutura de pastas a seguir é proposta, mas **não deve ser criada nesta task**.

**Backend (Go):**
```
backend/pkg/gamedata/
backend/pkg/gamedata/rules/
backend/pkg/gamedata/rules/races.go
backend/pkg/gamedata/rules/classes.go
backend/pkg/gamedata/rules/elements.go
backend/pkg/gamedata/rules/pvp.go
backend/pkg/gamedata/rules/advanced_classes.go
```

**Cliente (C#):**
```
scripts/GameData/
scripts/GameData/ClientRuleCatalog.cs
```

## 15. Ordem de implementação recomendada

As próximas tasks devem seguir uma ordem lógica para construir o sistema de regras:

- **Task 6B** — Backend Rule Registry Skeleton
- **Task 6C** — Race/Class/Element Constants
- **Task 6D** — Character Creation Rule Audit
- **Task 6E** — Class Selection Level 10 Backend Rule
- **Task 6F** — PvP Level Gate Rule
- **Task 6G** — Safe Zone Rule Contract
- **Task 6H** — Skill Progression Rule Plan
- **Task 6I** — Advanced Class Evolution Rule Plan
- **Task 6J** — Economy/Loot Audit Plan

## 16. Critérios de aceite futuros

Qualquer implementação de código baseada neste plano deve obrigatoriamente:

- Compilar sem erros.
- Incluir testes unitários e de integração no backend para validar as regras.
- Manter o princípio de arquitetura "server-authoritative".
- Não quebrar o funcionamento do cliente de debug atual.
- Não alterar o protocolo de rede sem um documento de planejamento e aprovação.
- Não expor senhas ou tokens em logs ou no cliente.
- Não permitir que o cliente crie ou modifique arbitrariamente itens, XP, level, skills, classes ou evoluções avançadas.

## 17. Status desta task

Esta task é exclusivamente de planejamento e documentação. Nenhuma alteração no comportamento do jogo, código ou assets foi realizada. Este documento servirá como guia para as próximas tasks de implementação.