# Windows Debug Local Test Instructions

## 1. Objetivo

Este documento ensina como testar localmente o cliente Windows Debug exportado do Godot para o projeto Light and Shadow.

- **Importante:** Este não é um build público ou distribuível, mas sim uma versão de desenvolvimento para validação de funcionalidades.
- O cliente se conectará ao backend local na porta padrão `127.0.0.1:8080`.

## 2. Pré-requisitos

- Sistema Operacional Windows.
- Docker Desktop instalado e em execução.
- Backend local do projeto rodando via Docker Compose.
- Cliente Windows Debug exportado e disponível em:
  `builds/windows-debug/LightAndShadow.exe`
- (Opcional) Pacote ZIP de distribuição local em:
  `builds/LightAndShadow-Windows-Debug.zip`

## 3. Subir backend local

Para iniciar todos os serviços do backend (Gateway, Auth, World, PostgreSQL, Redis), execute os seguintes comandos no PowerShell:

```powershell
cd "C:\Users\gabri\Desktop\Projeto Jogo Online\light-and-shadow-godot4-csharp-bootstrap\backend"
docker compose up --build
```

> **Nota:** O aviso do Docker sobre `version '3.8' is obsolete` é conhecido e não impede a execução. Você pode ignorá-lo.

## 4. Abrir cliente Windows Debug

Existem duas maneiras de executar o cliente para teste.

### Opção A: Executar diretamente

A forma mais simples é executar o cliente diretamente do diretório de exportação.

1. Navegue até a pasta `builds/windows-debug/`.
2. Execute o arquivo `LightAndShadow.exe`.

### Opção B: Extrair e testar o ZIP local

Esta opção simula o processo de distribuição e garante que todos os arquivos necessários foram empacotados corretamente.

1. Abra o PowerShell e execute os comandos abaixo para limpar, extrair e verificar o conteúdo do ZIP:

```powershell
cd "C:\Users\gabri\Desktop\Projeto Jogo Online\light-and-shadow-godot4-csharp-bootstrap"
Remove-Item -Recurse -Force "builds\zip-test" -ErrorAction SilentlyContinue
Expand-Archive -Path "builds\LightAndShadow-Windows-Debug.zip" -DestinationPath "builds\zip-test"
Get-ChildItem "builds\zip-test"
```

2. Após a extração, execute o cliente a partir da nova pasta:

```powershell
builds/zip-test/LightAndShadow.exe
```

## 5. Login de teste local

Use as seguintes credenciais para se conectar ao banco de dados local que roda no Docker:

- **Username:** `default_user`
- **Password:** `test123`

Estes dados são apenas para o ambiente de desenvolvimento e não funcionarão em servidores de produção.

## 6. Fluxo de teste

Siga os passos abaixo para validar o fluxo principal do cliente:

1.  Abra o cliente `LightAndShadow.exe`.
2.  Use as credenciais de teste para fazer login.
3.  Após o login bem-sucedido, clique no botão **Request Characters**.
4.  Selecione a personagem "Gabriela" na lista.
5.  Clique no botão **Select Character** para confirmar a entrada no mundo.
6.  Na cena `DebugWorldEntryScene`, verifique se os dados da sessão (ID da conta, nome do personagem) são exibidos corretamente.
7.  Verifique se o painel de snapshot exibe informações iniciais de inventário (`Inventory Sync`) e chunks (`Chunks Received`).
8.  Confirme se a visão do mundo (`DebugTileWorldView`) renderiza os chunks recebidos (blocos cinzas e vermelhos).
9.  Clique no botão **Send Debug Move**.
10. Observe o campo `Last Move Result`. Ele deve indicar `success=true` se o movimento for válido ou `success=false` se houver colisão.
11. Confirme que o marcador do jogador (bloco amarelo) só se move para a nova posição após o servidor confirmar o movimento (recebimento do `SC_MOVE_CONFIRM`).
12. Teste o botão **Back** para garantir que ele retorne à tela de autenticação.

## 7. Parar backend

Quando terminar os testes, pare os contêineres do Docker de forma segura com o seguinte comando:

```powershell
cd "C:\Users\gabri\Desktop\Projeto Jogo Online\light-and-shadow-godot4-csharp-bootstrap\backend"
docker compose down
```

> **Aviso:** Não use `docker compose down -v`. A flag `-v` apagará os volumes de dados, incluindo o banco local. Use esse comando somente se houver decisão explícita do diretor do projeto.

## 8. O que NÃO commitar

Os seguintes arquivos e pastas são gerados localmente e já estão configurados no `.gitignore`. Nunca os adicione ao controle de versão:

- `builds/`
- `builds/windows-debug/`
- `builds/LightAndShadow-Windows-Debug.zip`
- `builds/zip-test/`
- `.godot/`
- `.vscode/`
- `bin/`
- `obj/`

## 9. Troubleshooting

Problemas comuns e suas possíveis soluções:

- **Cliente não conecta:**
  - Verifique se o backend está rodando (`docker compose up`).
  - Confirme se o serviço `ls_gateway` está escutando na porta `127.0.0.1:8080`.

- **Login falha:**
  - Verifique se as credenciais `default_user` / `test123` estão corretas.
  - Verifique se o backend está rodando corretamente.
  - Verifique os logs do `auth-server`.
  - Envie os logs/status para auditoria antes de qualquer reset de banco.

- **Janela abre, mas não entra no mundo:**
  - Verifique os logs dos contêineres do backend (`docker compose logs -f`) para identificar possíveis erros no `world-server` ou `auth-server`.

- **Movimento retorna `success=false`:**
  - Isso não é necessariamente um bug. O servidor realiza validações autoritativas de colisão e velocidade. Um resultado `false` pode indicar que o movimento foi bloqueado por um obstáculo no mapa do servidor.

- **ZIP não abre ou está corrompido:**
  - Tente exportar o projeto novamente pelo Godot.
  - Como alternativa, teste o executável direto da pasta `builds/windows-debug/`.

## 10. Status esperado

Ao final do teste, o seguinte fluxo deve funcionar corretamente:

- Login de usuário.
- Listagem e seleção de personagens.
- Entrada na cena de mundo de debug.
- Recebimento de pacotes de sincronização (inventário, chunks).
- Renderização do mapa de debug.
- Envio de requisição de movimento.
- Atualização da posição do jogador apenas após confirmação do servidor.