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

func Build<<.ModelName>>Query(filters <<.ModelName>>Filters) orm.Query {
	query := facades.Orm().Query().Model(&models.<<.ModelName>>{})
<<range .SearchableFields>>
	if filters.<<.PascalName>> != "" {
		<<if eq .SearchType "like">>
		query = query.Where("<<.Name>> LIKE ?", "%"+filters.<<.PascalName>>+"%")
		<<else if eq .SearchType "=">>
		query = query.Where("<<.Name>> = ?", filters.<<.PascalName>>)
		<<else if eq .SearchType ">" >>
		query = query.Where("<<.Name>> > ?", filters.<<.PascalName>>)
		<<else if eq .SearchType ">=" >>
		query = query.Where("<<.Name>> >= ?", filters.<<.PascalName>>)
		<<else if eq .SearchType "<" >>
		query = query.Where("<<.Name>> < ?", filters.<<.PascalName>>)
		<<else if eq .SearchType "<=" >>
		query = query.Where("<<.Name>> <= ?", filters.<<.PascalName>>)
		<<else if eq .SearchType "!=" >>
		query = query.Where("<<.Name>> != ?", filters.<<.PascalName>>)
		<<else if eq .SearchType "in">>
		query = query.Where("<<.Name>> IN ?", filters.<<.PascalName>>)
		<<else>>
		query = query.Where("<<.Name>> LIKE ?", "%"+filters.<<.PascalName>>+"%")
		<<end>>
	}
<<- end>>

	return query
}

func (s *<<.ServiceName>>Impl) GetByID(id uint) (*models.<<.ModelName>>, error) {
	var item models.<<.ModelName>>
	if err := facades.Orm().Query().Where("id", id).FirstOrFail(&item); err != nil {
		return nil, apperrors.NewBusinessError("<<.ModuleName>>_not_found", "<<.ModelName>> not found").WithError(err)
	}
	return &item, nil
}

func (s *<<.ServiceName>>Impl) GetList(filters <<.ModelName>>Filters, page, pageSize int) ([]models.<<.ModelName>>, int64, error) {
	query := Build<<.ModelName>>Query(filters)

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
<<range .FormFields>>
		<<.FieldName>>: req.<<.FieldName>>,
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

<<range .FormFields>>
	if req.<<.FieldName>> != nil {
		item.<<.FieldName>> = *req.<<.FieldName>>
	}
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
		return err
	}
	return nil
}
<<end>>
