# R1 Technical Validation Phase Closure

Data: 2026-07-07

## Contexto

Este documento fecha a fase R1 de validação técnica do projeto Light and Shadow.

A fase teve como objetivo validar o primeiro fluxo técnico jogável server-authoritative do MMORPG, conectando autenticação, seleção de personagem, entrada no mundo debug, movimentação, combate, morte de alvo, estado visual, loot debug, sincronização de inventário, retry do alvo e correção de movimento autoritativo.

Esta fase não representa gameplay final. Ela fecha o bootstrap técnico validado em runtime.

## Escopo validado

O fluxo técnico validado foi:

    login -> character list -> character select -> world entry -> movement -> attack -> damage -> death -> visual dead state -> debug loot -> inventory sync -> retry -> revive -> repeat kill -> movement correction

## Ambiente validado

- Login: `default_user`
- Senha: `test123`
- Personagem: `Gabriela`
- Cena: `DebugWorldEntryScene`
- Backend: Docker Compose
- Gateway TCP: `8080`
- Auth: `8081`
- World: `8082`
- Client: Godot 4 C# / .NET 8

## Protocolos validados

- Login: `1002`
- Character list: `1004`
- Character select: `1006`
- Move request: `2004`
- Move confirm: `2005`
- Chunk data: `2006`
- Attack request: `3000`
- Damage event: `3002`
- Target dead: `3003`
- Inventory sync: `4001`

## Commits de fechamento da fase

- `c3d05f0 Mark debug target dead visually`
- `304c87c Add debug loot after Orc Elite death`
- `8455de4 Document debug loot validation closure`
- `81fba2a Add debug Orc Elite retry flow`
- `b213399 Document Orc Elite retry validation closure`
- `74a2cf1 Apply authoritative movement corrections`
- `1627243 Document movement correction validation closure`

## Validações fechadas

### Login e entrada no mundo

Validado:

- login `default_user / test123`;
- seleção da personagem `Gabriela`;
- entrada na `DebugWorldEntryScene`;
- carregamento de inventário;
- envio de posição inicial;
- streaming inicial de chunks;
- autosave de estado ativo.

### Movimento server-authoritative

Validado:

- envio de `CS_MOVE_REQUEST`;
- resposta `SC_MOVE_CONFIRM`;
- aplicação da posição confirmada pelo servidor;
- correção de rubberband aplicada no client mesmo quando `Success = false`;
- ausência de loop recorrente de warnings `Client out of sync/rubberbanded` após correção;
- persistência de posição atualizada.

### Combate debug

Validado:

- botão `Attack Orc_Elite`;
- envio de `CS_ATTACK_REQUEST`;
- gateway recebendo opcode `3000`;
- backend processando dano;
- client recebendo `SC_DAMAGE_EVENT` opcode `3002`;
- morte do `Orc_Elite` no backend;
- envio de `SC_TARGET_DEAD` opcode `3003`.

### Estado visual de morte

Validado:

- client recebe `3003`;
- `Last Action Result` mostra target dead recebido;
- `Orc_Elite` muda visualmente de vermelho para cinza;
- estado visual morto é mantido no debug world view.

### Loot debug e inventário

Validado:

- backend concede loot debug após morte do `Orc_Elite`;
- item concedido: `sword_t1_rusty x1`;
- backend envia `SC_INVENTORY_SYNC` opcode `4001`;
- `items_count` aumenta;
- autosave persiste estado atualizado.

### Retry/respawn debug do Orc_Elite

Validado:

- após morte, clicar novamente em `Attack Orc_Elite` aciona retry;
- backend detecta `Health <= 0`;
- backend chama revive server-authoritative;
- client reseta estado visual local;
- `Orc_Elite` volta visualmente para vermelho;
- novo ciclo de ataque e morte funciona;
- novo `3003` e novo `4001` são enviados;
- debug loot é concedido novamente;
- autosave persiste nova versão.

## Fechamento técnico

