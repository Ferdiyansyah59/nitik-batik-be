package repository

import (
	"batik/entity"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository interface {
	InsertUser(user entity.User) entity.User
	UpdateUser(user entity.User) entity.User
	VerifyCredential(enail string, password string) interface{}
	IsDuplicateEmail(email string) (tx *gorm.DB)
	FindByEmail(email string) entity.User
	ProfileUser(email string) entity.User
	GetAllUser(page, limit int, search string) ([]entity.User, int64, error)
	FindByID(id string) (entity.User, error)
}

type userConnection struct {
	connection *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userConnection{
		connection: db,
	}
}

func (db *userConnection) InsertUser(user entity.User) entity.User {
	user.Password = hashAndSalt([]byte(user.Password))
	db.connection.Save(&user)
	return user
}

func (db *userConnection) UpdateUser(user entity.User) entity.User {
	if user.Password != "" {
		user.Password = hashAndSalt([]byte(user.Password))
	} else {
		var tempUser entity.User
		db.connection.Find(&tempUser, user.ID)
		user.Password = tempUser.Password
	}
	db.connection.Save(&user)
	return user
}

func (db *userConnection) VerifyCredential(email string, password string) interface{} {
	var user entity.User
	res := db.connection.Where("email = ?", email).Take(&user)
	if res.Error == nil {
		return user
	}

	return nil
}

func (db *userConnection) IsDuplicateEmail(email string) (ex *gorm.DB) {
	var user entity.User
	return db.connection.Where("email = ?", email).Take(&user)
}

func (db *userConnection) FindByEmail(email string) entity.User {
	var user entity.User
	db.connection.Where("email = ?", email).Take(&user)
	return user
}

// GET USER
func (db *userConnection) ProfileUser(email string) entity.User {
	var user entity.User
	db.connection.Where("email = ?", email).First(&user)
	return user
}

// GET ALL USER
func (db *userConnection) GetAllUser(page, limit int, search string) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	query := db.connection.Model(&entity.User{})
	
	// Apply search filter if provided
	if search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_At DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)

	if err != nil {
		log.Println(err)
		panic("Failed to hash the password")
	}

	return string(hash)
}


func (r *userConnection) FindByID(id string) (entity.User, error) {
    var user entity.User
    result := r.connection.Where("id = ?", id).First(&user)
    if result.Error != nil {
        return entity.User{}, result.Error
    }
    return user, nil
}