// controller/product_controller.go
package controller

import (
	"batik/dto"
	"batik/helper"
	"batik/service"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type ProductController interface {
	GetProductsByStoreID(c *gin.Context)
	GetProductsByStoreIDPublic(c *gin.Context)
	CreateProduct(c *gin.Context)
	GetProductBySlug(c *gin.Context)
	UpdateProduct(c *gin.Context)
	DeleteProduct(c *gin.Context)
	AddProductImage(c *gin.Context)
	DeleteProductImage(c *gin.Context)
	GetAllPublicProduct(c *gin.Context)
	GetLatestProduct(c *gin.Context)
	GetDetailProduct(c *gin.Context)
}

type productController struct {
	productService service.ProductService
	storeService   service.StoreService
	jwtService     service.JWTService
	authService    service.AuthService
}

func NewProductController(productService service.ProductService, storeService service.StoreService, jwtService service.JWTService, authService service.AuthService) ProductController {
	return &productController{
		productService: productService,
		storeService:   storeService,
		jwtService:     jwtService,
		authService:    authService,
	}
}

func (ctrl *productController) GetProductsByStoreID(c *gin.Context) {
	// Fix parameter name - gunakan "id" sesuai routing
	storeIDParam := c.Param("id")
	storeID, err := strconv.Atoi(storeIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, "ID toko tidak valid", nil))
		return
	}
	
	// Dapatkan token dari header untuk validasi ownership
	authHeader := c.GetHeader("Authorization")
	
	// Dapatkan user ID dari token
	userID, err := ctrl.getUserFromToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.BuildResponse(false, "User tidak terautentikasi: "+err.Error(), nil))
		return
	}
	
	// Validasi kepemilikan toko
	store, err := ctrl.storeService.GetStoreByID(strconv.Itoa(storeID))
	if err != nil {
		c.JSON(http.StatusNotFound, helper.BuildResponse(false, "Toko tidak ditemukan", nil))
		return
	}
	
	if strconv.Itoa(store.UserID) != userID {
		c.JSON(http.StatusForbidden, helper.BuildResponse(false, "Anda tidak memiliki akses ke toko ini", nil))
		return
	}
	
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	search := c.Query("search")
	
	// Ambil produk dengan pagination
	products, pagination, err := ctrl.productService.GetAllProductByStore(storeID, page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.BuildResponse(false, "Gagal mendapatkan produk", nil))
		return
	}
	
	// Konversi entity ke DTO response
	var productResponses []dto.ProductCardResponse
	for _, product := range products {
		productResponses = append(productResponses, dto.ProductCardResponse{
			ID:          product.ID,
			Slug:        product.Slug,
			Name:        product.Name,
			Description: product.Description, // Tambahkan description untuk owner
			Harga:       product.Harga,
			StoreID:     product.StoreID,
			Thumbnail:   product.Thumbnail,
			CreatedAt:   product.CreatedAt,
		})
	}
	
	// Return response dengan pagination dan store info
	data := map[string]interface{}{
		"store":      store,
		"products":   productResponses,
		"pagination": pagination,
	}
	
	c.JSON(http.StatusOK, helper.BuildResponse(true, "Daftar produk toko berhasil diambil", data))
}

// GetProductsByStoreIDPublic - Untuk public access (tanpa auth)
func (ctrl *productController) GetProductsByStoreIDPublic(c *gin.Context) {
	storeIDParam := c.Param("id")
	storeID, err := strconv.Atoi(storeIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, "ID toko tidak valid", nil))
		return
	}
	
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	search := c.Query("search")
	
	// Ambil produk dengan pagination
	products, pagination, err := ctrl.productService.GetAllProductByStore(storeID, page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.BuildResponse(false, "Gagal mendapatkan produk", nil))
		return
	}
	
	// Konversi entity ke DTO response (tanpa description untuk public)
	var productResponses []dto.ProductCardResponse
	for _, product := range products {
		productResponses = append(productResponses, dto.ProductCardResponse{
			ID:        product.ID,
			Slug:      product.Slug,
			Name:      product.Name,
			Harga:     product.Harga,
			StoreID:   product.StoreID,
			Thumbnail: product.Thumbnail,
			CreatedAt: product.CreatedAt,
		})
	}
	
	// Get store info untuk public
	store, _ := ctrl.storeService.GetStoreByID(strconv.Itoa(storeID))
	
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"id":          store.ID,
			"name":        store.Name,
			"description": store.Description,
			"avatar":      store.Avatar,
			"banner":      store.Banner,
		},
		"products":   productResponses,
		"pagination": pagination,
	}
	
	c.JSON(http.StatusOK, helper.BuildResponse(true, "Katalog produk toko", data))
}

