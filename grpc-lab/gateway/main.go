package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	userpb "github.com/radieske/tech-lab/grpc-lab/proto"
	"google.golang.org/grpc"
)

type Gateway struct {
	grpcClient userpb.UserServiceClient
}

func NewGateway(grpcAddr string) (*Gateway, error) {
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure()) // local/dev; em prod usar TLS
	if err != nil {
		return nil, err
	}
	return &Gateway{
		grpcClient: userpb.NewUserServiceClient(conn),
	}, nil
}

func (g *Gateway) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	// URL esperada: /profiles?id=123   (simples para não depender de router)
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	// Deadline curto no HTTP → gRPC
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	resp, err := g.grpcClient.GetUser(ctx, &userpb.GetUserRequest{Id: id})
	if err != nil {
		http.Error(w, "upstream error: "+err.Error(), http.StatusBadGateway)
		return
	}

	out := map[string]string{
		"id":    id,
		"name":  resp.GetName(),
		"email": resp.GetEmail(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func main() {
	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}
	gw, err := NewGateway(grpcAddr)
	if err != nil {
		log.Fatalf("init gateway: %v", err)
	}

	http.HandleFunc("/profiles", gw.handleGetProfile)

	addr := ":8080"
	log.Printf("HTTP gateway listening on %s (gRPC upstream: %s)", addr, grpcAddr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
