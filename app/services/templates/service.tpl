package services

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/requests/admin"
	"goravel/app/models"
)

type <<.ServiceName>> interface {
	GetByID(id uint) (*models.<<.ModelName>>, error)
	GetList(filters <<.ModelName>>Filters, page, pageSize int) ([]models.<<.ModelName>>, int64, error)
<<if .HasExport>>
	GetAll<<.ModelName>>ForExport(filters <<.ModelName>>Filters) ([]models.<<.ModelName>>, error)
<<end>>
<<if .HasCreate>>
	Create(req *admin.<<.RequestCreateName>>) (*models.<<.ModelName>>, error)
<<end>>
<<if .HasEdit>>
	Update(id uint, req *admin.<<.RequestUpdateName>>) (*models.<<.ModelName>>, error)
<<end>>
<<if .HasDelete>>
	Delete(id uint) error
<<end>>
}

type <<.ModelName>>Filters struct {
<<range .SearchableFields>>
	<<.PascalName>> string
	<<- if or (eq .SearchUIType "daterange") (eq .SearchUIType "datetimerange")>>
	<<.PascalName>>Start string
	<<.PascalName>>End   string
	<<- end>>
<<- end>>
}

// Build<<.ModelName>>FiltersFromHTTP 与列表/导出共用：从 GET query 或 POST body 读取筛选字段
func Build<<.ModelName>>FiltersFromHTTP(ctx http.Context) <<.ModelName>>Filters {
	return <<.ModelName>>Filters{
<<- range .SearchableFields>>
		<<.PascalName>>: ctx.Request().Input("<<.Name>>", ctx.Request().Query("<<.Name>>", "")),
		<<- if or (eq .SearchUIType "daterange") (eq .SearchUIType "datetimerange")>>
		<<.PascalName>>Start: ctx.Request().Input("<<.Name>>_start", ctx.Request().Query("<<.Name>>_start", "")),
		<<.PascalName>>End: ctx.Request().Input("<<.Name>>_end", ctx.Request().Query("<<.Name>>_end", "")),
		<<- end>>
<<- end>>
	}
}

type <<.ServiceName>>Impl struct{}

func New<<.ServiceName>>() <<.ServiceName>> {
	return &<<.ServiceName>>Impl{}
}

func (s *<<.ServiceName>>Impl) withRelations(query orm.Query) orm.Query {
<<- range .FormFields>>
<<- if .Relation>>
	query = query.With("<<.Relation.Name>>")
<<- end>>
<<- end>>
	return query
}

// Build<<.ModelName>>Query 构建<<.ModelName>>查询（列表与导出共用）
func Build<<.ModelName>>Query(filters <<.ModelName>>Filters) orm.Query {
	query := facades.Orm().Query().Model(&models.<<.ModelName>>{})
<<- range .SearchableFields>>
	<<- if or (eq .SearchUIType "daterange") (eq .SearchUIType "datetimerange")>>
	if filters.<<.PascalName>>Start != "" {
		query = query.Where("<<.Name>> >= ?", filters.<<.PascalName>>Start)
	}
	if filters.<<.PascalName>>End != "" {
		query = query.Where("<<.Name>> <= ?", filters.<<.PascalName>>End)
	}
	<<- end>>
	if filters.<<.PascalName>> != "" {
		<<- if eq .SearchType "like">>
		query = query.Where("<<.Name>> LIKE ?", "%"+filters.<<.PascalName>>+"%")
		<<- else if eq .SearchType "=">>
		query = query.Where("<<.Name>> = ?", filters.<<.PascalName>>)
		<<- else if eq .SearchType ">" >>
		query = query.Where("<<.Name>> > ?", filters.<<.PascalName>>)
		<<- else if eq .SearchType ">=" >>
		query = query.Where("<<.Name>> >= ?", filters.<<.PascalName>>)
		<<- else if eq .SearchType "<" >>
		query = query.Where("<<.Name>> < ?", filters.<<.PascalName>>)
		<<- else if eq .SearchType "<=" >>
		query = query.Where("<<.Name>> <= ?", filters.<<.PascalName>>)
		<<- else if eq .SearchType "!=" >>
		query = query.Where("<<.Name>> != ?", filters.<<.PascalName>>)
		<<- else if eq .SearchType "in">>
		query = query.Where("<<.Name>> IN ?", filters.<<.PascalName>>)
		<<- else>>
		query = query.Where("<<.Name>> LIKE ?", "%"+filters.<<.PascalName>>+"%")
		<<- end>>
	}
<<- end>>

	return query
}

func (s *<<.ServiceName>>Impl) GetByID(id uint) (*models.<<.ModelName>>, error) {
	var item models.<<.ModelName>>
	query := s.withRelations(facades.Orm().Query().Model(&models.<<.ModelName>>{})).Where("id", id)
	if err := query.FirstOrFail(&item); err != nil {
		return nil, apperrors.ErrRecordNotFound.WithError(err)
	}
	return &item, nil
}

func (s *<<.ServiceName>>Impl) GetList(filters <<.ModelName>>Filters, page, pageSize int) ([]models.<<.ModelName>>, int64, error) {
	query := s.withRelations(Build<<.ModelName>>Query(filters))

	var list []models.<<.ModelName>>
	var total int64
	if err := query.Order("id desc").Paginate(page, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

<<if .HasExport>>
func (s *<<.ServiceName>>Impl) GetAll<<.ModelName>>ForExport(filters <<.ModelName>>Filters) ([]models.<<.ModelName>>, error) {
	query := s.withRelations(Build<<.ModelName>>Query(filters))

	var list []models.<<.ModelName>>
	if err := query.Order("id desc").Find(&list); err != nil {
		return nil, apperrors.ErrQueryFailed.WithError(err)
	}

	return list, nil
}
<<end>>

<<if .HasCreate>>
func (s *<<.ServiceName>>Impl) Create(req *admin.<<.RequestCreateName>>) (*models.<<.ModelName>>, error) {
	item := &models.<<.ModelName>>{
<<- range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
	<<- if and .Relation (eq .Relation.RelationType "belongsTo")>>
		<<.FieldName>>: uint(req.<<.FieldName>>),
	<<- else>>
		<<.FieldName>>: req.<<.FieldName>>,
	<<- end>>
<<- end>>
<<- end>>
	}

	if err := facades.Orm().Query().Create(item); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	return item, nil
}
<<end>>

<<if .HasEdit>>
func (s *<<.ServiceName>>Impl) Update(id uint, req *admin.<<.RequestUpdateName>>) (*models.<<.ModelName>>, error) {
	item, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

<<- range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
	if req.<<.FieldName>> != nil {
	<<- if and .Relation (eq .Relation.RelationType "belongsTo")>>
		item.<<.FieldName>> = uint(*req.<<.FieldName>>)
	<<- else>>
		item.<<.FieldName>> = *req.<<.FieldName>>
	<<- end>>
	}
<<- end>>
<<- end>>

	if err := facades.Orm().Query().Save(item); err != nil {
		return nil, apperrors.ErrUpdateFailed.WithError(err)
	}

	return item, nil
}
<<end>>

<<if .HasDelete>>
func (s *<<.ServiceName>>Impl) Delete(id uint) error {
	if _, err := s.GetByID(id); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Where("id", id).Delete(&models.<<.ModelName>>{}); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
<<end>>
