package service

import (
	"batik/entity"
	"batik/repository"
)

type ProductCategoryService interface {
	GetProductCategory() ([]entity.ProductCategory, error)
}

type productCategoryService struct {
	pc repository.ProductCategoryRepository
}

func NewProductCategoryService(pcRepo repository.ProductCategoryRepository) ProductCategoryService {
	return &productCategoryService {
		pc: pcRepo,
	}
}

func (serv *productCategoryService) GetProductCategory() ([]entity.ProductCategory, error) {
	res, err := serv.pc.GetProductCategory()
	return res, err
}