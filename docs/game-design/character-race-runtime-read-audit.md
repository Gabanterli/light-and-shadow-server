# Character Race Runtime Read Audit

## 1. Objetivo

Este documento audita o estado atual do runtime do backend e do client Godot para determinar o impacto e a estratégia segura para ler e expor o campo `race_id`, recém-adicionado ao schema do banco de dados. O objetivo é planejar as próximas tasks de integração sem alterar código, protocolo ou comportamento nesta task.

## 2. Contexto

- A Task R1-D adicionou a coluna `race_id VARCHAR(32) NOT NULL DEFAULT 'human'` à tabela `characters` via migration.
- A Task R1-E validou que o fluxo de login, seleção e carregamento de personagens existentes (como "Gabriela") continua funcionando, com o personagem legado recebendo o `race_id` padrão `human`.
- A Task R1-F alinhou o método `InitSchema` com a migration, garantindo a criação da coluna e do índice `idx_characters_race_id` em ambientes de desenvolvimento.
- O campo `race_id` existe no banco de dados, mas ainda não foi integrado à lógica de leitura do runtime.

## 3. Arquivos auditados

- `backend/pkg/persistence/persistence_manager.go`:
  - `LoadCharacter(...)`
  - `SaveCharacter(...)`
  - `ListCharactersByAccount(...)`
  - `CharacterSummary` (struct)
  - `InitSchema()`
- `backend/cmd/gateway/main.go`:
  - Fluxo de `CS_CHAR_LIST_REQUEST`
  - Fluxo de `CS_CHAR_SELECT_REQUEST`
- `backend/pkg/protocol/protocol.go`:
  - Opcodes `SC_CHAR_LIST_RESPONSE` e `SC_CHAR_SELECT_RESPONSE`.
- Documentação de protocolo e design:
  - `backend/docs/protocol/protocolo-binario-auth-personagem.md`
  - `docs/game-design/character-creation-runtime-integration-plan.md`

## 4. Estado atual da leitura de race_id

- **`race_id` existe no banco?**
  - Sim. Foi adicionado pela migration e é garantido por `InitSchema`.

- **`race_id` é carregado por `LoadCharacter`?**
  - Não. A auditoria em `persistence_manager.go` confirma que a query `SELECT` e o `row.Scan` dentro de `LoadCharacter` **não** leem a coluna `race_id`. O `combat.EntityStats` retornado não é populado com a raça.

- **`race_id` é retornado por `ListCharactersByAccount`?**
  - Não. A query `SELECT` em `ListCharactersByAccount` busca apenas `id, name, class, level, account_id`.

- **`race_id` aparece em `CharacterSummary`?**
  - Não. A struct `CharacterSummary` em `persistence_manager.go` contém apenas `ID`, `Name`, `Class`, `Level` e `Account`.

- **`race_id` aparece no protocolo?**
  - Não. Os pacotes `SC_CHAR_LIST_RESPONSE` e `SC_CHAR_SELECT_RESPONSE` não incluem um campo para raça.

- **`race_id` aparece no Godot?**
  - Não. Como o protocolo não envia a informação, o cliente Godot C# não tem como ler ou exibir a raça do personagem.

## 5. Impacto no backend

- **`LoadCharacter`:** Para que a raça seja usada no runtime (ex: para bônus raciais futuros), a query `SELECT` e o `row.Scan` precisam ser atualizados. Além disso, a struct `combat.EntityStats` precisaria de um campo `RaceID` para armazenar o valor.

- **`SaveCharacter`:** A query `UPDATE` em `SaveCharacter` **não** atualiza o campo `race_id`. Isso é o comportamento correto, pois a raça é imutável após a criação. O campo é preservado no banco de dados sem ser sobrescrito.

- **`ListCharactersByAccount`:** Para exibir a raça na tela de seleção de personagem, a query `SELECT` precisaria ser alterada para incluir `race_id`.

- **Fallback legado de criação:** O `INSERT` de fallback dentro de `LoadCharacter` **não** define um `race_id` explícito. Ele depende do `DEFAULT 'human'` definido no schema do banco de dados. Isso é um risco, pois a criação oficial futura deve validar a raça explicitamente contra o Rule Registry.

- **`CharacterSummary`:** Para expor a raça na lista de personagens, a struct `CharacterSummary` precisaria ser modificada para incluir um campo `RaceID string`.

## 6. Impacto no protocolo

- **`CS_CHAR_LIST_REQUEST`:** Nenhum impacto. A requisição não muda.

- **`SC_CHAR_LIST_RESPONSE`:** Grande impacto. Se `CharacterSummary` for alterado para incluir `race_id`, o payload binário de `SC_CHAR_LIST_RESPONSE` mudará. O cliente Godot atual, que espera o formato antigo, quebraria ao tentar desserializar o pacote, causando um erro de login ou uma falha na exibição da lista de personagens.

