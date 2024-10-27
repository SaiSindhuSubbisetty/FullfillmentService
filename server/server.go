package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "fullfillment-service/proto"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Order struct {
	OrderID          string `gorm:"primaryKey"`
	DeliveryPersonID string
	Status           string
	CreatedAt        int64
	UpdatedAt        int64
}

type DeliveryPerson struct {
	DeliveryPersonID string `gorm:"primaryKey"`
	Name             string
	Location         string
	Status           string
}

// Global DB connection
var db *gorm.DB

// Init DB connection
func initDB() {
	dsn := "host=localhost user=postgres password=1234 dbname=fulfillmentdb port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	// Remove db.AutoMigrate as migrations will be handled separately using SQL files
}

// FulfillmentServiceServer implements the gRPC methods
type FulfillmentServiceServer struct {
	pb.UnimplementedFulfillmentServiceServer
	db *gorm.DB
}

// AssignOrder assigns an order to the nearest available delivery person
func (s *FulfillmentServiceServer) AssignOrder(ctx context.Context, req *pb.AssignOrderRequest) (*pb.AssignOrderResponse, error) {
	// Find the nearest available delivery person
	var nearestDeliveryPerson DeliveryPerson
	err := db.Raw(`
		SELECT * FROM delivery_persons
		WHERE status = 'AVAILABLE'
		ORDER BY ST_Distance(location::geometry, ST_Point(40.748817, -73.985428)::geometry) ASC
		LIMIT 1
	`).Scan(&nearestDeliveryPerson).Error

	if err != nil || nearestDeliveryPerson.DeliveryPersonID == "" {
		return &pb.AssignOrderResponse{Status: "FAILED"}, fmt.Errorf("no available delivery person found")
	}

	// Assign the delivery person to the order
	order := Order{
		OrderID:          req.OrderId,
		DeliveryPersonID: nearestDeliveryPerson.DeliveryPersonID,
		Status:           "ASSIGNED",
	}
	if err := db.Create(&order).Error; err != nil {
		return &pb.AssignOrderResponse{Status: "FAILED"}, err
	}

	// Update delivery person status to "BUSY"
	nearestDeliveryPerson.Status = "BUSY"
	if err := db.Save(&nearestDeliveryPerson).Error; err != nil {
		return &pb.AssignOrderResponse{Status: "FAILED"}, err
	}

	return &pb.AssignOrderResponse{Status: "ASSIGNED"}, nil
}

// GetOrderStatus retrieves the status of an order
func (s *FulfillmentServiceServer) GetOrderStatus(ctx context.Context, req *pb.GetOrderStatusRequest) (*pb.GetOrderStatusResponse, error) {
	var order Order
	if err := db.First(&order, "order_id = ?", req.OrderId).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}

	return &pb.GetOrderStatusResponse{
		OrderId: req.OrderId,
		Status:  order.Status,
	}, nil
}

// UpdateOrderStatus updates the status of an order
func (s *FulfillmentServiceServer) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	// Update the order status in the database
	var order Order
	if err := db.First(&order, "order_id = ?", req.OrderId).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}

	order.Status = req.Status
	if err := db.Save(&order).Error; err != nil {
		return nil, fmt.Errorf("failed to update order status")
	}

	// Handle status transitions
	if req.Status == "DELIVERED" {
		// Mark the delivery person as available after delivery
		var deliveryPerson DeliveryPerson
		if err := db.First(&deliveryPerson, "delivery_person_id = ?", order.DeliveryPersonID).Error; err != nil {
			return nil, fmt.Errorf("failed to find delivery person")
		}
		deliveryPerson.Status = "AVAILABLE"
		if err := db.Save(&deliveryPerson).Error; err != nil {
			return nil, fmt.Errorf("failed to update delivery person status")
		}
	} else if req.Status == "IN_PROGRESS" {
		// If status is IN_PROGRESS, we ensure the delivery person remains BUSY
		var deliveryPerson DeliveryPerson
		if err := db.First(&deliveryPerson, "delivery_person_id = ?", order.DeliveryPersonID).Error; err != nil {
			return nil, fmt.Errorf("failed to find delivery person")
		}
		deliveryPerson.Status = "BUSY"
		if err := db.Save(&deliveryPerson).Error; err != nil {
			return nil, fmt.Errorf("failed to update delivery person status")
		}
	}

	return &pb.UpdateOrderStatusResponse{Status: "UPDATED"}, nil
}

// GetOrdersByDeliveryPerson retrieves all orders assigned to a delivery person
func (s *FulfillmentServiceServer) GetOrdersByDeliveryPerson(ctx context.Context, req *pb.GetOrdersByDeliveryPersonRequest) (*pb.GetOrdersByDeliveryPersonResponse, error) {
	var orders []Order
	if err := db.Where("delivery_person_id = ?", req.DeliveryPersonId).Find(&orders).Error; err != nil {
		return nil, err
	}

	// Convert orders to the proto format
	var protoOrders []*pb.Order
	for _, order := range orders {
		protoOrders = append(protoOrders, &pb.Order{
			OrderId: order.OrderID,
			Status:  order.Status,
		})
	}

	return &pb.GetOrdersByDeliveryPersonResponse{Orders: protoOrders}, nil
}

func main() {
	initDB()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFulfillmentServiceServer(grpcServer, &FulfillmentServiceServer{})

	fmt.Println("Fulfillment Service is running on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
