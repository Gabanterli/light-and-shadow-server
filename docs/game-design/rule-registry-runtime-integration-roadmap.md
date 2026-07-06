# Rule Registry Runtime Integration Roadmap

## 1. Objetivo

Este documento consolida e organiza a ordem recomendada para as futuras tasks de integração do **Backend Rule Registry** com os sistemas de jogo em tempo real (runtime). Ele serve como um roadmap mestre para garantir que as regras puras, já definidas e testadas, sejam implementadas de forma segura, sequencial e lógica.

## 2. Estado atual

- As regras de negócio puras já existem e estão testadas no package `backend/pkg/gamedata/rules`.
- Os testes unitários para cada regra garantem seu comportamento em isolamento.
- Documentos de planejamento detalhados para cada integração individual já foram criados.
- **Nenhuma integração runtime completa foi feita ainda.** As regras puras não estão sendo consumidas pelos sistemas de jogo.
- Esta task é exclusivamente de documentação e não altera o comportamento do jogo.

## 3. Princípios obrigatórios

Toda e qualquer integração futura deve aderir estritamente aos seguintes princípios:

- **Backend Server-Authoritative:** O backend é a única fonte da verdade.
- **Intenção do Cliente:** O cliente apenas envia intenções (ex: "quero atacar"). Ele nunca envia o resultado ou o estado final (ex: "meu HP agora é 80").
- **Estado Real do Servidor:** O backend sempre carrega o estado real e autoritativo de um personagem, entidade, item, mapa ou progressão a partir de fontes confiáveis do servidor. Dados enviados pelo cliente sobre level, class, element, ZoneType, item, XP, posição ou qualquer outro estado autoritativo são ignorados.
- **Validação Pré-Mutação:** A validação de uma regra deve ocorrer **antes** de qualquer mutação de estado (dano, loot, mudança de classe, etc.) e antes de qualquer cálculo custoso.
- **Testes Obrigatórios:** Cada fase de integração deve ser acompanhada por uma suíte de testes de integração robusta que valide os cenários de sucesso e falha.

## 4. Ordem recomendada das futuras integrações runtime

A integração deve ocorrer em fases sequenciais para construir sobre uma base estável e minimizar riscos.

### Fase R1 — Character Creation Runtime Integration

- **Motivo:** É a fundação de tudo. Garante que todo personagem que entra no mundo já nasce em conformidade com as regras mais básicas (raça oficial, classe `novice`, sem `ogre`, sem `air`/`water`). Sem isso, as outras regras não têm um estado inicial confiável para validar.

### Fase R2 — Class Selection Runtime Integration

- **Motivo:** Depende diretamente de um personagem existente (Fase R1) e de seu estado de progressão (nível). É o primeiro grande marco de progressão do jogador e deve ser seguro antes de introduzir mecânicas mais complexas.

### Fase R3 — PvP/Safe Zone Runtime Integration

- **Motivo:** Deve ser implementada antes de expandir o sistema de combate. Garante a segurança dos novos jogadores (bloqueio de PvP abaixo do nível 10) e a integridade das áreas sociais (cidades como safe zones). É um pré-requisito para um ambiente de jogo justo.

### Fase R4 — Skill Progression Runtime Integration

- **Motivo:** Depende de ações de jogo reais (combate, profissões) serem validadas. Com as regras de combate e zona já no lugar (Fase R3), podemos garantir que a progressão de skill só ocorra em contextos válidos, prevenindo farming em safe zones ou exploits de macro.

### Fase R5 — Economy/Loot Runtime Integration

- **Motivo:** Esta é a fase de maior risco, pois lida diretamente com a geração e transferência de valor (itens). Envolve lógica complexa de transações, prevenção de duplicação (dupe) e auditoria. Deve ser feita após os sistemas de interação básica estarem estabilizados, e requer uma base de persistência muito robusta.

### Fase R6 — Advanced Class Evolution Runtime Integration

- **Motivo:** É a integração mais complexa, pois depende de múltiplos sistemas estarem maduros: personagem (R1), classe base (R2), sistema de quests, sistema de afinidade e progressão de nível. Deve ser a última fase, pois consolida vários aspectos da jornada de um personagem de alto nível.

## 5. Detalhar Fase R1 — Character Creation

- O handler de criação de personagem no backend validará a `race_id` da intenção do cliente contra o catálogo oficial.
- Rejeitará explicitamente `ogre` ou qualquer raça inválida.
- Ignorará qualquer `class_id` enviado pelo cliente e definirá autoritativamente a classe como `novice`.
- Ignorará e rejeitará qualquer `element_id` enviado, bloqueando `air` e `water`.
- O backend definirá todos os atributos iniciais: stats, spawn point e inventário.
- A persistência do novo personagem será uma operação atômica.

## 6. Detalhar Fase R2 — Class Selection

- O handler de seleção de classe receberá a `target_class` da intenção do cliente.
- Carregará o nível e a classe atuais **reais** do personagem do banco de dados.
- Invocará `rules.CanSelectBaseClass` com os dados do servidor.
- A regra garantirá que o nível é `10+`, a classe atual é `novice` e a classe alvo é uma das 5 classes base oficiais.
- A persistência usará uma query transacional e condicional (ex: `WHERE class_id = 'novice'`) para evitar race conditions.

