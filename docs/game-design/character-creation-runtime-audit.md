# Character Creation Runtime Audit

## 1. Objetivo

Este documento audita o estado atual do runtime relacionado à criação, listagem, seleção e carregamento de personagens no backend do Light and Shadow. O objetivo é preparar a futura integração segura das regras oficiais de criação de personagem definidas na Fase 6, sem alterar código, protocolo, banco, client ou comportamento runtime nesta task.

## 2. Estado atual encontrado

A auditoria encontrou que **não existe fluxo oficial de criação de personagem exposto ao client** no protocolo atual.

O que existe hoje:

- fluxo de login;
- fluxo de listagem de personagens por conta;
- fluxo de seleção de personagem;
- validação de ownership antes do select;
- carregamento de personagem persistido;
- fallback interno em `LoadCharacter` que pode criar uma linha padrão se o personagem não for encontrado;
- schema/tabela `characters` existente;
- tabela `inventories` existente;
- seed/conta padrão para teste;
- personagem local/de teste usado no fluxo de debug, como `Gabriela`.

Portanto, a conclusão correta é:

- **não existe character creation oficial via protocolo/client ainda**;
- **existe infraestrutura parcial de personagens no banco e na persistência**;
- **existe fallback de criação padrão dentro do carregamento**, mas ele não é o fluxo oficial de criação de personagem e não aplica o contrato completo da Fase 6.

## 3. Arquivos auditados

### `backend/migrations/0002_create_characters.up.sql`

Existe migration específica para a tabela `characters`.

A tabela contém base de persistência para personagem, incluindo:

- `id`;
- `account_id`;
- `name`;
- `class`;
- `level`;
- `experience`;
- `posX`;
- `posY`;
- `posZ`;
- timestamps;
- índices para `account_id` e `name`.

Ponto importante: a auditoria não confirmou uma coluna oficial `race_id` nessa migration base. Isso será um ponto crítico para a futura implementação de criação real de personagem, porque a regra oficial exige seleção e validação de raça jogável.

### `backend/migrations/0003_create_inventories.up.sql`

Existe migration para inventários, vinculada a `characters`.

Isso indica que a criação real de personagem precisará considerar inventário inicial autoritativo e persistência consistente entre personagem e itens iniciais.

### `backend/pkg/persistence/persistence_manager.go`

Arquivo central encontrado na auditoria.

Funções relevantes:

- `ListCharactersByAccount(accountID int)`;
- `CharacterBelongsToAccount(accountID int, characterName string)`;
- `LoadCharacter(playerID string)`;
- `SaveCharacter(...)`;
- `InitSchema()`.

Achados importantes:

- `ListCharactersByAccount` lista `id`, `name`, `class`, `level`, `account_id` da tabela `characters`.
- `CharacterBelongsToAccount` valida ownership usando `account_id` e `name`.
- `LoadCharacter` carrega dados autoritativos de personagem e inventário.
- `LoadCharacter` possui fallback que cria um personagem padrão quando não encontra a linha no banco.
- Esse fallback usa `account_id = 1`, `class = 'Novice'`, `level = 1`, posição inicial e inventário padrão.
- Esse fallback não valida raça oficial, não registra `race_id`, não aplica o contrato oficial completo e não deve ser tratado como criação oficial de personagem.

### `backend/cmd/gateway/main.go`

Arquivo central do fluxo runtime atual.

Achados relevantes:

- `CS_CHAR_LIST_REQUEST` exige autenticação.
- O gateway chama `s.persistenceMgr.ListCharactersByAccount(authenticatedAccountID)`.
- `CS_CHAR_SELECT_REQUEST` exige autenticação.
- O gateway lê o nome do personagem selecionado.
- O gateway valida ownership com `s.persistenceMgr.CharacterBelongsToAccount(authenticatedAccountID, selectedCharacterName)`.
- Somente após ownership válido, o gateway chama `s.persistenceMgr.LoadCharacter(characterID)`.
- Depois do select, o backend inicializa inventário, XP, movement, AOI, quests, combat manager e envia `SC_CHAR_SELECT_RESPONSE`, `SC_INVENTORY_SYNC` e chunks iniciais.

Esse fluxo é de **seleção/carregamento**, não de criação oficial.

### `backend/pkg/protocol/protocol.go`

O protocolo atual possui opcodes de character list/select:

- `CS_CHAR_LIST_REQUEST = 1004`;
- `SC_CHAR_LIST_RESPONSE = 1005`;
- `CS_CHAR_SELECT_REQUEST = 1006`;
- `SC_CHAR_SELECT_RESPONSE = 1007`.

