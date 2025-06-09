// service/product_service.go
package service

import (
	"batik/dto"
	"batik/entity"
	"batik/repository"
	"batik/utils"
	"errors"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type ProductService interface {
	GetAllProductByStore(storeID, page, limit int, search string) ([]entity.Product, *utils.Pagination, error)
	CreateProduct(c *gin.Context, productDTO dto.CreateProductDTO, files []*multipart.FileHeader) (entity.Product, error)
	GetProductByID(id int) (entity.Product, error)
	GetProductBySlug(slug string) (entity.Product, error)
	UpdateProductWithImages(c *gin.Context, slug string, productDTO dto.UpdateProductDTO, files []*multipart.FileHeader, imagesToDelete []string) (entity.Product, error)
	UpdateProduct(c *gin.Context, slug string, productDTO dto.UpdateProductDTO) (entity.Product, error)
	DeleteProduct(slug string) error
	AddProductImage(c *gin.Context, slug string, file *multipart.FileHeader) (entity.ProductImage, error)
	DeleteProductImage(slug string, imageID int) error
	GetAllPublicProduct(page, limit int, search string) ([]dto.PublicProductCard, *utils.Pagination, error)
	GetLatestProduct()([]entity.ProductCard, error)
	GetDetailProduct(slug string)(entity.ProductCard, error)
	GetAllPublicProductByCategory(slug string, page, limit int) ([]dto.PublicProductCard, *utils.Pagination, error)
}

type productService struct {
	productRepo      repository.ProductRepository
	productImageRepo repository.ProductImageRepository
}

func NewProductService(productRepo repository.ProductRepository, productImageRepo repository.ProductImageRepository) ProductService {
	return &productService{
		productRepo:      productRepo,
		productImageRepo: productImageRepo,
	}
}

func (s *productService) GetAllProductByStore(storeID, page, limit int, search string) ([]entity.Product, *utils.Pagination, error) {
	// Validasi pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 12
	}
	
	// Get products dengan pagination
	products, total, err := s.productRepo.GetAllProductByStore(storeID, page, limit, search)
	if err != nil {
		return nil, nil, err
	}
	
	// Create pagination data
	pagination := utils.NewPagination(page, limit, total)
	
	return products, pagination, nil
}

func (s *productService) CreateProduct(c *gin.Context, productDTO dto.CreateProductDTO, files []*multipart.FileHeader) (entity.Product, error) {
	// Validasi input
	if len(files) == 0 {
		return entity.Product{}, errors.New("minimal satu gambar produk diperlukan")
	}
	
	// Generate unique slug berdasarkan nama produk
	baseSlug := utils.GenerateSlug(productDTO.Name, "product")
	slug := utils.EnsureUniqueSlug(baseSlug, s.productRepo.IsSlugExists)
	
	// Membuat objek produk baru
	product := entity.Product{
		Name:        productDTO.Name,
		Slug:        slug,
		Description: productDTO.Description,
		Harga:       productDTO.Harga,
		StoreID:     productDTO.StoreID,
		CategoryID:  productDTO.CategoryID,
	}
	
	// Upload gambar pertama sebagai thumbnail
	firstFile := files[0]
	if err := utils.FileValidator(firstFile, 5*1024*1024); err != nil {
		return entity.Product{}, fmt.Errorf("validasi gambar thumbnail gagal: %v", err)
	}
	
	thumbnailPath, err := utils.UploadFile(c, firstFile, "uploads/product-images")
	if err != nil {
		return entity.Product{}, fmt.Errorf("gagal mengupload thumbnail: %v", err)
	}
	
	// Set thumbnail
	product.Thumbnail = thumbnailPath
	
	// Simpan produk ke database untuk mendapatkan ID
	createdProduct, err := s.productRepo.Create(product)
	if err != nil {
		// Hapus thumbnail jika gagal menyimpan produk
		utils.DeleteFileIfExists(thumbnailPath)
		return entity.Product{}, fmt.Errorf("gagal menyimpan produk: %v", err)
	}
	
	// Proses dan simpan semua gambar produk
	var productImages []entity.ProductImage
	
	for _, file := range files {
		// Validasi file
		if err := utils.FileValidator(file, 5*1024*1024); err != nil {
			continue // Skip file yang tidak valid
		}
		
		var imagePath string
		
		// Gunakan thumbnail yang sudah diupload jika ini file pertama
		if file == firstFile {
			imagePath = thumbnailPath
		} else {
			// Upload file tambahan
			imagePath, err = utils.UploadFile(c, file, "uploads/product-images")
			if err != nil {
				continue // Skip jika gagal upload
			}
		}
		
		// Buat entitas ProductImage
		productImage := entity.ProductImage{
			Image:     imagePath,
			ProductID: createdProduct.ID,
		}
		
		productImages = append(productImages, productImage)
	}
	
	// Simpan semua gambar produk ke database
	if err := s.productImageRepo.CreateBatch(productImages); err != nil {
		// Jika gagal menyimpan gambar, tetap kembalikan produk
		// gambar thumbnail sudah disimpan di tabel produk
		return createdProduct, fmt.Errorf("sebagian gambar gagal disimpan: %v", err)
	}
	
	// Ambil produk lengkap dengan gambarnya
	return s.GetProductByID(createdProduct.ID)
}

