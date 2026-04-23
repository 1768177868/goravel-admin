package services

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/requests/admin"
	"goravel/app/models"
)

type <<.ServiceName>> interface {
	GetByID(id uint) (*models.<<.ModelName>>, error)
	GetList(filters <<.ModelName>>Filters, page, pageSize int) ([]models.<<.ModelName>>, int64, error)
<<if .HasCreate>>
	Create(req *admin.<<.RequestCreateName>>) (*models.<<.ModelName>>, error)
<<end>>
<<if .HasEdit>>
	Update(id uint, req *admin.<<.RequestUpdateName>>) (*models.<<.ModelName>>, error)
<<end>>
<<if .HasDelete>>
	Delete(id uint) error
<<end>>
<<if .HasExport>>
	Export(filters <<.ModelName>>Filters) error
<<end>>
}

type <<.ModelName>>Filters struct {
<<range .SearchableFields>>
	<<.PascalName>> string
<<- end>>
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

func (s *<<.ServiceName>>Impl) build<<.ModelName>>Query(filters <<.ModelName>>Filters) orm.Query {
	query := facades.Orm().Query().Model(&models.<<.ModelName>>{})
<<- range .SearchableFields>>
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
	query := s.withRelations(s.build<<.ModelName>>Query(filters))

	var list []models.<<.ModelName>>
	var total int64
	if err := query.Order("id desc").Paginate(page, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

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
	if _, err := facades.Orm().Query().Where("id", id).Delete(&models.<<.ModelName>>{}); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
<<end>>

<<if .HasExport>>
func (s *<<.ServiceName>>Impl) Export(filters <<.ModelName>>Filters) error {
	_ = filters
	// TODO: implement export business logic (task dispatch / file generation).
	return nil
}
<<end>>
