# Character Creation Rule Contract

## 1. Objetivo

Este documento define o contrato técnico e as regras de negócio obrigatórias para a implementação do sistema de **criação de personagem**. Ele serve como um guia para a futura integração entre os serviços de runtime do backend e o **Rule Registry** já existente.

O objetivo é garantir que a criação de personagem seja um processo seguro, determinístico e 100% server-authoritative, prevenindo exploits e inconsistências desde o primeiro momento da jornada do jogador.

## 2. Estado atual

- O **Rule Registry** (`backend/pkg/gamedata/rules`) já está implementado e testado, contendo as regras puras do jogo.
- O catálogo oficial de raças, classes e elementos (`catalog.go`) já existe.
- A regra de que um personagem começa como `novice` e só escolhe a classe base no level 10 (`class_selection.go`) já existe.
- **Nenhuma integração runtime foi feita ainda.** O sistema de criação de personagem atual (se houver) não consome estas regras.
- Esta task é exclusivamente de documentação e não altera o comportamento do jogo.

## 3. Princípio server-authoritative

A criação de personagem deve seguir rigorosamente a arquitetura server-authoritative do projeto:

- O **cliente** apenas envia uma **intenção** de criar um personagem com dados mínimos (ex: nome, raça).
- O **backend** é responsável por **validar 100%** dos dados recebidos contra as regras do Rule Registry.
- O **backend** tem a autoridade final para **decidir** se a criação é permitida.
- O **backend** define todos os atributos iniciais do personagem (classe, nível, stats, inventário, posição).
- O **backend** persiste o novo personagem no banco de dados somente após a validação completa.
- O cliente **nunca** define o estado autoritativo de um novo personagem.

## 4. Contrato de criação de personagem

Uma futura requisição de criação de personagem, seja via API ou protocolo de rede, deverá conter conceitualmente os seguintes campos enviados pelo cliente:

- **Sessão/Conta Autenticada:** A identidade do jogador que está criando o personagem.
- `character_name`: O nome desejado para o personagem.
- `race_id`: O ID da raça escolhida (ex: `"human"`, `"forest_elf"`).
- **Aparência (Futuro):** Dados de customização visual, se o sistema existir.

A requisição do cliente **NÃO DEVE** conter:

- `class_id`: O jogador não escolhe a classe na criação.
- `element_id`: O jogador não escolhe um elemento na criação.

O backend, ao processar uma requisição válida, deverá definir os seguintes atributos iniciais de forma autoritativa:

- `class_id`: Obrigatoriamente definido como `rules.StartingClassNovice` (`"novice"`).
- `level`: Obrigatoriamente definido como `1`.
- **Afinidades Elementais:** Devem começar vazias ou com valor `0`.
- **Posição Inicial:** A coordenada de spawn deve ser definida pelo backend (ex: ponto de spawn de Ironhold Bastion).
- **Stats Iniciais:** HP, Mana e outros stats base devem ser definidos pelo backend, conforme a raça e a classe `novice`.
- **Inventário Inicial:** Itens iniciais (se houver) devem ser definidos e gerados pelo backend.

## 5. Raças permitidas

O campo `race_id` enviado pelo cliente deve ser validado contra o catálogo oficial. As únicas raças permitidas são:

- `human`
- `forest_elf`
- `dwarf`
- `ice_elf`
- `green_orc`

Qualquer outro valor para `race_id` deve ser **rejeitado** com um erro claro. Isso inclui, mas não se limita a:

- `ogre`
- `air`
- `water`
- `unknown`
- String vazia (`""`)

## 6. Classe inicial

A regra de classe inicial é absoluta e não pode ser contornada:

- Todo personagem criado no backend deve ter seu `class_id` definido como `novice`.
- O jogador **não pode** escolher `knight`, `mage`, `archer`, `assassin` ou `cleric` na tela de criação.
- A escolha da classe base será feita futuramente, no level 10, e essa lógica deverá usar a função `rules.CanSelectBaseClass` para validação.

## 7. Elementos na criação

A regra para elementos na criação é simples:

- Nenhum elemento é escolhido ou associado ao personagem no momento da criação.
- Os elementos oficiais (`fire`, `earth`, `ice`, `shadow`, `sacred`) pertencem a sistemas futuros de afinidade e evolução avançada.
- Os IDs `air` e `water` são inválidos e devem ser rejeitados em qualquer contexto futuro.

