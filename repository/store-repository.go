package repository

import (
	"batik/entity"
	"errors"

	"gorm.io/gorm"
)

type StoreRepository interface {
	CreateStore(store entity.Store) (entity.Store, error)
    FindByID(id string) (entity.Store, error)
    Update(store entity.Store) (entity.Store, error)
}

type storeRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) StoreRepository {
	return &storeRepository{
		db: db,
	}
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