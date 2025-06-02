package entity

import "time"

type Product struct {
	ID int `json:"id" gorm:"column:id"`
	Slug        string         `json:"slug" gorm:"column:slug;uniqueIndex"`
	Name string `json:"name" gorm:"column:name"`
	Description string `json:"description" gorm:"column:description"`
	Harga float64 `json:"harga" gorm:"column:harga"`
	StoreID int `json:"store_id" gorm:"column:store_id"`
	Thumbnail string `json:"thumbnail" gorm:"column:thumbnail"`
	Images      []ProductImage `json:"images" gorm:"foreignKey:ProductID"`
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at"`
}

