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

	// Service
	jwtService     service.JWTService     = service.NewJWTService()
	userService    service.UserService    = service.NewUserService(userRepository)
	authService    service.AuthService    = service.NewAuthServie(userRepository)
	articleService service.ArticleService = service.NewArticleService(articleRepository)
	storeService service.StoreService = service.NewStoreService(storeRepository)
	productService service.ProductService = service.NewProductService(productRepository, productImageRepository)

	// Controller
	userController    controller.UserController    = controller.NewUserController(userService, jwtService)
	authController    controller.AuthController    = controller.NewAuthController(authService, jwtService)
	articleController controller.ArticleController = controller.NewArticleController(articleService, jwtService)
	storeController controller.StoreController = controller.NewStoreController(storeService, jwtService, authService)
	productController controller.ProductController = controller.NewProductController(productService, storeService, jwtService, authService)

)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "false")
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

			protected.GET("/all-users", userController.GetAllUser)

			protected.POST("/articles", articleController.CreateArticle)
			protected.PUT("/articles/:id", articleController.UpdateArticle)
			protected.DELETE("/articles/:id", articleController.DeleteArticle)
			protected.POST("/upload", utils.UploadImage)


			// Store
			protected.POST("/store", storeController.CreateStore)
			protected.PUT("/store/:id", storeController.UpdateStore)


			// Product (dashboard)
			protected.POST("/product", productController.CreateProduct)
			protected.GET("/product/detail/:slug", productController.GetProductBySlug)
			protected.GET("/product/store/:id", productController.GetProductsByStoreID)
			protected.PUT("/product/:slug", productController.UpdateProduct)
			protected.DELETE("/product/:slug", productController.DeleteProduct)
			protected.POST("/product/image", productController.AddProductImage)
			protected.DELETE("/product/image/:id", productController.DeleteProductImage)
		}
	}

	r.Run(":8081")
}
