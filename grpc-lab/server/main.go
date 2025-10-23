package main

import (
	"context"
	"fmt"
	"log"
	"net"

	userpb "github.com/radieske/tech-lab/grpc-lab/proto" // importa o pacote gerado pelo protobuf a partir do user.proto.
	"google.golang.org/grpc"                             // runtime do gRPC (servidor, registradores).
)

type userServer struct {
	userpb.UnimplementedUserServiceServer
}

func (s *userServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	log.Printf("GetUser request id=%s", req.GetId())
	return &userpb.GetUserResponse{
		Name:  "Jonathan Radieske",
		Email: "mrdjona@gmail.com",
	}, nil
}

// Fluxo do servidor:
// Listener TCP → gRPC Server → Deserializa Protobuf → chama seu método → serializa resposta.
func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	s := grpc.NewServer()                              // Instancia o servidor gRPC.
	userpb.RegisterUserServiceServer(s, &userServer{}) // Registra seu userServer no gRPC.

	fmt.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
