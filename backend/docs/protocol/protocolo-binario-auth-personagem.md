# Light and Shadow — Protocolo Binário Auth/Personagem



Status: Ativo

Backend: Go

Client alvo: Godot 4 C#

Transporte: TCP

Formato: Binário

Endianess: Little Endian



## 1. Header padrão



Todo pacote TCP possui header fixo de 8 bytes:



\- size: uint16

\- opcode: uint16

\- sequence: uint32

\- payload: bytes



Layout:



uint16 size

uint16 opcode

uint32 sequence

byte\[] payload



Constantes:



HeaderSize = 8

MaxPacketSize = 16384



Todos os inteiros usam Little Endian.



## 2. Strings



Strings usam prefixo uint16 com tamanho, seguido por bytes UTF-8.



Layout:



uint16 length

byte\[length] utf8\_string



## 3. Status



Campos de status usam 1 byte:



0 = false

1 = true



## 4. Opcodes iniciais



1002 CS\_LOGIN\_REQUEST

1003 SC\_LOGIN\_RESPONSE

1004 CS\_CHAR\_LIST\_REQUEST

1005 SC\_CHAR\_LIST\_RESPONSE

1006 CS\_CHAR\_SELECT\_REQUEST

1007 SC\_CHAR\_SELECT\_RESPONSE



## 5. CS\_LOGIN\_REQUEST — 1002



Direção: Client -> Server



Payload:



string username

string password



Erros possíveis:



invalid\_login\_payload

invalid\_credentials



## 6. SC\_LOGIN\_RESPONSE — 1003



Direção: Server -> Client



Payload:



uint8 status

uint32 account\_id

string token

string error\_code



Sucesso:



status = 1

account\_id = 1

token = "session\_..."

error\_code = ""



Falha:



status = 0

account\_id = 0

token = ""

error\_code = "invalid\_credentials"



## 7. CS\_CHAR\_LIST\_REQUEST — 1004



Direção: Client -> Server



Payload vazio.



Pré-condição:



O client precisa estar autenticado na mesma conexão TCP.



## 8. SC\_CHAR\_LIST\_RESPONSE — 1005



Direção: Server -> Client



Payload:



uint8 status

string error\_code

uint16 character\_count



Para cada personagem:



string name

string class

uint32 level



Erros possíveis:



not\_authenticated

failed\_to\_list\_characters



## 9. CS\_CHAR\_SELECT\_REQUEST — 1006



Direção: Client -> Server



Payload:



string character\_name



Pré-condições:



Client autenticado.

Personagem pertence à conta autenticada.

Personagem carrega corretamente do PostgreSQL.



## 10. SC\_CHAR\_SELECT\_RESPONSE — 1007



Direção: Server -> Client



Payload:



uint8 status

string character\_name

string error\_code



Erros possíveis:



not\_authenticated

invalid\_character\_select\_payload

ownership\_validation\_failed

character\_not\_owned

character\_load\_failed



## 11. Fluxo inicial do client



1\. Conectar no Gateway TCP.

2\. Enviar CS\_LOGIN\_REQUEST 1002.

3\. Receber SC\_LOGIN\_RESPONSE 1003.

4\. Enviar CS\_CHAR\_LIST\_REQUEST 1004.

5\. Receber SC\_CHAR\_LIST\_RESPONSE 1005.

6\. Escolher personagem.

7\. Enviar CS\_CHAR\_SELECT\_REQUEST 1006.

8\. Receber SC\_CHAR\_SELECT\_RESPONSE 1007.

9\. Se sucesso, aguardar pacotes iniciais do mundo.



Depois da seleção bem-sucedida, o Gateway pode enviar:



4001 SC\_INVENTORY\_SYNC

2006 SC\_CHUNK\_DATA



## 12. Regras para o Godot C#



O client Godot C# precisa implementar:



WriteUInt16LE

WriteUInt32LE

ReadUInt16LE

ReadUInt32LE

WriteStringComPrefixoUInt16

ReadStringComPrefixoUInt16

BuildPacket

ReadPacket

EncodeLoginRequest

DecodeLoginResponse

DecodeCharacterListResponse

EncodeCharacterSelectRequest

DecodeCharacterSelectResponse



## 13. Regra arquitetural



Este documento é a fonte de verdade da integração inicial Godot C#.



Qualquer mudança no protocolo deve atualizar:



pkg/protocol/protocol.go

pkg/protocol/protocol\_response\_test.go

docs/protocol/protocolo-binario-auth-personagem.md

client Godot C#


---

## Atualização R1-I-B - Race ID na lista de personagens

A partir da Task R1-I-B, o pacote SC_CHAR_LIST_RESPONSE inclui o campo race_id para cada personagem retornado na lista.

Esta é uma breaking change controlada local/debug, pois o client Godot também deve ser atualizado na mesma task para ler o novo campo.

### Opcode afetado

- 1005 SC_CHAR_LIST_RESPONSE

### Formato por personagem

Cada entrada de personagem dentro de SC_CHAR_LIST_RESPONSE deve ser serializada exatamente nesta ordem:

1. string name
2. string class
3. uint32 level
4. string race_id

### Observações

- CS_CHAR_LIST_REQUEST permanece inalterado.
- CS_CHAR_SELECT_REQUEST permanece inalterado.
- SC_CHAR_SELECT_RESPONSE permanece inalterado.
- Nenhum opcode novo foi criado nesta task.
- Esta alteração não implementa criação de personagem.
- race_id representa a raça oficial persistida no banco, por exemplo human.
- Backend e Godot devem manter a mesma ordem de escrita/leitura: name, class, level, race_id.

