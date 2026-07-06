# Economy and Loot Runtime Integration Plan

## 1. Objetivo

Este documento planeja a futura integração das regras de negócio de **autoridade sobre a economia e o loot** no ambiente de execução (runtime) do backend. Ele detalha como a função pura `rules.CanApplyItemMutation` será o ponto de verificação central e obrigatório para toda e qualquer criação, transferência, movimentação ou remoção de itens no jogo, garantindo um sistema econômico seguro, auditável e à prova de exploits.

## 2. Estado atual

- As regras puras de autoridade sobre a economia já existem e estão testadas em `backend/pkg/gamedata/rules/economy_loot_authority.go`.
- Os testes unitários que validam a lógica da regra em isolamento já existem.
- **Nenhuma integração runtime foi feita ainda.** Nenhum sistema de jogo (inventário, loot, trade, etc.) invoca a função `CanApplyItemMutation`.
- Esta task é exclusivamente de documentação e não altera o comportamento do jogo.

## 3. Princípio server-authoritative

A integração deve seguir rigorosamente a arquitetura server-authoritative, onde o cliente é um terminal não confiável:

- O cliente **nunca** cria, confirma, move, duplica, destrói ou transfere um item de forma autoritativa.
- O cliente apenas envia **intenções**: "quero pegar o item X do loot", "quero mover o item Y para o slot Z", "quero oferecer o item A na troca", "quero receber minha recompensa da quest B".
- O backend valida cada intenção contra o estado real do jogo e as regras canônicas, decidindo se a mutação do item é permitida.

## 4. Fontes oficiais de mutação

Toda mutação de item deve ser categorizada por uma das seguintes fontes oficiais, conforme definido em `economy_loot_authority.go`:

- `ItemMutationSourceLootDrop`: Um item sendo gerado como loot de uma criatura ou container.
- `ItemMutationSourcePlayerTrade`: Um item sendo transferido entre dois jogadores.
- `ItemMutationSourceQuestReward`: Um item sendo concedido como recompensa de uma quest.
- `ItemMutationSourceInventoryMove`: Um item sendo movido entre slots ou containers do inventário de um jogador.

## 5. Fluxo futuro para `loot_drop`

O backend só deve permitir que um jogador colete um item de loot quando:

1.  A criatura/container foi destruída no servidor, gerando um evento de loot.
2.  A tabela de loot foi resolvida pelo backend, e um item foi gerado com um ID único.
3.  O jogador tem permissão para coletar o loot (ex: dono do drop, grupo, etc.).
4.  O jogador está a uma distância válida do cadáver/container.
5.  O item específico ainda não foi coletado por ninguém.
6.  A operação de mover o item do container de loot para o inventário do jogador é validada com `Source: ItemMutationSourceLootDrop`, `BackendGeneratedItem: true`, e `AuditLogged: true`.
7.  A persistência da remoção do item do container e da adição ao inventário do jogador é atômica para evitar coleta duplicada.

## 6. Fluxo futuro para `quest_reward`

O backend só deve conceder um item como recompensa de quest quando:

1.  O sistema de quests do servidor valida que o personagem cumpriu todos os requisitos da quest.
2.  A recompensa (item ID e quantidade) é lida da definição oficial da quest no servidor.
3.  O item é gerado pelo backend.
4.  O sistema valida que a recompensa (se for única) ainda não foi entregue para este personagem.
5.  A operação de adicionar o item ao inventário é validada com `Source: ItemMutationSourceQuestReward`, `BackendGeneratedItem: true`, e `AuditLogged: true`.
6.  A entrega da recompensa e a marcação da quest como concluída/recompensada são feitas em uma única transação no banco de dados.

## 7. Fluxo futuro para `player_trade`

A troca entre jogadores é um ponto crítico para exploits de duplicação e deve ser rigorosamente transacional:

