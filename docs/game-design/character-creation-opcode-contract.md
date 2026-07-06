# Light and Shadow — Character Creation Opcode Contract

## Status

- Planejado
- Não implementado ainda
- Futuro contrato backend + Godot
- Depende das decisões já tomadas em R1-D, R1-H e R1-I-B

## Contexto já existente

- O banco de dados já possui a coluna `characters.race_id`.
- O backend já consegue ler `race_id` internamente na função `LoadCharacter`.
- O protocolo de lista de personagens (`SC_CHAR_LIST_RESPONSE`) já expõe o campo `race_id`.
- A ordem dos campos na lista de personagens é: `name`, `class`, `level`, `race_id`.

## OpCodes planejados

- `1008 CS_CHAR_CREATE_REQUEST`
- `1009 SC_CHAR_CREATE_RESPONSE`

## CS_CHAR_CREATE_REQUEST

O cliente só pode enviar uma intenção mínima de criação. O payload da requisição deve conter apenas:

- `string desired_name`
- `string race_id`

O cliente **NÃO PODE** enviar nenhum dos seguintes campos. O backend deve ignorar ou rejeitar qualquer tentativa de enviá-los:

- `account_id`
- `level`
- `class`
- `base_class`
- `element`
- `stats`
- `position`
- `inventory`
- `gold`
- `skills`
- `flags`
- `spawn`
- `permissions`

## Regras autoritativas do backend

- O `account_id` do novo personagem vem exclusivamente da sessão autenticada no backend.
- Um personagem recém-criado sempre começa com a classe `novice`.
- O nível (`level`) inicial é sempre `1`.
- A classe base (`knight`, `mage`, etc.) não é escolhida na criação.
- A escolha da classe base só pode ser feita a partir do nível 10.
- Nenhum elemento é escolhido na criação.
- O `race_id` enviado pelo cliente precisa ser validado contra o Rule Registry.
- As raças jogáveis oficiais são: `human`, `forest_elf`, `dwarf`, `ice_elf`, `green_orc`.
- `ogre` não é uma raça jogável e deve ser rejeitado.
- `air` e `water` não são elementos oficiais e não participam do fluxo de criação.
- O ponto de spawn inicial é definido autoritativamente pelo servidor.
- O inventário inicial é definido autoritativamente pelo servidor.
- O ouro (`gold`) inicial é definido autoritativamente pelo servidor.
- Os stats iniciais (HP, Mana, etc.) são definidos autoritativamente pelo servidor.
- Nenhuma decisão autoritativa sobre o estado do personagem pode vir do cliente.

## Nome do personagem

O backend deverá validar o nome do personagem de forma rigorosa:

- O nome não pode ser vazio.
- O tamanho mínimo e máximo do nome são definidos pelo backend.
- Os caracteres permitidos são definidos pelo backend.
- O nome deve passar por um processo de normalização/canonicalização antes de ser persistido.
- A unicidade do nome deve ser garantida no banco de dados.
- Nomes reservados, ofensivos ou de sistema devem ser rejeitados.
- A implementação deve usar uma constraint `UNIQUE` no banco e uma transação para proteger contra race conditions na verificação de duplicidade.

## Transação

A futura implementação da criação de personagem deve ser transacional, seguindo esta ordem:

1. Validar a sessão do jogador.
2. Validar se a conta atingiu o limite de personagens.
3. Validar o nome do personagem.
4. Validar o `race_id` contra o Rule Registry.
5. Iniciar a transação no banco de dados.
6. Inserir o novo personagem na tabela `characters`.
7. Criar e inserir o inventário inicial na tabela `inventories`.
8. Criar qualquer outro estado inicial necessário.
9. Fazer o `COMMIT` da transação.
10. Em caso de qualquer erro, executar um `ROLLBACK` completo para garantir a consistência dos dados.

## Relação com LoadCharacter

A criação oficial de personagem **NÃO PODE** depender do fallback existente na função `LoadCharacter`.

- `LoadCharacter` deve ser usado apenas para carregar um personagem existente.
- A criação de um novo personagem deve ser um fluxo explícito e separado.
- Qualquer comportamento de auto-criação dentro de `LoadCharacter` deve ser tratado como um artefato legado e um risco técnico a ser removido futuramente.

## SC_CHAR_CREATE_RESPONSE

O payload de resposta do servidor (`SC_CHAR_CREATE_RESPONSE`) deve conter:

- `uint8 success`: `1` para sucesso, `0` para falha.
- `string error_code`: Um código de erro padronizado se `success` for `0`.
- `string name`: O nome do personagem criado (se sucesso).
- `string class`: A classe inicial do personagem (`novice`).
- `uint32 level`: O nível inicial do personagem (`1`).
- `string race_id`: A raça do personagem.

A ordem dos campos do resumo do personagem na resposta deve ser: `name`, `class`, `level`, `race_id`.

## Error codes planejados

- `not_authenticated`
- `invalid_name`
- `name_taken`
- `invalid_race`
- `character_limit_reached`
- `persistence_error`
- `internal_error`

## Pós-criação

- Criar um personagem não o seleciona automaticamente.
- Criar um personagem não faz o jogador entrar no mundo do jogo.
- Criar um personagem não envia os chunks do mapa.
- Após uma resposta de sucesso, o cliente pode solicitar a lista de personagens atualizada (`CS_CHAR_LIST_REQUEST`) para exibir o novo personagem.
- A seleção de personagem continua acontecendo exclusivamente através do fluxo `CS_CHAR_SELECT_REQUEST`.

## Segurança MMO

A implementação futura deve mitigar os seguintes riscos:

- Criação massiva/spam de personagens (requer rate limiting).
- Nomes duplicados por race condition (requer constraint `UNIQUE` e transação).
- Cliente tentando criar personagem com classe, nível ou stats adulterados (requer validação 100% server-authoritative).
- Exploração de inventário inicial (requer inventário definido pelo servidor).
- Abuso de economia por ouro inicial (requer ouro definido pelo servidor).
- Bypass de raça bloqueada (requer validação contra o Rule Registry).
- Inconsistência entre o protocolo do backend e a implementação do cliente Godot.
- Rollback incompleto, gerando um personagem sem inventário ou um inventário órfão.

## Fora de escopo

Esta task de documentação não envolve:

- Implementar o opcode no backend ou no cliente.
- Alterar o banco de dados.
- Criar a UI de criação de personagem.
- Alterar a cena `DebugAuthScene` no Godot.
- Implementar a seleção de classe ou de elemento.

## Critério de aceite futuro

A implementação futura dos opcodes `1008` e `1009` só poderá começar quando este contrato de protocolo estiver aprovado e commitado no repositório.