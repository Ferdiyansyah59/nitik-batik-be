package repository

import (
	"batik/entity"

	"gorm.io/gorm"
)

type ProductCategoryRepository interface {
	GetProductCategory() ([]entity.ProductCategory, error)
}

type productCategoryRepository struct {
	db *gorm.DB
}

func NewProductCategoryRepository(db *gorm.DB) ProductCategoryRepository {
	return &productCategoryRepository{
		db: db,
	}
}

func (r *productCategoryRepository) GetProductCategory() ([]entity.ProductCategory, error) {
	var articles []entity.ProductCategory

	if err := r.db.Find(&articles).Error; err != nil {
		return nil, err
	}

	return articles, nil
}