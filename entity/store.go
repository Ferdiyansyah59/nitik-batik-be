package entity

import "time"

type Store struct {
	ID uint64 `json:"id" gorm:"column:id"`
	Name string `json:"name" gorm:"column:name"`
	Description string `json:"description" gorm:"column:description"`
	Whatsapp string `json:"whatsapp" gorm:"column:whatsapp"`
	Alamat string `json:"alamat" gorm:"column:alamat"`
	UserID int `json:"user_id" gorm:"column:user_id"`
	Avatar string `json:"avatar" gorm:"column:avatar"`
	Banner string `json:"banner" gorm:"column:banner"`
	CreatedAt   time.Time `json:"created_At" gorm:"column:created_at"`
    UpdatedAt   time.Time `json:"updated_At" gorm:"column:updated_at"`
}