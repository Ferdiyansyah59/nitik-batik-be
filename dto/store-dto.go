package dto

type StoreDTO struct {
    Name        string `json:"name" form:"name" binding:"required,min=3,max=50"`
    Description string `json:"description" form:"description" binding:"required,min=10,max=500"`
    Whatsapp    string `json:"whatsapp" form:"whatsapp" binding:"required,min=8,max=15,e164"`
    Alamat      string `json:"alamat" form:"alamat" binding:"required,min=10,max=200"`
    UserID      int `json:"user_id" form:"user_id"`
}


type UpdateStoreDTO struct {
	Name        string `form:"name"`
	Description string `form:"description"`
	Whatsapp    string `form:"whatsapp"`
	Alamat      string `form:"alamat"`
}

// StoreImageDTO adalah data gambar yang diupload
type StoreImageDTO struct {
	Avatar string `json:"avatar,omitempty"`
	Banner string `json:"banner,omitempty"`
}