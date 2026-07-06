# Character Creation Runtime Integration Plan

## 1. Objetivo

Este documento transforma a auditoria técnica da Task R1-A em um plano de implementação detalhado para a futura integração do fluxo oficial de **criação de personagem** no runtime do backend. O objetivo é fornecer um guia técnico claro, seguro e sequencial para os desenvolvedores, garantindo que a implementação final esteja em total conformidade com o `character-creation-rule-contract.md` e a arquitetura server-authoritative do projeto.

## 2. Decisão técnica principal

A criação oficial de personagem deve ser um fluxo de primeira classe, explícito e separado do carregamento de personagem.

- **`LoadCharacter`:** A função `persistence.LoadCharacter` deve, em seu estado final, ser responsável **apenas** por carregar um personagem *existente*.
- **Criação Explícita:** A criação de um novo personagem deve ocorrer através de um novo método/handler dedicado (ex: `CreateCharacter`).
- **Fallback Legado:** O comportamento de criação implícita (fallback) dentro de `LoadCharacter` é considerado um risco técnico e um artefato de desenvolvimento inicial. Ele **não deve** ser usado para a criação oficial de personagem e deve ser auditado e potencialmente removido ou refatorado em uma task futura, separada da implementação inicial de criação.

## 3. Estado atual resumido

A auditoria da Task R1-A confirmou que:

- Existem fluxos de **listagem e seleção** de personagem via protocolo TCP.
- O `PersistenceManager` e as migrations para as tabelas `characters` e `inventories` já existem.
- **Não existe** um fluxo ou opcode oficial para **criação de personagem** exposto ao cliente.
- A coluna `race_id` provavelmente está ausente do schema atual da tabela `characters`, sendo um ponto crítico para a implementação.

## 4. Arquitetura proposta

A futura implementação seguirá uma arquitetura em camadas para garantir desacoplamento e segurança:

1.  **Gateway (Handler):** O `backend/cmd/gateway` receberá a intenção de criação do cliente através de um novo opcode de protocolo. Ele será responsável por autenticar a sessão e decodificar a requisição mínima.
2.  **Serviço (Service):** Uma nova camada de serviço (ex: `backend/pkg/character/service.go`) orquestrará a lógica de negócio. Ela validará o nome, chamará o `Rule Registry` para validar a raça e definirá todos os atributos iniciais autoritativos (`novice`, nível 1, stats, etc.).
3.  **Rule Registry:** O serviço de criação usará o catálogo oficial do `backend/pkg/gamedata/rules` para garantir que a `race_id` é válida e que nenhuma regra de design é violada.
4.  **Persistência (Persistence):** O `PersistenceManager` terá um novo método explícito para inserir o personagem e seu inventário inicial em uma única transação no banco de dados.
5.  **Resposta:** O Gateway enviará uma resposta de sucesso ou erro ao cliente. Após a criação, o fluxo de `CS_CHAR_LIST_REQUEST` deverá incluir o novo personagem.

## 5. Request mínima proposta

A futura requisição de criação de personagem enviada pelo cliente deve conter **apenas** o essencial:

- `character_name` (string)
- `race_id` (string, ex: "human")

Os seguintes campos são **explicitamente proibidos** de serem enviados pelo cliente na requisição de criação, e qualquer tentativa de enviá-los deve ser ignorada ou rejeitada pelo backend:

- `class`
- `level`
- `element`
- `stats` (HP, Mana, etc.)
- `inventory`
- `gold`
- `position` (posX, posY, posZ)
- `faction`
- `skill XP`
- `affinity`
- `subclass`

## 6. Schema/banco

A tabela `characters` já existe, mas a auditoria não confirmou a presença de uma coluna `race_id`.

**Recomendação:** Uma futura task de migration (`R1-C`) deve ser criada para adicionar a coluna `race_id` à tabela `characters`.

- A migration deve ser não destrutiva (`ALTER TABLE ... ADD COLUMN IF NOT EXISTS`).
- Ela deve definir um valor padrão (`DEFAULT`) ou permitir nulos para garantir retrocompatibilidade com personagens existentes (como "Gabriela"), que não possuem uma raça definida.
- A estratégia deve evitar a necessidade de `docker compose down -v`, preservando o estado do banco de dados de desenvolvimento.

