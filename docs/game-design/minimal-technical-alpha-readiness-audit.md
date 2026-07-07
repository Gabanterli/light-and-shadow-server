# Auditoria de Prontidão para Alpha Técnico Mínimo

## 1. Purpose

Este documento audita o estado técnico atual do projeto Light and Shadow para identificar as capacidades já validadas em tempo de execução e os gaps bloqueantes que impedem a realização de um Alpha Técnico mínimo e jogável. O objetivo é criar um snapshot claro do progresso e definir as próximas tarefas críticas.

## 2. Current Confirmed Runtime Capabilities

- **Autenticação e Sessão:** O cliente consegue se conectar, autenticar com credenciais de debug (`default_user` / `test123`), receber a lista de personagens e selecionar um personagem (`Gabriela`).
- **Entrada no Mundo:** A cena de debug do mundo (`DebugWorldEntryScene`) é carregada com sucesso após a seleção do personagem, e a sessão do jogador é mantida.
- **Sincronização de Estado:** O cliente recebe e processa pacotes de sincronização de inventário (`SC_INVENTORY_SYNC` - 4001) e de chunks do mapa (`SC_CHUNK_DATA` - 2006).
- **Movimento:** O ciclo de movimento cliente-servidor é funcional, com persistência de dados.
- **Comunicação de Combate (Parcial):** O cliente consegue enviar uma intenção de ataque (`CS_ATTACK_REQUEST` - 3000) que é recebida e processada pelo backend.

## 3. Combat Runtime Status

- **Envio de Intenção:** O cliente Godot possui os codecs de combate e um botão de debug que envia com sucesso o pacote `CS_ATTACK_REQUEST` (opcode 3000) para o backend.
- **Recebimento no Gateway:** O Gateway Server confirma o recebimento do pacote de ataque do cliente.
- **Validação Autoritativa:** Em um teste documentado, o backend rejeitou autoritativamente um ataque contra o `Orc_Elite` devido à regra de alcance (`Distância: 2.83m, Alcance: 1.00m`). Isso valida que a camada de regras do servidor está funcionando.
- **Ciclo Incompleto:** O ciclo completo de dano **não foi validado**. Nenhum evento `SC_DAMAGE_EVENT` (3002) ou `SC_TARGET_DEAD` (3003) foi recebido pelo cliente, pois a única tentativa de ataque foi bloqueada por uma regra de jogo.
- **Estado:** Apenas a primeira metade do fluxo de combate (intenção do cliente -> validação do servidor) foi confirmada. A segunda metade (execução da ação -> feedback para o cliente) ainda não foi testada.

## 4. Movement Runtime Status

- **Ciclo Completo:** O fluxo de movimento está funcional e validado.
- **Ação do Cliente:** O jogador pode iniciar o movimento via teclado (WASD) ou botão de debug.
- **Confirmação do Servidor:** O backend recebe a requisição, valida a posição, persiste o novo estado no banco de dados (PostgreSQL) e envia uma confirmação (`SC_MOVE_CONFIRM` - 2005) de volta para o cliente.
- **Estado:** Funcional e estável para os propósitos de um alpha técnico.

## 5. Character/Auth Runtime Status

- **Fluxo Completo:** O login, listagem e seleção de personagens estão funcionais.
- **Estado:** Funcional e estável.

## 6. Inventory/Chunk Runtime Status

- **Sincronização Inicial:** O cliente recebe e processa os pacotes de inventário e chunks, exibindo informações básicas na UI de debug e no mapa de tiles.
- **Estado:** Funcional para a sincronização inicial. A manipulação de inventário (mover, usar, dropar itens) não foi implementada ou validada.

## 7. Blocking Gaps for Minimal Technical Alpha

Os seguintes gaps técnicos impedem que o estado atual seja considerado um "Alpha Técnico Mínimo Jogável":

- **Validação do Ciclo de Dano:** É necessário executar um cenário de combate onde o ataque **não** seja rejeitado, para validar o recebimento e processamento do evento `SC_DAMAGE_EVENT` (3002).
- **Validação de Morte de Alvo:** Consequentemente, é preciso validar o recebimento do evento `SC_TARGET_DEAD` (3003) após um alvo ser derrotado.
- **Representação de Inimigos:** Não há nenhuma representação visual (mesmo que um simples quadrado vermelho) ou de estado para os inimigos no cliente, tornando o combate impossível de ser testado de forma interativa.
- **Seleção de Alvo:** Falta um mecanismo, mesmo que de debug (ex: um `LineEdit` para digitar o ID do alvo), para permitir que o jogador ataque alvos diferentes do `Orc_Elite` fixo.
- **Cenário de Teste de Combate:** É preciso resolver o problema de posição/alcance para permitir um teste de ataque bem-sucedido. Isso pode envolver mover o jogador, o alvo, ou ambos para posições adjacentes.
- **Feedback de Rejeição:** O cliente atualmente não tem um feedback claro quando uma ação de combate é rejeitada pelo servidor (ex: "Fora de alcance", "Sem linha de visão").
- **Validação de Loot:** O ciclo de loot (derrotar um inimigo e receber itens/ouro) não foi implementado nem validado.
- **Visibilidade de Stats:** Stats básicos de combate (HP do alvo, etc.) não são visíveis no cliente, dificultando a avaliação do progresso do combate.

## 8. Non-Blocking Debug/UX Debts

- O botão de ataque possui um alvo fixo (`Orc_Elite`).
- A UI de debug é funcional, mas não representa a experiência final do jogador.
- O feedback de ações (movimento, ataque) é limitado a logs de texto.

## 9. Required Next Tasks

Para alcançar o Alpha Técnico Mínimo, as próximas tarefas devem focar em fechar os gaps bloqueantes:

1.  Implementar uma representação visual mínima para NPCs/inimigos no `DebugTileWorldView`.
2.  Adicionar um `LineEdit` na UI de debug para seleção dinâmica de alvo.
3.  Criar um cenário de teste (posicionando jogador e alvo) para garantir um ataque dentro do alcance.
4.  Executar o teste de ataque e validar o recebimento e log do `SC_DAMAGE_EVENT` (3002).
5.  Executar o teste até a morte do alvo e validar o recebimento e log do `SC_TARGET_DEAD` (3003).
6.  Implementar um handler de feedback mínimo no cliente para exibir mensagens de rejeição do servidor.

## 10. Minimal Technical Alpha Definition

Um "Alpha Técnico Mínimo" para este projeto será alcançado quando um testador interno puder executar as seguintes ações de ponta a ponta:

1.  Logar no jogo.
2.  Selecionar um personagem e entrar no mundo.
3.  Mover-se pelo mapa.
4.  Visualizar um inimigo no mapa.
5.  Selecionar esse inimigo como alvo.
6.  Atacá-lo repetidamente.
7.  Receber feedback visual/log de que o dano está sendo aplicado (`SC_DAMAGE_EVENT`).
8.  Ver o inimigo ser removido ou marcado como morto após sua derrota (`SC_TARGET_DEAD`).

## 11. Closure Status

**Aberto.**

Esta auditoria confirma que, embora fundações críticas (auth, movimento, sync) estejam estáveis, o ciclo de jogabilidade principal (combate) ainda está incompleto. O projeto não está pronto para um Alpha Técnico.
