package repository

import (
	"batik/entity"
	"errors"

	"gorm.io/gorm"
)

type ProductRepository interface {
	GetAllProductByStore(storeID, page, limit int, search string) ([]entity.Product, int64, error)
	Create(product entity.Product) (entity.Product, error)
	FindByID(id int) (entity.Product, error)
	FindBySlug(slug string) (entity.Product, error)
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

func (r *productRepository) GetAllProductByStore(storeID, page, limit int, search string) ([]entity.Product, int64, error) {
	var products []entity.Product
	var total int64

	query := r.db.Model(&entity.Product{}).Where("store_id = ?", storeID)
	
	// Apply search filter if provided
	if search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination dengan preload images
	offset := (page - 1) * limit
	if err := query.Preload("Images").Offset(offset).Limit(limit).Order("created_at DESC").Find(&products).Error; err != nil {
		return nil, 0, err
	}
	
	return products, total, nil
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