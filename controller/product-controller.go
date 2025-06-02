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
	CreateProduct(c *gin.Context)
	GetProductBySlug(c *gin.Context)
	GetProductsByStoreID(c *gin.Context)
	UpdateProduct(c *gin.Context)
	DeleteProduct(c *gin.Context)
	AddProductImage(c *gin.Context)
	DeleteProductImage(c *gin.Context)
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
		Thumbnail:   product.Thumbnail,
		Images:      images,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
	
	c.JSON(http.StatusOK, helper.BuildResponse(true, "Produk ditemukan", response))
}

func (ctrl *productController) GetProductsByStoreID(c *gin.Context) {
	// Dapatkan store ID dari parameter URL
	storeIDParam := c.Param("storeId")
	storeID, err := strconv.Atoi(storeIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.BuildResponse(false, "ID toko tidak valid", nil))
		return
	}
	
	// Ambil semua produk di toko
	products, err := ctrl.productService.GetProductsByStoreID(storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.BuildResponse(false, "Gagal mendapatkan produk", nil))
		return
	}
	
	// Konversi entity ke DTO response (card response tanpa gambar tambahan)
	var productResponses []dto.ProductCardResponse
	for _, product := range products {
		productResponses = append(productResponses, dto.ProductCardResponse{
			ID:          product.ID,
			Slug:        product.Slug,
			Name:        product.Name,
			Harga:       product.Harga,
			StoreID:     product.StoreID,
			Thumbnail:   product.Thumbnail,
			CreatedAt:   product.CreatedAt,
		})
	}
	
	c.JSON(http.StatusOK, helper.BuildResponse(true, "Daftar produk", productResponses))
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