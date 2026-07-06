# Skill Progression Runtime Integration Plan

## 1. Objetivo

Este documento planeja a futura integração da regra de negócio de **progressão de skills** no ambiente de execução (runtime) do backend. Ele detalha como a função pura `rules.CanGrantSkillProgression` será o ponto central de validação para qualquer ganho de experiência em skills, garantindo um sistema seguro e à prova de exploits.

## 2. Estado atual

- As regras puras de elegibilidade para progressão de skill já existem e estão testadas em `backend/pkg/gamedata/rules/skill_progression.go`.
- Os testes unitários que validam a lógica da regra em isolamento já existem.
- **Nenhuma integração runtime foi feita ainda.** Nenhum sistema de jogo (combate, profissões, etc.) invoca a função `CanGrantSkillProgression`.
- Esta task é exclusivamente de documentação e não altera o comportamento do jogo.

## 3. Princípio server-authoritative

A integração deve seguir rigorosamente a arquitetura server-authoritative:

- O cliente **nunca** envia uma mensagem como "ganhei 10 de XP na skill de espada".
- O cliente apenas envia **intenções de ação**: "quero atacar o alvo X", "quero minerar o nó Y", "quero usar o manequim de treino Z".
- O backend processa a ação, valida sua legitimidade (o alvo existe? o jogador está perto? o cooldown acabou?) e, somente se a ação for bem-sucedida e considerada elegível, o backend **decide** se aquela ação gera progresso de skill.

## 4. Fontes oficiais de progressão

A progressão de skill só pode ser originada por uma das seguintes fontes, que devem ser validadas pelo backend:

- `combat_use`: Uso de uma skill em um contexto de combate válido (ex: acertar um inimigo).
- `training_use`: Uso de uma skill em um contexto de treino autorizado (ex: usar um manequim de treino em uma cidade).
- `profession_use`: Uso de uma habilidade de profissão (ex: minerar, coletar, forjar um item).

## 5. Fluxo futuro para `combat_use`

O backend só deve considerar conceder progressão de skill de combate após uma cadeia completa de validações:

1.  A sessão do jogador é real e o personagem está online.
2.  A ação de combate (ataque, cast) é uma ação real do servidor, não uma simulação do cliente.
3.  O alvo é válido (existe, está ao alcance, na linha de visão, etc.).
4.  As regras de combate foram permitidas (ex: não está em uma safe zone, o PvP level gate foi respeitado).
5.  O dano/efeito foi processado com sucesso pelo servidor.
6.  Somente após tudo isso, o backend prepara uma `SkillProgressionEligibilityRequest` com `Source: "combat_use"`, `ActionValidatedByBackend: true`, `ContextValidatedByBackend: true` e os dados de tempo e diminishing returns.
7.  A chamada a `CanGrantSkillProgression` determina se o ganho é finalmente permitido.

## 6. Fluxo futuro para `training_use`

O treino de skills (ex: em manequins) deve ser rigorosamente validado para evitar farming AFK/macro:

1.  O backend valida que o personagem está fisicamente próximo a uma estação/NPC/área de treino autorizada.
2.  A ação de "treinar" é validada para garantir que não é uma simulação do cliente (ex: o personagem está realmente interagindo com o objeto de treino).
3.  O sistema verifica o intervalo mínimo desde o último ganho.
4.  O sistema anti-macro/diminishing returns é consultado.
5.  Regras futuras podem impor limites de tempo ou de ganhos por sessão de treino.

## 7. Fluxo futuro para `profession_use`

A progressão em profissões (coleta, craft) deve estar atrelada a resultados econômicos reais:

1.  O backend valida que o recurso/nó de coleta (ex: veio de minério) ou a receita de craft existem e estão disponíveis.
2.  Validações de jogo como ferramenta equipada, distância do nó e cooldowns são verificadas.
3.  A ação de coleta/craft é processada pelo servidor e resulta em um item real sendo gerado ou modificado.
4.  Somente após a conclusão bem-sucedida da ação, o backend considera conceder a progressão, validando o intervalo mínimo e outras regras para evitar a geração de progresso sem impacto econômico real.

## 8. Ordem de validação recomendada

O handler de qualquer sistema que possa gerar progresso de skill deve seguir esta ordem:

