package main

import (
	"fullfillment-service/config"
	"fullfillment-service/internal/fulfillment"
	pb "fullfillment-service/proto"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	db := config.InitDB()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFulfillmentServiceServer(grpcServer, fulfillment.NewService(db))

	log.Println("Fulfillment Service is running on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