func (s *productService) GetProductByID(id int) (entity.Product, error) {
	return s.productRepo.FindByID(id)
}

func (s *productService) GetProductBySlug(slug string) (entity.Product, error) {
	return s.productRepo.FindBySlug(slug)
}


func (s *productService) UpdateProductWithImages(c *gin.Context, slug string, productDTO dto.UpdateProductDTO, files []*multipart.FileHeader, imagesToDelete []string) (entity.Product, error) {
	// Find product by slug
	product, err := s.productRepo.FindBySlug(slug)
	if err != nil {
		return entity.Product{}, fmt.Errorf("produk tidak ditemukan: %v", err)
	}
	
	log.Printf("📝 Updating product: %s (ID: %d)", product.Name, product.ID)
	log.Printf("🗑️ Images to delete: %v", imagesToDelete)
	log.Printf("🖼️ New images to add: %d", len(files))
	
	// ✅ Update basic product fields
	hasChanges := false
	
	if productDTO.Name != "" && productDTO.Name != product.Name {
		baseSlug := utils.GenerateSlug(productDTO.Name, "product")
		newSlug := utils.EnsureUniqueSlug(baseSlug, s.productRepo.IsSlugExists)
		product.Slug = newSlug
		product.Name = productDTO.Name
		hasChanges = true
	}
	
	if productDTO.Description != "" && productDTO.Description != product.Description {
		product.Description = productDTO.Description
		hasChanges = true
	}
	
	if productDTO.Harga > 0 && productDTO.Harga != product.Harga {
		product.Harga = productDTO.Harga
		hasChanges = true
	}
	
	if productDTO.CategoryID > 0 && productDTO.CategoryID != product.CategoryID {
		product.CategoryID = productDTO.CategoryID
		hasChanges = true
	}
	
	// ✅ CRITICAL FIX: Handle image operations with better transaction management
	imageOperationsPerformed := false
	thumbnailNeedsUpdate := false
	var deletedImagePaths []string
	
	// STEP 1: Get current images ONCE and use throughout
	currentImages, err := s.productImageRepo.FindByProductID(product.ID)
	if err != nil {
		log.Printf("❌ Error getting current images: %v", err)
		return entity.Product{}, fmt.Errorf("gagal mendapatkan gambar produk: %v", err)
	}
	
	log.Printf("📸 Current images in DB at start: %d", len(currentImages))
	for i, img := range currentImages {
		log.Printf("  Current Image %d: ID=%d, Path=%s", i+1, img.ID, img.Image)
	}
	
	// STEP 2: Process deletions using more robust method
	if len(imagesToDelete) > 0 {
		log.Printf("🗑️ STEP 2: Processing %d images for deletion", len(imagesToDelete))
		
		// Find IDs to delete by matching paths
		var idsToDelete []int
		var pathsToDelete []string
		
		for _, imageToDelete := range imagesToDelete {
			log.Printf("🔍 Looking for image to delete: %s", imageToDelete)
			
			// Find matching image record
			for _, currentImage := range currentImages {
				if currentImage.Image == imageToDelete {
					idsToDelete = append(idsToDelete, currentImage.ID)
					pathsToDelete = append(pathsToDelete, imageToDelete)
					log.Printf("✅ Found match: ID=%d, Path=%s", currentImage.ID, currentImage.Image)
					
					// Check if this is the thumbnail
					if product.Thumbnail == imageToDelete {
						thumbnailNeedsUpdate = true
						log.Printf("⚠️ Image to delete is the current thumbnail")
					}
					break
				}
			}
		}
		
		if len(idsToDelete) > 0 {
			log.Printf("🗑️ Deleting %d images with IDs: %v", len(idsToDelete), idsToDelete)
			
			// ✅ Use batch delete for better performance and consistency
			err := s.productImageRepo.DeleteMultiple(idsToDelete)
			if err != nil {
				log.Printf("❌ Failed to batch delete images: %v", err)
				return entity.Product{}, fmt.Errorf("gagal menghapus gambar dari database: %v", err)
			}
			
			// Update local tracking
			deletedImagePaths = pathsToDelete
			imageOperationsPerformed = true
			
			log.Printf("✅ Successfully batch deleted %d images", len(idsToDelete))
			
			// Clean up physical files
			for _, imagePath := range deletedImagePaths {
				utils.DeleteFileIfExistsProduct(imagePath)
				log.Printf("🗑️ Deleted physical file: %s", imagePath)
			}
		} else {
			log.Printf("⚠️ No matching images found for deletion")
		}
	}
	
	// STEP 3: Add new images if provided
	var firstNewImagePath string
	if len(files) > 0 {
		log.Printf("🖼️ STEP 3: Adding %d new images", len(files))
		
		var newImages []entity.ProductImage
		var uploadedFiles []string
		
		for i, file := range files {
			// Validate file
			if err := utils.FileValidatorProduct(file, 5*1024*1024); err != nil {
				log.Printf("❌ File validation failed for %s: %v", file.Filename, err)
				continue
			}
			
			// Upload file
			imagePath, err := utils.UploadFileproduct(c, file, "uploads/product-images")
			if err != nil {
				log.Printf("❌ Upload failed for %s: %v", file.Filename, err)
				continue
			}
			
			uploadedFiles = append(uploadedFiles, imagePath)
			
			if firstNewImagePath == "" {
				firstNewImagePath = imagePath
			}
			
			productImage := entity.ProductImage{
				Image:     imagePath,
				ProductID: product.ID,
			}
			
			newImages = append(newImages, productImage)
			log.Printf("🖼️ Prepared new image %d: %s", i+1, imagePath)
		}
		
		// Save new images to database using batch insert
		if len(newImages) > 0 {
			err := s.productImageRepo.CreateBatch(newImages)
			if err != nil {
				log.Printf("❌ Error saving new images: %v", err)
				// Cleanup uploaded files on error
				for _, filePath := range uploadedFiles {
					utils.DeleteFileIfExistsProduct(filePath)
				}
				return entity.Product{}, fmt.Errorf("gagal menyimpan gambar baru: %v", err)
			}
			
			log.Printf("✅ Successfully saved %d new images", len(newImages))
			imageOperationsPerformed = true
		}
	}
	
	// STEP 4: Update thumbnail if needed
	if thumbnailNeedsUpdate || product.Thumbnail == "" {
		log.Printf("🖼️ STEP 4: Updating thumbnail")
		
		if firstNewImagePath != "" {
			product.Thumbnail = firstNewImagePath
			hasChanges = true
			log.Printf("📝 Set thumbnail to new image: %s", product.Thumbnail)
		} else {
			// Get fresh remaining images from database
			remainingImages, err := s.productImageRepo.FindByProductID(product.ID)
			if err != nil {
				log.Printf("❌ Error getting remaining images: %v", err)
			} else {
				log.Printf("📸 Fresh remaining images count: %d", len(remainingImages))
				
				if len(remainingImages) > 0 {
					product.Thumbnail = remainingImages[0].Image
					hasChanges = true
					log.Printf("📝 Set thumbnail to remaining image: %s", product.Thumbnail)
				} else {
					product.Thumbnail = ""
					hasChanges = true
					log.Printf("📝 Cleared thumbnail - no images remaining")
				}
			}
		}
	}
	
	// STEP 5: Save product changes
	if hasChanges || imageOperationsPerformed {
		log.Printf("💾 STEP 5: Saving product changes")
		
		updatedProduct, err := s.productRepo.Update(product)
		if err != nil {
			log.Printf("❌ Error updating product: %v", err)
			return entity.Product{}, fmt.Errorf("gagal mengupdate produk: %v", err)
		}
		
		log.Printf("✅ Product updated in database")
		
		// ✅ Get fresh product data with images
		finalProduct, err := s.productRepo.FindByID(updatedProduct.ID)
		if err != nil {
			log.Printf("❌ Error getting fresh product data: %v", err)
			return updatedProduct, nil
		}
		
		// ✅ FINAL VERIFICATION with fresh database query
		finalImages, err := s.productImageRepo.FindByProductID(updatedProduct.ID)
		if err != nil {
			log.Printf("❌ Error in final verification: %v", err)
		} else {
			log.Printf("✅ FINAL VERIFICATION: Product now has %d images in database", len(finalImages))
			for i, img := range finalImages {
				log.Printf("  Final Image %d: ID=%d, Path=%s", i+1, img.ID, img.Image)
			}
			
			// ✅ Additional check: verify deleted images are really gone
			if len(imagesToDelete) > 0 {
				stillExisting := 0
				for _, deletedPath := range imagesToDelete {
					for _, finalImg := range finalImages {
						if finalImg.Image == deletedPath {
							stillExisting++
							log.Printf("❌ WARNING: Deleted image still exists: %s (ID: %d)", finalImg.Image, finalImg.ID)
						}
					}
				}
				
				if stillExisting > 0 {
					log.Printf("❌ CRITICAL: %d supposedly deleted images still exist in database!", stillExisting)
					// Optionally: try to delete them again
					for _, deletedPath := range imagesToDelete {
						err := s.productImageRepo.DeleteByImagePath(deletedPath)
						if err != nil {
							log.Printf("❌ Failed to cleanup remaining image %s: %v", deletedPath, err)
						} else {
							log.Printf("✅ Successfully cleaned up remaining image: %s", deletedPath)
						}
					}
				} else {
					log.Printf("✅ All deleted images confirmed removed from database")
				}
			}
		}
		
		return finalProduct, nil
	}
	
	log.Printf("ℹ️ No changes detected")
	return s.GetProductByID(product.ID)
}

