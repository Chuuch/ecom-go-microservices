package main

import (
	"context"
	"log"

	userService "github.com/Chuuch/ecom-microservices/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	grpcConn, err := grpc.NewClient("127.0.0.1:5001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Cannot connect to grpc server: %v", err)
	}

	defer grpcConn.Close()
	client := userService.NewUserServiceClient(grpcConn)
	ctx := context.Background()

	md := metadata.Pairs(
		"session_id", "fd048bf8-d867-4948-b00f-51b0d592ed13",
		"subsystem", "cli",
	)

	ctx = metadata.NewOutgoingContext(ctx, md)

	res, err := client.GetMe(ctx, &userService.GetMeRequest{})
	if err != nil {
		log.Fatalf("Cannot get me: %v", err)
	}

	log.Println("RESPONSE: ", res.String())

}
