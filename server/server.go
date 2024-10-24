package main

import (
	"context"
	pb "fullfillment-service/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedFulfillmentServiceServer
}

func (s *server) AssignOrder(ctx context.Context, req *pb.AssignOrderRequest) (*pb.AssignOrderResponse, error) {
	// Implement logic to assign order to delivery personnel
	return &pb.AssignOrderResponse{Status: "Order assigned"}, nil
}

func (s *server) GetOrderStatus(ctx context.Context, req *pb.GetOrderStatusRequest) (*pb.GetOrderStatusResponse, error) {
	// Implement logic to get order status
	return &pb.GetOrderStatusResponse{OrderId: req.OrderId, Status: "In progress"}, nil
}

func (s *server) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	// Implement logic to update order status
	return &pb.UpdateOrderStatusResponse{Status: "Order completed"}, nil
}

func (s *server) GetOrdersByDeliveryPerson(ctx context.Context, req *pb.GetOrdersByDeliveryPersonRequest) (*pb.GetOrdersByDeliveryPersonResponse, error) {
	// Implement logic to get all orders for a delivery person
	orders := []*pb.Order{
		{OrderId: "order1", Status: "Completed"},
		{OrderId: "order2", Status: "In progress"},
	}
	return &pb.GetOrdersByDeliveryPersonResponse{Orders: orders}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterFulfillmentServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
