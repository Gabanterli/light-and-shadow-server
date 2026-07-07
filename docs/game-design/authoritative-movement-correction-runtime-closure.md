# Authoritative Movement Correction Runtime Closure

Data: 2026-07-07

## Contexto

Este documento fecha a validação runtime da correção de movimento autoritativo aplicada após investigação dos warnings de rubberband na `DebugWorldEntryScene`.

A task validada foi:

    R1-M-A — Investigate debug movement rubberband warnings

O problema observado era uma sequência recorrente de warnings no backend:

    Authoritative movement validation failed (Client out of sync/rubberbanded)

## Causa identificada

O backend já retornava a posição autoritativa confirmada em `SC_MOVE_CONFIRM`, tanto em sucesso quanto em falha.

Porém, o client só aplicava `_currentConfirmedPos` quando `Success = true`.

Isso fazia o client continuar calculando novos movimentos a partir de uma posição local stale quando o servidor rejeitava algum movimento por cooldown, colisão, região, speedcheck ou estado autoritativo diferente.

## Correção aplicada

O client passou a aplicar sempre a posição confirmada pelo servidor em `SC_MOVE_CONFIRM`.

Quando `Success = true`, a posição é tratada como movimento aceito.

Quando `Success = false`, a posição é tratada como correção autoritativa de rubberband.

Também foi melhorado o log do backend para incluir:

- posição solicitada;
- posição confirmada;
- andar confirmado.

## Arquivos alterados na task relacionada

- `scripts/DebugWorldEntryController.cs`
- `backend/cmd/gateway/main.go`

## Commit relacionado

    74a2cf1 Apply authoritative movement corrections

## Ambiente validado

- Login: `default_user`
- Senha: `test123`
- Personagem: `Gabriela`
- Cena: `DebugWorldEntryScene`
- Movimento testado via teclado/botão
- Backend executado via Docker Compose
- Client Godot buildado com sucesso

## Protocolos envolvidos

- Move request: `2004` / `CS_MOVE_REQUEST`
- Move confirm: `2005` / `SC_MOVE_CONFIRM`
- Player update: `2001` / `SC_PLAYER_UPDATE`

## Resultado validado

Após a correção, múltiplos pacotes `CS_MOVE_REQUEST` foram processados sem repetição do loop de warnings:

    Authoritative movement validation failed (Client out of sync/rubberbanded)

Também foi confirmado que o estado persistiu corretamente:

    player: Gabriela
    x: 137
    y: 127
    new_version: 104

## Fechamento técnico

Esta validação confirma que o client debug agora respeita corretamente a autoridade do servidor para movimentação.

O fluxo corrigido é:

    move request -> server validation -> move confirm -> client applies confirmed position -> next move starts from authoritative state

Do ponto de vista MMO, isso reduz:

- drift client-side;
- rubberband em loop;
- falso positivo de speedhack;
- divergência entre marker visual e posição real do servidor;
- inconsistência entre movimento, combate e persistência.

## Limites desta validação

Esta task não altera as regras autoritativas de movimento.

Ela não muda:

- velocidade base;
- cooldown de movimento;
- colisão;
- validação de região;
- speedcheck;
- AOI;
- pathfinding final.

A correção é focada em sincronização client-side no debug world view.

## Pendências

- Input queue final de movimento
- Pathfinding final click-to-move
- Predição client-side real
- Interpolação visual
- Reconciliation avançado
- Separar debug movement UI do client gameplay final
- Logs estruturados por causa de rejeição: cooldown, colisão, região ou speedcheck