// getClaimsFromToken adalah helper untuk mendapatkan semua informasi dari token
func (ctrl *productController) getClaimsFromToken(authHeader string) (jwt.MapClaims, error) {
	// Validasi token
	token, err := ctrl.jwtService.ValidateToken(authHeader)
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
	
	return claims, nil
}

// getUserFromToken adalah helper untuk mendapatkan user dari token
func (ctrl *productController) getUserFromToken(authHeader string) (string, error) {
	// Get claims dari token
	claims, err := ctrl.getClaimsFromToken(authHeader)
	if err != nil {
		return "", err
	}
	
	// Ekstrak email dari claims
	email, ok := claims["email"].(string)
	if !ok {
		return "", fmt.Errorf("email not found in token claims")
	}
	
	// Get user from email
	user := ctrl.authService.FindByEmail(email)
	if user.ID == 0 {
		return "", fmt.Errorf("user not found")
	}
	
	return strconv.FormatUint(user.ID, 10), nil
}

func (ctrl *productController) CreateProduct(c *gin.Context) {
	// Dapatkan token dari header
	authHeader := c.GetHeader("Authorization")
	
	// Dapatkan user ID dari token
	userID, err := ctrl.getUserFromToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.BuildResponse(false, "User tidak terautentikasi: "+err.Error(), nil))
		return
	}
	
	// Bind data product dari form
	var productDTO dto.CreateProductDTO
	if err := c.ShouldBind(&productDTO); err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, "Data produk tidak valid", nil))
		return
	}
	
	// Validasi kepemilikan toko (pastikan user adalah pemilik toko)
	store, err := ctrl.storeService.GetStoreByID(strconv.Itoa(productDTO.StoreID))
	if err != nil {
		c.JSON(http.StatusNotFound, helper.BuildResponse(false, "Toko tidak ditemukan", nil))
		return
	}
	
	if strconv.Itoa(store.UserID) != userID {
		c.JSON(http.StatusForbidden, helper.BuildResponse(false, "Anda tidak memiliki akses ke toko ini", nil))
		return
	}
	
	// Dapatkan file gambar dari multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, "Gagal memproses form", nil))
		return
	}
	
	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, "Minimal satu gambar produk diperlukan", nil))
		return
	}
	
	// Panggil service untuk membuat produk
	product, err := ctrl.productService.CreateProduct(c, productDTO, files)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, err.Error(), nil))
		return
	}
	
	// Konversi entity ke DTO response
	var images []dto.ProductImageDTO
	for _, img := range product.Images {
		images = append(images, dto.ProductImageDTO{
			ID:    img.ID,
			Image: img.Image,
		})
	}
	
	response := dto.ProductResponse{
		ID:          product.ID,
		Slug:        product.Slug,
		Name:        product.Name,
		Description: product.Description,
		Harga:       product.Harga,
		StoreID:     product.StoreID,
		CategoryID:  product.CategoryID,
		Thumbnail:   product.Thumbnail,
		Images:      images,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
	
	c.JSON(http.StatusCreated, helper.BuildResponse(true, "Produk berhasil dibuat", response))
}

func (ctrl *productController) GetProductBySlug(c *gin.Context) {
	// Dapatkan slug produk dari parameter URL
	slug := c.Param("slug")
	
	// Ambil data produk
	product, err := ctrl.productService.GetProductBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, helper.BuildResponse(false, "Produk tidak ditemukan", nil))
		return
	}
	
	// Konversi entity ke DTO response
	var images []dto.ProductImageDTO
	for _, img := range product.Images {
		images = append(images, dto.ProductImageDTO{
			ID:    img.ID,
			Image: img.Image,
		})
	}
	
	response := dto.ProductResponse{
		ID:          product.ID,
		Slug:        product.Slug,
		Name:        product.Name,
		Description: product.Description,
		Harga:       product.Harga,
		StoreID:     product.StoreID,
		CategoryID:  product.CategoryID,
		Thumbnail:   product.Thumbnail,
		Images:      images,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
	
	c.JSON(http.StatusOK, helper.BuildResponse(true, "Produk ditemukan", response))
}