// ✅ FIXED: Update original UpdateProduct method to use new method
func (s *productService) UpdateProduct(c *gin.Context, slug string, productDTO dto.UpdateProductDTO) (entity.Product, error) {
	// Call the new method with empty files and imagesToDelete
	return s.UpdateProductWithImages(c, slug, productDTO, []*multipart.FileHeader{}, []string{})
}

func (s *productService) DeleteProduct(slug string) error {
	// Ambil produk untuk mendapatkan ID dan path thumbnail
	product, err := s.productRepo.FindBySlug(slug)
	if err != nil {
		return fmt.Errorf("produk tidak ditemukan: %v", err)
	}
	
	// Ambil semua gambar produk
	images, err := s.productImageRepo.FindByProductID(product.ID)
	if err != nil {
		return fmt.Errorf("gagal mengambil gambar produk: %v", err)
	}
	
	// Hapus semua file fisik
	for _, img := range images {
		utils.DeleteFileIfExists(img.Image)
	}
	
	// Juga hapus thumbnail jika bukan bagian dari images
	thumbnailExists := false
	for _, img := range images {
		if img.Image == product.Thumbnail {
			thumbnailExists = true
			break
		}
	}
	
	if !thumbnailExists {
		utils.DeleteFileIfExists(product.Thumbnail)
	}
	
	// Hapus produk dari database (akan menghapus semua gambar berkat ON DELETE CASCADE)
	return s.productRepo.Delete(product.ID)
}

