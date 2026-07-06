# Class Selection Runtime Integration Plan

## 1. Objetivo

Este documento planeja a futura integração da regra de negócio de **seleção de classe base** no ambiente de execução (runtime) do backend. Ele detalha como a função pura `rules.CanSelectBaseClass` será utilizada para validar as intenções dos jogadores de forma segura e consistente.

## 2. Estado atual

- A regra pura de seleção de classe já existe e está testada em `backend/pkg/gamedata/rules/class_selection.go`.
- Os testes unitários que validam a lógica da regra em isolamento já existem em `class_selection_test.go`.
- **Nenhuma integração runtime foi feita ainda.** Nenhum sistema de jogo (progressão, protocolo, etc.) invoca a função `CanSelectBaseClass`.
- Esta task é exclusivamente de documentação e não altera o comportamento do jogo.

## 3. Princípio server-authoritative

A integração deve seguir rigorosamente a arquitetura server-authoritative:

- O **cliente** apenas envia uma **intenção** de escolher uma classe (ex: "quero me tornar um `knight`").
- O **backend** carrega o estado **real e autoritativo** do personagem (nível e classe atual) diretamente de sua fonte da verdade (memória ou banco de dados).
- O **backend** valida a intenção usando `CanSelectBaseClass`, passando os dados reais do personagem, não os dados enviados pelo cliente.
- O **backend** persiste a mudança no banco de dados somente após a validação bem-sucedida.
- O cliente **nunca** define a classe de um personagem diretamente.

## 4. Regra oficial de seleção de classe

A lógica de negócio, já codificada no Rule Registry, é a seguinte:

- Todo personagem começa sua jornada como `novice`.
- O nível mínimo para escolher uma classe base é **10**.
- A classe base só pode ser escolhida **uma vez**. A regra valida se a classe atual do personagem ainda é `novice`.
- A classe alvo (`target_class`) deve ser uma das classes base oficiais.
- Após um personagem se tornar um `knight`, `mage`, `archer`, `assassin` ou `cleric`, ele não pode usar esta regra para trocar de classe novamente.

## 5. Classes base permitidas

A classe alvo da seleção deve ser uma das seguintes:

- `knight`
- `mage`
- `archer`
- `assassin`
- `cleric`

Qualquer outro valor enviado pelo cliente como classe alvo deve ser **rejeitado**. Isso inclui, mas não se limita a:

- `novice` (não se pode "escolher" ser um novice)
- `ogre`
- `air`
- `water`
- `paladin`
- `unknown`
- String vazia (`""`)

## 6. Fluxo runtime futuro recomendado

O processo de ponta a ponta para a seleção de classe deve ser:

1.  O cliente, ao detectar que o jogador atingiu o nível 10, exibe a UI de seleção de classe e envia uma intenção para o backend (ex: `ChooseBaseClassRequest{TargetClass: "knight"}`).
2.  O backend recebe a requisição e autentica a sessão do jogador.
3.  O backend identifica o `account_id` da sessão e o `character_id` ativo.
4.  O backend carrega o estado completo e autoritativo do personagem do banco de dados ou de um cache seguro.
5.  O backend verifica se o personagem carregado pertence de fato à conta da sessão.
6.  O backend lê o nível atual **real** do personagem (`character.Level`).
7.  O backend lê a classe atual **real** do personagem (`character.ClassID`).
8.  O backend extrai a classe alvo (`target_class`) da intenção do cliente.
9.  O backend invoca a regra: `err := rules.CanSelectBaseClass(character.Level, character.ClassID, target_class)`.
10. Se `err` não for `nil`, o backend rejeita a requisição e envia uma resposta de erro ao cliente, sem alterar nenhum estado.
11. Se `err` for `nil`, o backend inicia uma transação para persistir a nova classe no banco de dados (ex: `UPDATE characters SET class_id = ? WHERE id = ? AND class_id = 'novice'`).
12. O backend recalcula stats derivados (HP, Mana, etc.) que podem mudar com a nova classe.
13. O backend envia uma resposta de sucesso ao cliente.
14. Se o personagem estiver online, o backend atualiza seu estado em memória no servidor do mundo (world server).

## 7. Ordem de validação recomendada

O handler que processa a requisição deve seguir esta ordem de validação:

1.  A sessão do jogador é válida?
2.  O personagem associado à sessão existe?
3.  O personagem pertence à conta da sessão?
4.  O personagem está em um estado que permite interações (ex: não está em meio a outro processo crítico)?
5.  A requisição do cliente contém uma `target_class`?
6.  Chamar `rules.CanSelectBaseClass` com os dados **do servidor**.
7.  Se a validação passar, persistir a mudança em uma transação.
8.  Atualizar o estado do personagem em memória.
9.  Responder ao cliente.

