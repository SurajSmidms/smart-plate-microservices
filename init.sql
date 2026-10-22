CREATE DATABASE restaurant_db;
\c restaurant_db;

CREATE TABLE raw_materials (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    stock_quantity INT NOT NULL,
    min_threshold INT NOT NULL
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_name VARCHAR(100) NOT NULL,
    cuisine_type VARCHAR(100) NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    payment_status VARCHAR(50) DEFAULT 'PENDING',
    order_status VARCHAR(50) DEFAULT 'CREATED',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed initial ingredients for world cuisines (Italian, Mexican, etc.)
INSERT INTO raw_materials (name, stock_quantity, min_threshold) VALUES
('wheat', 50, 10),
('cheese', 6, 10),  -- Triggers initial scarcity alert condition
('tomato', 30, 8),
('sausages', 15, 5);