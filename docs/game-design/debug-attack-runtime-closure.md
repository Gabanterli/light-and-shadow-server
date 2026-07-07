# Fechamento Runtime da Validação de Ataque de Debug

## 1. Purpose

Este documento formaliza o fechamento da validação técnica do fluxo de envio de ataque a partir do botão de debug `Attack Orc_Elite`. O objetivo era garantir que o cliente, após a correção de um bug de cache de sessão, pudesse construir e enviar com sucesso um pacote de ataque (opcode 3000) para o backend.

## 2. Runtime Scenario

- **Commit de Referência:** `bc6e597` (Fix debug attack selected character cache)
- **Cena:** `DebugWorldEntryScene`
- **Controller:** `scripts/DebugWorldEntryController.cs`
- **Login:** `default_user` / `test123`
- **Personagem Selecionado:** `Gabriela`
- **Ação:** O botão `Attack Orc_Elite` foi clicado na UI de debug.

## 3. Evidence

Após a correção do cache do nome do personagem, o cliente conseguiu enviar a requisição de ataque.

- **Pacote Enviado:** `CS_ATTACK_REQUEST` (Opcode 3000)
- **Target ID:** `Orc_Elite`
- **WeaponType:** `debug_sword`

O log do Gateway Server confirmou o recebimento do pacote:

```
Received packet opcode=3000 size=32 seq=4.
```

## 4. Result

O cliente **não recebeu** um pacote `SC_DAMAGE_EVENT` (opcode 3002) como resposta. A ausência deste evento não foi uma falha, mas sim o resultado de uma rejeição autoritativa pelo backend.

## 5. Authoritative Rejection

O backend/Gateway combat handler recebeu o pacote CS_ATTACK_REQUEST e rejeitou a ação de forma autoritativa devido a uma falha na validação de alcance.

- **Motivo da Rejeição:** O alvo estava fora do alcance da arma utilizada.
- **Detalhes do Log do Backend:** `alvo fora de alcance para Espada. Distância: 2.83m, Alcance: 1.00m.`

Este comportamento é considerado **correto e esperado**, pois demonstra que a validação de regras de combate no servidor está funcionando e precede a execução de qualquer cálculo de dano.

## 6. What This Validates

- A correção no cache do nome do personagem selecionado (`_selectedCharacterNameForWorldEntry`) no `DebugWorldEntryController` foi bem-sucedida.
- O cliente é capaz de construir e enviar um pacote `CS_ATTACK_REQUEST` válido.
- O Gateway Server recebeu o pacote de combate opcode 3000 e o submeteu à validação autoritativa de combate.
- A camada autoritativa do backend aplicou a validação de alcance antes de qualquer cálculo de dano.

## 7. What This Does Not Validate Yet

- O ciclo completo de combate, incluindo o cálculo de dano bem-sucedido e o recebimento de um evento `SC_DAMAGE_EVENT` (3002).
- O recebimento de um evento `SC_TARGET_DEAD` (3003).
- A interface de combate final do jogo.
- Validações preditivas de alcance no lado do cliente.

## 8. Follow-up Technical Debt

O botão `Attack Orc_Elite` é uma implementação de debug com um alvo fixo. Para testes futuros, será necessário criar uma interface mais flexível, como um `LineEdit` para inserir dinamicamente o `TargetID`.

## 9. Closure Status

**Fechado.**

A validação técnica do envio do pacote de ataque de debug está concluída. O problema de cache do cliente foi resolvido (commit `bc6e597`) e o fluxo de comunicação até a camada de validação de regras do backend foi confirmado como funcional.
