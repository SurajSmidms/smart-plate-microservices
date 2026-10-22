package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "restaurant-grpc-system/proto"
)

type Gateway struct {
	orderClient pb.OrderServiceClient
	users       map[string]string // username -> role
	mu          sync.Mutex
}

type SignupRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"` // "ADMIN" or "USER" (defaults to USER)
}

// RBAC Middleware checking user role from headers
func (gw *Gateway) rbacMiddleware(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username := r.Header.Get("X-Username")
			if username == "" {
				http.Error(w, `{"error": "Unauthorized: Missing X-Username header"}`, http.StatusUnauthorized)
				return
			}

			gw.mu.Lock()
			role, exists := gw.users[username]
			gw.mu.Unlock()

			if !exists {
				http.Error(w, `{"error": "Unauthorized: User not registered"}`, http.StatusUnauthorized)
				return
			}

			if strings.ToUpper(role) != strings.ToUpper(requiredRole) {
				http.Error(w, `{"error": "Forbidden: Insufficient privileges"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	conn, err := grpc.Dial("order-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Order gRPC service: %v", err)
	}
	defer conn.Close()

	gw := &Gateway{
		orderClient: pb.NewOrderServiceClient(conn),
		users:       make(map[string]string),
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Route("/api", func(r chi.Router) {
		r.Post("/signup", gw.handleSignup)
		r.Post("/orders", gw.handleCreateOrder)
		r.Post("/ai/chat", gw.handleAIChat)

		// Secured Store Owner / Admin Routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(gw.rbacMiddleware("ADMIN"))
			r.Get("/inventory/status", gw.handleAdminInventoryStatus)
			r.Post("/inventory/restock", gw.handleAdminRestock)
		})
	})

	log.Println("API Gateway running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func (gw *Gateway) handleSignup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid payload"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, `{"error": "username is required"}`, http.StatusBadRequest)
		return
	}

	role := strings.ToUpper(req.Role)
	if role != "ADMIN" {
		role = "USER" // Default to USER (customer)
	}

	gw.mu.Lock()
	gw.users[req.Username] = role
	gw.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "User registered successfully",
		"username": req.Username,
		"role":     role,
	})
}

type OrderHTTPRequest struct {
	CustomerName string           `json:"customer_name"`
	CuisineType  string           `json:"cuisine_type"`
	Items        string           `json:"items"`
	Ingredients  map[string]int32 `json:"ingredients"`
	PaymentDone  bool             `json:"payment_done"`
}

func (gw *Gateway) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req OrderHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request payload"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := gw.orderClient.PlaceOrder(ctx, &pb.OrderRequest{
		CustomerName: req.CustomerName,
		CuisineType:  req.CuisineType,
		Items:        req.Items,
		Ingredients:  req.Ingredients,
		PaymentDone:  req.PaymentDone,
	})

	if err != nil {
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (gw *Gateway) handleAIChat(w http.ResponseWriter, r *http.Request) {
	// Proxy prompt to Python FastAPI AI agent service
	aiReqBody, _ := io.ReadAll(r.Body)
	resp, err := http.Post("http://ai-agent:8000/chat", "application/json", bytes.NewBuffer(aiReqBody))
	if err != nil {
		http.Error(w, `{"error": "failed to connect to AI agent service"}`, http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (gw *Gateway) handleAdminInventoryStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"store_owner_status": "OK",
		"current_inventory": map[string]int32{
			"wheat":   48,
			"cheese":  5,
			"tomato":  29,
			"sausage": 15,
		},
		"scarcity_alerts": []string{"Critical: cheese stock is at 5 units (below threshold of 10)"},
	})
}

func (gw *Gateway) handleAdminRestock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "SUCCESS",
		"message": "Store owner successfully restocked low inventory items.",
	})
}