package service

import (
	"batik/entity"
	"batik/repository"
	"batik/utils"
)


type UserService interface {
	GetAllUser(page, limit int, search string) ([]entity.User, *utils.Pagination, error)
}

type userService struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepository: userRepo,
	}
}

 func (s userService) GetAllUser(page, limit int, search string) ([]entity.User, *utils.Pagination, error) {
	// Ensure valid pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	
	// Get users from repository
	users, total, err := s.userRepository.GetAllUser(page, limit, search)
	if err != nil {
		return nil, nil, err
	}
	
	// Create pagination data
	pagination := utils.NewPagination(page, limit, total)
	
	return users, pagination, nil
 }