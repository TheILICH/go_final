package models

import (
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	User        User `gorm:"foreignkey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	UserID      uint
	OrderStatus OrderStatus `gorm:"type:varchar(100);not null"`
}

type OrderStatus string

const (
	PENDING   OrderStatus = "pending"
	ACCEPTED  OrderStatus = "accepted"
	READY     OrderStatus = "ready"
	OUT       OrderStatus = "out"
	DELIVERED OrderStatus = "delivered"
	CONFIRMED OrderStatus = "confirmed"
	CANCELED  OrderStatus = "canceled"
)

type OrderItems struct {
	gorm.Model
	Order     Order   `gorm:"foreignkey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Product   Product `gorm:"foreignkey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	OrderID   uint
	ProductID uint
	Quantity  int
	Price     int
}

type CartItemRequest struct {
	Product     Product `gorm:"foreignkey:ProductID"`
	OrderItemID uint    `json:"order_item_id,omitempty"`
	ProductID   uint    `json:"product_id"`
	Quantity    int     `json:"quantity"`
}

type CartRequest struct {
	CartItems []CartItemRequest `json:"order"`
}

// TODO: switch mechanism between order statuses
// TODO: check routes again
// TODO: exlude User, updatedAt from getCurrentOrder
