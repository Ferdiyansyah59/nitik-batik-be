package repository

import (
	"batik/entity"
	"fmt"
	"log"

	"gorm.io/gorm"
)

type ProductImageRepository interface {
	Create(image entity.ProductImage) (entity.ProductImage, error)
	CreateBatch(images []entity.ProductImage) error
	FindByProductID(productID int) ([]entity.ProductImage, error)
	DeleteByProductID(productID int) error
	Delete(id int) error
	DeleteMultiple(ids []int) error // ✅ NEW: Batch delete
	FindByID(id int) (entity.ProductImage, error)
	DeleteByImagePath(imagePath string) error // ✅ NEW: Delete by path
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
	if len(images) == 0 {
		return nil
	}
	
	// ✅ Use transaction for batch insert
	err := r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&images).Error
	})
	
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

// ✅ ENHANCED Delete with transaction and better verification
func (r *productImageRepository) Delete(id int) error {
	log.Printf("🗑️ Starting deletion for image ID: %d", id)
	
	// ✅ Use transaction to ensure consistency
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// First, verify the record exists
		var existingImage entity.ProductImage
		err := tx.Where("id = ?", id).First(&existingImage).Error
		if err != nil {
			log.Printf("❌ Image ID %d not found: %v", id, err)
			return fmt.Errorf("image with ID %d not found: %v", id, err)
		}
		
		log.Printf("✅ Found image to delete: ID=%d, Path=%s", existingImage.ID, existingImage.Image)
		
		// Delete the record
		result := tx.Delete(&entity.ProductImage{}, id)
		if result.Error != nil {
			log.Printf("❌ Database error during deletion: %v", result.Error)
			return fmt.Errorf("failed to delete image from database: %v", result.Error)
		}
		
		if result.RowsAffected == 0 {
			log.Printf("❌ No rows were deleted for ID %d", id)
			return fmt.Errorf("no rows were deleted for image ID %d", id)
		}
		
		log.Printf("✅ Successfully deleted from transaction - Rows affected: %d", result.RowsAffected)
		return nil
	})
	
	if err != nil {
		return err
	}
	
	// ✅ Final verification outside transaction
	var verifyImage entity.ProductImage
	verifyErr := r.db.Where("id = ?", id).First(&verifyImage).Error
	if verifyErr == nil {
		log.Printf("❌ CRITICAL ERROR: Image ID %d still exists after transaction!", id)
		return fmt.Errorf("image still exists after deletion - database operation failed")
	}
	
	log.Printf("✅ VERIFIED: Image ID %d successfully deleted from database", id)
	return nil
}

// ✅ NEW: Delete multiple images in single transaction
func (r *productImageRepository) DeleteMultiple(ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	
	log.Printf("🗑️ Starting batch deletion for %d images: %v", len(ids), ids)
	
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Verify all records exist first
		var existingImages []entity.ProductImage
		err := tx.Where("id IN ?", ids).Find(&existingImages).Error
		if err != nil {
			return fmt.Errorf("failed to find images: %v", err)
		}
		
		if len(existingImages) != len(ids) {
			return fmt.Errorf("some images not found - expected %d, found %d", len(ids), len(existingImages))
		}
		
		// Delete all at once
		result := tx.Where("id IN ?", ids).Delete(&entity.ProductImage{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete images: %v", result.Error)
		}
		
		if int(result.RowsAffected) != len(ids) {
			return fmt.Errorf("expected to delete %d rows, but deleted %d", len(ids), result.RowsAffected)
		}
		
		log.Printf("✅ Successfully batch deleted %d images", result.RowsAffected)
		return nil
	})
	
	return err
}

// ✅ NEW: Delete by image path (useful for cleanup)
func (r *productImageRepository) DeleteByImagePath(imagePath string) error {
	log.Printf("🗑️ Deleting image by path: %s", imagePath)
	
	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("image = ?", imagePath).Delete(&entity.ProductImage{})
		if result.Error != nil {
			return result.Error
		}
		
		log.Printf("✅ Deleted %d records with path: %s", result.RowsAffected, imagePath)
		return nil
	})
	
	return err
}

func (r *productImageRepository) FindByID(id int) (entity.ProductImage, error) {
	var image entity.ProductImage
	err := r.db.Where("id = ?", id).First(&image).Error
	return image, err
}