A auditoria não encontrou opcode oficial de criação de personagem, como `CS_CHAR_CREATE_REQUEST` ou equivalente.

### `backend/docs/protocol/protocolo-binario-auth-personagem.md`

Documento de protocolo atual para login, lista e seleção de personagem.

Qualquer mudança futura no protocolo de personagem precisa atualizar este documento junto com `pkg/protocol/protocol.go`, testes de protocolo e client Godot.

## 4. Fluxo atual provável

O fluxo atual de entrada no jogo é:

1. O client realiza login.
2. O backend autentica a conta e guarda `authenticatedAccountID`.
3. O client pede lista de personagens com `CS_CHAR_LIST_REQUEST`.
4. O gateway busca personagens reais no PostgreSQL por `account_id`.
5. O client escolhe um personagem pelo nome com `CS_CHAR_SELECT_REQUEST`.
6. O gateway valida se o personagem pertence à conta autenticada.
7. O gateway carrega personagem e inventário via `LoadCharacter`.
8. O gateway inicializa sistemas runtime: inventário, XP, movimento, AOI, quests, combat manager.
9. O backend envia resposta de seleção, inventário e chunks iniciais.

Esse fluxo não cria personagem novo oficialmente.

## 5. Lacunas atuais

Para existir criação real de personagem, ainda faltam:

- opcode ou endpoint oficial de criação, se for decidido via protocolo TCP;
- request/response documentados;
- handler no gateway ou serviço apropriado;
- validação de sessão autenticada;
- validação de nome;
- validação de raça oficial;
- bloqueio explícito de `ogre`;
- bloqueio de classe inicial enviada pelo client;
- bloqueio de elemento inicial enviado pelo client;
- bloqueio de `air` e `water`;
- definição autoritativa de `class = novice`;
- definição autoritativa de `level = 1`;
- definição autoritativa de stats iniciais;
- definição autoritativa de spawn inicial;
- definição autoritativa de inventário inicial;
- persistência transacional de personagem + inventário;
- regra de unicidade de nome;
- testes de integração.

## 6. Regras que deverão ser integradas futuramente

A futura implementação deve seguir o contrato oficial:

- personagem recém-criado começa como `novice`;
- classe base não é escolhida na criação;
- elemento não é escolhido na criação;
- raças válidas são `human`, `forest_elf`, `dwarf`, `ice_elf`, `green_orc`;
- `ogre` é bloqueado como raça jogável;
- `air` e `water` são bloqueados;
- backend define stats, spawn, inventário e persistência;
- client apenas envia intenção.

Regras/funções candidatas para uso:

- catálogo oficial de raças em `backend/pkg/gamedata/rules/catalog.go`;
- `rules.StartingClassNovice`;
- funções oficiais existentes do catálogo, se disponíveis;
- `Registry.Get` ou helper oficial, caso não exista função específica para raça.

Importante: se uma função exata como `rules.IsOfficialRace` não existir, ela não deve ser inventada em documentação runtime como se já existisse. A futura implementação deve confirmar os nomes reais antes de codar.

## 7. Riscos encontrados

### Risco 1 — criação implícita dentro de `LoadCharacter`

`LoadCharacter` possui fallback que cria personagem padrão quando não encontra uma linha.

Esse comportamento é perigoso para criação oficial porque:

- mistura carregamento com criação;
- usa `account_id = 1`;
- não valida raça;
- não aplica o contrato oficial completo;
- não diferencia criação intencional de ausência inesperada de dados;
- pode mascarar bug de ownership, seed ou persistência.

Na futura integração, criação oficial deve ser separada de carregamento.

### Risco 2 — ausência de `race_id` confirmada no schema base

A criação oficial exige raça jogável. Se a tabela atual não possui `race_id`, será necessário planejar migration ou estratégia de persistência antes da implementação.

### Risco 3 — campo `element` já existe no runtime

O schema/runtime possui campo `element`, mas o design oficial diz que elemento não é escolhido na criação inicial.

A futura implementação deve garantir que o client não consiga definir elemento inicial e que valores antigos como `Light`, `None`, `air` ou `water` não contaminem o contrato oficial.

### Risco 4 — protocolo sem create character

O protocolo atual cobre login, lista e seleção, mas não criação.

Se a futura implementação exigir novo opcode, será necessário atualizar:

