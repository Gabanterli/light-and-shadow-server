# PvP and Safe Zone Integration Plan

## 1. Objetivo

Este documento planeja a futura integração das regras de negócio de **Open PvP Level Gate** e **Safe Zone** no ambiente de execução (runtime) do backend. Ele detalha como as funções puras `rules.CanEngageOpenPvP` and `rules.CanCombatOccurInZone` serão utilizadas para validar ações hostis de forma segura, antes de qualquer cálculo de dano ou mutação de estado.

## 2. Estado atual

- As regras puras de PvP e Safe Zone já existem e estão testadas em `backend/pkg/gamedata/rules`.
- Os testes unitários que validam a lógica das regras em isolamento já existem.
- **Nenhuma integração runtime foi feita ainda.** Nenhum sistema de jogo (combate, mapa, etc.) invoca as funções de validação.
- Esta task é exclusivamente de documentação e não altera o comportamento do jogo.

## 3. Princípio server-authoritative

A integração deve seguir rigorosamente a arquitetura server-authoritative:

- O **cliente** apenas envia uma **intenção** de realizar uma ação hostil (ex: "quero atacar o alvo X").
- O **backend** carrega o estado **real e autoritativo** do atacante, do alvo e da zona do mapa a partir de suas fontes da verdade (memória ou banco de dados).
- O **backend** valida a intenção usando as regras de PvP Level Gate e Safe Zone **antes** de processar qualquer lógica de combate (dano, efeitos, etc.).
- O cliente **nunca** decide se uma ação de combate é permitida.

## 4. Regras oficiais de PvP

A lógica de negócio do Open PvP Level Gate, já codificada, é:

- O nível mínimo para engajamento em PvP aberto é **10**.
- O atacante deve ter `level >= 10`.
- O alvo (se for outro jogador) deve ter `level >= 10`.
- A validação deve ocorrer antes de qualquer cálculo de dano, cast de magia hostil, aplicação de efeito negativo (debuff) ou geração de ameaça (aggro) PvP.

## 5. Regras oficiais de Safe Zone

A lógica de negócio de Safe Zone, já codificada, é:

- `ZoneTypeSafe` bloqueia completamente qualquer forma de combate.
- `ZoneTypeCombat` permite combate.
- `ZoneTypeNeutral` permite combate (regras mais específicas, como facção vs facção, podem ser adicionadas futuramente sobre esta base).
- Uma `ZoneType` inválida deve ser tratada como uma condição de erro que bloqueia a ação.
- A `ZoneType` de uma área é definida exclusivamente pelo backend (via metadados do mapa) e nunca pela UI ou pelo cliente.

## 6. Fluxo runtime futuro recomendado para ataque direto

1.  O cliente envia uma intenção de ataque para o backend (ex: `AttackRequest{TargetID: "..."}`).
2.  O backend autentica a sessão/conexão do jogador.
3.  O backend identifica o personagem **real** do atacante associado à conexão.
4.  O backend identifica o personagem/monstro **real** do alvo.
5.  O backend verifica o estado de ambos (ex: se estão vivos, online, etc.).
6.  O backend carrega o nível **real** do atacante (`attacker.Level`).
7.  O backend carrega o nível **real** do alvo (`target.Level`).
8.  O backend determina a `ZoneType` **real** da posição relevante (ex: a posição do alvo).
9.  O backend chama `err := rules.CanCombatOccurInZone(...)`. Se houver erro, a ação é rejeitada.
10. Se o alvo for outro jogador, o backend chama `err := rules.CanEngageOpenPvP(...)`. Se houver erro, a ação é rejeitada.
11. Se ambas as validações passarem, o backend prossegue para as validações de jogo (range, linha de visão, cooldown, etc.).
12. Somente após todas as validações, o backend calcula e aplica o dano/efeitos.
13. O backend envia os resultados (eventos de dano, etc.) de forma segura para os clientes envolvidos.

