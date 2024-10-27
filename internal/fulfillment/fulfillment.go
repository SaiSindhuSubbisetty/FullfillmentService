package fulfillment

type Order struct {
	OrderID          string `gorm:"primaryKey"`
	DeliveryPersonID string
	Status           string
	CreatedAt        int64
	UpdatedAt        int64
}

type DeliveryPerson struct {
	DeliveryPersonID string `gorm:"column:delivery_person_id;primaryKey"`
	Name             string `gorm:"column:name"`
	Status           string `gorm:"column:status"`
	Location         *Point `gorm:"column:location"`
}

type Point struct {
	Lat float64
	Lng float64
}
