package fulfillment

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	pb "fullfillment-service/proto"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm DB: %v", err)
	}

	return gormDB, mock, db
}

func TestAssignOrder(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	service := NewService(db)

	t.Run("Success - Assign Order", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "delivery_people" WHERE status = \$1 ORDER BY ST_Distance`).
			WithArgs("AVAILABLE").
			WillReturnRows(sqlmock.NewRows([]string{"delivery_person_id", "status", "location"}).
				AddRow("dp1", "AVAILABLE", "(37.7749,-122.4194)"))

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "orders"`).
			WithArgs("order1", "dp1", "ASSIGNED").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`UPDATE "delivery_people" SET "status"=\$1 WHERE "delivery_person_id" = \$2`).
			WithArgs("BUSY", "dp1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		req := &pb.AssignOrderRequest{OrderId: "order1"}
		resp, err := service.AssignOrder(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, "ASSIGNED", resp.Status)
	})

	t.Run("Failure - No Available Delivery Person", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "delivery_people" WHERE status = \$1 ORDER BY ST_Distance`).
			WithArgs("AVAILABLE").
			WillReturnRows(sqlmock.NewRows([]string{"delivery_person_id", "status"}))

		req := &pb.AssignOrderRequest{OrderId: "order2"}
		resp, err := service.AssignOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, "FAILED", resp.Status)
	})

	t.Run("Failure - Database Error on Create", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "delivery_people" WHERE status = \$1 ORDER BY ST_Distance`).
			WithArgs("AVAILABLE").
			WillReturnRows(sqlmock.NewRows([]string{"delivery_person_id", "status"}).AddRow("dp1", "AVAILABLE"))

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "orders"`).
			WithArgs("order3", "dp1", "ASSIGNED").
			WillReturnError(errors.New("some database error"))
		mock.ExpectRollback()

		req := &pb.AssignOrderRequest{OrderId: "order3"}
		resp, err := service.AssignOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, "FAILED", resp.Status)
	})
}

func TestGetOrderStatus(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	service := NewService(db)

	t.Run("Success - Get Order Status", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE order_id = \$1 ORDER BY "orders"."order_id" LIMIT \$2`).
			WithArgs("order1", 1). // Include the limit argument
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "status"}).AddRow("order1", "DELIVERED"))

		req := &pb.GetOrderStatusRequest{OrderId: "order1"}
		resp, err := service.GetOrderStatus(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, "order1", resp.OrderId)
		assert.Equal(t, "DELIVERED", resp.Status)
	})

	t.Run("Failure - Order Not Found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE order_id = \$1`).
			WithArgs("order2").
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "status"}))

		req := &pb.GetOrderStatusRequest{OrderId: "order2"}
		resp, err := service.GetOrderStatus(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
func TestUpdateOrderStatus(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	service := NewService(db)

	t.Run("Success - Update Order Status to Delivered", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE order_id = \$1 ORDER BY "orders"."order_id" LIMIT \$2`).
			WithArgs("order1", 1).
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "delivery_person_id", "status"}).AddRow("order1", "dp1", "ASSIGNED"))

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "orders" SET "status"=\$1, "updated_at"=\$2 WHERE "order_id" = \$3`).
			WithArgs("DELIVERED", sqlmock.AnyArg(), "order1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`UPDATE "delivery_people" SET "status"=\$1 WHERE "delivery_person_id" = \$2`).
			WithArgs("AVAILABLE", "dp1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		req := &pb.UpdateOrderStatusRequest{OrderId: "order1", Status: "DELIVERED"}
		resp, err := service.UpdateOrderStatus(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, "UPDATED", resp.Status)
	})

	t.Run("Failure - Order Not Found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE order_id = \$1`).
			WithArgs("order2").
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "status"}))

		req := &pb.UpdateOrderStatusRequest{OrderId: "order2", Status: "DELIVERED"}
		resp, err := service.UpdateOrderStatus(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Failure - Delivery Person Not Found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE order_id = \$1`).
			WithArgs("order1").
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "delivery_person_id", "status"}).AddRow("order1", "dp1", "ASSIGNED"))

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "orders" SET "status"=\$1, "updated_at"=\$2 WHERE "order_id" = \$3`).
			WithArgs("IN_PROGRESS", sqlmock.AnyArg(), "order1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`UPDATE "delivery_people" SET "status"=\$1 WHERE "delivery_person_id" = \$2`).
			WithArgs("BUSY", "dp1").
			WillReturnError(errors.New("failed to find delivery person"))
		mock.ExpectRollback()

		req := &pb.UpdateOrderStatusRequest{OrderId: "order1", Status: "IN_PROGRESS"}
		resp, err := service.UpdateOrderStatus(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGetOrdersByDeliveryPerson(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	service := NewService(db)

	t.Run("Success - Get Orders by Delivery Person", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE delivery_person_id = \$1`).
			WithArgs("dp1").
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "delivery_person_id", "status"}).AddRow("order1", "dp1", "DELIVERED"))

		req := &pb.GetOrdersByDeliveryPersonRequest{DeliveryPersonId: "dp1"}
		resp, err := service.GetOrdersByDeliveryPerson(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Orders, 1)
		assert.Equal(t, "order1", resp.Orders[0].OrderId)
		assert.Equal(t, "DELIVERED", resp.Orders[0].Status)
	})

	t.Run("Failure - No Orders Found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE delivery_person_id = \$1`).
			WithArgs("dp2").
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "delivery_person_id", "status"}))

		req := &pb.GetOrdersByDeliveryPersonRequest{DeliveryPersonId: "dp2"}
		resp, err := service.GetOrdersByDeliveryPerson(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Orders, 0)
	})
}
