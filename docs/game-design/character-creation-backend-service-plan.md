# Light and Shadow — Character Creation Backend Service Plan

## Status

- Planejado
- Não implementado ainda
- Depende do contrato R1-J-A
- Futuro serviço backend server-authoritative

## Contexto

A implementação deste serviço se baseia no seguinte contexto já estabelecido:

- A coluna `characters.race_id` já existe no banco de dados.
- O backend já consegue ler `race_id` internamente na função `LoadCharacter`.
- O protocolo `SC_CHAR_LIST_RESPONSE` já expõe o campo `race_id`.
- O contrato de opcode (`character-creation-opcode-contract.md`) já planejou os opcodes `1008 CS_CHAR_CREATE_REQUEST` e `1009 SC_CHAR_CREATE_RESPONSE`.
- A requisição futura do cliente conterá apenas `desired_name` e `race_id`.

## Objetivo do serviço

O serviço futuro deve criar um novo personagem para uma conta autenticada de forma:

- **Autoritativa:** O servidor toma todas as decisões sobre o estado inicial do personagem.
- **Transacional:** A criação deve ser uma operação atômica para garantir a consistência dos dados.
- **Validada pelo Rule Registry:** Todas as regras de negócio (raças, classes, etc.) devem ser validadas contra o catálogo oficial.
- **Independente:** O fluxo deve ser explícito e separado do fallback de criação legado existente em `LoadCharacter`.
- **Segura contra duplicidade:** Nomes de personagem devem ser únicos em todo o servidor.
- **Segura contra dados parciais:** Uma falha no meio do processo não deve deixar dados inconsistentes no banco (ex: personagem sem inventário).

## Local planejado da lógica

A implementação futura deverá avaliar, antes de codificar, onde a lógica de criação será encapsulada. As opções são:

- **Opção 1:** Adicionar um novo método diretamente em `backend/pkg/persistence/persistence_manager.go`.
- **Opção 2:** Criar um novo serviço dedicado, por exemplo, em `backend/pkg/characters/creation_service.go`.

**Recomendação:**
- É preferível criar um serviço dedicado se a lógica de validação e orquestração crescer em complexidade.
- Regras de negócio complexas devem ser evitadas diretamente no handler do Gateway.
- O Gateway deve apenas decodificar a requisição, validar a sessão, chamar o serviço e codificar a resposta.

## Função planejada

A implementação futura pode usar uma assinatura conceitual como esta:

`CreateCharacterForAccount(ctx, accountID, desiredName, raceID) -> (*CharacterSummary, error_code)`

Onde:
- `accountID` vem exclusivamente da sessão autenticada no Gateway.
- `desiredName` vem da requisição do cliente.
- `raceID` vem da requisição do cliente, mas precisa ser validado pelo serviço.
- O retorno de sucesso (`CharacterSummary`) deve ser compatível com a ordem de campos do protocolo: `name`, `class`, `level`, `race_id`.

## Validação de sessão

A criação de personagem deve falhar com o código de erro `not_authenticated` quando:

- A conexão TCP não possui uma sessão de jogador válida.
- O token de sessão expirou.
- O `accountID` associado à sessão é zero ou inválido.
- A sessão não foi previamente confirmada pelo fluxo de login do Auth Server/Gateway.

## Validação de limite de personagens

O backend deve definir um limite de personagens por conta. A validação deve ocorrer antes de qualquer `INSERT` no banco de dados.

- **Erro planejado:** `character_limit_reached`

## Validação de nome

O backend deve validar o `desiredName` de forma rigorosa:

- Aplicar trim para remover espaços em branco no início e no fim.
- Rejeitar nomes vazios.
- Validar o comprimento mínimo e máximo definidos por constantes no backend.
- Validar se o nome contém apenas caracteres permitidos.
- Normalizar/canonicalizar o nome antes de comparar e persistir.
- Rejeitar nomes reservados, ofensivos ou de sistema.
- Garantir a unicidade no banco de dados.
- Proteger contra race conditions usando uma transação e uma constraint/índice `UNIQUE` na coluna `name`.

- **Erros planejados:** `invalid_name`, `name_taken`

## Validação de raça

O `race_id` enviado pelo cliente deve ser validado no Rule Registry.

- **Raças aceitas:** `human`, `forest_elf`, `dwarf`, `ice_elf`, `green_orc`.
- **Raças rejeitadas:** `ogre` e qualquer outra que não esteja na lista oficial.

Também deve ser documentado que:
- `air` e `water` não são elementos oficiais e não participam da criação.
- O elemento não é escolhido na criação de personagem.
- Um `race_id` vazio deve ser considerado inválido e rejeitado.

