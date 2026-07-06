# Gateway Runtime Config

## 1. Objetivo

Este documento ensina como configurar o endereço (host) e a porta do Gateway em tempo de execução, sem a necessidade de alterar o código-fonte ou recompilar o cliente. Isso facilita a conexão com diferentes ambientes de backend (local, LAN, remoto) para testes.

## 2. Comportamento padrão

Se o arquivo de configuração `gateway-config.json` não for encontrado ou for inválido, o cliente usará automaticamente os seguintes valores padrão:

- **Host:** `127.0.0.1`
- **Porta:** `8080`

Isso garante que o fluxo de desenvolvimento local continue funcionando sem nenhuma configuração extra.

## 3. Local do arquivo no editor

Ao executar o projeto diretamente pelo Godot Editor, o arquivo `gateway-config.json` deve ser colocado na **raiz do projeto**.

Exemplo de caminho absoluto:
```
C:\Users\gabri\Desktop\Projeto Jogo Online\light-and-shadow-godot4-csharp-bootstrap\gateway-config.json
```

## 4. Local do arquivo no build exportado

Em um build Windows já exportado, o arquivo `gateway-config.json` deve ser colocado na mesma pasta que o executável `LightAndShadow.exe`.

Exemplo no diretório de build:
```
builds/windows-debug/gateway-config.json
builds/windows-debug/LightAndShadow.exe
```

Exemplo em um ZIP extraído para teste:
```
builds/zip-test/gateway-config.json
builds/zip-test/LightAndShadow.exe
```

## 5. Formato do arquivo

O arquivo deve ser um JSON simples contendo apenas as chaves `host` e `port`.

```json
{
  "host": "127.0.0.1",
  "port": 8080
}
```

## 6. Exemplo para outro computador na LAN

Para conectar o cliente a um backend rodando em outra máquina na mesma rede local (LAN), basta alterar o `host` para o IP privado da máquina servidora.

```json
{
  "host": "192.168.0.10",
  "port": 8080
}
```

> **Aviso:** Para que a conexão funcione, o backend deve estar rodando na máquina `192.168.0.10` e as configurações de firewall e rede devem permitir tráfego TCP de entrada na porta `8080`.

## 7. Validação e fallback

O cliente realiza uma validação segura. Se qualquer uma das seguintes condições for verdadeira, ele reverterá para o padrão (`127.0.0.1:8080`):

- O arquivo `gateway-config.json` não existe.
- O arquivo não é um JSON válido.
- A chave `host` está ausente, vazia ou contém apenas espaços.
- A chave `port` está ausente ou o valor está fora do intervalo válido (1 a 65535).

## 8. Segurança

Este arquivo é destinado **apenas** para configuração de endereço.

- **NÃO** coloque senhas, tokens, `SessionToken` ou qualquer outra credencial no arquivo.
- **NÃO** commite o arquivo `gateway-config.json` no repositório se ele contiver IPs locais, privados ou de servidores temporários. Atualmente, trate esse arquivo como configuração local temporária, salvo decisão explícita do diretor do projeto.

## 9. O que NÃO fazer

- Não altere o código C# (`GatewayTcpClient.cs`, `DebugAuthController.cs`, etc.) para mudar o host/porta. Use o arquivo de configuração.
- Não altere o protocolo de comunicação.
- Não salve tokens ou senhas em disco.
- Não commite a pasta `builds/`.
- Não commite `gateway-config.json` sem uma decisão explícita do diretor do projeto.

## 10. Fluxo de teste

1.  Crie um arquivo `gateway-config.json` temporário na localização apropriada (raiz do projeto ou ao lado do `.exe`).
2.  Configure o `host` e `port` desejados.
3.  Inicie o backend no servidor de destino.
4.  Abra o cliente (pelo editor Godot ou executando o `.exe`).
5.  Verifique os logs iniciais do cliente para confirmar qual configuração foi carregada.
6.  Realize o fluxo de login, seleção de personagem ("Gabriela") e entrada no mundo.
7.  Teste o envio de movimento (`Send Debug Move`).
8.  Após o teste, apague o arquivo `gateway-config.json` se ele foi criado apenas para um teste temporário.

## 11. Status esperado

O cliente deve se conectar com sucesso ao Gateway configurado no arquivo `gateway-config.json`. Todo o fluxo de jogo (login, seleção de personagem, entrada no mundo, recebimento de chunks, movimento) deve funcionar exatamente como no ambiente padrão.

## 12. Relação com próximas fases

Esta funcionalidade é um passo preparatório crucial que habilita:

- Testes de cliente e servidor em máquinas diferentes na mesma rede (teste LAN).
- Conexão com um futuro servidor de desenvolvimento remoto (staging/dev).
- Distribuição interna de builds de debug para a equipe de QA, permitindo que eles testem contra diferentes ambientes sem precisar do código-fonte.
