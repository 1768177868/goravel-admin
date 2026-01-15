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

	Delete(id uint) error
}

type ArticleFilters struct {
	Name string

	Status string
}

type ArticleServiceImpl struct{}

func NewArticleService() ArticleService {
	return &ArticleServiceImpl{}
}

func BuildArticleQuery(filters ArticleFilters) orm.Query {
	query := facades.Orm().Query().Model(&models.Article{})

	if filters.Name != "" {

		query = query.Where("name LIKE ?", "%"+filters.Name+"%")

	}

	if filters.Status != "" {

		query = query.Where("status = ?", filters.Status)

	}

	return query
}

func (s *ArticleServiceImpl) GetByID(id uint) (*models.Article, error) {
	var item models.Article
	if err := facades.Orm().Query().Where("id", id).FirstOrFail(&item); err != nil {
		return nil, apperrors.NewBusinessError("article_not_found", "Article not found").WithError(err)
	}
	return &item, nil
}

func (s *ArticleServiceImpl) GetList(filters ArticleFilters, page, pageSize int) ([]models.Article, int64, error) {
	query := BuildArticleQuery(filters)

	var list []models.Article
	var total int64
	if err := query.Order("id desc").Paginate(page, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *ArticleServiceImpl) Create(req *admin.ArticleCreate) (*models.Article, error) {
	item := &models.Article{

		Name: req.Name,

		Status: req.Status,
	}

	if err := facades.Orm().Query().Create(item); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	return item, nil
}

func (s *ArticleServiceImpl) Delete(id uint) error {
	_, err := facades.Orm().Query().Where("id", id).Delete(&models.Article{})
	return err
}
