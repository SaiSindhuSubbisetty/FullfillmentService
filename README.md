# Fulfillment Service

The **Fulfillment Service** is a core microservice in a food delivery system that manages order delivery logistics. It assigns delivery personnel to orders, tracks their status and location, and facilitates communication through gRPC APIs.

---

## 📦 Features

- **Order Management**: Track the delivery status of orders.
- **Delivery Personnel Management**: Manage delivery agents, including their availability and real-time location.
- **gRPC APIs**: Expose high-performance APIs using gRPC for inter-service communication.
- **PostgreSQL Database**: Uses a relational database for persistence.
- **Database Migrations**: Powered by `golang-migrate` to manage schema changes.

---

## 🧱 Data Models

### `Order`
Represents a customer order and its fulfillment status.

```go
type Order struct {
	OrderID          string `gorm:"primaryKey"`
	DeliveryPersonID string
	Status           string
	CreatedAt        int64
	UpdatedAt        int64
}
```

### `DeliveryPerson`
Represents a delivery agent.

```go
type DeliveryPerson struct {
	DeliveryPersonID string `gorm:"column:delivery_person_id;primaryKey"`
	Name             string
	Status           string
	Location         *Point
}
```

### `Point`
Represents geolocation coordinates.

```go
type Point struct {
	Lat float64
	Lng float64
}
```

---

## 🚀 Getting Started

### ✅ Prerequisites

- Go 1.20+
- PostgreSQL
- `protoc` for gRPC and `.proto` file compilation
- `golang-migrate` for database migrations

### 🔧 Setup Instructions

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/SaiSindhuSubbisetty/fulfillment-service.git
   cd fulfillment-service
   ```

2. **Configure the Database**:
   Edit `config/config.go` to match your PostgreSQL setup.

3. **Run Migrations**:
   ```bash
   migrate -path migrations -database "postgres://username:password@localhost:5432/fulfillment_db?sslmode=disable" up
   ```

4. **Generate gRPC Code (if needed)**:
   ```bash
   protoc --go_out=. --go-grpc_out=. proto/fulfillment.proto
   ```

5. **Run the Service**:
   ```bash
   go run main.go
   ```

   Service will be available at `localhost:50051`.

---

## 🔌 gRPC APIs

The service exposes the following core methods via gRPC:

- `AssignDeliveryPerson`
- `UpdateOrderStatus`
- `TrackDeliveryPerson`
- `GetOrderDetails`

Refer to `proto/fulfillment.proto` for complete definitions.

---

## 🛠 Tech Stack

- **Go (Golang)**: Core language
- **gRPC**: Communication protocol
- **GORM**: ORM for database access
- **PostgreSQL**: Persistent storage
- **golang-migrate**: Database migrations

---

## 📂 Project Structure

```
├── config/              # Database configuration
├── internal/fulfillment # Core business logic
├── migrations/          # SQL migration files
├── proto/               # gRPC .proto files
├── main.go              # Application entry point
└── go.mod/go.sum        # Dependencies
```

---

## 📄 License

This project is licensed under the MIT License.