- **Erro planejado:** `invalid_race`

## Defaults autoritativos do personagem

O servidor deve definir unilateralmente todos os atributos iniciais:

- `class`: `novice`
- `level`: `1`
- `element`: Nenhum/nulo/não selecionado
- Posição de spawn inicial
- Atributos iniciais (Health, Mana, etc.)
- Inventário inicial
- Ouro (`gold`) inicial
- Flags iniciais
- Estado inicial de skills
- Estado inicial de quests, se necessário

O cliente não pode definir nenhum desses campos.

## Transação planejada

A sequência de operações de criação futura deve ser:

1. Validar a sessão do jogador.
2. Validar o limite de personagens da conta.
3. Normalizar e validar o nome do personagem.
4. Validar a raça no Rule Registry.
5. Abrir uma transação no banco de dados.
6. Inserir a nova linha na tabela `characters`.
7. Criar e inserir o inventário inicial na tabela `inventories`.
8. Criar qualquer outro estado inicial necessário (ex: skills, quests).
9. Fazer o `COMMIT` da transação.
10. Em caso de qualquer erro, executar um `ROLLBACK` completo.

## Integridade de dados

A implementação deve proteger contra os seguintes riscos de integridade:

- **Personagem sem inventário:** Mitigado por uma transação que engloba a criação do personagem e do seu inventário.
- **Inventário órfão:** Mitigado pela mesma transação.
- **Nome duplicado:** Mitigado por uma constraint `UNIQUE` no banco e tratamento de erro no serviço.
- **Personagem com raça inválida:** Mitigado pela validação no Rule Registry antes do `INSERT`.
- **Falha no meio da criação:** Mitigado pelo `ROLLBACK` transacional.
- **Tentativa de retry duplicado:** Mitigado pela verificação de unicidade de nome.
- **Concorrência de duas criações com mesmo nome:** Mitigado pela constraint `UNIQUE` e pela transação.

## Relação com `LoadCharacter`

- A função `LoadCharacter` **não deve** ser usada para o fluxo oficial de criação de personagem.
- O fallback de auto-criação existente em `LoadCharacter` é considerado um artefato legado e um risco técnico.
- A criação oficial deve ser um fluxo explícito e separado.
- Uma task futura pode ser aberta para remover ou desativar o fallback de `LoadCharacter`.

## Integração futura com protocolo

A implementação futura do fluxo de criação via TCP precisará alterar:

- `backend/pkg/protocol/protocol.go`
- `backend/pkg/protocol/protocol_response_test.go`
- `backend/cmd/gateway/main.go`
- `backend/docs/protocol/protocolo-binario-auth-personagem.md`
- `scripts/BinaryProtocol.cs`
- `scripts/DebugAuthController.cs` (ou a futura UI de criação)

Esta task de planejamento não altera nenhum desses arquivos.

## Testes backend obrigatórios futuros

- Criação de personagem com cada raça válida.
- Tentativa de criação com raça inválida (`ogre`, etc.).
- Tentativa de criação com nome vazio ou inválido.
- Tentativa de criação com nome duplicado.
- Tentativa de criação sem uma sessão autenticada.
- Validação do limite de personagens por conta.
- Simulação de falha na criação do inventário para garantir o rollback do personagem.
- Verificação de que a resposta de sucesso contém `name`, `class`, `level` e `race_id`.
- Garantir que `LoadCharacter` não é invocado como parte do fluxo de criação oficial.

## Segurança MMO

A implementação deve considerar os seguintes riscos:

- Spam de criação de personagens (requer rate limiting).
- Bypass de raças bloqueadas.
- Cliente tentando enviar `class`, `level` ou `stats` ocultos no payload.
- Abuso de economia através do ouro ou inventário inicial.
- "Name squatting" massivo.
- Race conditions na validação de nomes.
- Inconsistência entre o protocolo do backend e a implementação do cliente Godot.
- Ataques de negação de serviço (DoS) por requisições de criação repetidas.

## Fora de escopo

- Não implementar o opcode ou o serviço agora.
- Não alterar o banco de dados.
- Não alterar o cliente Godot.
- Não criar a UI de criação de personagem.
- Não remover o fallback de `LoadCharacter` nesta fase.
- Não implementar a seleção de classe ou elemento.

## Critério de aceite desta task

- O documento `character-creation-backend-service-plan.md` foi criado no caminho correto.
- Nenhuma alteração de código foi realizada.
- O plano deixa claro que a criação será transacional e server-authoritative.
- O plano mantém as regras de que um novo personagem é sempre `novice`, nível 1, sem elemento e sem classe base.