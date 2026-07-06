# Light and Shadow — Character Creation Persistence Runtime Verification

## Status

- Verificação
- Pós R1-L-B
- Criação ainda não exposta ao Gateway
- Criação ainda não exposta ao Godot

## Contexto

A task anterior (R1-L-B) adicionou a seguinte função skeleton na camada de persistência:

`CreateCharacterForAccount(ctx context.Context, accountID int, desiredName string, raceID string)`

Esta função:
- Valida `accountID`, `desiredName` e `raceID` de forma mínima.
- Abre uma transação no banco de dados.
- Insere um novo personagem na tabela `characters` com `class="novice"`, `level=1` e o `race_id` explícito.
- Cria um inventário inicial padrão.
- Insere os itens do inventário na tabela `inventories` dentro da mesma transação.
- Retorna um `CharacterSummary` em caso de sucesso.
- **Ainda não é chamada** por nenhum handler do Gateway ou pelo cliente Godot.

## Validações executadas

Para verificar a estabilidade e a não-regressão, os seguintes comandos foram executados:

- `go test ./pkg/persistence/...`
- `go test ./pkg/...`
- `docker pull alpine:3.18`
- `docker compose up --build -d`
- `docker logs ls_gateway`
- Consulta direta na tabela `characters` do PostgreSQL.

## Resultado dos testes Go

- O pacote `backend/pkg/persistence` compilou com sucesso.
- Os testes dos pacotes `protocol`, `gamedata/rules` e outros continuaram passando.
- O comando `go test ./pkg/...` foi concluído com sucesso.
- Nenhum erro de build foi introduzido pela nova função `CreateCharacterForAccount`.

## Resultado Docker

- A falha externa `502 Bad Gateway` do Docker Hub durante o build Docker do backend foi resolvida com um `docker pull alpine:3.18` prévio.
- As imagens do backend (`auth-server`, `world-server`, `gateway`) foram buildadas com sucesso.
- Os containers `postgres`, `redis`, `auth-server`, `world-server` e `gateway` iniciaram corretamente.
- O log do `ls_gateway` confirmou a conexão com o PostgreSQL e Redis.
- O log `PostgreSQL schema validated and upgraded successfully` foi observado.
- O Gateway TCP permaneceu escutando na porta `8080`.
- O endpoint de health HTTP permaneceu funcional na porta `9080`.

## Resultado banco

A seguinte consulta foi executada no banco de dados:

`SELECT id, name, class, level, race_id FROM characters ORDER BY id ASC;`

O resultado confirmou que o personagem de teste existente, "Gabriela", permaneceu válido e inalterado:

- `class`: `Novice`
- `level`: `1`
- `race_id`: `human`

## Não-regressão

A implementação da R1-L-B não alterou:

- O código do Gateway.
- O código do cliente Godot.
- O protocolo de rede runtime existente.
- O fluxo de listagem de personagens.
- O fluxo de seleção de personagens.
- As migrations do banco de dados.
- O Rule Registry.

## Limitação importante

Esta verificação **NÃO** prova a funcionalidade completa da criação de personagem. Os seguintes pontos ainda não foram implementados ou testados:

- Criação real de personagem via opcode `1008`.
- Validação completa de nome (tamanho, caracteres, palavras reservadas).
- Validação completa de `race_id` contra o Rule Registry.
- Validação de limite de personagens por conta.
- Teste de rollback em caso de falha na criação do inventário.
- Integração com o cliente Godot.
- UI de criação de personagem.

## Conclusão

O skeleton da função de persistência `CreateCharacterForAccount` está compilável, estável e não introduziu regressões no runtime existente do jogo. A base para a criação transacional de personagem está pronta, mas ainda precisa de uma camada de serviço, validação e integração com o Gateway para se tornar funcional.

## Próxima task recomendada

**R1-M-A — Document Character Creation Service Validation Plan**

**Objetivo futuro:** Planejar a camada de serviço que ficará entre o Gateway e a persistência. Este plano detalhará como as validações de nome, `race_id` (via Rule Registry) e limite de personagens serão implementadas de forma segura antes de expor o opcode ao cliente.

## Fora de escopo

- Não implementar o handler no Gateway.
- Não implementar a UI no Godot.
- Não criar um personagem via cliente.
- Não remover o fallback de `LoadCharacter`.
- Não criar novas migrations.
- Não alterar o Rule Registry.