package entity

import "time"

type Product struct {
	ID int `json:"id" gorm:"column:id"`
	Slug        string         `json:"slug" gorm:"column:slug;uniqueIndex"`
	Name string `json:"name" gorm:"column:name"`
	Description string `json:"description" gorm:"column:description"`
	Harga float64 `json:"harga" gorm:"column:harga"`
	StoreID int `json:"store_id" gorm:"column:store_id"`
	CategoryID int `json:"category_id" gorm:"column:category_id"`
	Thumbnail string `json:"thumbnail" gorm:"column:thumbnail"`
	Images      []ProductImage `json:"images" gorm:"foreignKey:ProductID"`
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at"`
}


type ProductCard struct {
	ID           int             `json:"id" gorm:"column:id;primaryKey"`
	Slug         string          `json:"slug" gorm:"column:slug;uniqueIndex"`
	Name         string          `json:"name" gorm:"column:name"`
	Description  string          `json:"description" gorm:"column:description"`
	Harga        float64         `json:"harga" gorm:"column:harga"`
	StoreID      int             `json:"store_id" gorm:"column:store_id"` // Foreign key ke tabel stores
	StoreName    string          `json:"store_name,omitempty" gorm:"column:store_name"`   // Akan diisi oleh query JOIN
	CategoryID   int             `json:"category_id" gorm:"column:category_id"`
	CategoryName string          `json:"category_name,omitempty" gorm:"column:category_name"`
	CategorySlug  string          `json:"category_slug,omitempty" gorm:"column:slug"`
	Thumbnail    string          `json:"thumbnail" gorm:"column:thumbnail"`
	Images       []ProductImage  `json:"images" gorm:"foreignKey:ProductID"`
	Category     ProductCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID;references:ID"`
	Store        Store           `json:"store,omitempty" gorm:"foreignKey:StoreID;references:ID"` // Relasi GORM ke Store
	CreatedAt    time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"column:updated_at"`
}


func (pc ProductCard) TableName() string {
	return "products"
}