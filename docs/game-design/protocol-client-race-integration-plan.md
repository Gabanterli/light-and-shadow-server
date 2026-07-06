# Protocol and Client Race Integration Plan

## 1. Objetivo

Este documento detalha o plano técnico para expor o campo `race_id` no protocolo de rede e integrá-lo ao cliente Godot C#. O objetivo é planejar as futuras alterações de forma coordenada, garantindo que o backend e o cliente evoluam juntos sem quebrar o fluxo de login e seleção de personagem. Esta task é exclusivamente de planejamento; nenhuma implementação será realizada.

## 2. Contexto

- A coluna `race_id` já existe na tabela `characters` do banco de dados, com um `DEFAULT 'human'`.
- A Task R1-H modificou o backend para que `LoadCharacter` leia o `race_id` do banco e o armazene internamente na struct `combat.EntityStats`.
- Atualmente, a informação da raça é carregada no backend, mas não é transmitida ao cliente em nenhum momento do fluxo de autenticação ou seleção de personagem.
- O cliente Godot é, portanto, completamente agnóstico à raça do personagem.

## 3. Estado atual

- **`race_id` existe no banco?**
  - Sim.

- **`race_id` entra no modelo interno backend?**
  - Sim, na struct `combat.EntityStats` ao chamar `LoadCharacter`.

- **`race_id` aparece em `CharacterSummary`?**
  - Não. A struct `persistence.CharacterSummary` contém apenas `ID`, `Name`, `Class`, `Level`, e `Account`.

- **`race_id` aparece em `SC_CHAR_LIST_RESPONSE`?**
  - Não. O payload de resposta da lista de personagens não inclui a raça.

- **`race_id` aparece em `SC_CHAR_SELECT_RESPONSE`?**
  - Não. O payload de resposta da seleção de personagem não inclui a raça.

- **`race_id` aparece no Godot?**
  - Não.

## 4. Backend impactado futuramente

- **`persistence.CharacterSummary` (`backend/pkg/persistence/persistence_manager.go`):**
  - A struct precisará ser estendida com um campo `RaceID string` para que a informação possa ser transportada da camada de persistência para a camada de protocolo.

- **`PersistenceManager.ListCharactersByAccount` (`backend/pkg/persistence/persistence_manager.go`):**
  - A query SQL `SELECT` precisará ser alterada para incluir a coluna `race_id`.
  - A chamada `rows.Scan` precisará ser atualizada para ler o valor de `race_id` e popular o novo campo na struct `CharacterSummary`.

- **Handler de `CS_CHAR_LIST_REQUEST` (`backend/cmd/gateway/main.go`):**
  - A struct `protocol.CharacterListEntry` precisará ser estendida com um campo `RaceID string`.
  - O loop que converte `persistence.CharacterSummary` para `protocol.CharacterListEntry` precisará copiar o valor de `RaceID`.
  - A função `protocol.EncodeCharacterListResponse` precisará ser atualizada para serializar o novo campo.

- **Handler de `CS_CHAR_SELECT_REQUEST` (`backend/cmd/gateway/main.go`):**
  - Nenhum impacto. A raça já estará disponível no cliente a partir da lista de personagens. Não é necessário adicioná-la à resposta de seleção.

- **Documentação de protocolo (`backend/docs/protocol/protocolo-binario-auth-personagem.md`):**
  - O documento precisará ser atualizado para refletir a adição do campo `race_id` no payload de `SC_CHAR_LIST_RESPONSE`.

## 5. Protocolo impactado futuramente

- **Opcode envolvido:** `SC_CHAR_LIST_RESPONSE` (1005).

- **Formato atual esperado (por personagem):**
  - `string name`
  - `string class`
  - `uint32 level`

- **Formato futuro proposto (por personagem):**
  - `string name`
  - `string class`
  - `uint32 level`
  - `string race_id`

- **Risco de quebra:**
  - Alto. Adicionar um novo campo ao final do payload de cada personagem em um loop mudará o tamanho e o layout do pacote `SC_CHAR_LIST_RESPONSE`. Um cliente antigo que não espera o campo `race_id` falhará ao desserializar o pacote, provavelmente resultando em um erro de `InvalidDataException` ou similar, impedindo o jogador de ver a lista de personagens.

