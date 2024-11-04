package fulfillment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderInitialization(t *testing.T) {
	order := Order{
		OrderID:          "order123",
		DeliveryPersonID: "dp123",
		Status:           "ASSIGNED",
		CreatedAt:        1622547800,
		UpdatedAt:        1622547900,
	}

	assert.Equal(t, "order123", order.OrderID, "OrderID should be 'order123'")
	assert.Equal(t, "dp123", order.DeliveryPersonID, "DeliveryPersonID should be 'dp123'")
	assert.Equal(t, "ASSIGNED", order.Status, "Status should be 'ASSIGNED'")
	assert.Equal(t, int64(1622547800), order.CreatedAt, "CreatedAt should be 1622547800")
	assert.Equal(t, int64(1622547900), order.UpdatedAt, "UpdatedAt should be 1622547900")
}

func TestOrderInitializationFailure(t *testing.T) {
	order := Order{}

	assert.Empty(t, order.OrderID, "OrderID should be empty")
	assert.Empty(t, order.DeliveryPersonID, "DeliveryPersonID should be empty")
	assert.Empty(t, order.Status, "Status should be empty")
	assert.Zero(t, order.CreatedAt, "CreatedAt should be zero")
	assert.Zero(t, order.UpdatedAt, "UpdatedAt should be zero")
}

func TestDeliveryPersonInitialization(t *testing.T) {
	location := &Point{Lat: 40.748817, Lng: -73.985428}
	deliveryPerson := DeliveryPerson{
		DeliveryPersonID: "dp123",
		Name:             "John Doe",
		Status:           "AVAILABLE",
		Location:         location,
	}

	assert.Equal(t, "dp123", deliveryPerson.DeliveryPersonID, "DeliveryPersonID should be 'dp123'")
	assert.Equal(t, "John Doe", deliveryPerson.Name, "Name should be 'John Doe'")
	assert.Equal(t, "AVAILABLE", deliveryPerson.Status, "Status should be 'AVAILABLE'")
	assert.Equal(t, location, deliveryPerson.Location, "Location should be the same as the initialized Point")
}

func TestDeliveryPersonInitializationFailure(t *testing.T) {
	deliveryPerson := DeliveryPerson{}

	assert.Empty(t, deliveryPerson.DeliveryPersonID, "DeliveryPersonID should be empty")
	assert.Empty(t, deliveryPerson.Name, "Name should be empty")
	assert.Empty(t, deliveryPerson.Status, "Status should be empty")
	assert.Nil(t, deliveryPerson.Location, "Location should be nil")
}

func TestPointInitialization(t *testing.T) {
	point := Point{Lat: 40.748817, Lng: -73.985428}

	assert.Equal(t, 40.748817, point.Lat, "Lat should be 40.748817")
	assert.Equal(t, -73.985428, point.Lng, "Lng should be -73.985428")
}

func TestPointInitializationFailure(t *testing.T) {
	point := Point{}

	assert.Zero(t, point.Lat, "Lat should be zero")
	assert.Zero(t, point.Lng, "Lng should be zero")
}
