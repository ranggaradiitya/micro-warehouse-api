# Warehouse & Point of Sales - Microservices API

Sistem manajemen warehouse dan point of sales berbasis microservices architecture menggunakan Golang, PostgreSQL, Redis, dan RabbitMQ.

## 📋 Daftar Isi

-   [Arsitektur](#arsitektur)
-   [Tech Stack](#tech-stack)
-   [Microservices](#microservices)
-   [Prerequisites](#prerequisites)
-   [Instalasi](#instalasi)
-   [Konfigurasi](#konfigurasi)
-   [Menjalankan Aplikasi](#menjalankan-aplikasi)
-   [API Endpoints](#api-endpoints)
-   [Database Schema](#database-schema)
-   [Monitoring](#monitoring)

## 🏗️ Arsitektur

Proyek ini menggunakan arsitektur microservices dengan komponen utama:

```
┌─────────────────┐
│   API Gateway   │ (Port 8080)
│  Rate Limiter   │
└────────┬────────┘
         │
    ┌────┴────────────────────────────────────┐
    │                                         │
    ▼              ▼         ▼        ▼       ▼
┌────────┐   ┌─────────┐ ┌──────┐ ┌─────┐ ┌────────┐
│  User  │   │ Product │ │Trans │ │Merch│ │Warehouse│
│Service │   │ Service │ │Serv  │ │Serv │ │ Service │
│  8081  │   │  8082   │ │ 8085 │ │8084 │ │  8083   │
└───┬────┘   └────┬────┘ └───┬──┘ └──┬──┘ └────┬────┘
    │             │           │       │         │
    └─────────────┴───────────┴───────┴─────────┘
                  │           │
            ┌─────┴─────┐ ┌──┴───────┐
            │   Redis   │ │ RabbitMQ │
            │   6379    │ │   5672   │
            └───────────┘ └──────────┘
```

## 🛠️ Tech Stack

### Backend

-   **Go** 1.21+ - Programming Language
-   **Fiber** v2 - Web Framework
-   **GORM** - ORM untuk database
-   **PostgreSQL** - Database untuk setiap service
-   **Redis** - Caching & Rate Limiting
-   **RabbitMQ** - Message Queue
-   **JWT** - Authentication & Authorization
-   **Docker & Docker Compose** - Containerization

### Tools & Libraries

-   **Cobra** - CLI Framework
-   **Godotenv** - Environment Variables
-   **Validator** - Data Validation
-   **Supabase** - Cloud Storage (Optional)

## 🔧 Microservices

### 1. API Gateway (Port 8080)

**Fungsi:**

-   Entry point untuk semua request
-   Authentication & Authorization
-   Rate Limiting dengan Redis
-   Request routing ke service yang tepat

**Fitur:**

-   JWT Token Validation
-   Redis-based Rate Limiter
-   CORS Configuration
-   Health Check Endpoint

### 2. User Service (Port 8081)

**Fungsi:**

-   Manajemen user (CRUD)
-   Authentication (Login)
-   Role Management
-   Assign Role ke User

**Database:** `warehouse_user_db` (Port 5432)

**Endpoints:**

-   `POST /api/v1/auth/login` - Login
-   `GET/POST/PUT/DELETE /api/v1/users/*` - User CRUD
-   `GET/POST/PUT/DELETE /api/v1/roles/*` - Role Management
-   `POST /api/v1/assign-role/*` - Assign Role

### 3. Product Service (Port 8082)

**Fungsi:**

-   Manajemen produk (CRUD)
-   Manajemen kategori produk
-   Upload gambar produk

**Database:** `warehouse_product_db` (Port 5433)

**Endpoints:**

-   `GET/POST/PUT/DELETE /api/v1/products/*` - Product CRUD
-   `GET/POST/PUT/DELETE /api/v1/categories/*` - Category CRUD
-   `POST /api/v1/upload-product/*` - Upload Product Image

### 4. Warehouse Service (Port 8083)

**Fungsi:**

-   Manajemen warehouse
-   Manajemen stok produk di warehouse
-   Transfer stok antar warehouse

**Database:** `warehouse_warehouse_db` (Port 5437)

**Endpoints:**

-   `GET/POST/PUT/DELETE /api/v1/warehouses/*` - Warehouse CRUD
-   `GET/POST/PUT/DELETE /api/v1/warehouse-products/*` - Warehouse Stock Management
-   `POST /api/v1/upload-warehouse/*` - Upload Warehouse Images

### 5. Merchant Service (Port 8084)

**Fungsi:**

-   Manajemen merchant/toko
-   Manajemen produk merchant
-   Integrasi dengan warehouse

**Database:** `warehouse_merchant_db` (Port 5435)

**Endpoints:**

-   `GET/POST/PUT/DELETE /api/v1/merchants/*` - Merchant CRUD
-   `GET/POST/PUT/DELETE /api/v1/merchant-products/*` - Merchant Product Management
-   `POST /api/v1/upload-merchant/*` - Upload Merchant Images

### 6. Transaction Service (Port 8085)

**Fungsi:**

-   Manajemen transaksi penjualan
-   Integrasi pembayaran (Midtrans)
-   Dashboard & reporting

**Database:** `warehouse_transaction_db` (Port 5434)

**Endpoints:**

-   `GET/POST/PUT/DELETE /api/v1/transactions/*` - Transaction CRUD
-   `POST /api/v1/midtrans/callback` - Midtrans Payment Callback
-   `GET /api/v1/dashboard/*` - Dashboard Data

### 7. Notification Service (Port 8086)

**Fungsi:**

-   Pengiriman email notification
-   RabbitMQ consumer untuk async notification

**Database:** `warehouse_notification_db` (Port 5436)

## 📦 Prerequisites

Pastikan tools berikut sudah terinstall:

-   **Go** 1.21 atau lebih tinggi
-   **Docker** & **Docker Compose**
-   **PostgreSQL** (jika run tanpa Docker)
-   **Redis** (jika run tanpa Docker)
-   **RabbitMQ** (jika run tanpa Docker)
-   **Git**

## 🚀 Instalasi

### 1. Clone Repository

```bash
git clone <repository-url>
cd micro-warehouse-api-main
```

### 2. Setup Environment Variables

Setiap service memiliki file `env.example`. Copy dan rename menjadi `.env`:

```bash
# API Gateway
cp api-gateway/env.example api-gateway/.env

# User Service
cp user-service/env.example user-service/.env

# Product Service
cp product-service/env.example product-service/.env

# Warehouse Service
cp warehouse-service/env.example warehouse-service/.env

# Merchant Service
cp merchant-service/env.example merchant-service/.env

# Transaction Service
cp transaction-service/env.example transaction-service/.env

# Notification Service
cp notification-service/env.example notification-service/.env
```

### 3. Konfigurasi Environment Variables

Edit file `.env` di setiap service sesuai kebutuhan. Contoh konfigurasi untuk `merchant-service/.env`:

```env
APP_ENV=development
APP_PORT=8084

DATABASE_PORT=5435
DATABASE_HOST=postgres-merchant
DATABASE_USER=postgres
DATABASE_PASSWORD=lokal
DATABASE_NAME=warehouse_merchant_db
DATABASE_MAX_OPEN_CONNECTION=100
DATABASE_MAX_IDLE_CONNECTION=20

RABBITMQ_HOST=warehouse_rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

REDIS_HOST=warehouse_redis
REDIS_PORT=6379

SUPABASE_URL=your-supabase-url
SUPABASE_KEY=your-supabase-key
SUPABASE_BUCKET=your-bucket-name

URL_API_GATEWAY=http://localhost:8080
```

## 🏃 Menjalankan Aplikasi

### Menggunakan Docker Compose (Recommended)

```bash
# Build dan jalankan semua services
docker-compose up -d --build

# Lihat logs
docker-compose logs -f

# Stop semua services
docker-compose down

# Stop dan hapus volumes (reset database)
docker-compose down -v
```

### Menjalankan Service Individual

```bash
# Masuk ke direktori service
cd user-service

# Install dependencies
go mod download

# Jalankan service
go run main.go start
```

## 🔍 API Endpoints

### Health Check

```bash
GET http://localhost:8080/health
```

Response:

```json
{
    "status": "OK",
    "message": "API Gateway is running",
    "services": {
        "user-service": "http://localhost:8081",
        "product-service": "http://localhost:8082",
        "warehouse-service": "http://localhost:8083",
        "merchant-service": "http://localhost:8084",
        "transaction-service": "http://localhost:8085",
        "notification-service": "http://localhost:8086"
    }
}
```

### Authentication

```bash
# Login
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

Response:

```json
{
    "status": "success",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIs...",
        "user": {
            "id": 1,
            "email": "user@example.com",
            "name": "John Doe"
        }
    }
}
```

### Protected Endpoints

Semua endpoint selain `/health` dan `/api/v1/auth/login` memerlukan JWT token:

```bash
# Example: Get All Products
GET http://localhost:8080/api/v1/products
Authorization: Bearer <your-jwt-token>
```

## 🗄️ Database Schema

Setiap service memiliki database terpisah:

| Database Name             | Port | Service              |
| ------------------------- | ---- | -------------------- |
| warehouse_user_db         | 5432 | User Service         |
| warehouse_product_db      | 5433 | Product Service      |
| warehouse_transaction_db  | 5434 | Transaction Service  |
| warehouse_merchant_db     | 5435 | Merchant Service     |
| warehouse_notification_db | 5436 | Notification Service |
| warehouse_warehouse_db    | 5437 | Warehouse Service    |

## 📊 Monitoring

### RabbitMQ Management Console

```
URL: http://localhost:15672
Username: guest
Password: guest
```

### Database Connections

Gunakan tool seperti DBeaver, pgAdmin, atau TablePlus:

```
Host: localhost
Port: 5432-5437 (sesuai service)
Username: postgres
Password: lokal
```

### Redis CLI

```bash
docker exec -it warehouse_redis redis-cli
```

## 🔒 Security Features

-   **JWT Authentication** - Semua endpoint protected
-   **Rate Limiting** - Redis-based rate limiter di API Gateway
-   **CORS** - Configured untuk cross-origin requests
-   **Request Validation** - Input validation di setiap service
-   **Internal Request Headers** - Komunikasi antar service diamankan

## 📝 Development

### Menambah Service Baru

1. Buat folder service baru
2. Setup struktur folder (app, cmd, configs, controller, dll)
3. Tambahkan konfigurasi di `docker-compose.yml`
4. Tambahkan routing di `api-gateway`

### Testing

```bash
# Unit tests
cd <service-name>
go test ./...

# Integration tests
go test -tags=integration ./...
```

### Build Production

```bash
# Build single service
cd <service-name>
go build -o bin/app main.go

# Build dengan Docker
docker build -t <service-name>:latest .
```

## 🐛 Troubleshooting

### Service tidak bisa connect ke database

```bash
# Check database health
docker-compose ps

# Restart database
docker-compose restart postgres-<service>
```

### Port sudah digunakan

Edit `docker-compose.yml` dan ubah port mapping:

```yaml
ports:
    - "8081:8081" # Ubah port kiri (host)
```

### Redis connection error

```bash
# Restart Redis
docker-compose restart redis

# Check Redis logs
docker-compose logs redis
```

## 📄 License

[MIT License](LICENSE)

## 📧 Contact

Untuk pertanyaan atau dukungan, hubungi adityarangga990@gmail.com

---

**Happy Coding! 🚀**