## 7. Fluxo runtime futuro recomendado para cast/skill hostil

1.  O cliente envia uma intenção de usar uma skill hostil (ex: `CastSkillRequest{SkillID: "...", TargetID: "..."}`).
2.  O backend valida a sessão e o personagem.
3.  O backend valida se o personagem possui a skill, se o cooldown está pronto e se tem recursos (mana) suficientes.
4.  O backend identifica a skill como hostil a partir de seus metadados.
5.  O backend identifica o(s) alvo(s) ou a área de efeito.
6.  Para cada alvo, o backend valida a `ZoneType` e o `PvP Level Gate` (se aplicável), conforme o fluxo de ataque direto.
7.  O backend deve rejeitar a ação **antes** de consumir o recurso (mana), a menos que o design do jogo especifique que o recurso é consumido mesmo em falha (decisão a ser tomada).
8.  Somente após todas as validações, o backend consome o recurso e aplica os efeitos da skill.

## 8. Ordem de validação recomendada

O handler de qualquer ação hostil deve seguir esta ordem de validação para máxima segurança e eficiência:

1.  Sessão/conexão é válida.
2.  Personagem atacante existe e está em um estado que permite atacar.
3.  Personagem atacante pertence à sessão/conexão.
4.  Alvo existe e está em um estado que permite ser atacado.
5.  Carregar a `ZoneType` real da área.
6.  **Validar `rules.CanCombatOccurInZone`**.
7.  Se o alvo for um jogador, **validar `rules.CanEngageOpenPvP`**.
8.  Validar mecânicas de jogo (range, linha de visão, cooldown, recursos).
9.  Calcular dano/efeitos.
10. Aplicar a mutação de estado (reduzir HP, etc.) em uma operação segura.
11. Emitir eventos para os clientes.

## 9. Regras de rejeição obrigatórias

Uma ação hostil deve falhar e ser interrompida imediatamente se:

- A sessão do jogador for inválida.
- O atacante ou o alvo não existirem ou não estiverem em um estado válido.
- O nível do atacante for menor que 10 em um contexto PvP.
- O nível do alvo for menor que 10 em um contexto PvP.
- A `ZoneType` for inválida.
- A `ZoneType` for `safe`.
- O cliente tentar enviar um nível, `ZoneType` ou qualquer outra flag de permissão falsa (esses dados devem ser ignorados).
- A tentativa de aplicar dano ocorrer antes da validação completa das regras.

## 10. Impacto futuro em combat runtime

- O pacote `backend/pkg/combat` será o principal ponto de integração.
- Ele deverá invocar as funções de validação do Rule Registry no início de qualquer processamento de ação hostil.
- As fórmulas de dano e a aplicação de efeitos só devem ser executadas se as validações de PvP e Safe Zone passarem.

## 11. Impacto futuro em pvp runtime

- O pacote `backend/pkg/pvp` pode ser criado para centralizar a lógica específica de PvP.
- As regras de PvP Level Gate devem ser aplicadas a **todas** as formas de hostilidade entre jogadores, incluindo ataques diretos, magias de alvo único, dano ao longo do tempo (DoT), dano em área (AoE), armadilhas (traps) e dano de summons.
- O multiplicador de dano PvP (ex: 70%) só deve ser aplicado **após** a validação de que a ação PvP é permitida.

## 12. Impacto futuro em map/chunk/region system

- O sistema de mapa do backend deverá carregar metadados para cada região ou chunk, incluindo a `ZoneType`.
- Cidades e postos avançados serão definidos como `ZoneTypeSafe`.
- Áreas de caça (hunts) e masmorras (dungeons) serão `ZoneTypeCombat` ou `ZoneTypeNeutral`.
- A `ZoneType` é uma informação autoritativa do servidor; o cliente pode recebê-la para exibir na UI, mas nunca para tomar decisões.

## 13. Impacto futuro em protocolo/API

