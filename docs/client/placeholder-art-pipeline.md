# Placeholder Art Pipeline

## 1. Objetivo

Este documento define as regras e convenções para a criação e importação de assets visuais temporários (placeholders) para o cliente do jogo. O objetivo é estabelecer um pipeline inicial antes da produção da arte final, garantindo consistência e organização.

## 2. Escopo

Esta fase de documentação define as convenções para:
- Tiles de mapa placeholder.
- Marcador de jogador (player marker) placeholder.
- Marcador de tile bloqueado ou com objeto placeholder.
- Convenção de estrutura de pastas.
- Convenção de nomes de arquivos.
- Escala visual de referência inicial.
- Regras de importação de assets no Godot.

O que **NÃO** está no escopo desta fase:
- Criação de arte final.
- Animações de personagens ou efeitos.
- Design final da interface do usuário (HUD).
- Criação do mapa final do jogo.
- Design final dos personagens.
- Funcionalidades de monetização.
- Mecânicas de gameplay final.

## 3. Relação com o renderer debug atual

Atualmente, o cliente renderiza o mundo de jogo a partir dos dados recebidos do servidor (`SC_CHUNK_DATA`) usando quadrados coloridos simples. O próximo passo evolutivo será substituir a aparência desses quadrados por tiles placeholder, sem alterar a lógica de renderização ou o controle autoritativo do servidor.

## 4. Fonte da verdade

A lógica de jogo permanece inalterada e é controlada pelo servidor:

- O **backend** é a fonte da verdade e envia os dados dos chunks para o cliente.
- Um tile com ID `0` é considerado caminhável (walkable).
- Um tile com ID `1` é considerado bloqueado (blocked).
- O **cliente** é responsável apenas pela representação visual desses dados.
- O cliente **não** decide sobre colisões. O movimento do marcador do jogador só é atualizado após receber a confirmação do servidor.

## 5. Tile size e escala

As seguintes regras de escala serão adotadas como referência inicial:

- **Tile Lógico:** A unidade base de design do jogo é um tile de **64x64 pixels**.
- **Escala de Placeholder:** Os assets placeholder podem ser criados em uma escala menor para facilitar a visualização em zoom out, mas devem respeitar a proporção.
- **Perspectiva:** Todos os sprites devem ser desenhados com uma perspectiva 2D top-down.
- **Legibilidade:** Os assets devem ser desenhados para manter a clareza visual, mesmo quando vistos de longe (zoom out).

## 6. Estrutura futura sugerida de pastas

A seguinte estrutura de pastas será criada em uma task futura para organizar os assets. **Não crie estas pastas agora.**

```
assets/
  placeholders/
    tiles/
    characters/
    objects/
    ui/
  source/
    prompts/
    references/
  final/
    tiles/
    characters/
    objects/
    ui/
```

- A pasta `placeholders` conterá todos os assets temporários.
- A pasta `final` será usada para a arte final aprovada.

## 7. Convenção de nomes

Os arquivos de assets devem seguir um padrão claro para facilitar a busca e a automação.

**Exemplos:**
```
tile_grass_placeholder_01.png
tile_dirt_placeholder_01.png
tile_stone_blocked_placeholder_01.png
player_marker_placeholder_01.png
object_wall_placeholder_01.png
```

**Regras:**
- Use nomes em `minúsculas` (lowercase).
- Separe palavras com `_` (underscore).
- Evite espaços ou caracteres especiais.
- Inclua a palavra `placeholder` no nome.
- Use numeração de versão (ex: `_01`, `_02`) quando houver variações.

## 8. Regras visuais de placeholder

- O visual deve ser simples e funcional, sem detalhes excessivos.
- A legibilidade é a prioridade máxima.
- Tiles caminháveis (walkable) e bloqueados (blocked) devem ser visualmente distintos à primeira vista.
- **Não** use arte copiada de outros jogos comerciais, especialmente Tibia.
- **Não** use assets com copyright duvidoso ou licenças restritivas.
- O objetivo é a funcionalidade, não a beleza. Não gaste tempo excessivo com polimento nesta fase.

## 9. Import settings no Godot

Quando os assets forem importados no futuro, as seguintes diretrizes devem ser seguidas:

- Para pixel art, o filtro de importação deve ser desativado para evitar um visual borrado (`TextureFilter` = `Nearest`).
- A escala dos assets deve ser mantida de forma consistente.
- Evite compressão com perdas (lossy) em assets críticos para a UI ou gameplay.
- Sempre revise as configurações de importação de um novo asset antes de fazer o commit.
- **Não altere nenhuma configuração de importação nesta task.**

## 10. Regras de Git

- Não commite a pasta `builds/` ou arquivos de build locais.
- Não commite arquivos temporários gerados pelo editor (ex: `.godot/`, `.vscode/`).
- Commite os arquivos `.uid` que o Godot gera para assets, cenas e scripts. Eles são necessários para manter as referências do projeto.
- Não adicione `*.uid` ao `.gitignore`.
- Não commite arquivos muito grandes (ex: > 5MB) sem uma decisão da equipe (para isso, o Git LFS pode ser considerado no futuro).
- Evite usar `git add .` ou `git add *`. Adicione arquivos específicos para ter controle sobre o que está sendo commitado.

## 11. Segurança e licenciamento

- Todos os assets placeholder devem ser de criação própria, gerados internamente (ex: com IA) ou obtidos de fontes com licenças claras e permissivas (ex: CC0).
- **É proibido usar sprites ou qualquer asset extraído do Tibia.**
- Não crie cópias fiéis de assets de outros jogos comerciais.
- Futuramente, guarde prompts de IA e referências na pasta `assets/source/` para rastreabilidade.
- Mantenha um registro da origem de cada asset.

## 12. Próxima task prevista

A próxima task deste pipeline será: **Task 5X-B — Create Placeholder Asset Folder Structure**.

O objetivo será criar a estrutura de pastas vazia (ou com arquivos `.gitkeep`) dentro de `assets/`, preparando o projeto para receber os novos arquivos de placeholder.

## 13. Depois da estrutura

Após a criação da estrutura de pastas, a task seguinte será: **Task 5Y — Replace Debug Squares With Placeholder Tiles**.

O objetivo será modificar o código de renderização para trocar os quadrados crus pelos tiles placeholder, mantendo toda a lógica de chunks e movimento server-authoritative intacta.