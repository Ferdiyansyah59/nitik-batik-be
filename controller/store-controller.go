package controller

import (
	"batik/dto"
	"batik/helper"
	"batik/service"
	"fmt"
	"net/http"
	"strconv"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// StoreController interface represents the store controller contract
type StoreController interface {
	CreateStore(c *gin.Context)
	// GetStoreByID(c *gin.Context)
	// GetStoreByUserID(c *gin.Context)
	UpdateStore(c *gin.Context)
	// DeleteStore(c *gin.Context)
	// GetAllStores(c *gin.Context)
	// UploadStoreImage(c *gin.Context) 
}

// storeController is the implementation of StoreController interface
type storeController struct {
	storeService service.StoreService
	jwtService   service.JWTService
	authService  service.AuthService
	// userRepository repository.UserRepository // Menggunakan AuthService bukan UserService
}

// NewStoreController creates a new instance of StoreController
func NewStoreController(storeService service.StoreService, jwtService service.JWTService, authService service.AuthService) StoreController {
	return &storeController{
		storeService: storeService,
		jwtService:   jwtService,
		authService:  authService,
	}
}

// getEmailFromToken adalah helper function untuk mengekstrak email dari token
// getUserIDFromToken adalah helper function untuk mengekstrak userID dari token
// getClaimsFromToken adalah helper untuk mendapatkan semua informasi dari token
func (c *storeController) getClaimsFromToken(authHeader string) (jwt.MapClaims, error) {
	// Validasi token
	token, err := c.jwtService.ValidateToken(authHeader)
	if err != nil {
		log.Printf("Token validation error: %v", err)
		return nil, err
	}
	
	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("Failed to parse claims")
		return nil, fmt.Errorf("failed to parse claims")
	}
	
	// Dump all claims for debugging
	log.Printf("======= ALL TOKEN CLAIMS =======")
	for key, value := range claims {
		log.Printf("%s: %v (type: %T)", key, value, value)
	}
	log.Printf("================================")
	
	return claims, nil
}

