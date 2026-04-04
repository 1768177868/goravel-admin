package services

import (
	"github.com/dromara/carbon/v2"
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/models"
)

type PositionService interface {
	GetByID(id uint) (*models.Position, error)
	GetList(filters PositionFilters, page, pageSize int) ([]models.Position, int64, error)
	HasAdmins(positionID uint) (bool, error)
	Create(name, code, remark string, status uint8, sort int) (*models.Position, error)
	Update(position *models.Position) error
	Delete(position *models.Position) error
}

type PositionFilters struct {
	Name      string
	Status    string
	StartTime string
	EndTime   string
	OrderBy   string
}

type PositionServiceImpl struct{}

func NewPositionService() PositionService {
	return &PositionServiceImpl{}
}

func (s *PositionServiceImpl) GetByID(id uint) (*models.Position, error) {
	var position models.Position
	if err := facades.Orm().Query().Where("id", id).First(&position); err != nil {
		return nil, apperrors.ErrPositionNotFound.WithError(err)
	}
	return &position, nil
}

func (s *PositionServiceImpl) GetList(filters PositionFilters, page, pageSize int) ([]models.Position, int64, error) {
	query := facades.Orm().Query().Model(&models.Position{})
	if filters.Name != "" {
		query = query.Where("name LIKE ?", "%"+filters.Name+"%")
	}
	if filters.Status != "" {
		query = query.Where("status", filters.Status)
	}
	if filters.StartTime != "" {
		query = query.Where("created_at >= ?", filters.StartTime)
	}
	if filters.EndTime != "" {
		query = query.Where("created_at <= ?", filters.EndTime)
	}
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "sort:asc,id:asc"
	}
	query = helpers.ApplySort(query, orderBy, "sort:asc,id:asc")
	var list []models.Position
	var total int64
	if err := query.Paginate(page, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *PositionServiceImpl) HasAdmins(positionID uint) (bool, error) {
	count, err := facades.Orm().Query().Model(&models.Admin{}).Where("position_id", positionID).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *PositionServiceImpl) Create(name, code, remark string, status uint8, sort int) (*models.Position, error) {
	position := &models.Position{}
	createData := map[string]any{
		"name":       name,
		"code":       code,
		"remark":     remark,
		"status":     status,
		"sort":       sort,
		"created_at": carbon.Now(),
		"updated_at": carbon.Now(),
	}
	if err := facades.Orm().Query().Model(position).Create(createData); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}
	return position, nil
}

func (s *PositionServiceImpl) Update(position *models.Position) error {
	if err := facades.Orm().Query().Save(position); err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}
	return nil
}

func (s *PositionServiceImpl) Delete(position *models.Position) error {
	if _, err := facades.Orm().Query().Delete(position); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