- **Decisão sobre versionamento ou breaking change controlado:**
  - Dado que o projeto ainda está em fase de desenvolvimento (debug/local) e não há clientes em produção, um versionamento complexo do protocolo é desnecessário. A abordagem mais pragmática é tratar isso como um **breaking change controlado**. A implementação futura deve atualizar o backend e o cliente Godot de forma coordenada, idealmente na mesma task ou em um único pull request.

## 6. Godot impactado futuramente

- **`scripts/BinaryProtocol.cs`:**
  - A classe `CharacterListEntryData` precisará ser estendida com uma propriedade `public string RaceId { get; set; }`.
  - O método `DecodeCharacterListResponse` precisará ser modificado. Dentro do loop `for`, após ler `level`, ele deverá ler o novo campo `race_id` e populá-lo no objeto `CharacterListEntryData`.

- **`scripts/AuthSession.cs`:**
  - Nenhum impacto direto. Esta classe armazena o estado da sessão autenticada, não a lista de personagens.

- **`scripts/DebugAuthController.cs`:**
  - O método `OnRequestCharactersButtonPressed` popula a UI. A linha que formata o texto do item na lista (`_characterList.AddItem(...)`) precisará ser atualizada para incluir a raça.
  - Formato sugerido para a UI de debug: `_characterList.AddItem($"{character.Name} (Lvl {character.Level} {character.Class} {character.RaceId})");`

- **`scripts/GatewayTcpClient.cs`:**
  - Nenhum impacto direto. Este cliente apenas envia e recebe pacotes brutos; a lógica de decodificação está em `BinaryProtocol.cs`.

## 7. Estratégia recomendada

A implementação futura deve ser uma mudança coordenada e atômica, seguindo esta ordem:

1.  **Backend:**
    - Alterar a struct `persistence.CharacterSummary` para incluir `RaceID`.
    - Alterar a função `ListCharactersByAccount` para ler e popular `RaceID`.
    - Alterar a struct `protocol.CharacterListEntry` para incluir `RaceID`.
    - Alterar o handler de `CS_CHAR_LIST_REQUEST` para passar `RaceID` para o protocolo.
    - Alterar `protocol.EncodeCharacterListResponse` para serializar o novo campo.

2.  **Documentação:**
    - Atualizar `backend/docs/protocol/protocolo-binario-auth-personagem.md` para refletir o novo formato do payload de `SC_CHAR_LIST_RESPONSE`.

3.  **Godot Client:**
    - Alterar a classe `CharacterListEntryData` em `BinaryProtocol.cs` para incluir `RaceId`.
    - Alterar o método `DecodeCharacterListResponse` em `BinaryProtocol.cs` para ler o novo campo.
    - Alterar `DebugAuthController.cs` para exibir a raça na UI da lista de personagens.

4.  **Validação:**
    - Compilar e testar o backend e o cliente juntos.

## 8. Critérios de aceite da implementação futura

- `gofmt` passa em todos os arquivos Go alterados.
- `go test ./pkg/...` passa com sucesso.
- O projeto Godot C# compila com `dotnet build`.
- O ambiente completo sobe com `docker compose up --build`.
- O fluxo de login com `default_user`/`test123` funciona.
- Ao solicitar a lista de personagens, o personagem "Gabriela" é exibido na UI do Godot com a raça "human".
- A seleção do personagem "Gabriela" funciona.
- Após a seleção, o inventário, os chunks do mapa e o movimento do personagem continuam funcionando normalmente.

## 9. Fora de escopo

A futura task de implementação **não** deve:
- Implementar a criação de personagem.
- Criar um novo opcode para criação de personagem.
- Alterar o schema do banco de dados (já foi feito).
- Alterar o `Rule Registry`.
- Alterar a lógica de combate, movimento ou economia.

## 10. Status

Este plano de integração está concluído. Ele fornece um roteiro claro e seguro para expor a raça do personagem ao cliente.

A próxima task de implementação deve ser **R1-I-B — Implement Protocol and Client Race Integration**, que executará as mudanças detalhadas neste documento de forma coordenada no backend e no cliente Godot.


