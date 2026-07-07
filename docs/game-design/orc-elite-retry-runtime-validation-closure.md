# Orc Elite Retry Runtime Validation Closure

Data: 2026-07-07

## Contexto

Este documento fecha a validação runtime do fluxo debug de retry/respawn técnico do alvo `Orc_Elite` na `DebugWorldEntryScene`.

A task validada foi:

    R1-S-D — Add respawn/retry flow for Orc_Elite debug hunt loop

O objetivo foi permitir repetir o hunt loop técnico sem reiniciar a stack e sem recriar manualmente o estado do alvo.

## Ambiente validado

- Login: `default_user`
- Senha: `test123`
- Personagem: `Gabriela`
- Cena: `DebugWorldEntryScene`
- Alvo: `Orc_Elite`
- Botão usado: `Attack Orc_Elite`

## Commit relacionado

    81fba2a Add debug Orc Elite retry flow

## Fluxo validado

1. Player entrou na `DebugWorldEntryScene`.
2. Player atacou `Orc_Elite` até a morte.
3. Backend enviou `SC_TARGET_DEAD`.
4. Client recebeu o opcode de morte.
5. `Orc_Elite` ficou visualmente cinza.
6. Player clicou novamente em `Attack Orc_Elite`.
7. Backend detectou que `Orc_Elite` estava morto.
8. Backend reviveu o alvo de forma server-authoritative.
9. Client resetou o estado visual local do `Orc_Elite`.
10. `Orc_Elite` voltou visualmente para vermelho.
11. Player atacou novamente até a segunda morte.
12. Backend enviou novo `SC_TARGET_DEAD`.
13. Backend enviou novo `SC_INVENTORY_SYNC`.
14. Debug loot foi concedido novamente.
15. Autosave persistiu o estado atualizado.

## Protocolos envolvidos

- Attack request: `3000` / `CS_ATTACK_REQUEST`
- Target dead: `3003` / `SC_TARGET_DEAD`
- Inventory sync: `4001` / `SC_INVENTORY_SYNC`

## Resultado validado

Eventos confirmados em runtime:

    Debug Orc Elite revived for retry flow
    Target dead packet sent to client
    Sent inventory sync packet to client
    Debug loot granted after target death

Também foi validado:

    items_count: 8 -> 9
    autosave persistiu estado
    new_version: 103

## Fechamento técnico

Esta validação confirma que o hunt loop técnico pode ser repetido:

    attack -> damage -> death -> visual dead state -> debug loot -> inventory sync -> retry -> revive -> attack -> death -> inventory sync

Do ponto de vista arquitetural, o retry foi validado com autoridade do servidor:

- o client não revive o alvo sozinho;
- o client apenas reseta o estado visual após enviar novo ataque;
- o backend verifica `Health <= 0`;
- o backend chama `ReviveEntity`;
- o próximo ataque é processado contra o alvo revivido;
- a nova morte gera novamente `3003`;
- o inventário é sincronizado novamente com `4001`.

## Limites desta validação

Este fluxo ainda é debug-only.

Ele não representa o sistema final de respawn do MMORPG.

Ele também não representa o sistema final de loot, corpse, party loot ou economia.

O fato de `items_count` subir novamente após retry é esperado nesta validação técnica, mas esse comportamento não deve ser usado como regra final de farm, economia ou progressão.

## Pendências

- Sistema real de respawn por spawn point
- Timers reais de respawn
- Estado de criatura por região/chunk
- Loot table real
- Corpse/container loot
- Party loot
- Anti-dupe/auditoria econômica final
- UI real de loot
- Separar debug hunt loop do comportamento final de PvE
- Investigar warnings de movimento `Client out of sync/rubberbanded` em task própria
