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
	AdminId string
	Title   string
	Content string
	Status  string
}

type ArticleServiceImpl struct{}

func NewArticleService() ArticleService {
	return &ArticleServiceImpl{}
}

func (s *ArticleServiceImpl) withRelations(query orm.Query) orm.Query {
	query = query.With("Admin")
	return query
}

func (s *ArticleServiceImpl) buildArticleQuery(filters ArticleFilters) orm.Query {
	query := facades.Orm().Query().Model(&models.Article{})
	if filters.AdminId != "" {
		query = query.Where("admin_id = ?", filters.AdminId)
	}
	if filters.Title != "" {
		query = query.Where("title LIKE ?", "%"+filters.Title+"%")
	}
	if filters.Content != "" {
		query = query.Where("content LIKE ?", "%"+filters.Content+"%")
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	return query
}

func (s *ArticleServiceImpl) GetByID(id uint) (*models.Article, error) {
	var item models.Article
	query := s.withRelations(facades.Orm().Query().Model(&models.Article{})).Where("id", id)
	if err := query.FirstOrFail(&item); err != nil {
		return nil, apperrors.ErrRecordNotFound.WithError(err)
	}
	return &item, nil
}

func (s *ArticleServiceImpl) GetList(filters ArticleFilters, page, pageSize int) ([]models.Article, int64, error) {
	query := s.withRelations(s.buildArticleQuery(filters))

	var list []models.Article
	var total int64
	if err := query.Order("id desc").Paginate(page, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *ArticleServiceImpl) Create(req *admin.ArticleCreate) (*models.Article, error) {
	item := &models.Article{
		AdminId: uint(req.AdminId),
		Title:   req.Title,
		Content: req.Content,
		Status:  req.Status,
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
	if req.AdminId != nil {
		item.AdminId = uint(*req.AdminId)
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

	if err := facades.Orm().Query().Save(item); err != nil {
		return nil, apperrors.ErrUpdateFailed.WithError(err)
	}

	return item, nil
}

func (s *ArticleServiceImpl) Delete(id uint) error {
	if _, err := facades.Orm().Query().Where("id", id).Delete(&models.Article{}); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
