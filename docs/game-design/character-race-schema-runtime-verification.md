# Character Race Schema Runtime Verification

## 1. Objetivo

Este documento registra a validação runtime da Task R1-E após a implementação da Task R1-D — Add Character Race Schema.

O objetivo foi confirmar que a adição de race_id ao schema de personagens não quebrou o backend, o banco local, o fluxo de login/list/select, o carregamento da personagem legada Gabriela, o inventário, os chunks ou o movimento debug no client Godot.

## 2. Estado validado

Commit validado:

- 9e6a703 Add character race schema

Arquivos alterados pela task anterior:

- backend/migrations/0011_add_character_race_id.up.sql
- backend/pkg/persistence/persistence_manager.go

## 3. Backend Docker

O backend foi iniciado com docker compose up --build.

Resultado observado:

- auth-server buildou e iniciou corretamente.
- world-server buildou e iniciou corretamente.
- gateway-server buildou e iniciou corretamente.
- PostgreSQL iniciou corretamente.
- Redis iniciou corretamente.
- Gateway TCP ficou disponível em 0.0.0.0:8080.
- Gateway Health ficou disponível em 9080.
- O log confirmou: PostgreSQL schema validated and upgraded successfully.

O aviso the attribute version is obsolete foi observado e continua não bloqueador.

## 4. Banco PostgreSQL

Banco correto identificado:

- light_and_shadow

A tabela characters foi inspecionada com psql.

Resultado confirmado:

- coluna race_id existe;
- tipo: character varying(32);
- nullable: not null;
- default: 'human'::character varying.

Consulta de personagens confirmou:

- Gabriela ainda existe;
- account_id = 1;
- class = Novice;
- level = 1;
- race_id = human.

Isso confirma que dados legados foram preservados e que não houve reset destrutivo de banco.

## 5. Client Godot

O client Godot debug foi executado contra o backend Docker.

Resultado observado na tela Debug World Entry:

- IsAuthenticated = True;
- IsCharacterSelected = True;
- AccountId = 1;
- SelectedCharacterName = Gabriela;
- Inv. Sync Received = True;
- Level = 1;
- HP = 600/600;
- Mana = 100/100;
- Chunks Received = 9;
- movimento debug confirmado com sucesso.

Isso confirma que race_id não quebrou login, listagem/seleção, load de personagem, inventário, chunks ou movimento.

## 6. Observação técnica sobre índice

A migration 0011_add_character_race_id.up.sql contém intenção de criar o índice idx_characters_race_id.

Porém, ao consultar pg_indexes, o índice idx_characters_race_id não apareceu no banco local atual.

Conclusão:

- a coluna race_id foi garantida por PersistenceManager.InitSchema();
- a migration SQL 0011 não parece ter sido executada automaticamente neste banco local existente;
- isso não bloqueia o runtime atual, pois ainda não existem buscas por raça;
- deve ser criada uma task pequena para alinhar InitSchema com o índice ou revisar o mecanismo de execução de migrations.

## 7. Critérios validados

- Backend compila e sobe com Docker.
- PostgreSQL inicia com banco existente.
- InitSchema valida e atualiza schema com sucesso.
- characters.race_id existe.
- race_id é NOT NULL.
- race_id possui default human.
- Gabriela foi preservada.
- Login/list/select/load continuam funcionando.
- Inventário é sincronizado.
- Chunks são recebidos.
- Movimento debug funciona.
- Nenhum reset de volume foi necessário.
- docker compose down -v não foi usado.

## 8. Riscos restantes

- idx_characters_race_id ainda não existe no banco local atual.
- Migration SQL e InitSchema ainda não estão totalmente equivalentes quanto ao índice.
- O campo race_id ainda não é carregado por LoadCharacter.
- O campo race_id ainda não aparece em CharacterSummary.
- Ainda não existe criação oficial de personagem.
- Ainda não existe validação runtime de raça pelo Rule Registry.
- Ainda não existe opcode de criação de personagem.

## 9. Recomendação

A validação runtime da Task R1-E está aprovada.

A próxima task recomendada é Task R1-F — Align Race Schema Index in InitSchema.

Objetivo da próxima task:

- garantir idx_characters_race_id também pelo InitSchema;
- manter migration e schema runtime alinhados;
- não alterar protocolo;
- não alterar Godot;
- não implementar criação de personagem ainda.

## 10. Status

Task R1-E validada com sucesso.

A alteração de schema race_id está funcional em runtime e preserva o fluxo atual do jogo.
