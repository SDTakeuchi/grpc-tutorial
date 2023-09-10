package main

import (
	"context"
	"fmt"
	"grpcserver/api/gen/proto"
	"grpcserver/api/handler"
	"log"
	"net"
	"os"
	"os/signal"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/reflection"
)

const (
	port        = 50051
	secretToken = "some_token"
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	// logger
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	grpczap.ReplaceGrpcLogger(zapLogger)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpcmiddleware.ChainUnaryServer(
				grpczap.UnaryServerInterceptor(zapLogger),
				grpcauth.UnaryServerInterceptor(auth),
			),
		),
	)
	proto.RegisterPancakeBakerServiceServer(
		server,
		handler.NewBakerHandler(),
	)
	reflection.Register(server)

	go func() {
		log.Printf("starting gRPC server on port %v", port)
		server.Serve(lis)
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("stopping gRPC server...")
	server.GracefulStop()
}

func auth(ctx context.Context) (context.Context, error) {
	token, err := grpcauth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	if token != secretToken {
		return nil, status.Errorf(codes.Unauthenticated, "invalid bearer token")
	}

	return context.WithValue(ctx, "UserName", "Thomas"), nil
}