func (s *productService) AddProductImage(c *gin.Context, slug string, file *multipart.FileHeader) (entity.ProductImage, error) {
	// Cari produk berdasarkan slug
	product, err := s.productRepo.FindBySlug(slug)
	if err != nil {
		return entity.ProductImage{}, fmt.Errorf("produk tidak ditemukan: %v", err)
	}
	
	// Validasi file
	if err := utils.FileValidator(file, 5*1024*1024); err != nil {
		return entity.ProductImage{}, fmt.Errorf("validasi gambar gagal: %v", err)
	}
	
	// Upload file
	imagePath, err := utils.UploadFile(c, file, "uploads/product-images")
	if err != nil {
		return entity.ProductImage{}, fmt.Errorf("gagal mengupload gambar: %v", err)
	}
	
	// Buat entitas ProductImage
	productImage := entity.ProductImage{
		Image:     imagePath,
		ProductID: product.ID,
	}
	
	// Simpan gambar ke database
	createdImage, err := s.productImageRepo.Create(productImage)
	if err != nil {
		// Hapus file jika gagal menyimpan ke database
		utils.DeleteFileIfExists(imagePath)
		return entity.ProductImage{}, fmt.Errorf("gagal menyimpan gambar: %v", err)
	}
	
	return createdImage, nil
}

