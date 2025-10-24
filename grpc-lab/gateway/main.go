package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	userpb "github.com/radieske/tech-lab/grpc-lab/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type Gateway struct {
	grpcClient userpb.UserServiceClient
}

func dialGRPC(grpcAddr string, useTLS bool) (*grpc.ClientConn, error) {
	if !useTLS {
		return grpc.Dial(grpcAddr, grpc.WithInsecure())
	}
	// TLS: usa cert do servidor; para autoassinado, skip verify (dev only) ou forneça CA
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // DEV APENAS
	}
	creds := credentials.NewTLS(tlsConfig)
	return grpc.Dial(grpcAddr, grpc.WithTransportCredentials(creds))
}

func NewGateway(grpcAddr string, useTLS bool) (*Gateway, error) {
	// conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure()) // local/dev; em prod usar TLS
	conn, err := dialGRPC(grpcAddr, useTLS)
	if err != nil {
		return nil, err
	}
	return &Gateway{
		grpcClient: userpb.NewUserServiceClient(conn),
	}, nil
}

func grpcErrorToHTTP(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}
	st, ok := status.FromError(err)
	if !ok {
		return http.StatusBadGateway, "unknown upstream error"
	}
	switch st.Code() {
	case codes.InvalidArgument:
		return http.StatusBadRequest, st.Message()
	case codes.NotFound:
		return http.StatusNotFound, st.Message()
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, st.Message()
	case codes.PermissionDenied, codes.Unauthenticated:
		return http.StatusUnauthorized, st.Message()
	case codes.ResourceExhausted, codes.Unavailable:
		return http.StatusServiceUnavailable, st.Message()
	default:
		return http.StatusBadGateway, st.Message()
	}
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
		code, msg := grpcErrorToHTTP(err)
		http.Error(w, msg, code)
		return
	}
	out := map[string]string{
		"id":    id,
		"name":  resp.GetName(),
		"email": resp.GetEmail(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// Server-stream → NDJSON
func (g *Gateway) handleStreamProfiles(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := int32(3)
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = int32(v)
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stream, err := g.grpcClient.ListUsers(ctx, &userpb.ListUsersRequest{Limit: limit})
	if err != nil {
		code, msg := grpcErrorToHTTP(err)
		http.Error(w, msg, code)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	bw := bufio.NewWriter(w)
	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}
		row := map[string]string{
			"name":  resp.GetName(),
			"email": resp.GetEmail(),
		}
		_ = json.NewEncoder(bw).Encode(row) // uma linha JSON por item
		bw.Flush()
	}
}

func main() {
	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}
	useTLS := os.Getenv("USE_TLS") == "true"

	gw, err := NewGateway(grpcAddr, useTLS)
	if err != nil {
		log.Fatalf("init gateway: %v", err)
	}

	http.HandleFunc("/profiles", gw.handleGetProfile)
	http.HandleFunc("/profiles/stream", gw.handleStreamProfiles)

	addr := ":8080"
	log.Printf("HTTP gateway listening on %s (gRPC upstream: %s, TLS=%v)", addr, grpcAddr, useTLS)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
