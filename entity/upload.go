package entity

import "time"

type UploadImageResponse struct {
	ImageURL  string    `json:"imageUrl"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	UploadedAt time.Time `json:"uploadedAt"`
}

type UploadSellerImage struct {
	Avatar string `json:"avatar"`
	Banner string `json:"banner"`
	Filename string `json:"filename"`
	Size int64 `json:"size"`
	UploadAt time.Time `json:"uploadAt"`
}