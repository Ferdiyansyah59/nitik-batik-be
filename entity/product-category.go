package entity

type ProductCategory struct {
	ID int `json:"id" gorm:"column:id"`
	CategoryName string `json:"category_name" gorm:"column:category_name"`
	Slug string `json:"slug" gorm:"column:slug"`
}

func (pc ProductCategory) TableName() string {
	return "category_catalog"
}