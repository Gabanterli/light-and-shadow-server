# Light and Shadow — Character Creation Persistence Audit

## Status

- Auditoria
- Não implementa criação de personagem
- Preparação para futura implementação do serviço `CreateCharacterForAccount`

## Contexto

Esta auditoria é realizada com base no seguinte contexto já estabelecido:

- O contrato dos opcodes `1008 CS_CHAR_CREATE_REQUEST` e `1009 SC_CHAR_CREATE_RESPONSE` foi definido.
- Os codecs binários para os opcodes `1008` e `1009` já foram implementados no pacote de protocolo do backend.
- A coluna `race_id` já existe no schema do banco de dados.
- O protocolo `SC_CHAR_LIST_RESPONSE` já expõe o campo `race_id`.
- Existe um plano técnico para a criação de um serviço de backend transacional e autoritativo.

## Escopo da auditoria

Esta auditoria foca exclusivamente na camada de persistência (`persistence_manager.go` e migrations) e nos riscos associados à futura implementação do fluxo de criação de personagem. Não são propostas implementações detalhadas nem alterações de código.

## Estado atual da tabela `characters`

A auditoria do código em `persistence_manager.go` (`InitSchema`) e das migrations revela os seguintes detalhes sobre a tabela `characters`:

- **Campos relevantes existentes:**
  - `id`: `SERIAL PRIMARY KEY`
  - `account_id`: `INT NOT NULL` (referencia `accounts`)
  - `name`: `VARCHAR(32) UNIQUE NOT NULL`
  - `class`: `VARCHAR(20) NOT NULL`
  - `level`: `INT DEFAULT 1`
  - `race_id`: `VARCHAR(32) NOT NULL DEFAULT 'human'`
  - Posição: `posX`, `posY`, `posZ` (todos `FLOAT`)
  - `gold`: `BIGINT`
  - Stats: `health`, `max_health`, `mana`, `max_mana`, etc. (todos `DOUBLE PRECISION`)

- **Índices e Constraints:**
  - A coluna `name` possui uma constraint `UNIQUE NOT NULL`, o que garante a unicidade do nome no nível do banco de dados.
  - Existem índices em `account_id`, `name` e `race_id`.

## Estado atual da tabela `inventories`

- **Persistência:** O inventário é persistido na tabela `inventories`.
- **Relação:** A tabela está vinculada à tabela `characters` através da coluna `character_id`, que é uma chave estrangeira.
- **Campos relevantes:** `id`, `character_id`, `slot_index`, `item_id`, `quantity`, `durability`.
- **Carregamento:** A função `LoadCharacter` carrega todos os itens de um personagem consultando a tabela `inventories` por `character_id`.
- **Salvamento:** A função `SaveCharacter` executa uma operação de "delete-then-insert": ela apaga todos os itens do inventário do personagem e, em seguida, insere o estado atual dos itens. Esta operação é realizada dentro de uma transação.

## `LoadCharacter`

- **Função atual:** Carrega o estado completo de um personagem do banco de dados, incluindo todos os atributos de combate, posição, ouro, experiência, versão e itens do inventário.
- **Uso de `race_id`:** A função lê a coluna `race_id` do banco e a popula no campo `RaceID` da struct `combat.EntityStats`.
- **Fallback/Auto-criação:** Existe um bloco de código `if errors.Is(err, sql.ErrNoRows)` que é acionado quando um personagem não é encontrado. Este bloco insere um novo personagem padrão na tabela `characters` e popula seu inventário inicial.
- **Riscos do Fallback:**
  - **Mistura de Responsabilidades:** Combina a lógica de carregamento com a de criação, o que é uma má prática.
  - **Segurança:** O fallback usa um `account_id` fixo (`1`) e não valida a propriedade da conta que está tentando carregar o personagem.
  - **Contrato Violado:** Não valida o nome, não valida a raça contra o Rule Registry e não aplica as regras de negócio definidas no contrato de criação.
  - **Mascaramento de Erros:** Pode esconder problemas de dados ou de lógica, criando um personagem em vez de retornar um erro de "não encontrado".
- **Conclusão:** A criação oficial de personagem deve ser um fluxo explícito e separado para garantir segurança, validação e conformidade com as regras.

## `SaveCharacter`

- **Função atual:** Salva o estado completo de um personagem *existente*.
- **O que salva:** Salva posição (`posX`, `posY`, `posZ`), inventário completo, ouro (`gold`), versão (`version`), experiência e todos os stats de combate.
- **Dependência:** A função depende que o personagem já exista no banco para obter seu `charID` e aplicar o `UPDATE`. Ela não foi projetada para criar novos personagens.
- **Uso como criação:** Não deve ser usada para criação, pois sua lógica é baseada em `UPDATE` e `DELETE/INSERT` para um `character_id` já existente.

## `ListCharactersByAccount`

