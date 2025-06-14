package repository

import (
	"batik/entity"
	"errors"

	"gorm.io/gorm"
)

type StoreRepository interface {
	CreateStore(store entity.Store) (entity.Store, error)
    FindByID(id string) (entity.Store, error)
    FindByUserID(userID int) (entity.Store, error) 
    FindAll() ([]entity.Store, error)  
    Update(store entity.Store) (entity.Store, error)
    GetAllStoreData(page, limit int, search string) ([]entity.Store, int64, error)
}

type storeRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) StoreRepository {
	return &storeRepository{
		db: db,
	}
}


func (r *storeRepository) FindByUserID(userID int) (entity.Store, error) {
    var store entity.Store
    err := r.db.Where("user_id = ?", userID).First(&store).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return store, errors.New("store not found for this user")
        }
        return store, err
    }
    
    return store, nil
}

func (r *storeRepository) FindAll() ([]entity.Store, error) {
    var stores []entity.Store
    err := r.db.Find(&stores).Error
    return stores, err
}

func (r *storeRepository) CreateStore(store entity.Store) (entity.Store, error) {
    var existingStore entity.Store
    result := r.db.Where("name = ?", store.Name).First(&existingStore)
    
    if result.Error == nil {
        return entity.Store{}, errors.New("store with this name already exists")
    }
    
    if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return entity.Store{}, result.Error
    }
    
    if err := r.db.Create(&store).Error; err != nil {
        return entity.Store{}, err
    }
    
    return store, nil
}


func (r *storeRepository) FindByID(id string) (entity.Store, error) {
    var store entity.Store
    err := r.db.Where("id = ?", id).First(&store).Error
    return store, err
}

func (r *storeRepository) Update(store entity.Store) (entity.Store, error) {
    err := r.db.Save(&store).Error
    return store, err
}

func (r *storeRepository) GetAllStoreData(page, limit int, search string) ([]entity.Store, int64, error) {
    var stores []entity.Store
    var total int64

    query := r.db.Model(&entity.Store{})

    if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}
	
	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_At DESC").Find(&stores).Error; err != nil {
		return nil, 0, err
	}
	
	return stores, total, nil
}