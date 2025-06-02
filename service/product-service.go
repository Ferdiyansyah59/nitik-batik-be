// service/product_service.go
package service

import (
	"batik/dto"
	"batik/entity"
	"batik/repository"
	"batik/utils"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type ProductService interface {
	GetAllProductByStore(storeID, page, limit int, search string) ([]entity.Product, *utils.Pagination, error)
	CreateProduct(c *gin.Context, productDTO dto.CreateProductDTO, files []*multipart.FileHeader) (entity.Product, error)
	GetProductByID(id int) (entity.Product, error)
	GetProductBySlug(slug string) (entity.Product, error)
	UpdateProduct(c *gin.Context, slug string, productDTO dto.UpdateProductDTO) (entity.Product, error)
	DeleteProduct(slug string) error
	AddProductImage(c *gin.Context, slug string, file *multipart.FileHeader) (entity.ProductImage, error)
	DeleteProductImage(slug string, imageID int) error
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


func (s *productService) UpdateProduct(c *gin.Context, slug string, productDTO dto.UpdateProductDTO) (entity.Product, error) {
	// Cari produk berdasarkan slug
	product, err := s.productRepo.FindBySlug(slug)
	if err != nil {
		return entity.Product{}, fmt.Errorf("produk tidak ditemukan: %v", err)
	}
	
	// Update field-field jika ada perubahan
	if productDTO.Name != "" && productDTO.Name != product.Name {
		// Generate slug baru jika nama berubah
		baseSlug := utils.GenerateSlug(product.Name, "product")
		newSlug := utils.EnsureUniqueSlug(baseSlug, s.productRepo.IsSlugExists)
		product.Slug = newSlug
		product.Name = productDTO.Name
	}
	
	if productDTO.Description != "" {
		product.Description = productDTO.Description
	}
	
	if productDTO.Harga > 0 {
		product.Harga = productDTO.Harga
	}
	
	// Simpan perubahan
	updatedProduct, err := s.productRepo.Update(product)
	if err != nil {
		return entity.Product{}, fmt.Errorf("gagal mengupdate produk: %v", err)
	}
	
	return updatedProduct, nil
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