1.  Ambos os jogadores existem e estão em um estado que permite a troca.
2.  Os itens oferecidos por cada jogador existem em seus respectivos inventários e são bloqueados/reservados para a troca.
3.  Ambos os jogadores devem confirmar a mesma versão final da janela de troca. Qualquer alteração nos itens ou quantidades por uma das partes deve invalidar a confirmação de ambos.
4.  O commit final da troca deve ser uma operação **atômica** no backend, validada com `Source: ItemMutationSourcePlayerTrade`, `Transactional: true`, e `AuditLogged: true`.
5.  Se qualquer parte da troca falhar (ex: um jogador desconectar, inventário cheio), a transação inteira deve sofrer rollback, retornando os itens bloqueados aos seus donos originais, sem perdas ou duplicações.
6.  Um log de auditoria detalhado deve registrar a origem, destino, itens e quantidades de cada troca bem-sucedida.

## 8. Fluxo futuro para `inventory_move`

Mesmo uma simples movimentação de item no inventário deve ser validada pelo servidor:

1.  O backend valida que o personagem é o dono do item e dos containers de origem e destino.
2.  O item existe no slot de origem no estado autoritativo do servidor.
3.  O slot de destino é válido e compatível com o item.
4.  As regras de empilhamento (stacking) são respeitadas.
5.  O item não está bloqueado por outro sistema (trade, craft, etc.).
6.  A operação é validada com `Source: ItemMutationSourceInventoryMove` e `AuditLogged: true` (especialmente para movimentos que podem ter impacto, como mover de/para um banco compartilhado).

## 9. Ordem de validação recomendada

O handler de qualquer sistema que cause uma mutação de item deve seguir esta ordem:

1.  Validar a sessão e carregar o estado autoritativo do personagem e do item.
2.  Identificar a `ItemMutationSource` oficial.
3.  Validar a intenção do jogador e o contexto da ação (ex: o loot existe? a quest foi concluída?).
4.  Construir a `ItemMutationAuthorityRequest` com os flags corretos (`BackendValidated`, `BackendGeneratedItem`, `Transactional`, `AuditLogged`).
5.  Chamar `err := rules.CanApplyItemMutation(request)`.
6.  Se houver erro, rejeitar a operação.
7.  Se a validação passar, executar a mutação de estado de forma segura (transacional/idempotente).
8.  Persistir o novo estado.
9.  Emitir um evento para o cliente informando o resultado **somente após o sucesso da persistência**.

## 10. Regras de rejeição obrigatórias

Uma mutação de item deve ser **bloqueada** se:

- A `Source` for inválida, como `client_create`, `admin_spawn`, `debug_grant` ou `offline_macro`.
- O cliente tentar criar um item (`ClientCreatedItem: true`).
- O cliente tentar definir o resultado de uma mutação (ex: "agora eu tenho 10 poções").
- O item de origem não existir no estado autoritativo do servidor.
- O jogador não for o dono do item/inventário.
- O loot já tiver sido coletado.
- A recompensa da quest já tiver sido entregue.
- Uma troca (`player_trade`) não for processada de forma transacional.
- Uma mutação obrigatória de auditoria não for registrada.

## 11. Impacto futuro em inventory runtime

- O pacote `backend/pkg/inventory` deverá ser o principal ponto de integração, centralizando a lógica de adição, remoção e movimentação de itens, e invocando `CanApplyItemMutation` antes de cada operação.

## 12. Impacto futuro em loot runtime

- O sistema de loot, ao ser acionado pela morte de uma criatura, deverá gerar os itens no backend e associá-los a um container de loot. A coleta por um jogador será uma operação de `inventory_move` validada com a `Source` `loot_drop`.

## 13. Impacto futuro em trade/economy runtime

- O pacote `backend/pkg/economy` deverá conter a implementação do sistema de troca transacional, garantindo a atomicidade e chamando `CanApplyItemMutation` com `Transactional: true`.

## 14. Impacto futuro em quest runtime