- **Função atual:** Lista os personagens de uma conta, retornando `id`, `name`, `class`, `level`, `account_id` e `race_id`.
- **Compatibilidade:** Os campos retornados são compatíveis com a struct `protocol.CharacterListEntry` e, portanto, com o payload do `SC_CHAR_LIST_RESPONSE`.
- **Riscos:** Se um personagem fosse criado sem um `race_id` válido (por exemplo, `NULL` ou vazio), o `SC_CHAR_LIST_RESPONSE` poderia enviar um valor inesperado ao cliente, potencialmente causando erros na UI. O `DEFAULT 'human'` atual mitiga isso para dados legados.

## Inventário inicial

- **Lógica atual:** Não há uma lógica clara e separada para "inventário inicial". O fallback de `LoadCharacter` cria um inventário padrão chamando `inventory.NewPlayerInventory()` e persistindo seus itens.
- **Riscos:** Criar um personagem sem um inventário associado deixaria o estado do jogador inconsistente.
- **Necessidade futura:** A criação de personagem deve ser transacional, garantindo que a inserção na tabela `characters` e a inserção dos itens iniciais na tabela `inventories` ocorram de forma atômica (tudo ou nada).

## Nome único

- **Validação atual:** Não há validação explícita de nome na camada de serviço ou aplicação.
- **Constraint no banco:** A auditoria confirmou que a tabela `characters` possui uma constraint `UNIQUE NOT NULL` na coluna `name`. Isso protege contra nomes duplicados em nível de banco de dados e previne race conditions.
- **Necessidades futuras:**
  - A camada de serviço precisará normalizar/canonicalizar o nome antes de tentar a inserção.
  - A camada de serviço precisará tratar o erro de violação de constraint única do banco e traduzi-lo para um erro amigável (`name_taken`).

## Race ID

- **Existência:** A coluna `race_id` foi encontrada no método `InitSchema` com `VARCHAR(32) NOT NULL DEFAULT 'human'`.
- **Índice:** Um índice (`idx_characters_race_id`) também foi encontrado.
- **Necessidade futura:** A camada de serviço deve validar o `race_id` recebido do cliente contra o Rule Registry antes de passá-lo para a camada de persistência.
- **Riscos:** Sem essa validação, um valor inválido (como `ogre`, vazio ou malicioso) poderia ser persistido se a constraint `DEFAULT` fosse removida ou alterada, violando as regras do jogo.

## Transação futura

A futura implementação da persistência para criação de personagem deve garantir uma transação única que englobe:

1.  `INSERT` na tabela `characters`.
2.  `INSERT` de todos os itens do inventário inicial na tabela `inventories`.
3.  `INSERT` em quaisquer outras tabelas de estado inicial (ex: `character_professions`, `character_quests`).
4.  `COMMIT` somente se todas as operações forem bem-sucedidas.
5.  `ROLLBACK` completo em caso de qualquer erro, garantindo que não existam personagens órfãos sem inventário ou inventários órfãos sem personagem.

## Gaps encontrados

- O serviço `CreateCharacterForAccount` ainda não existe.
- O Gateway ainda não possui um handler para o opcode `CS_CHAR_CREATE_REQUEST`.
- O cliente Godot ainda não envia a requisição de criação.
- A validação de nome (formato, comprimento, palavras reservadas) ainda não foi implementada na camada de serviço.
- A validação de raça contra o Rule Registry para o fluxo de criação ainda não foi implementada.
- A política de limite de personagens por conta ainda não foi implementada.
- O fallback de criação em `LoadCharacter` ainda não foi removido ou desativado, representando um risco técnico.
- A confirmação da constraint de nome único foi feita, mas a lógica de tratamento de erro no serviço ainda precisa ser criada.

## Recomendação para próxima implementação

A próxima task de implementação de código deve ser focada no backend, com escopo reduzido, para criar a base da camada de persistência.

**Próxima task sugerida:** `R1-L-B — Add Character Creation Persistence Skeleton`

Esta task criaria a função `CreateCharacterForAccount` no `PersistenceManager` com a lógica transacional de `INSERT`, mas ainda sem integração com o Gateway ou validações complexas.

## Fora de escopo

- Não implementar o serviço de criação.
- Não alterar o schema do banco de dados.
- Não criar uma nova migration.
- Não alterar o Gateway.
- Não alterar o cliente Godot.
- Não criar a UI de criação.
- Não remover o fallback de `LoadCharacter` ainda.

## Critério de aceite desta task

- O documento `character-creation-persistence-audit.md` foi criado no caminho correto.
- Nenhuma alteração de código foi realizada.
- A auditoria identifica os riscos de persistência relacionados ao fallback de `LoadCharacter` e à falta de validação.
- A auditoria prepara o caminho para a futura implementação de uma criação de personagem transacional e segura.