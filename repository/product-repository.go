package repository

import (
	"batik/entity"
	"errors"

	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product entity.Product) (entity.Product, error)
	FindByID(id int) (entity.Product, error)
	FindBySlug(slug string) (entity.Product, error)
	FindByStoreID(storeID int) ([]entity.Product, error)
	Update(product entity.Product) (entity.Product, error)
	Delete(id int) error
	IsSlugExists(slug string) bool
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository {
		db: db,
	}
}

func (r *productRepository) Create(product entity.Product) (entity.Product, error) {
	err := r.db.Create(&product).Error
	return product, err
}

func (r *productRepository) FindByID(id int) (entity.Product, error) {
	var product entity.Product
	err := r.db.Preload("Images").Where("id = ?", id).First(&product).Error

	return product, err
}

func (r *productRepository) FindBySlug(slug string) (entity.Product, error) {
	var product entity.Product

	err := r.db.Preload("Images").Where("slug = ?", slug).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return product, errors.New("product not found")
		}
		return product, err
	}
	return product, nil
}

func (r *productRepository) FindByStoreID(storeID int) ([]entity.Product, error) {
	var products []entity.Product
	err := r.db.Where("store_id = ?", storeID).Find(&products).Error
	return products, err
}

func (r *productRepository) Update(product entity.Product) (entity.Product, error) {
	err := r.db.Save(&product).Error
	return product, err
}

func (r *productRepository) Delete(id int) error {
	err := r.db.Delete(&entity.Product{}, id).Error
	return err
}

func (r *productRepository) IsSlugExists(slug string) bool {
	var count int64
	r.db.Model(&entity.Product{}).Where("slug = ?", slug).Count(&count)

	return count > 0
}