- O sistema de quests, ao validar a conclusão de uma tarefa, chamará o sistema de inventário para conceder o item, que por sua vez usará `CanApplyItemMutation` com a `Source` `quest_reward`.

## 15. Impacto futuro em persistence/database

- **Transações:** São essenciais para operações multi-passo como trocas e recompensas de quests.
- **Locks ou Optimistic Concurrency:** Mecanismos para prevenir que duas operações modifiquem o mesmo item/inventário simultaneamente.
- **Idempotência:** Requisições de mutação devem ser idempotentes sempre que possível para evitar duplicação em caso de reenvio por falha de rede.
- **Audit Log:** Uma tabela separada e durável para registrar todas as mutações de itens, garantindo rastreabilidade.

## 16. Impacto futuro em anti-dupe/anti-exploit

Este plano previne os exploits mais comuns em MMOs:

- **Dupe por reconexão/lag:** A validação server-side e a persistência transacional impedem que uma ação seja processada duas vezes.
- **Dupe por trade:** A exigência de atomicidade e confirmação mútua bloqueia a maioria dos exploits de troca.
- **Loot/Quest Reward Duplicado:** A verificação de estado (item já coletado, quest já recompensada) antes da mutação previne ganhos duplicados.
- **Item "Fantasma" (Ghost Item):** A autoridade do servidor garante que o estado do inventário é consistente, eliminando itens que só existem no cliente.
- **Criação de Itens pelo Cliente:** A regra `ErrClientCreatedItemRejected` é a defesa fundamental contra clientes modificados.

## 17. Impacto futuro no client

- O cliente é um "visualizador" do estado do inventário. Ele recebe atualizações do servidor e renderiza o resultado.
- O cliente não decide o sucesso de uma coleta, troca ou movimentação. Ele apenas envia a intenção e aguarda a resposta autoritativa do servidor.
- O cliente deve ser capaz de tratar respostas de erro (ex: "Inventário cheio", "Troca cancelada") e atualizar a UI de acordo.

## 18. Testes futuros recomendados

A futura implementação deve incluir testes de integração que cubram:

- Coletar um item de loot (deve funcionar uma vez, falhar na segunda).
- Tentar criar um item via `client_create` (deve ser rejeitado).
- Receber uma recompensa de quest (deve funcionar uma vez, falhar na segunda se for única).
- Realizar uma troca bem-sucedida.
- Cancelar uma troca no meio e verificar se os itens retornam corretamente.
- Simular uma falha de rede durante uma troca e verificar se não há duplicação ou perda de itens.
- Mover um item para um slot inválido (deve ser rejeitado).
- Tentar uma mutação sem o flag de auditoria obrigatório (deve ser rejeitada).

## 19. Critérios de aceite para futura implementação

A futura task de implementação será considerada concluída quando:

- O código do backend compilar com sucesso.
- Todos os testes unitários e de integração passarem.
- O cliente for comprovadamente incapaz de criar ou duplicar itens.
- Todas as fontes de mutação de item (`loot`, `trade`, `quest`, `move`) usarem `rules.CanApplyItemMutation` como ponto de verificação.
- A persistência de itens for transacional e segura contra race conditions.
- A troca entre jogadores for atômica.
- Logs de auditoria forem gerados para todas as mutações econômicas relevantes.

## 20. O que esta task NÃO faz

- Não implementa código Go, C# ou qualquer outra linguagem.
- Não altera os pacotes `backend/pkg/inventory`, `economy`, `loot`, `trade` ou `quest`.
- Não altera o banco de dados, o protocolo de rede ou o cliente.
- Não cria endpoints ou handlers de rede.
- Não altera as regras puras já existentes no Rule Registry.

## 21. Status

Este documento é o plano técnico oficial e preparatório para a futura integração dos sistemas de economia e loot. Ele garante que a implementação será segura, robusta e alinhada com a arquitetura server-authoritative, protegendo a integridade da economia do jogo.