1.  Validar a sessão e carregar o estado autoritativo do personagem.
2.  Processar a ação do jogador (ataque, craft, etc.) e validar seu sucesso no contexto do jogo.
3.  Se a ação foi bem-sucedida, construir a `SkillProgressionEligibilityRequest`.
4.  Validar a `Source` da progressão.
5.  Confirmar que a `ActionValidatedByBackend` e `ContextValidatedByBackend` são `true`.
6.  Validar o `MillisecondsSinceLastGain` contra o `SkillProgressionMinimumIntervalMilliseconds`.
7.  Validar se `DiminishingReturnsBlocked` é `false`.
8.  Se `CanGrantSkillProgression` retornar `nil`, calcular o ganho de progresso.
9.  Persistir o novo valor da skill de forma segura (transacional/idempotente).
10. Emitir um evento para o cliente informando sobre o ganho.

## 9. Regras de rejeição obrigatórias

A progressão de skill deve ser **bloqueada** se:

- A `Source` for inválida (ex: "client_grant").
- A ação principal (ataque, craft) não foi validada pelo backend.
- O contexto da ação foi inválido (ex: atacar o ar, minerar um nó que não existe).
- O intervalo desde o último ganho for menor que 1000ms.
- O sistema de diminishing returns/anti-macro estiver bloqueando o ganho.
- O cliente tentar enviar um ganho de XP ou um novo nível de skill diretamente.
- Uma macro ou spam for detectado pela frequência das ações.

## 10. Impacto futuro em combat runtime

- O pacote `backend/pkg/combat` deverá, após aplicar dano ou um efeito de skill com sucesso, invocar o sistema de progressão para tentar conceder XP para a skill ou arma utilizada.

## 11. Impacto futuro em profession/economy runtime

- Os pacotes `backend/pkg/professions` e `backend/pkg/economy` deverão, após uma coleta ou craft bem-sucedido, invocar o sistema de progressão para a respectiva skill de profissão.

## 12. Impacto futuro em persistence/database

- A persistência dos níveis de skill deve ser segura. Para evitar ganhos duplicados em caso de falha de rede, a operação pode precisar ser idempotente. Para evitar perda de progresso, pode ser agrupada em transações com outras mutações de estado.

## 13. Impacto futuro em anti-macro/anti-bot

A `SkillProgressionEligibilityRequest` é a base para um sistema anti-macro. Sinais de alerta que podem ser usados para setar `DiminishingReturnsBlocked = true` incluem:

- Frequência de ações impossível para um humano.
- Repetição de ações com timing perfeito por longos períodos.
- Ações realizadas fora de um contexto válido (ex: tentar atacar repetidamente um alvo que já morreu).
- Ganhos de skill em intervalos consistentemente próximos do limite mínimo de 1000ms.

## 14. Impacto futuro no client

- O cliente é um receptor passivo de informações de progresso. Ele pode receber um evento do servidor como `SkillProgressed{SkillID: "sword", NewXP: 1234, NewLevel: 15}`.
- Com base nesse evento, o cliente atualiza a UI (barra de XP, nível da skill).
- O cliente **não** calcula o ganho nem confirma a progressão.

## 15. Testes futuros recomendados

A futura implementação deve incluir testes de integração que cubram:

- Um `combat_use` válido (acertar um inimigo) concede progresso.
- Uma tentativa de ganho com `source` inválida é rejeitada.
- Uma ação não validada (ex: errar um ataque) não concede progresso.
- Um contexto não validado (ex: atacar o ar) não concede progresso.
- Duas ações válidas em menos de 1000ms só concedem progresso na primeira.
- Um ganho bloqueado por `DiminishingReturnsBlocked` é rejeitado.
- Uma tentativa de ganho via `profession_use` sem um nó de recurso real é rejeitada.

## 16. Critérios de aceite para futura implementação

A futura task de implementação será considerada concluída quando:

- O código do backend compilar com sucesso.
- Todos os testes unitários e de integração passarem.
- Nenhuma progressão de skill puder ser originada ou validada pelo cliente.
- Todas as fontes de ganho (`combat`, `training`, `profession`) utilizarem a função `rules.CanGrantSkillProgression` como ponto de verificação final.
- A persistência dos níveis de skill for segura contra race conditions e falhas.
- Os eventos enviados ao cliente forem apenas o **resultado** de uma decisão do servidor.

## 17. O que esta task NÃO faz

- Não implementa código Go, C# ou qualquer outra linguagem.
- Não altera o `backend/pkg/combat`, `professions` ou `economy`.
- Não altera o banco de dados, o protocolo de rede ou o cliente.
- Não cria endpoints ou handlers de rede.
- Não altera as regras puras já existentes no Rule Registry.

## 18. Status

Este documento é o plano técnico oficial e preparatório para a futura integração do sistema de progressão de skills. Ele garante que a implementação será segura, robusta e alinhada com a arquitetura server-authoritative do projeto, prevenindo exploits comuns em MMOs.