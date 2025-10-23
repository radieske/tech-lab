# gRPC

## N1 — Conceitos e fundamentos 

### 1. O que é gRPC

gRPC (Google Remote Procedure Call) é um framework de comunicação entre serviços baseado no protocolo HTTP/2 e Protocol Buffers (Protobuf).

Enquanto REST trabalha com recursos e JSON, o gRPC trabalha com métodos e mensagens binárias — o que o torna mais rápido, seguro e fortemente tipado.

Em resumo:

- Protocol Buffers → definem a estrutura dos dados.
- Service definition → define métodos disponíveis.
- gRPC runtime → gera código cliente/servidor automaticamente.

Quando usar:

- Comunicação entre microserviços internos.
- Ambientes de alta performance e baixa latência.
- Necessidade de streaming bidirecional (ex: chat, telemetry).

Quando evitar:

- APIs públicas (difícil debugar, menos compatível).
- Casos onde integração humana/manual via HTTP é comum.

### 2. Comparativo rápido — REST vs gRPC

| Aspecto | REST | gRPC |
| --- | --- | --- |
| Protocolo | HTTP 1.1 | HTTP/2 |
| Payload | JSON | Protobuf (binário) |
| Comunicação | Requisição-resposta | Requisição-resposta e streaming |
| Performance | Boa | Excelente |
| Tipagem | Fraca | Forte (com geração de código) |
| Uso típico | APIs públicas | Comunicação interna entre microserviços |

### 3. Protocol Buffers (Protobuf)

É o formato de serialização usado pelo gRPC.

Funciona como um contrato: define mensagens (dados) e serviços (métodos).

Exemplo simples:

```protobuf
syntax = "proto3";

package user;

service UserService {
  rpc GetUser (GetUserRequest) returns (GetUserResponse);
}

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  string name = 1;
  string email = 2;
}
```

### 4. Recursos de apoio

- Artigo: [gRPC Overview (Google)](https://grpc.io/docs/what-is-grpc/introduction/)
- Artigo: [O guia completo do gRPC parte 1: O que é gRPC?](https://blog.lsantos.dev/guia-grpc-1/)
- Artigo: [gRPC: Framework RPC de Alto Desempenho](https://api7.ai/pt/learning-center/api-101/what-is-grpc)

### 5. Colinha

O gRPC usa Protocol Buffers (Protobuf) como IDL (interface definition language).

## N2 — Aplicação mínima

### 1. O que é o UnimplementedUserServiceServer

Quando você gera código com o protoc, ele cria no arquivo user_grpc.pb.go uma interface que define o contrato do servidor, e uma struct vazia chamada UnimplementedUserServiceServer.

No trecho gerado (resumido) você veria algo assim:

```go
type UserServiceServer interface {
    GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
    mustEmbedUnimplementedUserServiceServer()
}

type UnimplementedUserServiceServer struct{}

func (UnimplementedUserServiceServer) GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error) {
    return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}
func (UnimplementedUserServiceServer) mustEmbedUnimplementedUserServiceServer() {}
```

É basicamente uma implementação-vazia da interface, usada como “base class” no Go.

Quando você cria o seu servidor:

```go
type userServer struct {
    userpb.UnimplementedUserServiceServer
}
```

Você está dizendo:

“Minha struct userServer embute a implementação padrão dos métodos, que apenas retornam ‘não implementado’. Eu depois vou sobrescrever os métodos que quero realmente implementar.”

Assim, se no futuro você atualizar o .proto e adicionar novos métodos (por exemplo UpdateUser), o seu código continua compilando, porque a implementação embutida já fornece um “stub” para os métodos novos.

Se você não embutisse isso, o Go obrigaria a implementar todos os métodos da interface sempre que ela mudasse — e isso quebraria o build de cada serviço a cada novo método adicionado.

Em resumo:

- Ele garante compatibilidade futura.
- Evita que o build quebre quando o contrato evolui.
- Retorna um erro gRPC padrão (codes.Unimplemented) se alguém chamar um método que você ainda não implementou.

### 2. protoc × protobuf: qual a diferença?

Eles estão relacionados, mas não são a mesma coisa.

| Termo | O que é | Papel |
| --- | --- | --- |
| Protocol Buffers (protobuf) | É o formato de serialização e a linguagem de definição de mensagens criada pelo Google. | Define como os dados são descritos (.proto) e como são codificados (binário eficiente). |
| protoc | É o compilador do protobuf. | Lê os arquivos .proto e gera código para várias linguagens (Go, Java, Python, etc.). |

Em outras palavras:

- protobuf → a tecnologia / especificação
- protoc → a ferramenta (CLI) que lê .proto e gera os .pb.go, .pb.java, etc.

Quando você roda:

```bash
protoc --go_out=... --go-grpc_out=... proto/user.proto
```

Você está:

1. Usando o compilador protoc;
2. Para processar um arquivo escrito na linguagem protobuf (user.proto);
3. Gerando código Go que implementa as estruturas e interfaces definidas lá.

Exemplo prático

1. Você escreve:
```protobuf
message User {
  string name = 1;
}
```
2. O Protobuf define o significado disso.
3. O protoc lê esse arquivo e gera:
```go
type User struct {
    Name string `protobuf:"bytes,1,opt,name=name,proto3"`
}
```

com toda a lógica de serialização binária embutida.

Resumo rápido

| Termo | Analogia |
| --- | --- |
| protobuf | a “linguagem” e “gramática” que define seus dados |
| .proto | o arquivo-fonte onde você escreve o contrato |
| protoc | o compilador que traduz o .proto para código Go, Java etc. |
| .pb.go | o código gerado a partir disso |

## N3 — Mini-case prático
### Integração HTTP → gRPC

### 1. Objetivo

Expandir o laboratório gRPC criando uma aplicação realista, onde um **gateway HTTP** consome um **serviço gRPC**.  
O objetivo é simular o padrão comum em arquiteturas modernas: um *frontend HTTP* expõe endpoints REST, enquanto a comunicação interna entre serviços é feita via gRPC.

```
grpc-lab/
├── client
│   └── main.go
├── docker-compose.yml
├── Dockerfile.gateway
├── Dockerfile.server
├── gateway
│   └── main.go
├── go.mod
├── go.sum
├── Makefile
├── proto
│   ├── user_grpc.pb.go
│   ├── user.pb.go
│   └── user.proto
├── readme.md
└── server
    └── main.go
```

---

### 2. Resultado

Com o N3, o projeto passou a possuir:
- Um serviço gRPC funcional (UserService);
- Um gateway HTTP que converte chamadas REST → gRPC;
- Suporte a Docker e Docker Compose para execução unificada;
- Padronização de comandos via Makefile;
- Fluxo de comunicação em três camadas: HTTP → gRPC → Resposta JSON.