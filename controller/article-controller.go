package controller

import (
	"batik/entity"
	"batik/helper"
	"batik/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ArticleController interface represents the article controller contract
type ArticleController interface {
	GetAllArticles(c *gin.Context)
	GetLatestArticles(c *gin.Context)
	GetArticleByID(c *gin.Context)
	GetArticleBySlug(c *gin.Context)
	CreateArticle(c *gin.Context)
	UpdateArticle(c *gin.Context)
	DeleteArticle(c *gin.Context)
	SearchArticles(c *gin.Context)
}

// articleController is the implementation of ArticleController interface
type articleController struct {
	articleService service.ArticleService
	jwtService     service.JWTService
}

// NewArticleController creates a new instance of ArticleController
func NewArticleController(articleService service.ArticleService, jwtService service.JWTService) ArticleController {
	return &articleController{
		articleService: articleService,
		jwtService:     jwtService,
	}
}

// GetAllArticles handles request to get all articles with pagination
func (c *articleController) GetAllArticles(ctx *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	search := ctx.Query("search")
	
	// Get articles from service
	articles, pagination, err := c.articleService.GetAllArticles(page, limit, search)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to fetch articles", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Return response - combine articles and pagination into a single data object
	data := map[string]interface{}{
		"articles":   articles,
		"pagination": pagination,
	}
	response := helper.BuildResponse(true, "Articles fetched successfully", data)
	ctx.JSON(http.StatusOK, response)
}

func (c *articleController) GetLatestArticles(ctx *gin.Context) {
	articles, err := c.articleService.GetLatestArticles()
	if err != nil {
		response := helper.BuildErrorResponse("Failed to fetch latest articles", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	response := helper.BuildResponse(
		true,
		"Latest Artiles fethed successfully",
		articles,
	)
	ctx.JSON(http.StatusOK, response)
}

// GetArticleByID handles request to get an article by ID
func (c *articleController) GetArticleByID(ctx *gin.Context) {
	// Parse article ID
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response := helper.BuildErrorResponse("Invalid ID format", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Get article from service
	article, err := c.articleService.GetArticleByID(id)
	if err != nil {
		response := helper.BuildErrorResponse("Article not found", err.Error(), nil)
		ctx.JSON(http.StatusNotFound, response)
		return
	}
	
	// Return response
	response := helper.BuildResponse(true, "Article fetched successfully", article)
	ctx.JSON(http.StatusOK, response)
}

// GetArticleBySlug handles request to get an article by slug
func (c *articleController) GetArticleBySlug(ctx *gin.Context) {
	// Get slug from path
	slug := ctx.Param("slug")
	
	// Get article from service
	article, err := c.articleService.GetArticleBySlug(slug)
	if err != nil {
		response := helper.BuildErrorResponse("Article not found", err.Error(), nil)
		ctx.JSON(http.StatusNotFound, response)
		return
	}
	
	// Return response
	response := helper.BuildResponse(true, "Article fetched successfully", article)
	ctx.JSON(http.StatusOK, response)
}

// CreateArticle handles request to create a new article
func (c *articleController) CreateArticle(ctx *gin.Context) {
	// Parse request body
	var article entity.Article
	if err := ctx.ShouldBindJSON(&article); err != nil {
		response := helper.BuildErrorResponse("Failed to parse request", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Validate required fields
	if article.Title == "" || article.Description == "" {
		response := helper.BuildErrorResponse("Title and description are required", "Validation error", nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Create article through service
	result, err := c.articleService.CreateArticle(article)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to create article", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Return response
	response := helper.BuildResponse(true, "Article created successfully", result)
	ctx.JSON(http.StatusCreated, response)
}


// CreateArticle handles request to create a new article with form-data
// func (c *articleController) CreateArticle(ctx *gin.Context) {
//     // Parse form data
//     if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
//         response := helper.BuildErrorResponse("Failed to parse form data", err.Error(), nil)
//         ctx.JSON(http.StatusBadRequest, response)
//         return
//     }
    
//     // Get form values
//     title := ctx.PostForm("title")
//     description := ctx.PostForm("description")
//     excerpt := ctx.PostForm("excerpt")
    
//     // Validate required fields
//     if title == "" || description == "" {
//         response := helper.BuildErrorResponse("Title and description are required", "Validation error", nil)
//         ctx.JSON(http.StatusBadRequest, response)
//         return
//     }
    
//     // Handle image upload if available
//     var imageUrl string
//     file, header, err := ctx.Request.FormFile("imageUrl")
//     if err == nil { // Image was uploaded
//         defer file.Close()
        
//         // Generate unique filename
//         filename := fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(header.Filename))
        
//         // Ensure upload directory exists
//         uploadDir := "uploads/images"
//         if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
//             if err := os.MkdirAll(uploadDir, 0755); err != nil {
//                 response := helper.BuildErrorResponse("Failed to create upload directory", err.Error(), nil)
//                 ctx.JSON(http.StatusInternalServerError, response)
//                 return
//             }
//         }
        
//         // Save the file
//         filepath := filepath.Join(uploadDir, filename)
//         out, err := os.Create(filepath)
//         if err != nil {
//             response := helper.BuildErrorResponse("Failed to save image", err.Error(), nil)
//             ctx.JSON(http.StatusInternalServerError, response)
//             return
//         }
//         defer out.Close()
        
//         // Copy the uploaded file to the destination file
//         _, err = io.Copy(out, file)
//         if err != nil {
//             response := helper.BuildErrorResponse("Failed to save image", err.Error(), nil)
//             ctx.JSON(http.StatusInternalServerError, response)
//             return
//         }
        
//         // Set image URL
//         imageUrl = fmt.Sprintf("/%s/%s", uploadDir, filename)
//     }
    
//     // Create article object
//     article := entity.Article{
//         Title:       title,
//         Description: description,
//         Excerpt:     excerpt,
//         ImageURL:    imageUrl,
//     }
    
//     // Create article through service
//     result, err := c.articleService.CreateArticle(article)
//     if err != nil {
//         response := helper.BuildErrorResponse("Failed to create article", err.Error(), nil)
//         ctx.JSON(http.StatusBadRequest, response)
//         return
//     }
    
//     // Return response
//     response := helper.BuildResponse(true, "Article created successfully", result)
//     ctx.JSON(http.StatusCreated, response)
// }

// UpdateArticle handles request to update an existing article
func (c *articleController) UpdateArticle(ctx *gin.Context) {
	// Parse article ID
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response := helper.BuildErrorResponse("Invalid ID format", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Parse request body
	var article entity.Article
	if err := ctx.ShouldBindJSON(&article); err != nil {
		response := helper.BuildErrorResponse("Failed to parse request", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Update article through service
	result, err := c.articleService.UpdateArticle(id, article)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to update article", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Return response
	response := helper.BuildResponse(true, "Article updated successfully", result)
	ctx.JSON(http.StatusOK, response)
}

// DeleteArticle handles request to delete an article
func (c *articleController) DeleteArticle(ctx *gin.Context) {
	// Parse article ID
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response := helper.BuildErrorResponse("Invalid ID format", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Delete article through service
	err = c.articleService.DeleteArticle(id)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to delete article", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Return response
	response := helper.BuildResponse(true, "Article deleted successfully", nil)
	ctx.JSON(http.StatusOK, response)
}

// SearchArticles handles request to search for articles
func (c *articleController) SearchArticles(ctx *gin.Context) {
	// Parse search parameters
	query := ctx.Query("q")
	if query == "" {
		response := helper.BuildErrorResponse("Search query is required", "Missing query parameter", nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	
	// Search articles through service
	articles, pagination, err := c.articleService.SearchArticles(query, page, limit)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to search articles", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	
	// Return response - combine articles and pagination into a single data object
	data := map[string]interface{}{
		"articles":   articles,
		"pagination": pagination,
	}
	response := helper.BuildResponse(true, "Articles searched successfully", data)
	ctx.JSON(http.StatusOK, response)
}