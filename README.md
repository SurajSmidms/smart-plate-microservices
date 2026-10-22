# smart-plate-microservices: Restaurant & Inventory Microservices Platform

A production-grade, multi-language microservices platform designed for modern restaurant operations, real-time inventory tracking, and intelligent AI assistance. This system pairs high-performance **Go** microservices with a robust **Python FastAPI AI agent layer**, orchestrated seamlessly via Docker Compose.

---

## Architecture Overview

The system is decoupled into specialized services communicating via HTTP REST, gRPC, and container-level networking:

* **API Gateway (`Go / Chi Framework` on Port `8080`)**: Serves as the single entry point, managing routing, payload validation, RBAC security headers, and proxying requests.
* **Order Service (`Go / gRPC` on Port `50052`)**: Manages core order processing, enforcing payment verification gatekeeping before coordinating raw material deductions.
* **Inventory Service (`Go / gRPC` on Port `50051`)**: Tracks live stock levels, handles automated restocking workflows, and logs raw material scarcity alerts.
* **AI Agent Layer (`Python / FastAPI` on Port `8000`)**: Powered by the Google Gemini API (`google-genai SDK`) and governed by custom behavioral specifications (`features.txt`) to assist customers and store owners.

---

## Core Features

* **Role-Based Access Control (RBAC)**: Secure multi-tenant endpoints separating standard **Customers** (`USER`) from **Store Owners** (`ADMIN`).
* **Payment Gatekeeping**: Enforces strict pre-order transaction verification; unverified orders trigger immediate `402 Payment Required` rejections.
* **Smart Inventory Tracking**: Real-time gRPC communication between services to update raw material counts and trigger scarcity warnings.
* **AI-Powered Assistant**: Natural language endpoint governed by behavior guidelines for answering menu inquiries, recipe recommendations, and management queries.

---
