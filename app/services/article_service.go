package services

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/requests/admin"
	"goravel/app/models"
)

type ArticleService interface {
	GetByID(id uint) (*models.Article, error)
	GetList(filters ArticleFilters, page, pageSize int) ([]models.Article, int64, error)

	GetAllArticleForExport(filters ArticleFilters) ([]models.Article, error)

	Create(req *admin.ArticleCreate) (*models.Article, error)

	Update(id uint, req *admin.ArticleUpdate) (*models.Article, error)

	Delete(id uint) error
}

type ArticleFilters struct {
	AdminId        string
	Title          string
	Content        string
	Status         string
	CreatedAt      string
	CreatedAtStart string
	CreatedAtEnd   string
	UpdatedAt      string
	UpdatedAtStart string
	UpdatedAtEnd   string
}

// BuildArticleFiltersFromHTTP is shared by list/export and reads filters from GET query or POST body.
func BuildArticleFiltersFromHTTP(ctx http.Context) ArticleFilters {
	return ArticleFilters{
		AdminId:        ctx.Request().Input("admin_id", ctx.Request().Query("admin_id", "")),
		Title:          ctx.Request().Input("title", ctx.Request().Query("title", "")),
		Content:        ctx.Request().Input("content", ctx.Request().Query("content", "")),
		Status:         ctx.Request().Input("status", ctx.Request().Query("status", "")),
		CreatedAt:      ctx.Request().Input("created_at", ctx.Request().Query("created_at", "")),
		CreatedAtStart: ctx.Request().Input("created_at_start", ctx.Request().Query("created_at_start", "")),
		CreatedAtEnd:   ctx.Request().Input("created_at_end", ctx.Request().Query("created_at_end", "")),
		UpdatedAt:      ctx.Request().Input("updated_at", ctx.Request().Query("updated_at", "")),
		UpdatedAtStart: ctx.Request().Input("updated_at_start", ctx.Request().Query("updated_at_start", "")),
		UpdatedAtEnd:   ctx.Request().Input("updated_at_end", ctx.Request().Query("updated_at_end", "")),
	}
}

type ArticleServiceImpl struct{}

func NewArticleService() ArticleService {
	return &ArticleServiceImpl{}
}

func (s *ArticleServiceImpl) withRelations(query orm.Query) orm.Query {
	query = query.With("Admin")
	return query
}

// BuildArticleQuery builds the Article query shared by list/export.
func BuildArticleQuery(filters ArticleFilters) orm.Query {
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
	if filters.CreatedAtStart != "" {
		query = query.Where("created_at >= ?", filters.CreatedAtStart)
	}
	if filters.CreatedAtEnd != "" {
		query = query.Where("created_at <= ?", filters.CreatedAtEnd)
	}
	if filters.CreatedAt != "" {
		query = query.Where("created_at = ?", filters.CreatedAt)
	}
	if filters.UpdatedAtStart != "" {
		query = query.Where("updated_at >= ?", filters.UpdatedAtStart)
	}
	if filters.UpdatedAtEnd != "" {
		query = query.Where("updated_at <= ?", filters.UpdatedAtEnd)
	}
	if filters.UpdatedAt != "" {
		query = query.Where("updated_at = ?", filters.UpdatedAt)
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
	query := s.withRelations(BuildArticleQuery(filters))

	var list []models.Article
	var total int64
	if err := query.Order("id desc").Paginate(page, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *ArticleServiceImpl) GetAllArticleForExport(filters ArticleFilters) ([]models.Article, error) {
	query := s.withRelations(BuildArticleQuery(filters))

	var list []models.Article
	if err := query.Order("id desc").Find(&list); err != nil {
		return nil, apperrors.ErrQueryFailed.WithError(err)
	}

	return list, nil
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
	if _, err := s.GetByID(id); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Where("id", id).Delete(&models.Article{}); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
