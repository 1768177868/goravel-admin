package services

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/requests/admin"
	"goravel/app/models"
)

type ArticleService interface {
	GetByID(id uint) (*models.Article, error)
	GetList(filters ArticleFilters, page, pageSize int) ([]models.Article, int64, error)

	Create(req *admin.ArticleCreate) (*models.Article, error)

	Update(id uint, req *admin.ArticleUpdate) (*models.Article, error)

	Delete(id uint) error
}

type ArticleFilters struct {
	Title     string
	Content   string
	Status    string
	AdminId   string
	CreatedAt string
	UpdatedAt string
}

type ArticleServiceImpl struct{}

func NewArticleService() ArticleService {
	return &ArticleServiceImpl{}
}

func (s *ArticleServiceImpl) BuildArticleQuery(filters ArticleFilters) orm.Query {
	query := facades.Orm().Query().Model(&models.Article{})

	if filters.Title != "" {

		query = query.Where("title LIKE ?", "%"+filters.Title+"%")

	}
	if filters.Content != "" {

		query = query.Where("content = ?", filters.Content)

	}
	if filters.Status != "" {

		query = query.Where("status = ?", filters.Status)

	}
	if filters.AdminId != "" {

		query = query.Where("admin_id = ?", filters.AdminId)

	}
	if filters.CreatedAt != "" {

		query = query.Where("created_at = ?", filters.CreatedAt)

	}
	if filters.UpdatedAt != "" {

		query = query.Where("updated_at = ?", filters.UpdatedAt)

	}

	return query
}

func (s *ArticleServiceImpl) GetByID(id uint) (*models.Article, error) {
	var item models.Article
	query := facades.Orm().Query().Where("id", id)

	query = query.With("Admin")

	if err := query.FirstOrFail(&item); err != nil {
		return nil, apperrors.NewBusinessError("article_not_found", "Article not found").WithError(err)
	}
	return &item, nil
}

func (s *ArticleServiceImpl) GetList(filters ArticleFilters, page, pageSize int) ([]models.Article, int64, error) {
	query := s.BuildArticleQuery(filters)

	query = query.With("Admin")

	var list []models.Article
	var total int64
	if err := query.Order("id desc").Paginate(page, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *ArticleServiceImpl) Create(req *admin.ArticleCreate) (*models.Article, error) {
	item := &models.Article{

		Title: req.Title,

		Content: req.Content,

		Status: req.Status,

		AdminId: uint(req.AdminId),
	}

	if err := facades.Orm().Query().Create(item); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	return item, nil
}

func (s *ArticleServiceImpl) Update(id uint, req *admin.ArticleUpdate) (*models.Article, error) {
	item, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {

		item.Title = *req.Title

	}
	if req.Content != nil {

		item.Content = *req.Content

	}
	if req.Status != nil {

		item.Status = *req.Status

	}
	if req.AdminId != nil {

		item.AdminId = uint(*req.AdminId)

	}

	if err := facades.Orm().Query().Save(item); err != nil {
		return nil, apperrors.ErrUpdateFailed.WithError(err)
	}

	return item, nil
}

func (s *ArticleServiceImpl) Delete(id uint) error {
	if _, err := facades.Orm().Query().Where("id", id).Delete(&models.Article{}); err != nil {
		return err
	}
	return nil
}