- **`CS_CHAR_SELECT_REQUEST`:** Nenhum impacto. A requisição não muda.

- **`SC_CHAR_SELECT_RESPONSE`:** Impacto potencial. Se a raça for adicionada ao payload de seleção, o cliente também precisaria ser atualizado para ler essa informação.

## 7. Impacto no Godot

A auditoria indica que o cliente Godot C# atual é agnóstico à raça do personagem. Uma futura integração exigiria alterações coordenadas.

- **`BinaryProtocol.cs` (ou equivalente):** O código de desserialização para `SC_CHAR_LIST_RESPONSE` precisaria ser atualizado para ler o novo campo `race_id` (string).

- **`AuthSession.cs` (ou equivalente):** A struct ou classe que armazena os dados do personagem na lista (`CharacterSummary` do lado do cliente) precisaria ser estendida com um campo para a raça.

- **`DebugAuthController.cs` (ou equivalente):** A lógica que popula a UI de seleção de personagem precisaria ser atualizada para exibir a raça ao lado do nome, classe e nível.

- **`GatewayTcpClient.cs`:** Nenhuma mudança direta, mas seria o orquestrador que lida com a conexão e os pacotes que agora teriam um formato diferente.

## 8. Riscos

- **Quebra de protocolo:** Adicionar `race_id` a `SC_CHAR_LIST_RESPONSE` sem uma estratégia de versionamento quebrará todos os clientes existentes.

- **Cliente antigo:** Um cliente com a versão antiga do protocolo não conseguirá mais se conectar ou listar personagens se o backend for atualizado de forma incompatível.

- **Inconsistência entre banco e runtime:** Se `LoadCharacter` não carregar `race_id`, o runtime (ex: `combat.EntityStats`) não terá a informação, mesmo que ela exista no banco. Qualquer lógica futura que dependa da raça falhará ou usará um valor zero/padrão incorreto.

- **Fallback criando personagem sem raça oficial:** O fallback em `LoadCharacter` continua criando um personagem `human` por `DEFAULT`, sem passar pela validação do `Rule Registry`. Isso precisa ser endereçado quando a criação oficial for implementada.

- **Usar `Novice` vs `novice`:** O fallback em `LoadCharacter` ainda usa a string `'Novice'` (maiúscula), enquanto o `Rule Registry` define `novice` (minúscula). A criação oficial deve usar a constante do `Rule Registry`.

- **`race_id` não validado pelo Rule Registry:** O `LoadCharacter` não valida o `race_id` vindo do banco. Embora o risco seja baixo (pois a criação oficial o validará), é um ponto a ser observado.

## 9. Recomendação de ordem segura

Para expor a raça do personagem de forma segura e incremental, as próximas tasks devem seguir esta ordem:

1.  **Task R1-H — Backend Race Read Implementation:**
    - Adicionar o campo `RaceID string` à struct `combat.EntityStats`.
    - Atualizar a query e o `Scan` em `LoadCharacter` para ler a coluna `race_id` do banco e popular o novo campo em `EntityStats`.
    - **Não** alterar `ListCharactersByAccount` ou `CharacterSummary` ainda para evitar quebra de protocolo.

2.  **Task R1-I — Protocol and Client Race Integration:**
    - Adicionar o campo `RaceID string` à struct `CharacterSummary`.
    - Atualizar `ListCharactersByAccount` para ler e retornar `race_id`.
    - Atualizar o handler de `SC_CHAR_LIST_RESPONSE` no gateway para serializar o novo campo.
    - Atualizar o cliente Godot C# (`BinaryProtocol`, `AuthSession`, UI) para ler e exibir a raça na tela de seleção de personagem. Esta deve ser uma mudança coordenada.

3.  **Task R1-J — Character Creation Service Implementation:**
    - Somente após a leitura e exibição da raça estarem funcionando, implementar o serviço de criação de personagem, que usará o `Rule Registry` para validação.

4.  **Task R1-K — Character Creation Protocol/Opcode:**
    - Implementar o novo opcode `CS_CHAR_CREATE_REQUEST` e a integração final com o cliente.

## 10. Fora de escopo

Esta task de auditoria **não** realiza nenhuma das seguintes ações:

- Não implementa a criação de personagem.
- Não altera o protocolo de rede ou o cliente Godot.
- Não altera o schema do banco de dados.
- Não remove o fallback de criação legado em `LoadCharacter`.
- Não altera nenhum arquivo de código Go ou C#.

## 11. Status

Esta auditoria está concluída. Ela mapeia os pontos de impacto e os riscos associados à leitura do campo `race_id`. O projeto está pronto para iniciar a próxima task do roadmap, **R1-H — Backend Race Read Implementation**, que consiste em integrar a leitura da raça ao modelo de dados interno do backend de forma segura e sem quebrar o protocolo atual.


