# Character Creation Schema Decision

## 1. Objetivo

Este documento define e decide a estratégia de schema do banco de dados para suportar a futura implementação do fluxo oficial de **criação de personagem**. O foco é detalhar as mudanças necessárias na tabela `characters`, especialmente a adição da coluna `race_id`, e garantir a compatibilidade com dados legados, sem implementar a migration ou alterar o código runtime nesta task.

## 2. Estado atual do schema

A auditoria da Task R1-A confirmou o seguinte estado para a tabela `characters`:

- A tabela já existe e é criada pela migration `backend/migrations/0002_create_characters.up.sql`.
- Ela possui colunas essenciais como `account_id`, `name`, `class`, `level`, `experience`, e `posX`, `posY`, `posZ`.
- Colunas adicionais de combate e progressão (ex: `health`, `mana`, `faction`, `gold`) são garantidas em tempo de execução pelo método `PersistenceManager.InitSchema()`.
- A tabela possui uma coluna legada `element`.
- A auditoria **não confirmou** a existência de uma coluna `race_id`, o que é uma lacuna crítica para o contrato de criação de personagem.
- A tabela `inventories` referencia `characters` através de `character_id`, então criação oficial futura precisa preservar consistência entre personagem e inventário.

## 3. Decisão principal

Fica decidido que:

- Uma coluna `race_id` **será adicionada** à tabela `characters` em uma futura task de migration.
- O tipo da coluna será `VARCHAR(32)` para acomodar os IDs de raça oficiais (ex: "forest_elf").
- Esta coluna armazenará a raça jogável oficial escolhida pelo jogador no momento da criação.
- O valor de `race_id` será **obrigatoriamente validado** pelo backend contra o `Rule Registry` antes de qualquer inserção no banco de dados.
- Raças não-jogáveis como `ogre` serão rejeitadas na camada de serviço e nunca chegarão à camada de persistência.

## 4. Estratégia para dados legados

Personagens existentes, como o de teste "Gabriela", não possuem uma `race_id` definida. A estratégia para lidar com isso deve ser segura e não destrutiva.

**Recomendação:**

- A futura migration deve ser **não destrutiva**. Comandos como `DROP TABLE` são proibidos. A operação deve ser um `ALTER TABLE`.
- O processo **não deve exigir** a recriação de volumes Docker (`docker compose down -v`).
- Para garantir que o login e a seleção de personagens existentes não quebrem, a coluna `race_id` deve ser adicionada de uma forma que não invalide as linhas existentes.

## 5. Estratégia recomendada para `race_id`

Avaliando as opções para a nova coluna `race_id`:

- **Opção A — `race_id VARCHAR(32) NULL`:** Permite que a coluna seja adicionada sem afetar linhas existentes, que permanecerão com o valor `NULL`. A lógica da aplicação teria que tratar o caso nulo.
- **Opção B — `race_id VARCHAR(32) NOT NULL DEFAULT 'human'`:** Adiciona a coluna com uma restrição `NOT NULL` e aplica um valor padrão ('human') a todas as linhas existentes. É mais simples para a aplicação, pois não há nulos para tratar, mas atribui uma raça a personagens legados que pode não ser a "correta".
- **Opção C — Tabela separada `races`:** Cria uma nova tabela e usa uma chave estrangeira. É uma abordagem mais normalizada, mas excessivamente complexa para o estado atual do projeto, onde as raças são um conjunto fixo e pequeno.

**Decisão Recomendada: Opção B Modificada.**

A estratégia mais segura e pragmática é adicionar a coluna com uma restrição `NOT NULL` e um `DEFAULT` seguro.

`ALTER TABLE characters ADD COLUMN IF NOT EXISTS race_id VARCHAR(32) NOT NULL DEFAULT 'human';`

**Justificativa:**
- **Segurança:** Evita valores `NULL` no banco, simplificando a lógica da aplicação que não precisará tratar o caso de um personagem sem raça.
- **Compatibilidade:** Personagens legados como "Gabriela" receberão a raça 'human' por padrão, o que é aceitável para um personagem de teste, e continuarão funcionando.
- **Simplicidade e Avanço Incremental:** É uma mudança atômica e simples que resolve o problema imediato sem introduzir complexidade desnecessária.

## 6. Decisão sobre `class`

A coluna `class` já existe. A auditoria notou um risco de mismatch entre o valor legado `Novice` (maiúsculo) e o valor oficial `novice` (minúsculo) da constante `rules.StartingClassNovice`.

**Decisão:**
- A futura implementação do serviço de criação de personagem deve **obrigatoriamente** usar a constante `rules.StartingClassNovice` para garantir a normalização para `novice` em todas as novas criações.
- Nenhuma alteração será feita nos dados legados (`Novice`) nesta fase. Uma futura task de limpeza de dados pode ser planejada, se necessário.

## 7. Decisão sobre `element`

A coluna `element` existe como um artefato legado. O contrato de criação proíbe a escolha de elemento.