## 8. Ordem de validação recomendada

A futura implementação do handler de criação de personagem deve seguir esta ordem de validação para garantir segurança e erros determinísticos:

1.  Validar se a sessão/conta do jogador está autenticada e autorizada a criar um personagem.
2.  Validar o formato do `character_name` (ex: comprimento, caracteres permitidos).
3.  Validar a disponibilidade do `character_name` (verificar se já não existe no banco de dados).
4.  Validar o `race_id` contra o catálogo oficial do Rule Registry.
5.  **Rejeitar** a requisição se o cliente enviar qualquer valor para `class_id`.
6.  **Rejeitar** a requisição se o cliente enviar qualquer valor para `element_id`.
7.  Se todas as validações passarem, o backend define os atributos iniciais: `class_id = "novice"`, `level = 1`, etc.
8.  Persistir o novo personagem no banco de dados dentro de uma transação.
9.  Retornar uma resposta de sucesso ao cliente.

## 9. Regras de rejeição obrigatórias

A criação de um personagem deve falhar se qualquer uma das seguintes condições for verdadeira:

- A sessão do jogador é inválida.
- O nome do personagem é inválido (formato, comprimento).
- O nome do personagem já está em uso.
- O `race_id` enviado não é uma das 5 raças oficiais.
- O `race_id` é `ogre`.
- O cliente envia um `class_id` na requisição.
- O cliente envia um `element_id` na requisição.
- O cliente tenta definir qualquer atributo inicial (nível, stats, posição, inventário, moeda).

## 10. Riscos que o contrato previne

Aderir a este contrato mitiga os seguintes riscos de segurança e design:

- **Criação de Personagens Inválidos:** Impede que um `ogre` jogável seja criado e impede o uso de elementos inválidos como `air` ou `water`.
- **Progressão Acelerada:** Impede que um jogador comece como `knight` no nível 1, pulando a fase `novice`.
- **Stats Manipulados:** Impede que um cliente modificado crie um personagem com HP, mana ou outros atributos alterados.
- **Vantagem Injusta:** Impede que um jogador escolha um ponto de spawn vantajoso ou comece com itens/moeda que não deveria ter.
- **Inconsistência de Jogo:** Garante que todos os jogadores comecem sob as mesmas regras, mantendo a integridade do design do jogo.

## 11. Integrações futuras necessárias

Para implementar este contrato, as seguintes tasks serão necessárias:

- Implementação do handler/endpoint de criação de personagem no backend.
- Lógica de persistência transacional para salvar o novo personagem.
- Sistema de verificação de unicidade de nome.
- Desenvolvimento da UI de criação de personagem no cliente.
- Mecanismo para o cliente obter uma cópia read-only do catálogo de raças para a UI.
- Testes de integração e de unidade para todos os cenários de criação (válidos e inválidos).
- Documentação do protocolo de rede, se forem necessárias novas mensagens.

## 12. Critérios de aceite para futura implementação

A futura task de implementação será considerada concluída quando:

- O código do backend compilar com sucesso.
- Todos os testes unitários e de integração passarem.
- O endpoint de criação de personagem rejeitar raças inválidas, incluindo `ogre`.
- O endpoint sempre definir a classe inicial como `novice`, independentemente do que o cliente enviar.
- O endpoint rejeitar qualquer `class_id` ou `element_id` enviado pelo cliente.
- O cliente de debug atual continuar funcionando.
- O protocolo de rede não for alterado sem um documento de planejamento aprovado.
- Nenhuma senha, token ou dado sensível for exposto em logs.

## 13. O que esta task NÃO faz

- Não cria código Go, C# ou qualquer outra linguagem.
- Não altera o banco de dados.
- Não altera o protocolo de rede.
- Não altera o cliente.
- Não cria um endpoint de API.
- Não implementa a lógica de criação de personagem.

## 14. Status

Este documento é o contrato técnico oficial e preparatório para a futura implementação do sistema de criação de personagem. Ele garante que a integração com o Rule Registry será feita de forma segura e alinhada com a arquitetura server-authoritative do projeto.
