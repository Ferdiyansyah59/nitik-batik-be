package controller

import (
	"batik/helper"
	"batik/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController interface {
	GetAllUser(ctx *gin.Context)
}

type userController struct {
	userService service.UserService
	jwtService service.JWTService
}

func NewUserController(userService service.UserService, jwtService service.JWTService) UserController {
	return &userController {
		userService: userService,
		jwtService: jwtService,
	}
}

func (c userController) GetAllUser(ctx *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	search := ctx.Query("search")
	
	// Get users from service
	users, pagination, err := c.userService.GetAllUser(page, limit, search)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to fetch users", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Return response - combine users and pagination into a single data object
	data := map[string]interface{}{
		"users":   users,
		"pagination": pagination,
	}
	response := helper.BuildResponse(true, "Users fetched successfully", data)
	ctx.JSON(http.StatusOK, response)
}