- As requisições de ataque/cast do cliente **não devem** conter o nível do atacante ou a `ZoneType` como dados confiáveis.
- O backend deve ter respostas de erro específicas para `ErrCombatBlockedInSafeZone` e `ErrPvPAttackerLevelTooLow`/`ErrPvPTargetLevelTooLow`, para que o cliente possa dar feedback claro ao jogador.
- Qualquer mudança no protocolo deve ser documentada previamente.

## 14. Impacto futuro no client

- O cliente pode usar a informação de `ZoneType` recebida do servidor para exibir um aviso na UI (ex: "Zona Segura").
- O cliente pode exibir um aviso como "PvP será habilitado no nível 10".
- O cliente deve tratar as respostas de erro do backend e impedir que o jogador continue tentando a ação bloqueada.
- O cliente **nunca** deve aplicar dano ou efeitos de forma local e autoritativa. Animações podem ser preditivas, mas a confirmação do resultado vem do servidor.

## 15. Riscos que o plano previne

- **Player Killing de Baixo Nível (PKing):** Impede que jogadores matem ou sejam mortos em PvP antes do nível 10.
- **Combate em Cidades:** Impede qualquer forma de combate em áreas designadas como seguras.
- **Exploits de Cliente:** Impede que um cliente modificado ignore as restrições de zona ou nível.
- **Cálculo Desnecessário:** Economiza recursos do servidor ao bloquear ações inválidas antes de executar cálculos complexos de combate.
- **Consistência de Regras:** Garante que todas as formas de hostilidade (AoE, DoT, etc.) respeitem as mesmas regras de zona e PvP.

## 16. Testes futuros recomendados

A futura implementação deve incluir testes de integração que cubram:

- Ataque de um jogador nível 9 contra um nível 10 (deve ser rejeitado).
- Ataque de um jogador nível 10 contra um nível 9 (deve ser rejeitado).
- Ataque entre dois jogadores de nível 10 em uma `ZoneTypeCombat` (deve ser aceito).
- Ataque entre dois jogadores de nível 10 em uma `ZoneTypeSafe` (deve ser rejeitado).
- Ataque de um jogador contra um monstro em uma `ZoneTypeSafe` (deve ser rejeitado).
- Ataque em uma `ZoneTypeCombat` e `ZoneTypeNeutral` (deve ser aceito).
- Tentativa de ataque em uma `ZoneType` inválida (deve ser rejeitada).
- Validação de que um cliente enviando um nível ou `ZoneType` falsos tem esses dados ignorados pelo servidor.
- Validação de que a fórmula de dano não é executada se a validação de PvP ou Safe Zone falhar.
- Testes futuros para AoE, DoT e summons devem reutilizar a mesma lógica de validação.

## 17. Critérios de aceite para futura implementação

A futura task de implementação será considerada concluída quando:

- O código do backend compilar com sucesso.
- Todos os testes unitários e de integração passarem.
- Nenhuma ação hostil (ataque, cast, etc.) aplicar dano ou efeitos antes de passar pelas validações de PvP e Safe Zone.
- A regra de Safe Zone bloquear efetivamente o combate no backend.
- A regra de PvP Level Gate bloquear efetivamente o combate envolvendo jogadores abaixo do nível 10.
- O cliente de debug atual não quebrar.
- O protocolo não for alterado sem um documento de planejamento aprovado.
- Nenhum dado sensível (tokens, senhas) for exposto em logs.

## 18. O que esta task NÃO faz

- Não cria código Go, C# ou qualquer outra linguagem.
- Não altera o `backend/pkg/combat` ou `backend/pkg/pvp`.
- Não implementa o sistema de mapa/chunks.
- Não altera o banco de dados, o protocolo de rede ou o cliente.
- Não altera o comportamento runtime do jogo.

## 19. Status

Este documento é o plano técnico oficial e preparatório para a futura integração dos sistemas de PvP e Safe Zone. Ele garante que a implementação será segura, robusta e alinhada com a arquitetura server-authoritative do projeto.