func (ctrl *productController) UpdateProduct(c *gin.Context) {
	// Dapatkan slug dari parameter URL
	slug := c.Param("slug")
	
	// Dapatkan token dari header
	authHeader := c.GetHeader("Authorization")
	
	// Dapatkan user ID dari token
	userID, err := ctrl.getUserFromToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.BuildResponse(false, "User tidak terautentikasi: "+err.Error(), nil))
		return
	}
	
	// Ambil produk untuk validasi kepemilikan
	product, err := ctrl.productService.GetProductBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, helper.BuildResponse(false, "Produk tidak ditemukan", nil))
		return
	}
	
	// Validasi kepemilikan toko
	store, err := ctrl.storeService.GetStoreByID(strconv.Itoa(product.StoreID))
	if err != nil {
		c.JSON(http.StatusNotFound, helper.BuildResponse(false, "Toko tidak ditemukan", nil))
		return
	}
	
	if strconv.Itoa(store.UserID) != userID {
		c.JSON(http.StatusForbidden, helper.BuildResponse(false, "Anda tidak memiliki akses untuk mengubah produk ini", nil))
		return
	}
	
	// Bind data update
	var updateDTO dto.UpdateProductDTO
	if err := c.ShouldBind(&updateDTO); err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, "Data update tidak valid", nil))
		return
	}
	
	// Update produk
	updatedProduct, err := ctrl.productService.UpdateProduct(c, slug, updateDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, err.Error(), nil))
		return
	}
	
	// Ambil data lengkap produk setelah update
	completeProduct, err := ctrl.productService.GetProductByID(updatedProduct.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.BuildResponse(false, "Produk berhasil diupdate tetapi gagal mengambil data lengkap", nil))
		return
	}
	
	// Konversi entity ke DTO response
	var images []dto.ProductImageDTO
	for _, img := range completeProduct.Images {
		images = append(images, dto.ProductImageDTO{
			ID:    img.ID,
			Image: img.Image,
		})
	}
	
	response := dto.ProductResponse{
		ID:          completeProduct.ID,
		Slug:        completeProduct.Slug,
		Name:        completeProduct.Name,
		Description: completeProduct.Description,
		Harga:       completeProduct.Harga,
		StoreID:     completeProduct.StoreID,
		Thumbnail:   completeProduct.Thumbnail,
		Images:      images,
		CreatedAt:   completeProduct.CreatedAt,
		UpdatedAt:   completeProduct.UpdatedAt,
	}
	
	c.JSON(http.StatusOK, helper.BuildResponse(true, "Produk berhasil diupdate", response))
}

func (ctrl *productController) DeleteProduct(c *gin.Context) {
	// Dapatkan slug produk dari parameter URL
	slug := c.Param("slug")
	
	// Dapatkan token dari header
	authHeader := c.GetHeader("Authorization")
	
	// Dapatkan user ID dari token
	userID, err := ctrl.getUserFromToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.BuildResponse(false, "User tidak terautentikasi: "+err.Error(), nil))
		return
	}
	
	// Ambil produk untuk validasi kepemilikan
	product, err := ctrl.productService.GetProductBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, helper.BuildResponse(false, "Produk tidak ditemukan", nil))
		return
	}
	
	// Validasi kepemilikan toko
	store, err := ctrl.storeService.GetStoreByID(strconv.Itoa(product.StoreID))
	if err != nil {
		c.JSON(http.StatusNotFound, helper.BuildResponse(false, "Toko tidak ditemukan", nil))
		return
	}
	
	if strconv.Itoa(store.UserID) != userID {
		c.JSON(http.StatusForbidden, helper.BuildResponse(false, "Anda tidak memiliki akses untuk menghapus produk ini", nil))
		return
	}
	
	// Hapus produk
	if err := ctrl.productService.DeleteProduct(slug); err != nil {
		c.JSON(http.StatusInternalServerError, helper.BuildResponse(false, err.Error(), nil))
		return
	}
	
	c.JSON(http.StatusOK, helper.BuildResponse(true, "Produk berhasil dihapus", nil))
}

