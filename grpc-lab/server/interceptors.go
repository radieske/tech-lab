package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Unary interceptor: loga início/fim, duração e código de status, com recovery.
func unaryLoggingRecoveryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	start := time.Now()

	defer func() {
		if r := recover(); r != nil {
			err = status.Errorf( // retorna INTERNAL
				13, // codes.Internal
				"panic recovered: %v", r,
			)
			fmt.Printf("[PANIC] method=%s err=%v\nstack=%s\n", info.FullMethod, err, string(debug.Stack()))
		}
	}()

	resp, err = handler(ctx, req)
	dur := time.Since(start)

	st := status.Convert(err)
	fmt.Printf("[UNARY] method=%s code=%s dur=%s err=%v\n", info.FullMethod, st.Code(), dur, err)
	observeUnary(info.FullMethod, st.Code().String(), dur) // métrica
	return resp, err
}

// Stream interceptor: semelhante ao unary, porém para streaming.
func streamLoggingRecoveryInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) (err error) {
	start := time.Now()

	defer func() {
		if r := recover(); r != nil {
			err = status.Errorf(13, "panic recovered: %v", r) // INTERNAL
			fmt.Printf("[PANIC-STREAM] method=%s err=%v\nstack=%s\n", info.FullMethod, err, string(debug.Stack()))
		}
	}()

	err = handler(srv, ss)
	dur := time.Since(start)

	st := status.Convert(err)
	fmt.Printf("[STREAM] method=%s code=%s dur=%s err=%v\n", info.FullMethod, st.Code(), dur, err)
	observeStream(info.FullMethod, st.Code().String(), dur) // métrica
	return err
}