**Decisão:**
- A futura implementação do serviço de criação deve **ignorar** qualquer `element_id` enviado pelo cliente.
- Ao criar um novo personagem, o backend deve inserir um valor padrão e neutro para a coluna `element`, como `'None'` ou `''`, para satisfazer o schema, mas sem atribuir poder elemental.
- A rejeição de `air` e `water` será garantida pela camada de serviço, que não permitirá que esses valores cheguem à persistência.

## 8. Decisão sobre inventário inicial

A tabela `inventories` já existe e está vinculada a `characters`.

**Decisão:**
- A futura implementação do método `CreateCharacter` no `PersistenceManager` deve ocorrer dentro de uma **transação de banco de dados**.
- A transação deve incluir a inserção na tabela `characters` e a inserção de todos os itens do inventário inicial na tabela `inventories`.
- Se a inserção do inventário falhar por qualquer motivo, a transação inteira deve sofrer `ROLLBACK`, garantindo que nenhum personagem "órfão" sem inventário seja criado.

## 9. Decisão sobre `LoadCharacter`

O fallback de criação implícita em `LoadCharacter` é um risco.

**Decisão:**
- O fluxo oficial de criação de personagem **não usará** `LoadCharacter`. Um método explícito será criado.
- O fallback em `LoadCharacter` **não será removido nesta fase** para não quebrar o fluxo de desenvolvimento local.
- Uma task de "dívida técnica" deve ser criada para, no futuro, refatorar `LoadCharacter` para que ele apenas carregue personagens existentes, retornando um erro se o personagem não for encontrado.

## 10. Futuro arquivo de migration

A futura task de migration (R1-D) deverá criar um arquivo SQL com a seguinte intenção:

```sql
-- Intenção para a futura migration
-- Adiciona a coluna race_id se ela não existir, com um default seguro.
ALTER TABLE characters ADD COLUMN IF NOT EXISTS race_id VARCHAR(32) NOT NULL DEFAULT 'human';

-- Considerar adicionar um índice no futuro, se as buscas por raça se tornarem comuns.
-- CREATE INDEX IF NOT EXISTS idx_characters_race_id ON characters(race_id);
```

## 11. Futuras alterações em `InitSchema`

O método `PersistenceManager.InitSchema()` também garante colunas em tempo de execução. A futura task de implementação (`R1-D` ou `R1-F`) deverá atualizar este método para também garantir a existência da coluna `race_id`, mantendo a consistência com a migration.

## 12. Futuras alterações em `LoadCharacter`/`ListCharacters`

- `LoadCharacter`: Deverá ser atualizado para ler a nova coluna `race_id` e populá-la na struct de stats do personagem.
- `ListCharactersByAccount`: Poderá ser atualizado para incluir `race_id` na `CharacterSummary`, se a UI de seleção de personagem precisar exibir essa informação. Isso pode exigir uma mudança de protocolo coordenada.

## 13. Critérios de aceite para a futura migration

- A migration não deve apagar dados existentes.
- O fluxo de login/list/select para personagens legados (como "Gabriela") deve continuar funcionando.
- A coluna `race_id` deve ser adicionada com sucesso à tabela `characters`.
- Todos os testes de persistência existentes devem passar após a migration.
- O backend deve iniciar corretamente com `docker compose up` após a migration ser aplicada.

## 14. Riscos e mitigação

- **Perda de dados legados:** Mitigada pelo uso de uma migration não destrutiva com `ALTER TABLE`.
- **Quebra de personagem legado:** Mitigada pela aplicação de um default seguro, garantindo que personagens existentes continuem logando/listando/selecionando.
- **`race_id` ausente:** Mitigado pela criação de uma task de migration dedicada como pré-requisito para a implementação do serviço.
- **Default errado:** Mitigado por decisão explícita de usar apenas raça oficial e por validação posterior via Rule Registry.
- **Aceitar `ogre`:** Mitigado por validação server-authoritative no serviço de criação antes da persistência.
- **Mismatch `Novice`/`novice`:** Mitigado pela normalização para `novice` na nova lógica de criação, isolando o dado legado.
- **`element` legado:** Mitigado pela regra de que criação oficial não aceita elemento do client e por normalização futura planejada.
- **Protocolo/Godot afetados indevidamente:** Mitigado por manter esta task restrita a documentação/schema decision, sem alterar protocolo ou client.
- **Migration e `InitSchema` divergentes:** Mitigado por exigir que a futura task de schema atualize migration e `PersistenceManager.InitSchema()` de forma consistente.
- **Reset acidental de banco local:** Mitigado pela proibição explícita de `docker compose down -v`.

## 15. O que esta task NÃO faz

- Não cria o arquivo de migration SQL.
- Não altera o banco de dados.
- Não altera o código Go do backend.
- Não altera o protocolo, o cliente ou o comportamento runtime do jogo.

## 16. Status

Este documento estabelece as decisões técnicas de schema necessárias para a implementação da criação de personagem. Ele está pronto para orientar a próxima sub-task do roadmap, R1-D, que será a criação segura da migration/schema real para `race_id`.
