FROM golang:1.22-alpine AS builder
WORKDIR /app

# Copy module definitions first
COPY go.mod go.sum ./
RUN go mod download

# Copy all source files into the container
COPY . .
RUN go mod tidy

# Build all Go microservices
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway api-gateway/cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service order-service/cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /inventory-service inventory-service/cmd/main.go

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /api-gateway /app/api-gateway
COPY --from=builder /order-service /app/order-service
COPY --from=builder /inventory-service /app/inventory-service
COPY --from=builder /app/features.txt /app/features.txt