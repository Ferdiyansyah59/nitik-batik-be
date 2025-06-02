package service

import (
	"batik/dto"
	"batik/entity"
	"batik/repository"
	"batik/utils"
	"fmt"
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

	// Proses avatar jika ada
	avatarFile, err := c.FormFile("avatar")
	if err == nil {
		// Validasi file
		if err := utils.FileValidator(avatarFile, 5*1024*1024); err != nil {
			return entity.Store{}, fmt.Errorf("validasi avatar gagal: %v", err)
		}

		// Hapus avatar lama jika ada
		if store.Avatar != "" {
			utils.DeleteFileIfExists(store.Avatar)
		}

		// Upload avatar baru
		avatarPath, err := utils.UploadFile(c, avatarFile, "uploads/store-avatar")
		if err != nil {
			return entity.Store{}, fmt.Errorf("gagal upload avatar: %v", err)
		}

		store.Avatar = avatarPath
	}

	// Proses banner jika ada
	bannerFile, err := c.FormFile("banner")
	if err == nil {
		// Validasi file
		if err := utils.FileValidator(bannerFile, 5*1024*1024); err != nil {
			return entity.Store{}, fmt.Errorf("validasi banner gagal: %v", err)
		}

		// Hapus banner lama jika ada
		if store.Banner != "" {
			utils.DeleteFileIfExists(store.Banner)
		}

		// Upload banner baru
		bannerPath, err := utils.UploadFile(c, bannerFile, "uploads/store-banner")
		if err != nil {
			return entity.Store{}, fmt.Errorf("gagal upload banner: %v", err)
		}

		store.Banner = bannerPath
	}

	// Update timestamp
	store.UpdatedAt = time.Now()

	// Simpan ke database
	updatedStore, err := s.storeRepository.Update(store)
	if err != nil {
		return entity.Store{}, fmt.Errorf("gagal menyimpan data toko: %v", err)
	}

	return updatedStore, nil
}


func (s *storeService) GetStoreByID(id string) (entity.Store, error) {
	store, err := s.storeRepository.FindByID(id)

	return store, err
}