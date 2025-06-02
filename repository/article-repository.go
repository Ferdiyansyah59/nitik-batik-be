package repository

import (
	"batik/entity"

	"gorm.io/gorm"
)

// ArticleRepository interface represents the article repository contract
type ArticleRepository interface {
	GetAllArticles(page, limit int, search string) ([]entity.Article, int64, error)
	GetLatestArticles() ([]entity.Article, error)
	GetArticleByID(id uint64) (entity.Article, error)
	GetArticleBySlug(slug string) (entity.Article, error)
	CreateArticle(article entity.Article) (entity.Article, error)
	UpdateArticle(article entity.Article) (entity.Article, error)
	DeleteArticle(id uint64) error
	SearchArticles(query string, page, limit int) ([]entity.Article, int64, error)
	SlugExists(slug string) bool
}

// articleRepository is the implementation of ArticleRepository interface
type articleRepository struct {
	db *gorm.DB
}

// NewArticleRepository creates a new instance of ArticleRepository
func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepository{
		db: db,
	}
}

// GetAllArticles retrieves all articles with pagination
func (r *articleRepository) GetAllArticles(page, limit int, search string) ([]entity.Article, int64, error) {
	var articles []entity.Article
	var total int64

	query := r.db.Model(&entity.Article{})
	
	// Apply search filter if provided
	if search != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	
	return articles, total, nil
}

// Get Latest Articles
func (r *articleRepository) GetLatestArticles() ([]entity.Article, error) {
	var articles []entity.Article

	if err := r.db.Limit(5).Order("created_at desc").Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

// GetArticleByID retrieves an article by its ID
func (r *articleRepository) GetArticleByID(id uint64) (entity.Article, error) {
	var article entity.Article
	
	if err := r.db.Where("id = ?", id).First(&article).Error; err != nil {
		return entity.Article{}, err
	}
	
	return article, nil
}

// GetArticleBySlug retrieves an article by its slug
func (r *articleRepository) GetArticleBySlug(slug string) (entity.Article, error) {
	var article entity.Article
	
	if err := r.db.Where("slug = ?", slug).First(&article).Error; err != nil {
		return entity.Article{}, err
	}
	
	return article, nil
}

// CreateArticle adds a new article
func (r *articleRepository) CreateArticle(article entity.Article) (entity.Article, error) {
	if err := r.db.Create(&article).Error; err != nil {
		return entity.Article{}, err
	}
	
	return article, nil
}

// UpdateArticle updates an existing article
func (r *articleRepository) UpdateArticle(article entity.Article) (entity.Article, error) {
	if err := r.db.Save(&article).Error; err != nil {
		return entity.Article{}, err
	}
	
	return article, nil
}

// DeleteArticle removes an article
func (r *articleRepository) DeleteArticle(id uint64) error {
	return r.db.Delete(&entity.Article{}, id).Error
}

// SearchArticles searches for articles by query
func (r *articleRepository) SearchArticles(query string, page, limit int) ([]entity.Article, int64, error) {
	var articles []entity.Article
	var total int64
	
	// Search in title, description, and excerpt
	searchQuery := r.db.Model(&entity.Article{}).
		Where("title LIKE ? OR description LIKE ? OR excerpt LIKE ?", 
			"%"+query+"%", "%"+query+"%", "%"+query+"%")
	
	// Count total matching records
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination
	offset := (page - 1) * limit
	if err := searchQuery.Offset(offset).Limit(limit).Order("created_at DESC").Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	
	return articles, total, nil
}

// SlugExists checks if a slug already exists
func (r *articleRepository) SlugExists(slug string) bool {
	var count int64
	r.db.Model(&entity.Article{}).Where("slug = ?", slug).Count(&count)
	return count > 0
}



// type ArticleRepository interface {
// 	GetAllArticle() []entity.Article
// 	GetArticleByKey(title string) []entity.Article
// }

// type articleConnection struct {
// 	connection *gorm.DB
// }

// func NewArticleRepository(db *gorm.DB) ArticleRepository {
// 	return &articleConnection{
// 		connection: db,
// 	}
// }

// func (db *articleConnection) GetAllArticle() []entity.Article {
// 	var articles []entity.Article
// 	db.connection.Last(&articles)
// 	return articles
// }

// func (db *articleConnection) GetArticleByKey(title string) []entity.Article {
// 	var articles []entity.Article
// 	db.connection.Where("title LIKE ?", "%"+title+"%").Find(&articles)
// 	return articles
// }
