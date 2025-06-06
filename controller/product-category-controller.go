package controller

import (
	"batik/helper"
	"batik/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductCategoryController interface {
	GetProductCategory(ctx *gin.Context)
}

type productCategoryController struct {
	pcService service.ProductCategoryService
}

func NewProductCategoryController(pcService service.ProductCategoryService) ProductCategoryController {
	return &productCategoryController {
		pcService: pcService,
	}
}

func (c productCategoryController) GetProductCategory(ctx *gin.Context) {
	products, err := c.pcService.GetProductCategory()

	if err != nil {
		res := helper.BuildErrorResponse("Gagal menampilkan data", err.Error(), nil)
		ctx.JSON(http.StatusNotFound, res)
	}

	res := helper.BuildResponse(true, "Berhasil menampilkan data", products)
	ctx.JSON(http.StatusOK, res)
}