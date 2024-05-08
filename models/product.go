package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name        string `json:"name" gorm:"unique"`
	Quantity    int    `json:"quantity"`
	Description string `json:"description"`
	Price       int    `json:"price"`
}