func (ctrl *productController) AddProductImage(c *gin.Context) {
	// Dapatkan slug produk dari parameter URL
	slug := c.Param("slug")
	
	// Dapatkan token dari header
	authHeader := c.GetHeader("Authorization")
	
	// Dapatkan user ID dari token
	userID, err := ctrl.getUserFromToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.BuildResponse(false, "User tidak terautentikasi: "+err.Error(), nil))
		return
	}
	
	// Ambil produk untuk validasi kepemilikan
	product, err := ctrl.productService.GetProductBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, helper.BuildResponse(false, "Produk tidak ditemukan", nil))
		return
	}
	
	// Validasi kepemilikan toko
	store, err := ctrl.storeService.GetStoreByID(strconv.Itoa(product.StoreID))
	if err != nil {
		c.JSON(http.StatusNotFound, helper.BuildResponse(false, "Toko tidak ditemukan", nil))
		return
	}
	
	if strconv.Itoa(store.UserID) != userID {
		c.JSON(http.StatusForbidden, helper.BuildResponse(false, "Anda tidak memiliki akses untuk menambah gambar produk ini", nil))
		return
	}
	
	// Dapatkan file gambar
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, "Tidak ada file gambar yang diberikan", nil))
		return
	}
	
	// Tambahkan gambar ke produk
	productImage, err := ctrl.productService.AddProductImage(c, slug, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, err.Error(), nil))
		return
	}
	
	c.JSON(http.StatusCreated, helper.BuildResponse(true, "Gambar produk berhasil ditambahkan", dto.ProductImageDTO{
		ID:    productImage.ID,
		Image: productImage.Image,
	}))
}

func (ctrl *productController) DeleteProductImage(c *gin.Context) {
	// Dapatkan slug produk dan ID gambar dari parameter URL
	slug := c.Param("slug")
	imageIDParam := c.Param("imageId")
	
	imageID, err := strconv.Atoi(imageIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, "ID gambar tidak valid", nil))
		return
	}
	
	// Dapatkan token dari header
	authHeader := c.GetHeader("Authorization")
	
	// Dapatkan user ID dari token
	userID, err := ctrl.getUserFromToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.BuildResponse(false, "User tidak terautentikasi: "+err.Error(), nil))
		return
	}
	
	// Ambil produk untuk validasi kepemilikan
	product, err := ctrl.productService.GetProductBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, helper.BuildResponse(false, "Produk tidak ditemukan", nil))
		return
	}
	
	// Validasi kepemilikan toko
	store, err := ctrl.storeService.GetStoreByID(strconv.Itoa(product.StoreID))
	if err != nil {
		c.JSON(http.StatusNotFound, helper.BuildResponse(false, "Toko tidak ditemukan", nil))
		return
	}
	
	if strconv.Itoa(store.UserID) != userID {
		c.JSON(http.StatusForbidden, helper.BuildResponse(false, "Anda tidak memiliki akses untuk menghapus gambar produk ini", nil))
		return
	}
	
	// Hapus gambar produk
	if err := ctrl.productService.DeleteProductImage(slug, imageID); err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, err.Error(), nil))
		return
	}
	
	c.JSON(http.StatusOK, helper.BuildResponse(true, "Gambar produk berhasil dihapus", nil))
}

func (ctrl *productController) GetAllPublicProduct(c *gin.Context) {
	page, errPage := strconv.Atoi(c.DefaultQuery("page", "1"))
	if errPage != nil || page < 1 {
		page = 1
	}
	limit, errLimit := strconv.Atoi(c.DefaultQuery("limit", "40"))
	if errLimit != nil || limit < 1 {
		limit = 40 // Batas default
	}

	products, pagination, err := ctrl.productService.GetAllPublicProduct(page, limit)
	if err != nil {
		response := helper.BuildErrorResponse("Gagal mengambil produk dengan detail", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, response)
		return
	}
    
    data := map[string]interface{}{
		"products":   products,
		"pagination": pagination,
	}

	if len(products) == 0 {
		response := helper.BuildResponse(true, "Tidak ada produk yang ditemukan", data)
		c.JSON(http.StatusOK, response)
		return
	}

	response := helper.BuildResponse(true, "Produk dengan detail berhasil diambil", data)
	c.JSON(http.StatusOK, response)
}

func (ctrl *productController) GetLatestProduct(c *gin.Context) {
	products, err := ctrl.productService.GetLatestProduct()

	if err != nil {
		res := helper.BuildErrorResponse("Gagal menampilkan data", err.Error(), nil)
		c.JSON(http.StatusNotFound, res)
	}

	res := helper.BuildResponse(true, "Berhasil menampilkan data", products)
	c.JSON(http.StatusOK, res)
}

func (ctrl *productController) GetDetailProduct(c *gin.Context) {
	slug := c.Param("slug")
	products, err := ctrl.productService.GetDetailProduct(slug)

	if err != nil {
		res := helper.BuildErrorResponse("Gagal menampilkan data", err.Error(), nil)
		c.JSON(http.StatusNotFound, res)
	}

	res := helper.BuildResponse(true, "Berhasil menampilkan data", products)
	c.JSON(http.StatusOK, res)
}