## 7. Protocolo

O protocolo atual não possui um opcode para criação de personagem.

**Recomendação:**

- **Opção A (Recomendada):** Criar um novo opcode TCP oficial para a criação de personagem (ex: `CS_CHAR_CREATE_REQUEST` e `SC_CHAR_CREATE_RESPONSE`). Esta é a abordagem mais limpa e segura, separando claramente as intenções do cliente.
- **Opção B (Não Recomendada):** Adiar a criação de um fluxo via TCP e focar apenas na implementação do serviço e persistência no backend, para ser usado por ferramentas administrativas. Esta opção não entrega a funcionalidade ao jogador.

Se a Opção A for escolhida, uma futura task de protocolo (`R1-E`) precisará atualizar os seguintes arquivos:

- `backend/pkg/protocol/protocol.go`
- `backend/pkg/protocol/protocol_response_test.go`
- `backend/docs/protocol/protocolo-binario-auth-personagem.md`
- Código do cliente Godot C# que lida com o protocolo.

## 8. PersistenceManager

**Recomendação:** Uma futura task de persistência (`R1-D`) deve criar um novo método explícito no `PersistenceManager`, como:

`CreateCharacter(ctx context.Context, accountID int, name string, raceID string, initialStats CharacterStats, initialInventory []*Item) (*Character, error)`

Este método deve:
- Receber o `account_id` autenticado, não um valor fixo.
- Iniciar uma transação no banco de dados.
- Inserir a nova linha na tabela `characters` com os dados validados pelo serviço.
- Inserir as linhas correspondentes ao inventário inicial na tabela `inventories`.
- Fazer `ROLLBACK` da transação se a inserção do inventário falhar, para evitar um personagem "órfão".
- Utilizar a constraint `UNIQUE` da coluna `name` para rejeitar nomes duplicados de forma atômica.

## 9. LoadCharacter

**Recomendação:**

- O fallback de criação implícita dentro de `LoadCharacter` **não deve ser removido nesta fase inicial** para não quebrar o fluxo de desenvolvimento local existente.
- Uma task futura de refatoração técnica deve ser criada para auditar e remover esse fallback, fazendo com que `LoadCharacter` retorne um erro (`sql.ErrNoRows`) se o personagem não for encontrado, como esperado de uma função de carregamento.

## 10. Rule Registry

O novo serviço de criação de personagem deve integrar as regras da seguinte forma:

- Carregar/consultar o catálogo oficial do Rule Registry usando os nomes reais existentes no código. Antes da implementação, confirmar se a integração correta será via `Registry`, helper de catálogo, função específica de raça ou outro ponto já existente.
- Validar a `race_id` recebida do cliente contra a categoria oficial de raças do Rule Registry, usando os nomes reais de tipos/categorias existentes no código. Não inventar helper ou categoria inexistente durante a implementação.
- Rejeitar explicitamente `ogre`.
- Definir a classe inicial autoritativamente usando a constante `rules.StartingClassNovice`.
- Rejeitar qualquer tentativa de definir um elemento na criação.

## 11. Fluxo futuro recomendado

1.  O cliente, após autenticação, envia uma intenção de criação (`CS_CHAR_CREATE_REQUEST`) com `character_name` e `race_id`.
2.  O Gateway valida a sessão e decodifica a requisição.
3.  O Gateway chama o serviço de criação de personagem.
4.  O serviço valida o nome (formato, disponibilidade).
5.  O serviço valida a `race_id` contra o `Rule Registry`.
6.  O serviço define autoritativamente todos os atributos iniciais: `class="novice"`, `level=1`, stats, spawn, inventário.
7.  O serviço chama o `PersistenceManager` para inserir o personagem e seu inventário em uma única transação.
8.  Se a persistência for bem-sucedida, o serviço retorna sucesso.
9.  O Gateway envia uma resposta de sucesso (`SC_CHAR_CREATE_RESPONSE`) ao cliente.
10. O cliente pode então solicitar a lista de personagens atualizada, que agora incluirá o novo personagem.

## 12. Testes obrigatórios futuros

