# Plano de Cenário para Validação de Combate de Debug

## 1. Purpose

Este documento planeja o próximo cenário técnico necessário para validar o ciclo de combate de ponta a ponta. O objetivo é criar um ambiente de teste controlado que permita ao cliente enviar um ataque bem-sucedido (dentro do alcance) e receber e processar um evento de dano real (`SC_DAMAGE_EVENT`, opcode 3002) do backend.

## 2. Current Runtime Evidence

A validação técnica da *intenção* de ataque já foi concluída e documentada. O estado atual é:

- O cliente Godot envia com sucesso o pacote `CS_ATTACK_REQUEST` (opcode 3000).
- O Gateway Server recebe e processa este pacote.
- O backend executa validações autoritativas sobre a requisição.

## 3. Problem

O único teste de ataque documentado foi contra o alvo `Orc_Elite`. Esta tentativa foi autoritativamente rejeitada pelo backend devido à regra de alcance (`Distância: 2.83m, Alcance: 1.00m`).

Como resultado, o fluxo de cálculo de dano, a geração do pacote `SC_DAMAGE_EVENT` (3002) e a eventual morte do alvo (`SC_TARGET_DEAD` - 3003) nunca foram acionados. O ciclo de combate permanece validado apenas pela metade.

## 4. Required Successful Combat Scenario

Para considerar o ciclo de combate minimamente validado, o seguinte cenário precisa ser executado com sucesso:

1.  O jogador, a partir do cliente de debug, inicia um ataque contra um alvo.
2.  A posição do jogador e do alvo deve estar dentro do alcance da arma utilizada.
3.  O cliente envia o pacote `CS_ATTACK_REQUEST` (3000).
4.  O backend valida a requisição, confirma que está dentro do alcance, calcula o dano e envia um pacote `SC_DAMAGE_EVENT` (3002) de volta ao cliente.
5.  O cliente Godot recebe, decodifica e exibe o conteúdo do pacote `SC_DAMAGE_EVENT` no log de debug.

## 5. Options Considered

Para resolver o problema de alcance e criar um cenário de teste bem-sucedido, as seguintes opções foram consideradas:

1.  **Aproximar o Jogador do Alvo:** Modificar o cliente de debug para permitir que o jogador se mova para coordenadas adjacentes ao `Orc_Elite` antes de atacar.
2.  **Aproximar o Alvo do Jogador:** Modificar a lógica de spawn do backend para que o `Orc_Elite` apareça em uma posição adjacente ao ponto de spawn do jogador.
3.  **Usar Arma de Longo Alcance:** Modificar o `weaponType` enviado no `CS_ATTACK_REQUEST` para um tipo de arma de debug com alcance maior (ex: `debug_bow` com 10m de alcance).
4.  **Criar Alvo de Teste Adjacente:** Criar um novo alvo técnico no backend (ex: `Training_Dummy`) que é programado para aparecer sempre em uma posição adjacente ao jogador.

## 6. Recommended Approach

A abordagem recomendada é a **Opção 1: Aproximar o Jogador do Alvo**.

**Justificativa:**
- **Conservadora e de Baixo Impacto:** Esta abordagem exige apenas alterações no cliente de debug (`DebugWorldEntryController.cs`), sem tocar na lógica de combate, spawn ou dados de itens do backend.
- **Controle do Cliente:** Mantém o controle do teste no lado do cliente, que é o ambiente de depuração. O testador pode se mover para a posição correta e então iniciar o ataque, simulando um cenário de jogo real de forma controlada.
- **Reutilizável:** A capacidade de mover o jogador para coordenadas específicas pode ser útil para outros cenários de teste no futuro.
- **Evita Débito Técnico:** Evita a criação de alvos ou armas de debug que precisariam ser limpos ou gerenciados posteriormente.

A implementação futura deve ser uma task separada e focada em usar movimento debug validado pelo servidor para posicionar o jogador próximo ao alvo, sem criar bypass de teleporte nem violar o modelo server authoritative.

## 7. Files That May Be Changed In Future Implementation

- `scripts/DebugWorldEntryController.cs`: Para adicionar lógica debug de movimentação controlada usando o fluxo normal de movimento validado pelo servidor.
- `scenes/DebugWorldEntryScene.tscn`: Para adicionar botão ou campos debug de posicionamento/movimento controlado.

## 8. Files That Must Not Be Changed Yet

- `backend/*`: Nenhuma alteração no backend é necessária para a abordagem recomendada.
- `scripts/BinaryProtocol.cs`: O protocolo de ataque já está completo.
- `scripts/GatewayTcpClient.cs`: O método de envio de ataque já existe.
- Qualquer arquivo relacionado à UI de combate final, HUD, ou sistemas de target lock.

## 9. Validation Criteria

A futura task de implementação será considerada bem-sucedida quando:

- O jogador puder usar a nova funcionalidade de debug para se posicionar ao lado do `Orc_Elite`.
- Ao clicar no botão "Attack Orc_Elite" a partir da posição adjacente, o cliente receber e logar um pacote `SC_DAMAGE_EVENT` (3002) com dados de dano válidos.

## 10. Risks

- O cálculo de dano no backend pode conter um bug que impeça a geração do evento `SC_DAMAGE_EVENT`.
- A posição do `Orc_Elite` no backend pode não ser a esperada, dificultando o posicionamento adjacente.
- O pacote `SC_DAMAGE_EVENT` pode ser enviado, mas o cliente pode falhar ao decodificá-lo.

## 11. Closure Status

**Fechado.**

Este plano está concluído e fornece uma estratégia clara e de baixo risco para avançar na validação do sistema de combate. A próxima etapa é criar uma task de implementação para executar a abordagem recomendada.
