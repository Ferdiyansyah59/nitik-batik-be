package service

import (
	"batik/dto"
	"batik/entity"
	"batik/repository"
	"batik/utils"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// StoreService interface represents the store service contract
type StoreService interface {
	CreateStore(storeDTO dto.StoreDTO) (entity.Store, error)
	Update(c *gin.Context, storeID string, userID string, storeDTO dto.UpdateStoreDTO) (entity.Store, error)
	GetStoreByID(id string) (entity.Store, error)
	GetStoreByUserID(userID int) (entity.Store, error) 
	GetAllStores() ([]entity.Store, error)   
	GetAllStoreData(page, limit int, search string) ([]entity.Store, *utils.Pagination, error)          
}

// storeService is the implementation of StoreService interface
type storeService struct {
	storeRepository repository.StoreRepository
}

// NewStoreService creates a new instance of StoreService
func NewStoreService(repo repository.StoreRepository) StoreService {
	return &storeService{
		storeRepository: repo,
	}
}


func (s *storeService) GetStoreByUserID(userID int) (entity.Store, error) {
	return s.storeRepository.FindByUserID(userID)
}

// GetAllStores retrieves all stores
func (s *storeService) GetAllStores() ([]entity.Store, error) {
	return s.storeRepository.FindAll()
}

// CreateStore transforms DTO to entity and creates a new store
func (s *storeService) CreateStore(storeDTO dto.StoreDTO) (entity.Store, error) {
	// Transform DTO to entity
	now := time.Now()
	store := entity.Store{
		Name:        storeDTO.Name,
		Description: storeDTO.Description,
		Whatsapp:    storeDTO.Whatsapp,
		Alamat:      storeDTO.Alamat,
		UserID:      storeDTO.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	// Call repository to persist the entity
	return s.storeRepository.CreateStore(store)
}

func (s *storeService) Update(c *gin.Context, storeID string, userID string, storeDTO dto.UpdateStoreDTO) (entity.Store, error) {
	// Cari toko berdasarkan ID
	store, err := s.storeRepository.FindByID(storeID)
	if err != nil {
		return entity.Store{}, fmt.Errorf("toko tidak ditemukan: %v", err)
	}

	// Validasi kepemilikan
	if strconv.Itoa(store.UserID) != userID {
		return entity.Store{}, fmt.Errorf("anda tidak memiliki akses untuk mengubah toko ini")
	}

	// Update data toko
	if storeDTO.Name != "" {
		store.Name = storeDTO.Name
	}
	if storeDTO.Description != "" {
		store.Description = storeDTO.Description
	}
	if storeDTO.Whatsapp != "" {
		store.Whatsapp = storeDTO.Whatsapp
	}
	if storeDTO.Alamat != "" {
		store.Alamat = storeDTO.Alamat
	}

	// ✅ PERBAIKAN: Tambah logging dan debugging untuk file upload
	log.Printf("🔍 Checking for avatar file...")
	
	// Proses avatar jika ada
	avatarFile, err := c.FormFile("avatar")
	if err != nil {
		log.Printf("ℹ️ No avatar file found or error: %v", err)
	} else {
		log.Printf("✅ Avatar file found: %s, Size: %d bytes", avatarFile.Filename, avatarFile.Size)
		
		// Validasi file
		if err := utils.FileValidator(avatarFile, 5*1024*1024); err != nil {
			log.Printf("❌ Avatar validation failed: %v", err)
			return entity.Store{}, fmt.Errorf("validasi avatar gagal: %v", err)
		}

		// Hapus avatar lama jika ada
		if store.Avatar != "" {
			log.Printf("🗑️ Deleting old avatar: %s", store.Avatar)
			utils.DeleteFileIfExists(store.Avatar)
		}

		// Upload avatar baru
		log.Printf("📤 Uploading new avatar...")
		avatarPath, err := utils.UploadFile(c, avatarFile, "uploads/store-avatar")
		if err != nil {
			log.Printf("❌ Avatar upload failed: %v", err)
			return entity.Store{}, fmt.Errorf("gagal upload avatar: %v", err)
		}

		log.Printf("✅ Avatar uploaded successfully: %s", avatarPath)
		store.Avatar = avatarPath
	}

	// ✅ PERBAIKAN: Tambah logging untuk banner
	log.Printf("🔍 Checking for banner file...")
	
	// Proses banner jika ada
	bannerFile, err := c.FormFile("banner")
	if err != nil {
		log.Printf("ℹ️ No banner file found or error: %v", err)
	} else {
		log.Printf("✅ Banner file found: %s, Size: %d bytes", bannerFile.Filename, bannerFile.Size)
		
		// Validasi file
		if err := utils.FileValidator(bannerFile, 5*1024*1024); err != nil {
			log.Printf("❌ Banner validation failed: %v", err)
			return entity.Store{}, fmt.Errorf("validasi banner gagal: %v", err)
		}

		// Hapus banner lama jika ada
		if store.Banner != "" {
			log.Printf("🗑️ Deleting old banner: %s", store.Banner)
			utils.DeleteFileIfExists(store.Banner)
		}

		// Upload banner baru
		log.Printf("📤 Uploading new banner...")
		bannerPath, err := utils.UploadFile(c, bannerFile, "uploads/store-banner")
		if err != nil {
			log.Printf("❌ Banner upload failed: %v", err)
			return entity.Store{}, fmt.Errorf("gagal upload banner: %v", err)
		}

		log.Printf("✅ Banner uploaded successfully: %s", bannerPath)
		store.Banner = bannerPath
	}

	// Update timestamp
	store.UpdatedAt = time.Now()

	// ✅ PERBAIKAN: Log sebelum menyimpan ke database
	log.Printf("💾 Saving store to database...")
	log.Printf("Store data before save: Avatar=%s, Banner=%s", store.Avatar, store.Banner)

	// Simpan ke database
	updatedStore, err := s.storeRepository.Update(store)
	if err != nil {
		log.Printf("❌ Database save failed: %v", err)
		return entity.Store{}, fmt.Errorf("gagal menyimpan data toko: %v", err)
	}

	log.Printf("✅ Store updated successfully in database")
	log.Printf("Updated store data: Avatar=%s, Banner=%s", updatedStore.Avatar, updatedStore.Banner)

	return updatedStore, nil
}


func (s *storeService) GetStoreByID(id string) (entity.Store, error) {
	store, err := s.storeRepository.FindByID(id)

	return store, err
}


func (s *storeService) GetAllStoreData(page, limit int, search string) ([]entity.Store, *utils.Pagination, error) {
	// Ensure valid pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	
	// Get users from repository
	users, total, err := s.storeRepository.GetAllStoreData(page, limit, search)
	if err != nil {
		return nil, nil, err
	}
	
	// Create pagination data
	pagination := utils.NewPagination(page, limit, total)
	
	return users, pagination, nil
 }