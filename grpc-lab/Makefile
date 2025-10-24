# O Makefile automatiza os comandos para você não ter que decorar flags.

# Variáveis para reuso (módulo e caminhos).
MODULE := github.com/radieske/tech-lab/grpc-lab
PROTO_DIR := proto
PROTO_FILES := $(PROTO_DIR)/user.proto

.PHONY: proto tidy run-server run-client clean

# Alvo que roda o protoc.
# -I . -I proto: inclui a raiz e a pasta proto como includes.
# --go_out=paths=source_relative:. → gera .pb.go espelhando o caminho de entrada (por isso passamos proto/user.proto para cair em proto/).
# --go-grpc_out=paths=source_relative:. → idem, mas para os stubs gRPC.
# $(PROTO_FILES) → usar proto/user.proto garante que o destino seja proto/.
proto:
	protoc -I . -I $(PROTO_DIR) \
	  --go_out=paths=source_relative:. \
	  --go-grpc_out=paths=source_relative:. \
	  $(PROTO_FILES)

# Atalhos para arrumar dependências e rodar server/cliente.
tidy:
	go mod tidy

run-server:
	USE_TLS=false go run ./server

run-gateway:
	USE_TLS=false GRPC_ADDR=localhost:50051 go run ./gateway

run-client:
	go run ./client

run-chat:
	go run ./client/chat/chat_client.go

clean:
	rm -rf github.com
	rm -f $(PROTO_DIR)/*.pb.go $(PROTO_DIR)/*_grpc.pb.go