func (s *productService) DeleteProductImage(slug string, imageID int) error {
	// Cek dulu apakah produk ada
	product, err := s.productRepo.FindBySlug(slug)
	if err != nil {
		return fmt.Errorf("produk tidak ditemukan: %v", err)
	}
	
	// Ambil gambar yang akan dihapus
	images, err := s.productImageRepo.FindByProductID(product.ID)
	if err != nil {
		return fmt.Errorf("gagal mengambil gambar produk: %v", err)
	}
	
	// Temukan gambar dengan ID yang sesuai
	var targetImage entity.ProductImage
	var targetFound bool
	
	for _, img := range images {
		if img.ID == imageID {
			targetImage = img
			targetFound = true
			break
		}
	}
	
	if !targetFound {
		return fmt.Errorf("gambar dengan ID %d tidak ditemukan untuk produk ini", imageID)
	}
	
	// Cek jika gambar ini adalah thumbnail
	if product.Thumbnail == targetImage.Image {
		// Jika ini adalah gambar thumbnail dan satu-satunya gambar, tolak penghapusan
		if len(images) == 1 {
			return errors.New("tidak dapat menghapus satu-satunya gambar produk")
		}
		
		// Jika ini adalah thumbnail tapi ada gambar lain, set gambar lain sebagai thumbnail
		var newThumbnail string
		for _, img := range images {
			if img.ID != imageID {
				newThumbnail = img.Image
				break
			}
		}
		
		// Update thumbnail produk
		product.Thumbnail = newThumbnail
		if _, err := s.productRepo.Update(product); err != nil {
			return fmt.Errorf("gagal mengupdate thumbnail produk: %v", err)
		}
	}
	
	// Hapus file fisik
	utils.DeleteFileIfExists(targetImage.Image)
	
	// Hapus gambar dari database
	return s.productImageRepo.Delete(imageID)
}

// service/product-service.go

func (s *productService) GetAllPublicProduct(page, limit int, search string) ([]dto.PublicProductCard, *utils.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 40 // Batas default, bisa disesuaikan
	}

	// Teruskan parameter 'search' ke pemanggilan repositori
	products, total, err := s.productRepo.GetAllPublicProduct(page, limit, search)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mendapatkan semua produk dengan detail lengkap: %w", err)
	}

	var publicProductCard []dto.PublicProductCard
	for _, p := range products {
		publicProductCard = append(publicProductCard, dto.PublicProductCard{
			ID:           p.ID,
			Slug:         p.Slug,
			Name:         p.Name,
			Harga:        p.Harga,
			StoreID:      p.StoreID,
			StoreName:    p.StoreName,
			CategoryID:   p.CategoryID,
			CategoryName: p.CategoryName,
			CategorySlug: p.CategorySlug,
			Thumbnail:    p.Thumbnail,
			CreatedAt:    p.CreatedAt,
		})
	}

	pagination := utils.NewPagination(page, limit, total)
	return publicProductCard, pagination, nil
}

func (s *productService) GetLatestProduct()([]entity.ProductCard, error) {
	res, err := s.productRepo.GetLatestProduct()
	return res, err
}

func (s *productService) GetDetailProduct(slug string)(entity.ProductCard, error) {
	res, err := s.productRepo.GetDetailProduct(slug)
	return res, err
}

func (s *productService) GetAllPublicProductByCategory(slug string, page, limit int) ([]dto.PublicProductCard, *utils.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 40 // Batas default, bisa disesuaikan
	}

	products, total, err := s.productRepo.GetAllPublicProductByCategory(slug, page, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mendapatkan semua produk dengan detail lengkap: %w", err)
	}

	var publicProductCard []dto.PublicProductCard
	for _, p := range products {
		publicProductCard = append(publicProductCard, dto.PublicProductCard{
			ID:            p.ID,
			Slug:          p.Slug,
			Name:          p.Name,
			Harga:         p.Harga,
			StoreID:       p.StoreID,
			StoreName:     p.StoreName,   
			CategoryID:    p.CategoryID,
			CategoryName:  p.CategoryName,
			CategorySlug:  p.CategorySlug,
			Thumbnail:     p.Thumbnail,
			CreatedAt:     p.CreatedAt,
		})
	}

	pagination := utils.NewPagination(page, limit, total)
	return publicProductCard, pagination, nil
}