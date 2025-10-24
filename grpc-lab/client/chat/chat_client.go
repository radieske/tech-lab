package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	userpb "github.com/radieske/tech-lab/grpc-lab/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := userpb.NewUserServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := c.Chat(ctx)
	if err != nil {
		panic(err)
	}

	// goroutine para receber
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				fmt.Println("recv err:", err)
				return
			}
			fmt.Printf("[server] %s (%d)\n", in.GetText(), in.GetTsUnix())
		}
	}()

	// enviar a partir do terminal
	sc := bufio.NewScanner(os.Stdin)
	fmt.Println("Digite mensagens (Ctrl+C para sair):")
	for sc.Scan() {
		msg := sc.Text()
		if msg == "" {
			continue
		}
		if err := stream.Send(&userpb.ChatMessage{
			From:   "client",
			Text:   msg,
			TsUnix: time.Now().Unix(),
		}); err != nil {
			fmt.Println("send err:", err)
			return
		}
	}
}