- `backend/pkg/protocol/protocol.go`;
- testes de protocolo;
- `backend/docs/protocol/protocolo-binario-auth-personagem.md`;
- client Godot C#.

### Risco 5 — client authority

Sem validação server-authoritative, um client modificado poderia tentar criar personagem com:

- raça inválida;
- `ogre`;
- classe inicial `knight`;
- elemento inicial `fire`, `air` ou `water`;
- stats adulterados;
- spawn adulterado;
- inventário inicial adulterado.

## 8. Arquivos candidatos para futura Task R1-B

### Backend runtime

- `backend/cmd/gateway/main.go`
  - candidato se a criação for via Gateway TCP.

### Persistência

- `backend/pkg/persistence/persistence_manager.go`
  - candidato para extrair/criar método explícito de criação, evitando usar `LoadCharacter` como criação implícita.

### Banco/migrations

- `backend/migrations/0002_create_characters.up.sql`
  - schema base existente.
- nova migration futura pode ser necessária se for preciso adicionar `race_id`, normalizar `class`, revisar `element` ou criar constraints.

### Protocolo

Somente se for decidido criar fluxo TCP oficial:

- `backend/pkg/protocol/protocol.go`;
- `backend/pkg/protocol/protocol_response_test.go`;
- `backend/docs/protocol/protocolo-binario-auth-personagem.md`;
- client Godot C#.

### Testes

- testes de persistence;
- testes de gateway/handler;
- testes de protocolo, se houver novo opcode;
- testes de integração via Docker.

### Documentação

- documentação de protocolo, se houver mudança;
- plano de implementação R1-B;
- contrato de criação, caso precise ser refinado.

## 9. Estratégia segura para futura implementação

A implementação futura deve seguir esta ordem:

1. Confirmar schema real de `characters`.
2. Decidir se será necessário adicionar `race_id`.
3. Separar criação oficial de carregamento.
4. Auditar e possivelmente remover/limitar o auto-create dentro de `LoadCharacter`.
5. Definir request mínima de criação.
6. Validar sessão autenticada.
7. Validar nome.
8. Validar raça oficial usando o Rule Registry.
9. Bloquear `ogre`.
10. Rejeitar ou ignorar qualquer classe enviada pelo client.
11. Rejeitar qualquer elemento enviado pelo client.
12. Bloquear `air` e `water`.
13. Forçar `class = novice` pelo backend.
14. Forçar `level = 1` pelo backend.
15. Definir stats/spawn/inventory iniciais no backend.
16. Persistir personagem e inventário em transação.
17. Criar testes.
18. Rodar `gofmt` e `go test` via Docker.
19. Atualizar protocolo e client somente se a task aprovar novo opcode.

## 10. Testes futuros recomendados

- criar personagem com `human`;
- criar personagem com `forest_elf`;
- criar personagem com `dwarf`;
- criar personagem com `ice_elf`;
- criar personagem com `green_orc`;
- rejeitar `ogre`;
- rejeitar raça inexistente;
- rejeitar tentativa de classe inicial `knight`;
- rejeitar tentativa de elemento inicial `fire`;
- rejeitar `air`;
- rejeitar `water`;
- garantir classe final `novice`;
- garantir level inicial `1`;
- garantir spawn definido pelo backend;
- garantir inventário inicial definido pelo backend;
- garantir nome duplicado rejeitado;
- garantir que personagem criado pertence ao `account_id` autenticado;
- garantir que falha na criação do inventário faz rollback do personagem;
- garantir que `LoadCharacter` não cria personagem oficial sem fluxo explícito.

## 11. O que esta task NÃO faz

- Não implementa criação de personagem.
- Não altera backend runtime.
- Não altera client.
- Não altera protocolo.
- Não altera banco.
- Não cria endpoint.
- Não cria migration.
- Não altera regras existentes.
- Não remove fallback de `LoadCharacter`.
- Não muda comportamento do jogo.

## 12. Status

Este documento é uma auditoria técnica preparatória para a futura integração runtime de criação de personagem.

A próxima task segura deve ser uma das duas opções, conforme decisão técnica:

- **Task R1-B — Character Creation Runtime Integration Plan**, se quisermos planejar a implementação com mais detalhe antes de codar;
- **Task R1-B — Character Creation Runtime Implementation**, se o escopo for reduzido, os arquivos permitidos forem definidos e os testes forem obrigatórios.

Como o protocolo atual não possui opcode de criação e o schema pode precisar de `race_id`, a recomendação mais segura é fazer primeiro uma task de plano técnico R1-B antes de implementar.
