package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "restaurant-grpc-system/proto"
	"google.golang.org/grpc"
)

type InventoryServer struct {
	pb.UnimplementedInventoryServiceServer
}

// In-memory stock store representing world cuisine raw materials (Italian, Mexican, etc.)
var stockStore = map[string]int32{
	"wheat":   50,
	"cheese":  6,  // Below threshold -> triggers scarcity WhatsApp alert
	"tomato":  30,
	"sausage": 15,
}

var thresholds = map[string]int32{
	"wheat":   10,
	"cheese":  10,
	"tomato":  8,
	"sausage": 5,
}

// DeductStock updates inventory counts and evaluates safety thresholds for WhatsApp alerts
func (s *InventoryServer) DeductStock(ctx context.Context, req *pb.DeductRequest) (*pb.DeductResponse, error) {
	var alerts []string

	for item, qty := range req.Ingredients {
		current, exists := stockStore[item]
		if !exists {
			return nil, fmt.Errorf("raw material '%s' not found in inventory catalog", item)
		}

		newStock := current - qty
		stockStore[item] = newStock
		log.Printf("[INVENTORY] Updated item '%s': deducted %d units. Remaining stock: %d", item, qty, newStock)

		// Check if stock has dropped below or equal to the safety threshold
		if threshold, hasThreshold := thresholds[item]; hasThreshold && newStock <= threshold {
			alertMsg := fmt.Sprintf("🚨 WHATSAPP ALERT TO ADMIN: Critical scarcity for '%s'! Remaining stock: %d units (Threshold: %d). Immediate restock required.", item, newStock, threshold)
			log.Println(alertMsg)
			alerts = append(alerts, alertMsg)
		}
	}

	return &pb.DeductResponse{
		Success:        true,
		Message:        "Inventory successfully updated and evaluated via gRPC",
		ScarcityAlerts: alerts,
	}, nil
}

// CheckScarcity allows external inspection of current low-stock items
func (s *InventoryServer) CheckScarcity(ctx context.Context, req *pb.ScarcityRequest) (*pb.ScarcityResponse, error) {
	var alerts []string
	for item, stock := range stockStore {
		if stock <= thresholds[item] {
			alerts = append(alerts, fmt.Sprintf("Scarcity Warning: '%s' has %d units left.", item, stock))
		}
	}
	return &pb.ScarcityResponse{Alerts: alerts}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterInventoryServiceServer(grpcServer, &InventoryServer{})

	log.Println("Inventory gRPC Service running on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}