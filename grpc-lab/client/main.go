package main

import (
	"context"
	"log"
	"time"

	userpb "github.com/radieske/tech-lab/grpc-lab/proto" // Semelhante ao servidor, mas aqui você usa o stub de cliente gerado pelo protoc.
	"google.golang.org/grpc"
)

// Fluxo do cliente:
// Cria stub → monta request → chama RPC → recebe response tipado.
func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	c := userpb.NewUserServiceClient(conn) // Cria o cliente tipado para o seu serviço (gerado pelo protoc).

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // Cria um timeout de 2s para a chamada (boa prática).
	defer cancel()

	resp, err := c.GetUser(ctx, &userpb.GetUserRequest{Id: "123"}) // Faz a RPC: serializa GetUserRequest, envia via HTTP/2, recebe GetUserResponse.
	if err != nil {
		log.Fatalf("GetUser: %v", err)
	}
	log.Printf("User: %s <%s>", resp.GetName(), resp.GetEmail()) // Usa os getters gerados (Protobuf gera métodos GetX).
}
