// dto/product_dto.go
package dto

import "time"

type CreateProductDTO struct {
	Name        string  `form:"name" binding:"required"`
	Description string  `form:"description" binding:"required"`
	Harga       float64 `form:"harga" binding:"required,gt=0"`
	StoreID     int     `form:"store_id" binding:"required"`
	CategoryID  int		`form:"category_id" binding:"required"`
}

type UpdateProductDTO struct {
	Name        string  `json:"name" form:"name"`
	Description string  `json:"description" form:"description"`
	Harga       float64 `json:"harga" form:"harga" binding:"omitempty,gt=0"`
	CategoryID  int     `json:"category_id" form:"category_id"` 
}

type ProductImageDTO struct {
	ID    int    `json:"id,omitempty"`
	Image string `json:"image"`
}

type ProductResponse struct {
	ID          int               `json:"id"`
	Slug        string            `json:"slug"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Harga       float64           `json:"harga"`
	StoreID     int               `json:"store_id"`
	CategoryID  int				  `json:"category_id"`
	Thumbnail   string            `json:"thumbnail"`
	Images      []ProductImageDTO `json:"images"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// DTO untuk menampilkan card produk (ringkas tanpa gambar tambahan)
type ProductCardResponse struct {
	ID          int       `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Harga       float64   `json:"harga"`
	StoreID     int       `json:"store_id"`
	Thumbnail   string    `json:"thumbnail"`
	CreatedAt   time.Time `json:"created_at"`
}

type PublicProductCard struct {
	ID           int       `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Harga        float64   `json:"harga"`
	StoreID      int       `json:"store_id"`
	StoreName    string    `json:"store_name,omitempty"` 
	CategoryID   int       `json:"category_id,omitempty"`
	CategoryName string    `json:"category_name,omitempty"`
	CategorySlug string    `json:"category_slug,omitempty"`
	Thumbnail    string    `json:"thumbnail"`
	CreatedAt    time.Time `json:"created_at"`
}