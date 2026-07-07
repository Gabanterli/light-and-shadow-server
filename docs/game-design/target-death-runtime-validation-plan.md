# Plano de Validação Runtime da Morte de Alvo

## 1. Purpose

Este documento planeja o cenário técnico para a próxima fase de validação do sistema de combate. O objetivo é confirmar que, após um alvo ser derrotado, o cliente recebe e processa corretamente o pacote `SC_TARGET_DEAD` (opcode 3003), completando assim o ciclo de vida básico do combate (ataque -> dano -> morte).

## 2. Current Validated Combat State

- **Commit de Referência:** `02a614d` (Document debug combat damage runtime closure)
- **Envio de Ataque:** O cliente envia com sucesso o pacote `CS_ATTACK_REQUEST` (opcode 3000).
- **Validação de Alcance:** O backend valida autoritativamente o alcance do ataque, rejeitando-o quando necessário.
- **Evento de Dano:** Quando o ataque está dentro do alcance, o backend processa o dano e envia um pacote `SC_DAMAGE_EVENT` (opcode 3002), que o cliente recebe e decodifica com sucesso.

O ciclo de dano está funcional, mas o resultado final do combate (a morte do alvo) ainda não foi validado.

## 3. Target Death Validation Goal

O objetivo principal é validar o recebimento e processamento do pacote `SC_TARGET_DEAD` (opcode 3003) no cliente. Isso confirmará que o backend está rastreando corretamente a vida do alvo e notificando os clientes quando ele é derrotado.

## 4. Runtime Scenario

- **Cena:** `DebugWorldEntryScene`
- **Personagem:** `Gabriela`
- **Alvo:** `Orc_Elite` (HP: 500, conforme `backend/cmd/gateway/main.go`)
- **Arma:** `debug_sword`
- **Ação:**
  1. O jogador usará as ferramentas de movimento de debug para se posicionar adjacente ao `Orc_Elite`.
  2. O jogador clicará repetidamente no botão `Attack Orc_Elite`.
  3. Os ataques continuarão até que a vida do `Orc_Elite` chegue a zero.

## 5. Expected Packet Flow

1.  O cliente envia um pacote `CS_ATTACK_REQUEST` (3000).
2.  O servidor responde com um pacote `SC_DAMAGE_EVENT` (3002).
3.  Os passos 1 e 2 se repetem a cada clique no botão de ataque.
4.  Após dano suficiente ser aplicado e a vida do `Orc_Elite` chegar a zero ou menos, o servidor deve enviar um pacote `SC_TARGET_DEAD` (3003) para o cliente.
5.  O cliente deve receber e logar o conteúdo do pacote `SC_TARGET_DEAD`.

## 6. Validation Criteria

A validação será considerada bem-sucedida quando:

- O log de pacotes do cliente mostrar o recebimento de múltiplos eventos `SC_DAMAGE_EVENT`.
- O log de pacotes do cliente mostrar o recebimento de um único evento `SC_TARGET_DEAD` contendo o ID `Orc_Elite`.
- O cliente não travar ou apresentar erros durante ou após o recebimento do pacote de morte.

## 7. What This Will Not Validate Yet

- **Loot e Recompensas:** A geração ou recebimento de loot, experiência (XP) ou qualquer outra recompensa pós-morte.
- **Feedback Visual:** A remoção do `Orc_Elite` do mundo do jogo ou qualquer animação de morte.
- **Respawn:** A lógica de respawn do `Orc_Elite` após ser derrotado.

## 8. Risks

- O backend pode não enviar o pacote `SC_TARGET_DEAD` mesmo após a vida do alvo chegar a zero (bug na lógica de combate).
- O cliente pode falhar ao decodificar o pacote `SC_TARGET_DEAD`.
- O dano da `debug_sword` pode ser muito baixo, tornando o teste excessivamente longo.

## 9. Follow-up Tasks

1.  Implementar uma representação visual mínima para a morte de inimigos no `DebugTileWorldView` (ex: remover o tile do inimigo).
2.  Planejar a validação do próximo ciclo: loot e recompensas.
3.  Refatorar a UI de debug para permitir a seleção dinâmica de alvos.

## 10. Closure Status

**Fechado como plano.**

Este plano está pronto para guiar a próxima sessão de validação técnica. A validação runtime do opcode 3003 ainda está pendente.