## 7. Detalhar Fase R3 — PvP/Safe Zone

- O sistema de combate, antes de processar qualquer dano, cast ou efeito hostil, invocará `rules.CanCombatOccurInZone` usando a `ZoneType` real da área (obtida dos metadados do mapa do servidor).
- Se o alvo for um jogador, o sistema também invocará `rules.CanEngageOpenPvP`, usando os níveis reais do atacante e do alvo.
- Uma falha em qualquer uma dessas validações interromperá a ação hostil imediatamente.

## 8. Detalhar Fase R4 — Skill Progression

- Após uma ação de jogo ser validada e executada com sucesso pelo servidor (um ataque que acertou, um craft que foi concluído), o sistema de progressão será acionado.
- Ele construirá uma `SkillProgressionEligibilityRequest` com a `Source` correta (`combat_use`, `profession_use`, etc.), `ActionValidatedByBackend: true`, `ContextValidatedByBackend: true`, e os dados de tempo e anti-macro.
- A chamada a `rules.CanGrantSkillProgression` será o portão final antes de calcular e persistir o ganho de XP da skill.

## 9. Detalhar Fase R5 — Economy/Loot

- Todo sistema que modifica um inventário (loot, trade, quest, move) deverá invocar `rules.CanApplyItemMutation`.
- A `Source` correta será identificada (`loot_drop`, `player_trade`, etc.).
- O sistema garantirá que `ClientCreatedItem` é sempre `false`.
- Para `loot_drop` e `quest_reward`, o flag `BackendGeneratedItem` será `true`.
- Para `player_trade`, o flag `Transactional` será `true`, e a operação será atômica.
- O flag `AuditLogged` será `true` para garantir que um log de auditoria seja gravado.

## 10. Detalhar Fase R6 — Advanced Class Evolution

- O handler de evolução de classe validará a intenção do jogador usando `rules.CanEvolveAdvancedClass`.
- Ele carregará o nível real do personagem, o nível de afinidade, o status da quest de evolução, a classe base e o elemento desejado.
- A regra garantirá que todos os requisitos (nível 100+, afinidade 100+, quest concluída, classe e elemento oficiais) sejam atendidos antes de permitir a evolução.

## 11. Dependências entre sistemas

- **Character Creation (R1)** é a base para todos os outros.
- **Class Selection (R2)** depende de um personagem existente (R1).
- **Advanced Evolution (R6)** depende da classe base (R2) e de sistemas de quest/afinidade.
- **PvP/Safe Zone (R3)** depende do estado do personagem (nível) e do mapa.
- **Skill Progression (R4)** depende de ações validadas por outros sistemas, como o de combate (R3).
- **Economy/Loot (R5)** depende de uma infraestrutura de persistência e inventário muito robusta.

## 12. Riscos globais

Aderir a este roadmap mitiga os seguintes riscos de arquitetura em um MMO:

- **Client Authority:** A validação em cada fase impede que o cliente dite o estado do jogo.
- **Race Conditions:** Planos de integração exigem lógica transacional e condicional para prevenir estados inconsistentes.
- **Duplication (Dupe) Exploits:** A fase de economia (R5) foca especificamente em transações atômicas para prevenir a duplicação de itens.
- **Macro/Botting:** A fase de progressão (R4) introduz as bases para sistemas anti-macro.
- **Rollback Parcial:** A ênfase em transações atômicas previne que apenas parte de uma operação seja concluída (ex: item removido de um jogador, mas não adicionado ao outro).
- **Validação Pós-Mutação:** A ordem de validação em cada fase garante que a verificação ocorra *antes* da mudança de estado.

## 13. Política de implementação futura

- Cada fase de integração (R1-R6) será dividida em uma ou mais tasks de escopo pequeno.
- Cada task deve começar com um `git status` limpo.
- Cada task deve ter um conjunto claro de arquivos permitidos.
- Testes de integração são obrigatórios para cada nova implementação de runtime.
- Commits devem ser atômicos e representar uma única mudança lógica.
- Não misturar diferentes fases de integração ou refatorações não relacionadas na mesma task.

## 14. Critérios globais de aceite

Uma task de integração runtime só será considerada "concluída" quando:

- O backend compilar e todos os testes (unitários e de integração) passarem.
- A regra pura correspondente for chamada no ponto correto do fluxo de execução.
- A validação ocorrer **antes** de qualquer mutação de estado persistente.
- O cliente for comprovadamente incapaz de contornar a regra.
- Erros de validação forem tratados corretamente, retornando feedback ao sistema chamador.
- Logs de auditoria forem implementados para ações críticas (economia, progressão).

## 15. O que este roadmap NÃO faz

- Não implementa código Go, C# ou qualquer outra linguagem.
- Não altera o comportamento runtime do jogo.
- Não altera o banco de dados, o protocolo de rede ou o cliente.
- Não cria endpoints ou handlers de rede.
- Não altera as regras puras já existentes no Rule Registry.

## 16. Status

Este documento é o roadmap oficial para orientar as próximas fases de desenvolvimento, garantindo que a integração do Backend Rule Registry com os sistemas de jogo seja feita de forma incremental, segura e robusta.
