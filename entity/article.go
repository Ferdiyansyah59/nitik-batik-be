package entity

import "time"

type Article struct {
    ID          uint64    `json:"id" gorm:"column:id"`
    Title       string    `json:"title" gorm:"column:title"`
    Slug        string    `json:"slug" gorm:"column:slug"`
    Description string    `json:"description" gorm:"column:description"`
    Excerpt     string    `json:"excerpt" gorm:"column:excerpt"`
    ImageURL    string    `json:"imageUrl" gorm:"column:image_url"`  // Tambahkan tag gorm
    CreatedAt   time.Time `json:"created_At" gorm:"column:created_at"`
    UpdatedAt   time.Time `json:"updated_At" gorm:"column:updated_at"`
}