- Testes de integração para criar um personagem com cada uma das 5 raças oficiais.
- Teste para rejeitar a criação com a raça `ogre`.
- Teste para rejeitar a criação com uma raça inexistente.
- Teste para rejeitar uma requisição que tente enviar uma classe inicial.
- Teste para garantir que todo personagem criado tenha a classe `novice` e nível `1`.
- Teste para garantir que o `account_id` do novo personagem corresponde ao da sessão autenticada.
- Teste para rejeitar a criação com um nome duplicado.
- Teste para simular uma falha na inserção do inventário e garantir que o personagem não seja criado (rollback).

## 13. Ordem de implementação recomendada

A implementação da Fase R1 deve ser dividida em sub-tasks para manter o escopo gerenciável:

- **R1-C:** Planejamento e execução da migration do banco de dados (adicionar `race_id`).
- **R1-D:** Implementação do método `CreateCharacter` no `PersistenceManager` e seus testes unitários.
- **R1-E:** Definição e implementação do novo opcode de protocolo (se a Opção A for escolhida).
- **R1-F:** Implementação do serviço de criação e do handler no Gateway.
- **R1-G:** Implementação dos testes de integração de ponta a ponta.
- **R1-H:** Integração no cliente Godot (somente após o backend estar completo e validado).

## 14. Arquivos candidatos para futura implementação

- **Migrations:** `backend/migrations/` (novo arquivo de migration).
- **Persistence:** `backend/pkg/persistence/persistence_manager.go`.
- **Gateway:** `backend/cmd/gateway/main.go` e um novo arquivo em `backend/cmd/gateway/handlers/`.
- **Service:** Novo pacote e arquivo, ex: `backend/pkg/character/service.go`.
- **Protocol:** `backend/pkg/protocol/protocol.go` e `backend/pkg/protocol/protocol_response_test.go`.
- **Tests:** Novos arquivos de teste de integração.
- **Docs:** `backend/docs/protocol/protocolo-binario-auth-personagem.md`.
- **Client:** Código C# do cliente Godot.

## 15. Critérios de aceite para implementação futura

A integração da criação de personagem será considerada concluída quando:

- Um jogador puder criar um personagem com uma raça válida através do cliente.
- Todas as regras do contrato de criação forem aplicadas e validadas no backend.
- Personagens com raças, classes ou atributos iniciais inválidos forem consistentemente rejeitados.
- O novo personagem for persistido corretamente no banco de dados e aparecer na lista de personagens.
- Todos os novos testes de integração passarem.

## 16. Riscos e mitigação

- **Client Authority:** Mitigado pela arquitetura server-authoritative, onde o backend ignora o estado enviado pelo cliente e valida tudo.
- **`race_id` Ausente:** Mitigado pela criação de uma task de migration dedicada antes da implementação do runtime.
- **Protocolo quebrando Godot:** Mitigado por task própria de protocolo, atualizando backend, testes, documentação e client Godot C# em conjunto.
- **Fallback de `LoadCharacter`:** Mitigado pela criação de um fluxo de criação explícito e separado. O risco legado será tratado em uma task de refatoração futura.
- **Rollback Incompleto:** Mitigado pelo uso de transações de banco de dados no método futuro de criação.
- **Duplicidade de nome:** Mitigada por constraint única no banco e tratamento explícito de erro no backend.
- **Dados Legados:** A migration não destrutiva garante que personagens de teste existentes não sejam perdidos.
- **Mismatch `Novice` versus `novice`:** Mitigado por normalização futura usando a constante oficial do Rule Registry e testes de compatibilidade com dados legados.
- **`element` legado versus ausência de elemento na criação:** Mitigado por regra explícita de que criação oficial não aceita elemento do client e por plano futuro de normalização do campo legado.

## 17. O que esta task NÃO faz

- Não implementa a criação de personagem.
- Não altera o banco de dados, protocolo ou cliente.
- Não cria uma migration.
- Não remove o fallback de `LoadCharacter`.
- Não altera o comportamento atual do jogo.

## 18. Status

Este documento é o plano técnico oficial para orientar as próximas sub-tasks de implementação da criação de personagem. Ele fornece uma estratégia clara e segura para integrar a primeira regra do Rule Registry ao runtime do jogo.

