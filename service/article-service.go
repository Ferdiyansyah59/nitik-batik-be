package service

import (
	"batik/entity"
	"batik/repository"
	"batik/utils"
	"time"
)

// ArticleService interface represents the article service contract
type ArticleService interface {
	GetAllArticles(page, limit int, search string) ([]entity.Article, *utils.Pagination, error)
	GetLatestArticles() ([]entity.Article, error)
	GetArticleByID(id uint64) (entity.Article, error)
	GetArticleBySlug(slug string) (entity.Article, error)
	CreateArticle(article entity.Article) (entity.Article, error)
	UpdateArticle(id uint64, article entity.Article) (entity.Article, error)
	DeleteArticle(id uint64) error
	SearchArticles(query string, page, limit int) ([]entity.Article, *utils.Pagination, error)
}

// articleService is the implementation of ArticleService interface
type articleService struct {
	articleRepository repository.ArticleRepository
}

// NewArticleService creates a new instance of ArticleService
func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		articleRepository: repo,
	}
}

// GetAllArticles retrieves all articles with pagination
func (s *articleService) GetAllArticles(page, limit int, search string) ([]entity.Article, *utils.Pagination, error) {
	// Ensure valid pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	
	// Get articles from repository
	articles, total, err := s.articleRepository.GetAllArticles(page, limit, search)
	if err != nil {
		return nil, nil, err
	}
	
	// Create pagination data
	pagination := utils.NewPagination(page, limit, total)
	
	return articles, pagination, nil
}

// Get Latest Article
func (s *articleService) GetLatestArticles() ([]entity.Article, error) {
	articles, err := s.articleRepository.GetLatestArticles()
	if err != nil {
		return nil, err
	}
	return articles, nil
}

// GetArticleByID retrieves an article by its ID
func (s *articleService) GetArticleByID(id uint64) (entity.Article, error) {
	return s.articleRepository.GetArticleByID(id)
}

// GetArticleBySlug retrieves an article by its slug
func (s *articleService) GetArticleBySlug(slug string) (entity.Article, error) {
	return s.articleRepository.GetArticleBySlug(slug)
}

// CreateArticle adds a new article
func (s *articleService) CreateArticle(article entity.Article) (entity.Article, error) {
	// Generate slug from title
	baseSlug := utils.GenerateSlug(article.Title, "article")
	uniqueSlug := utils.EnsureUniqueSlug(baseSlug, s.articleRepository.SlugExists)
	article.Slug = uniqueSlug
	
	// Generate excerpt if not provided
	if article.Excerpt == "" {
		article.Excerpt = utils.GenerateExcerpt(article.Description, 150)
	}
	
	// Set timestamps
	now := time.Now()
	article.CreatedAt = now
	article.UpdatedAt = now
	
	return s.articleRepository.CreateArticle(article)
}

// UpdateArticle updates an existing article
func (s *articleService) UpdateArticle(id uint64, articleData entity.Article) (entity.Article, error) {
	// Get existing article
	existingArticle, err := s.articleRepository.GetArticleByID(id)
	if err != nil {
		return entity.Article{}, err
	}
	
	// Update fields
	existingArticle.Title = articleData.Title
	existingArticle.Description = articleData.Description
	
	// Update excerpt if provided, or generate from description
	if articleData.Excerpt != "" {
		existingArticle.Excerpt = articleData.Excerpt
	} else {
		existingArticle.Excerpt = utils.GenerateExcerpt(articleData.Description, 150)
	}
	
	// Update image URL if provided
	if articleData.ImageURL != "" {
		existingArticle.ImageURL = articleData.ImageURL
	}
	
	// Update slug if title changed
	if existingArticle.Title != articleData.Title {
		baseSlug := utils.GenerateSlug(articleData.Title, "article")
		// Check function to exclude the current article's slug
		checkSlugExists := func(slug string) bool {
			if slug == existingArticle.Slug {
				return false // Current slug is okay to use
			}
			return s.articleRepository.SlugExists(slug)
		}
		existingArticle.Slug = utils.EnsureUniqueSlug(baseSlug, checkSlugExists)
	}
	
	// Update timestamp
	existingArticle.UpdatedAt = time.Now()
	
	return s.articleRepository.UpdateArticle(existingArticle)
}

// DeleteArticle removes an article
func (s *articleService) DeleteArticle(id uint64) error {
	return s.articleRepository.DeleteArticle(id)
}

// SearchArticles searches for articles by query
func (s *articleService) SearchArticles(query string, page, limit int) ([]entity.Article, *utils.Pagination, error) {
	// Ensure valid pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	
	// Search articles
	articles, total, err := s.articleRepository.SearchArticles(query, page, limit)
	if err != nil {
		return nil, nil, err
	}
	
	// Create pagination data
	pagination := utils.NewPagination(page, limit, total)
	
	return articles, pagination, nil
}

// type ArticleService interface {
// 	GetAllArticle() []entity.Article
// 	GetArticleByKey(title string) []entity.Article
// }

// type articleService struct {
// 	articleRepository repository.ArticleRepository
// }

// func NewArticleService(artRepo repository.ArticleRepository) ArticleService {
// 	return &articleService{
// 		articleRepository: artRepo,
// 	}
// }

// func (serv *articleService) GetAllArticle() []entity.Article {
// 	return serv.articleRepository.GetAllArticle()
// }

// func (serv *articleService) GetArticleByKey(title string) []entity.Article {
// 	return serv.articleRepository.GetArticleByKey(title)
// }