// CreateStore dengan logging yang lebih detail
func (c *storeController) CreateStore(ctx *gin.Context) {
	var storeDTO dto.StoreDTO
	
	// Parse request body
	if err := ctx.ShouldBindJSON(&storeDTO); err != nil {
		log.Printf("Error parsing request: %v", err)
		response := helper.BuildErrorResponse("Failed to parse request", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Dapatkan token dari header
	authHeader := ctx.GetHeader("Authorization")
	log.Printf("Auth header received: %s", authHeader)
	
	// Get all claims from token
	claims, err := c.getClaimsFromToken(authHeader)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process token", err.Error(), nil)
		ctx.JSON(http.StatusUnauthorized, response)
		return
	}
	
	// Ekstrak email dari claims
	email, ok := claims["email"].(string)
	if !ok {
		log.Printf("Email not found in token claims")
		response := helper.BuildErrorResponse("Invalid token format", "Email not found in token", nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	log.Printf("Email from token: %s", email)
	
	// Get user from email
	user := c.authService.FindByEmail(email)
	log.Printf("User found: ID=%v, Email=%s, Role=%s", user.ID, user.Email, user.Role)
	
	// Periksa apakah user valid
	if user.ID == 0 {
		log.Printf("User not found with email: %s", email)
		response := helper.BuildErrorResponse("User not found", "No user found with email from token", nil)
		ctx.JSON(http.StatusNotFound, response)
		return
	}
	
	// Periksa role dengan debugging detail
	if user.Role != "penjual" {
		log.Printf("Role check failed: '%s' != 'penjual'", user.Role)
		// Check untuk whitespace atau hidden chars
		log.Printf("Role string bytes: %v", []byte(user.Role))
		log.Printf("Expected bytes: %v", []byte("penjual"))
		response := helper.BuildErrorResponse("Unauthorized", "Only users with seller role can create a store", nil)
		ctx.JSON(http.StatusForbidden, response)
		return
	}
	
	// Set userID dari user yang ditemukan
	storeDTO.UserID = int(user.ID)
	log.Printf("Setting UserID in DTO: %d", storeDTO.UserID)
	
	// Create store through service
	result, err := c.storeService.CreateStore(storeDTO)
	if err != nil {
		log.Printf("Store creation error: %v", err)
		response := helper.BuildErrorResponse("Failed to create store", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Return response
	response := helper.BuildResponse(true, "Store created successfully", result)
	ctx.JSON(http.StatusCreated, response)
}

func (c *storeController) UpdateStore(ctx *gin.Context) {
	// Ambil ID toko dari parameter URL
	storeID := ctx.Param("id")
	if storeID == "" {
		ctx.JSON(http.StatusBadRequest, helper.BuildResponse(false, "ID toko tidak diberikan", nil))
		return
	}

	// Ambil ID user dari context (disimpan oleh middleware auth)

	authHeader := ctx.GetHeader("Authorization")
	log.Printf("Auth header received: %s", authHeader)
	// Get all claims from token
	claims, err := c.getClaimsFromToken(authHeader)
	log.Println("Ini dari toket ", claims["email"])

	email, _ := claims["email"].(string)

	user := c.authService.FindByEmail(email)
	log.Printf("User found: ID=%v, Email=%s, Role=%s", user.ID, user.Email, user.Role)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process token", err.Error(), nil)
		ctx.JSON(http.StatusUnauthorized, response)
		return
	}

	userID := strconv.FormatUint(user.ID, 10)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, helper.BuildResponse(false, "User tidak terautentikasi", nil))
		return
	}

	

	// Bind data dari form
	var storeDTO dto.UpdateStoreDTO
	if err := ctx.ShouldBind(&storeDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, helper.BuildResponse(false, "Data tidak valid", nil))
		return
	}

	// Panggil service untuk update toko
	updatedStore, err := c.storeService.Update(ctx, storeID, userID, storeDTO)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, helper.BuildResponse(false, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, helper.BuildResponse(true, "Toko berhasil diperbarui", updatedStore))
}


// // GetStoreByID handles request to get a store by ID
// func (c *storeController) GetStoreByID(ctx *gin.Context) {
// 	// Parse store ID
// 	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Invalid ID format", err.Error(), nil)
// 		ctx.JSON(http.StatusBadRequest, response)
// 		return
// 	}
	
// 	// Get store from service
// 	store, err := c.storeService.GetStoreByID(id)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Store not found", err.Error(), nil)
// 		ctx.JSON(http.StatusNotFound, response)
// 		return
// 	}
	
// 	// Return response
// 	response := helper.BuildResponse(true, "Store fetched successfully", store)
// 	ctx.JSON(http.StatusOK, response)
// }

// // GetStoreByUserID handles request to get a store by UserID
// func (c *storeController) GetStoreByUserID(ctx *gin.Context) {
// 	// Get userID from path
// 	userID := ctx.Param("user_id")
	
// 	// Get store from service
// 	store, err := c.storeService.GetStoreByUserID(userID)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Store not found", err.Error(), nil)
// 		ctx.JSON(http.StatusNotFound, response)
// 		return
// 	}
	
// 	// Return response
// 	response := helper.BuildResponse(true, "Store fetched successfully", store)
// 	ctx.JSON(http.StatusOK, response)
// }

// // UpdateStore handles request to update an existing store
// func (c *storeController) UpdateStore(ctx *gin.Context) {
// 	// Parse store ID
// 	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Invalid ID format", err.Error(), nil)
// 		ctx.JSON(http.StatusBadRequest, response)
// 		return
// 	}
	
// 	// Parse request body
// 	var storeDTO dto.StoreDTO
// 	if err := ctx.ShouldBindJSON(&storeDTO); err != nil {
// 		response := helper.BuildErrorResponse("Failed to parse request", err.Error(), nil)
// 		ctx.JSON(http.StatusBadRequest, response)
// 		return
// 	}
	
// 	// Dapatkan email dari token
// 	authHeader := ctx.GetHeader("Authorization")
// 	email, err := c.getEmailFromToken(authHeader)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Failed to process token", err.Error(), nil)
// 		ctx.JSON(http.StatusUnauthorized, response)
// 		return
// 	}
	
// 	// Dapatkan user berdasarkan email
// 	user := c.authService.FindByEmail(email)
	
// 	// Verify the existing store belongs to this user
// 	existingStore, err := c.storeService.GetStoreByID(id)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Store not found", err.Error(), nil)
// 		ctx.JSON(http.StatusNotFound, response)
// 		return
// 	}
	
// 	if existingStore.UserID != strconv.FormatUint(user.ID, 10) {
// 		response := helper.BuildErrorResponse("Not authorized", "You can only update your own store", nil)
// 		ctx.JSON(http.StatusForbidden, response)
// 		return
// 	}
	
// 	// Set userID dari user yang ditemukan (untuk memastikan konsistensi)
// 	storeDTO.UserID = strconv.FormatUint(user.ID, 10)
	
// 	// Update store through service
// 	result, err := c.storeService.UpdateStore(id, storeDTO)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Failed to update store", err.Error(), nil)
// 		ctx.JSON(http.StatusBadRequest, response)
// 		return
// 	}
	
// 	// Return response
// 	response := helper.BuildResponse(true, "Store updated successfully", result)
// 	ctx.JSON(http.StatusOK, response)
// }

// // DeleteStore handles request to delete a store
// func (c *storeController) DeleteStore(ctx *gin.Context) {
// 	// Parse store ID
// 	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Invalid ID format", err.Error(), nil)
// 		ctx.JSON(http.StatusBadRequest, response)
// 		return
// 	}
	
// 	// Dapatkan email dari token
// 	authHeader := ctx.GetHeader("Authorization")
// 	email, err := c.getEmailFromToken(authHeader)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Failed to process token", err.Error(), nil)
// 		ctx.JSON(http.StatusUnauthorized, response)
// 		return
// 	}
	
// 	// Dapatkan user berdasarkan email
// 	user := c.authService.FindByEmail(email)
	
// 	// Verify the existing store belongs to this user
// 	existingStore, err := c.storeService.GetStoreByID(id)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Store not found", err.Error(), nil)
// 		ctx.JSON(http.StatusNotFound, response)
// 		return
// 	}
	
// 	if existingStore.UserID != strconv.FormatUint(user.ID, 10) {
// 		response := helper.BuildErrorResponse("Not authorized", "You can only delete your own store", nil)
// 		ctx.JSON(http.StatusForbidden, response)
// 		return
// 	}
	
// 	// Delete store through service
// 	err = c.storeService.DeleteStore(id)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Failed to delete store", err.Error(), nil)
// 		ctx.JSON(http.StatusBadRequest, response)
// 		return
// 	}
	
// 	// Return response
// 	response := helper.BuildResponse(true, "Store deleted successfully", nil)
// 	ctx.JSON(http.StatusOK, response)
// }

// // GetAllStores handles request to get all stores with pagination
// func (c *storeController) GetAllStores(ctx *gin.Context) {
// 	// Parse pagination parameters
// 	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
// 	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	
// 	// Get stores from service
// 	stores, total, err := c.storeService.GetAllStores(page, limit)
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Failed to fetch stores", err.Error(), nil)
// 		ctx.JSON(http.StatusBadRequest, response)
// 		return
// 	}
	
// 	// Create pagination
// 	pagination := utils.NewPagination(page, limit, total)
	
// 	// Return response
// 	data := map[string]interface{}{
// 		"stores":     stores,
// 		"pagination": pagination,
// 	}
// 	response := helper.BuildResponse(true, "Stores fetched successfully", data)
// 	ctx.JSON(http.StatusOK, response)
// }

// // UploadStoreImage handles avatar/banner image uploads for a store
// func (c *storeController) UploadStoreImage(ctx *gin.Context) {
// 	// Method stub - akan diimplementasikan sesuai kebutuhan
// }