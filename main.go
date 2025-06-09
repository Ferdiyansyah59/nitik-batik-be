package main

import (
	"batik/config"
	"batik/controller"
	"batik/middleware"
	"batik/repository"
	"batik/service"
	"batik/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	db *gorm.DB = config.SetupDatabaseConnection()

	// Repo
	userRepository    repository.UserRepository    = repository.NewUserRepository(db)
	articleRepository repository.ArticleRepository = repository.NewArticleRepository(db)
	storeRepository repository.StoreRepository = repository.NewStoreRepository(db)
	productRepository repository.ProductRepository = repository.NewProductRepository(db)
	productImageRepository repository.ProductImageRepository = repository.NewProductImageRepository(db)
	productCategoryRepository repository.ProductCategoryRepository = repository.NewProductCategoryRepository(db)

	// Service
	jwtService     service.JWTService     = service.NewJWTService()
	userService    service.UserService    = service.NewUserService(userRepository)
	authService    service.AuthService    = service.NewAuthServie(userRepository)
	articleService service.ArticleService = service.NewArticleService(articleRepository)
	storeService service.StoreService = service.NewStoreService(storeRepository)
	productService service.ProductService = service.NewProductService(productRepository, productImageRepository)
	productCategoryService service.ProductCategoryService = service.NewProductCategoryService(productCategoryRepository)

	// Controller
	userController    controller.UserController    = controller.NewUserController(userService, jwtService)
	authController    controller.AuthController    = controller.NewAuthController(authService, jwtService)
	articleController controller.ArticleController = controller.NewArticleController(articleService, jwtService)
	storeController controller.StoreController = controller.NewStoreController(storeService, jwtService, authService)
	productController controller.ProductController = controller.NewProductController(productService, storeService, jwtService, authService)
	productCategoryController controller.ProductCategoryController = controller.NewProductCategoryController(productCategoryService)

)

func CORSMiddleware() gin.HandlerFunc {
    // Define allowed origins
    allowedOrigins := []string{
        "http://localhost:3000",
        "http://82.112.230.106:1801",
        "https://nitikbatik.ferdirns.com",
    }
    
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        
        // Check if origin is in allowed list
        for _, allowedOrigin := range allowedOrigins {
            if origin == allowedOrigin {
                c.Header("Access-Control-Allow-Origin", origin)
                break
            }
        }
        
        c.Header("Access-Control-Allow-Credentials", "true") // biasanya true kalau specific origins
        c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT, DELETE")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}

func main() {
	defer config.CloseDatabaseConnection(db)
	r := gin.Default()
	r.Use(CORSMiddleware())


	// Serve static files (images)
	r.Static("/uploads", "./uploads")
	
	// Additional specific routes for subdirectories if needed
	// r.Static("/uploads/images", "./uploads/images")
	// r.Static("/uploads/product-images", "./uploads/product-images")
	// r.Static("/uploads/store-avatar", "./uploads/store-avatar")
	// r.Static("/uploads/store-banner", "./uploads/store-banner")
	
	// // For serving other static assets if needed
	// r.Static("/assets", "./assets")
	// r.Static("/public", "./public")

	authRoutes := r.Group("api")
	{
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/register", authController.Register)
	}

	// articleRoutes := r.Group("api", middleware.AuthorizeJWT(jwtService))
	// {
	// 	// Untuk menampilkan semua data
	// 	articleRoutes.GET("/getAllArticles", articleController.GetAllArticle)
	// 	// Untuk output klasifikasi
	// 	articleRoutes.GET("/getArticleWithKey/:title", articleController.GetArticleByKey)
	// }

	userRoutes := r.Group("api") 
	{
		protected := userRoutes.Group("", middleware.AuthorizeJWT(jwtService))
		{
			protected.GET("/all-users", userController.GetAllUser)
		}
	}

	storeAuth := r.Group("api")
	{
		storeAuth.GET("/stores", storeController.GetAllStores)                   
		storeAuth.GET("/store/user/:user_id", storeController.GetStoreByUserID)

				// Protected routes (require JWT authentication)
		protected := storeAuth.Group("", middleware.AuthorizeJWT(jwtService))
		{
			// Store
			protected.GET("/stores-data", storeController.GetAllStoreData)
			protected.POST("/store", storeController.CreateStore)
			protected.PUT("/store/:id", storeController.UpdateStore)
			protected.GET("/store/:id", storeController.GetStoreByID)
		}
	}

	productRoutes := r.Group("api") 
	{
				// Product
		productRoutes.GET("/products", productController.GetAllPublicProduct) 
		productRoutes.GET("/latest-products", productController.GetLatestProduct)
		productRoutes.GET("/store/:id/products", productController.GetProductsByStoreIDPublic)
		productRoutes.GET("/product/:slug", productController.GetDetailProduct)
		productRoutes.GET("/products/category/:slug", productController.GetAllPublicProductByCategory)
		productRoutes.GET("/products/store/:id", productController.GetPublicProductsByStoreID)

		protected := productRoutes.Group("", middleware.AuthorizeJWT(jwtService))
		{
			// Product (dashboard)
			protected.POST("/product", productController.CreateProduct)
			protected.GET("/product/detail/:slug", productController.GetProductBySlug)
			protected.GET("/my-store/:id/products", productController.GetProductsByStoreID)
			protected.PUT("/product/:slug", productController.UpdateProduct)
			protected.DELETE("/product/:slug", productController.DeleteProduct)
			protected.POST("/product/image", productController.AddProductImage)
			protected.DELETE("/product/image/:id", productController.DeleteProductImage)
		}
	}

	productCategory := r.Group("api")
	{
		productCategory.GET("/product-category", productCategoryController.GetProductCategory)
	}



	articleRoutes := r.Group("api")
	{
		// Public routes
		articleRoutes.GET("/articles", articleController.GetAllArticles)
		articleRoutes.GET("/latest-articles", articleController.GetLatestArticles)
		articleRoutes.GET("/articles/:id", articleController.GetArticleByID)
		articleRoutes.GET("/articles/slug/:slug", articleController.GetArticleBySlug)
		articleRoutes.GET("/articles/search", articleController.SearchArticles)
		
		// Protected routes (require JWT authentication)
		protected := articleRoutes.Group("", middleware.AuthorizeJWT(jwtService))
		{
			protected.POST("/articles", articleController.CreateArticle)
			protected.PUT("/articles/:id", articleController.UpdateArticle)
			protected.DELETE("/articles/:id", articleController.DeleteArticle)
			protected.POST("/upload", utils.UploadImage)
		}
	}

	r.Run(":1815")
}
