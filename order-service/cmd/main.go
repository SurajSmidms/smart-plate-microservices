package main

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	pb "restaurant-grpc-system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	inventoryClient pb.InventoryServiceClient
}

// PlaceOrder verifies payment status and delegates inventory deduction via gRPC
func (s *OrderServer) PlaceOrder(ctx context.Context, req *pb.OrderRequest) (*pb.OrderResponse, error) {
	// Rule: Once the payment status is done, only then should the order be placed
	if !req.PaymentDone {
		return nil, errors.New("payment status pending: order rejected. Payment must be verified before placement")
	}

	log.Printf("[ORDER] Processing order for customer '%s' (%s - %s)", req.CustomerName, req.CuisineType, req.Items)

	// Internal gRPC call to Inventory Service to deduct raw materials
	invCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	deductRes, err := s.inventoryClient.DeductStock(invCtx, &pb.DeductRequest{
		Ingredients: req.Ingredients,
	})
	if err != nil {
		return nil, errors.New("failed to update inventory via gRPC: " + err.Error())
	}

	// Log any scarcity alerts received from the inventory service response
	for _, alert := range deductRes.ScarcityAlerts {
		log.Printf("[ALERT PROCESSED] %s", alert)
	}

	return &pb.OrderResponse{
		Status:  "SUCCESS",
		Message: "Order placed successfully, payment verified, and raw materials updated via gRPC channels.",
	}, nil
}

func main() {
	// Establish internal gRPC connection to the Inventory Service on port 50051
	conn, err := grpc.Dial("inventory-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect to inventory gRPC service: %v", err)
	}
	defer conn.Close()

	invClient := pb.NewInventoryServiceClient(conn)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen on port 50052: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, &OrderServer{inventoryClient: invClient})

	log.Println("Order gRPC Service running on port 50052...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}