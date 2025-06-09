package entity

import "time"

type ProductImage struct {
	ID        int       `json:"id" gorm:"column:id;primaryKey"`
	Image     string    `json:"image" gorm:"column:image"`
	ProductID int       `json:"product_id" gorm:"column:product_id"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}
