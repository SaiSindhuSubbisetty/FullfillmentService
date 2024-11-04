package fulfillment

import (
	"context"
	"fmt"
	pb "fullfillment-service/proto"
	"gorm.io/gorm"
)

type OrderService struct {
	db *gorm.DB
	pb.UnimplementedFulfillmentServiceServer
}

func NewService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) AssignOrder(ctx context.Context, req *pb.AssignOrderRequest) (*pb.AssignOrderResponse, error) {
	var nearestDeliveryPerson DeliveryPerson

	err := s.db.Model(&DeliveryPerson{}).
		Where("status = ?", "AVAILABLE").
		Order("ST_Distance(location::geometry, ST_Point(40.748817, -73.985428)::geometry) ASC").
		Limit(1).
		First(&nearestDeliveryPerson).Error

	if err != nil || nearestDeliveryPerson.DeliveryPersonID == "" {
		return &pb.AssignOrderResponse{Status: "FAILED"}, fmt.Errorf("no available delivery person found")
	}

	order := Order{
		OrderID:          req.OrderId,
		DeliveryPersonID: nearestDeliveryPerson.DeliveryPersonID,
		Status:           "ASSIGNED",
	}
	if err := s.db.Create(&order).Error; err != nil {
		return &pb.AssignOrderResponse{Status: "FAILED"}, err
	}

	nearestDeliveryPerson.Status = "BUSY"
	if err := s.db.Save(&nearestDeliveryPerson).Error; err != nil {
		return &pb.AssignOrderResponse{Status: "FAILED"}, err
	}

	return &pb.AssignOrderResponse{Status: "ASSIGNED"}, nil
}

func (s *OrderService) GetOrderStatus(ctx context.Context, req *pb.GetOrderStatusRequest) (*pb.GetOrderStatusResponse, error) {
	var order Order
	if err := s.db.First(&order, "order_id = ?", req.OrderId).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}

	return &pb.GetOrderStatusResponse{
		OrderId: req.OrderId,
		Status:  order.Status,
	}, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	var order Order
	if err := s.db.First(&order, "order_id = ?", req.OrderId).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}

	order.Status = req.Status
	if err := s.db.Save(&order).Error; err != nil {
		return nil, fmt.Errorf("failed to update order status")
	}

	var deliveryPerson DeliveryPerson
	if err := s.db.First(&deliveryPerson, "delivery_person_id = ?", order.DeliveryPersonID).Error; err != nil {
		return nil, fmt.Errorf("failed to find delivery person")
	}

	if req.Status == "DELIVERED" {
		deliveryPerson.Status = "AVAILABLE"
	} else if req.Status == "IN_PROGRESS" {
		deliveryPerson.Status = "BUSY"
	}

	if err := s.db.Save(&deliveryPerson).Error; err != nil {
		return nil, fmt.Errorf("failed to update delivery person status")
	}

	return &pb.UpdateOrderStatusResponse{Status: "UPDATED"}, nil
}

func (s *OrderService) GetOrdersByDeliveryPerson(ctx context.Context, req *pb.GetOrdersByDeliveryPersonRequest) (*pb.GetOrdersByDeliveryPersonResponse, error) {
	var orders []Order
	if err := s.db.Where("delivery_person_id = ?", req.DeliveryPersonId).Find(&orders).Error; err != nil {
		return nil, err
	}

	var protoOrders []*pb.Order
	for _, order := range orders {
		protoOrders = append(protoOrders, &pb.Order{
			OrderId: order.OrderID,
			Status:  order.Status,
		})
	}

	return &pb.GetOrdersByDeliveryPersonResponse{Orders: protoOrders}, nil
}
