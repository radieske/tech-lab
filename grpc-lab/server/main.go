package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	userpb "github.com/radieske/tech-lab/grpc-lab/proto" // importa o pacote gerado pelo protobuf a partir do user.proto.
	"google.golang.org/grpc"                             // runtime do gRPC (servidor, registradores).
	"google.golang.org/grpc/credentials"
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

func (s *userServer) ListUsers(req *userpb.ListUsersRequest, stream userpb.UserService_ListUsersServer) error {
	limit := req.GetLimit()
	if limit <= 0 {
		limit = 3
	}
	for i := int32(1); i <= limit; i++ {
		resp := &userpb.GetUserResponse{
			Name:  fmt.Sprintf("User %d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
		time.Sleep(250 * time.Millisecond) // simula trabalho
	}
	return nil
}

func (s *userServer) Chat(stream userpb.UserService_ChatServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			return err // EOF ou erro real
		}
		out := &userpb.ChatMessage{
			From:   "server",
			Text:   "echo: " + in.GetText(),
			TsUnix: time.Now().Unix(),
		}
		if err := stream.Send(out); err != nil {
			return err
		}
	}
}

func loadTLSCredentials() (grpc.ServerOption, error) {
	// Server TLS only (não mTLS). Usa certs gerados.
	certFile := "certs/server.crt"
	keyFile := "certs/server.key"

	// opcional: carregar CA custom se desejar validar clientes
	var tlsConfig *tls.Config
	// server cert
	servCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load server cert: %w", err)
	}
	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{servCert},
	}
	return grpc.Creds(credentials.NewTLS(tlsConfig)), nil
}

// Fluxo do servidor:
// Listener TCP → gRPC Server → Deserializa Protobuf → chama seu método → serializa resposta.
func main() {
	startMetricsServer() // /metrics em :9090

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	var opts []grpc.ServerOption
	// interceptors
	opts = append(opts,
		grpc.UnaryInterceptor(unaryLoggingRecoveryInterceptor),
		grpc.StreamInterceptor(streamLoggingRecoveryInterceptor),
	)

	// TLS opcional via env: USE_TLS=true
	if os.Getenv("USE_TLS") == "true" {
		credOpt, err := loadTLSCredentials()
		if err != nil {
			log.Fatalf("tls: %v", err)
		}
		opts = append(opts, credOpt)
	}

	s := grpc.NewServer(opts...)
	userpb.RegisterUserServiceServer(s, &userServer{})

	fmt.Println("gRPC server listening on :50051; metrics on :9090; TLS:", os.Getenv("USE_TLS"))
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