A fase R1 confirma que o projeto já possui um hunt loop técnico mínimo validado:

    attack -> damage -> death -> visual dead state -> debug loot -> inventory sync -> retry -> repeat

Também confirma que o movimento debug respeita autoridade do servidor:

    move request -> server validation -> move confirm -> client applies authoritative position

Do ponto de vista arquitetural MMO, isso prova integração mínima entre:

- client Godot;
- gateway TCP;
- auth server;
- world server;
- PostgreSQL;
- Redis;
- sistema de movimento;
- sistema de combate;
- inventário;
- autosave;
- protocolo binário;
- debug UI.

## O que ainda NÃO é final

Esta fase não fecha sistemas finais de gameplay.

Ainda são debug-only:

- `Orc_Elite` fixo;
- botão `Attack Orc_Elite`;
- loot fixo `sword_t1_rusty`;
- revive/retry técnico;
- visual marker vermelho/cinza;
- debug world view;
- fluxo de inventário usado apenas para validação técnica.

Ainda não são finais:

- loot table real;
- corpse/container loot;
- ownership de corpse;
- party loot;
- split de loot;
- economia player-driven;
- auditoria anti-dupe;
- respawn real por spawn point;
- timer de respawn;
- AI de criatura;
- pathfinding final;
- click-to-move final;
- predição/reconciliation final;
- UI real de combate;
- UI real de loot;
- sistema de hunt/progressão final.

## Riscos arquiteturais pendentes

### Economia e anti-dupe

O loot debug pode gerar item repetido a cada morte/retry. Isso é aceitável para validação técnica, mas não pode virar regra final.

Risco final:

- farm infinito;
- duplicação de itens;
- inflação econômica;
- inconsistência entre inventário e autosave.

### Respawn

O revive atual é técnico e acionado por retry de ataque. Não representa respawn MMO real.

Risco final:

- respawn sem timer;
- spawn contestável sem ownership;
- abuso de spawn por player;
- inconsistência em múltiplos players no mesmo AOI.

### Movimento

A correção atual fecha drift no debug flow, mas não substitui sistema final de movimentação.

Risco final:

- input spam;
- colisão client/server divergente;
- pathfinding local divergente;
- exploração de cooldown;
- problemas de interpolação em rede.

### Combate

O combate atual valida dano/morte/eventos, mas ainda não fecha sistema final de PvE.

Risco final:

- target state compartilhado entre players;
- morte duplicada;
- evento de loot duplicado;
- falta de AI/aggro real para criatura;
- falta de regras de contribuição.

## Checklist de entrada da próxima fase

A próxima fase deve começar somente com estas premissas claras:

- manter debug-only separado do gameplay final;
- não transformar loot fixo em sistema final;
- não transformar retry técnico em respawn final;
- não misturar UI debug com UI final;
- cada sistema final deve ter autoridade server-side;
- todo loot final precisa de auditoria anti-dupe;
- todo respawn final precisa de spawn state server-side;
- toda criatura real precisa de estado próprio por spawn;
- todo fluxo econômico precisa de logs/auditoria;
- toda alteração deve seguir uma task por commit.

## Próximas frentes recomendadas

### Frente 1 — Respawn real de criatura

Criar sistema server-side de spawn state com:

- spawn ID;
- creature ID;
- posição;
- alive/dead state;
- death timestamp;
- respawn delay;
- AOI notification.

### Frente 2 — Loot table real

Criar drop table server-side com:

- item ID;
- quantidade;
- chance;
- raridade;
- regra de elegibilidade;
- auditoria.

### Frente 3 — Corpse/container loot

Criar corpse temporário com:

- ownership;
- timeout;
- container inventory;
- permission checks;
- anti-dupe.

### Frente 4 — UI real de loot

Criar UI que consome inventário/corpse de forma final, separada da debug UI.

## Recomendação de corte da fase

A fase R1 está tecnicamente encerrada quando este documento estiver commitado em `origin/master`.

A partir daqui, qualquer nova implementação deve sair do modo debug-only e começar a migrar para sistemas finais com escopo controlado.
