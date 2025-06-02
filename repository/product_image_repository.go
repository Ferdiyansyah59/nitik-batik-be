package repository

import (
	"batik/entity"

	"gorm.io/gorm"
)

type ProductImageRepository interface {
	Create(image entity.ProductImage) (entity.ProductImage, error)
	CreateBatch(images []entity.ProductImage) error
	FindByProductID(productID int) ([]entity.ProductImage, error)
	DeleteByProductID(productID int) error
	Delete(id int) error
}

type productImageRepository struct {
	db *gorm.DB
}

func NewProductImageRepository(db *gorm.DB) ProductImageRepository {
	return &productImageRepository {
		db: db,
	}
}

func (r *productImageRepository) Create(image entity.ProductImage) (entity.ProductImage, error) {
	err := r.db.Create(&image).Error
	return image, err
}

func (r *productImageRepository) CreateBatch(images []entity.ProductImage) error {
	err := r.db.Create(&images).Error
	return err
}

func (r *productImageRepository) FindByProductID(productID int) ([]entity.ProductImage, error) {
	var images []entity.ProductImage
	err := r.db.Where("product_id = ?", productID).Find(&images).Error
	return images, err
}

func (r *productImageRepository) DeleteByProductID(productID int) error {
	err := r.db.Where("product_id = ?", productID).Delete(&entity.ProductImage{}).Error
	return err
}

func (r *productImageRepository) Delete(id int) error {
	err := r.db.Delete(&entity.ProductImage{}, id).Error
	return err
}