## 8. Regras de rejeição obrigatórias

A seleção de classe deve falhar se:

- A sessão do jogador for inválida.
- O personagem não for encontrado.
- O personagem não pertencer à conta que fez a requisição.
- O personagem tiver nível menor que 10.
- A classe atual do personagem **não** for `novice`.
- A `target_class` for inválida (ex: `ogre`, `paladin`).
- A `target_class` for `novice`.
- O cliente tentar enviar um nível ou classe atual falsos na requisição (esses dados devem ser ignorados e lidos do estado do servidor).

## 9. Impacto futuro em banco

- A tabela `characters` precisará de um campo `class_id` (ou similar).
- A operação de `UPDATE` para alterar a classe deve ser transacional.
- Para prevenir race conditions (ex: duas requisições simultâneas), a query de `UPDATE` deve incluir uma condição `WHERE class_id = 'novice'`. Isso garante que a troca só possa ocorrer uma vez, de forma atômica.
- Um log de auditoria deve ser gerado para registrar a data, hora, personagem e a classe escolhida.

## 10. Impacto futuro em protocolo/API

- Uma nova mensagem de requisição (ex: `ChooseBaseClassRequest`) será necessária.
- Esta requisição deve conter **apenas** o `target_class`. Ela **não deve** conter o nível ou a classe atual do personagem, pois esses dados são autoritativos do servidor.
- O backend deve ter mensagens de resposta para sucesso e para os diferentes tipos de erro (`ErrClassSelectionLevelTooLow`, `ErrClassSelectionAlreadyChosen`, etc.).
- Qualquer mudança no protocolo deve ser documentada em um `protocol-change-plan.md` antes da implementação.

## 11. Impacto futuro no client

- O cliente deve ter uma UI para a seleção de classe, que idealmente se torna visível quando o personagem atinge o nível 10.
- A UI pode listar as classes base oficiais, obtendo esses dados de uma cópia read-only do Rule Registry.
- O cliente pode desabilitar o botão de confirmação se os requisitos não forem atendidos (apenas para melhorar a experiência do usuário - UX), mas a validação final e decisiva é sempre do backend.
- O cliente deve ser capaz de tratar as respostas de sucesso e erro do backend, exibindo feedback apropriado ao jogador.

## 12. Impacto futuro em stats/progression

- A mudança de `novice` para uma classe base provavelmente alterará os stats base ou derivados do personagem.
- O recálculo de HP, Mana, e outros atributos deve ser feito **exclusivamente no backend** após a persistência da nova classe.
- A nova classe deve liberar o acesso a novas árvores de skills ou habilidades. Essa liberação deve ser controlada pelo backend.

## 13. Riscos que o plano previne

- **Seleção Prematura:** Impede que um jogador escolha uma classe antes do nível 10.
- **Troca de Classe Indevida:** Impede que um jogador que já é `knight` tente se tornar um `mage`.
- **Manipulação de Dados:** Impede que um cliente malicioso forje seu nível ou classe atual para enganar o sistema.
- **Seleção de Classe Inválida:** Impede a escolha de classes que não existem ou não são permitidas, como `paladin` ou `ogre`.
- **Race Conditions:** A abordagem transacional com `WHERE class_id = 'novice'` previne que o jogador consiga escolher duas classes em requisições simultâneas.

## 14. Testes futuros recomendados

A futura implementação deve incluir testes de integração que cubram:

- Um personagem de nível 9 tentando escolher uma classe (deve ser rejeitado).
- Um personagem de nível 10 com a classe `novice` (deve ser aceito).
- Um personagem de nível 11 com a classe `knight` tentando escolher `mage` (deve ser rejeitado).
- A escolha de cada uma das 5 classes base oficiais (deve ser aceita).
- A tentativa de escolher `novice`, `ogre`, `air`, `water` ou `paladin` (deve ser rejeitada).
- Uma tentativa de escolher uma classe para um personagem que não pertence à conta da sessão (deve ser rejeitada).
- Testes de concorrência para garantir que duas requisições simultâneas não resultem em um estado inconsistente.

## 15. O que esta task NÃO faz

- Não cria código Go, C# ou qualquer outra linguagem.
- Não altera o banco de dados.
- Não altera o protocolo de rede.
- Não altera o cliente.
- Não cria um endpoint de API.
- Não implementa a lógica de seleção de classe.

## 16. Status

Este documento é o plano técnico oficial e preparatório para a futura integração do sistema de seleção de classe base. Ele garante que a implementação será segura, robusta e alinhada com a arquitetura server-authoritative do projeto.
