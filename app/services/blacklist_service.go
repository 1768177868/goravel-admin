package services

import (
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/models"
)

type BlacklistService interface {
	// GetByID 根据ID获取黑名单
	GetByID(id uint) (*models.Blacklist, error)
	// GetList 获取黑名单列表
	GetList(filters BlacklistFilters, page, pageSize int) ([]models.Blacklist, int64, error)
	// Create 创建黑名单
	Create(ip, remark string, status uint8) (*models.Blacklist, error)
	// Update 更新黑名单
	Update(blacklist *models.Blacklist) error
	// Delete 删除黑名单
	Delete(blacklist *models.Blacklist) error
}

// BlacklistFilters 黑名单查询过滤器
type BlacklistFilters struct {
	IP        string
	Status    string
	StartTime string
	EndTime   string
	OrderBy   string
}

type BlacklistServiceImpl struct {
}

func NewBlacklistService() BlacklistService {
	return &BlacklistServiceImpl{}
}

// GetByID 根据ID获取黑名单
func (s *BlacklistServiceImpl) GetByID(id uint) (*models.Blacklist, error) {
	var blacklist models.Blacklist
	if err := facades.Orm().Query().Where("id", id).FirstOrFail(&blacklist); err != nil {
		return nil, apperrors.ErrBlacklistNotFound.WithError(err)
	}
	return &blacklist, nil
}

// GetList 获取黑名单列表
func (s *BlacklistServiceImpl) GetList(filters BlacklistFilters, page, pageSize int) ([]models.Blacklist, int64, error) {
	query := facades.Orm().Query().Model(&models.Blacklist{})

	// 应用筛选条件
	if filters.IP != "" {
		query = query.Where("ip LIKE ?", "%"+filters.IP+"%")
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

	// 应用排序
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "id:desc"
	}
	query = helpers.ApplySort(query, orderBy, "id:desc")

	// 分页查询
	var blacklists []models.Blacklist
	var total int64
	if err := query.Paginate(page, pageSize, &blacklists, &total); err != nil {
		return nil, 0, err
	}

	return blacklists, total, nil
}

// Create 创建黑名单
func (s *BlacklistServiceImpl) Create(ip, remark string, status uint8) (*models.Blacklist, error) {
	blacklist := &models.Blacklist{}
	createData := map[string]any{
		"ip":     ip,
		"remark": remark,
		"status": status,
	}

	if err := facades.Orm().Query().Model(blacklist).Create(createData); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	return blacklist, nil
}

// Update 更新黑名单
func (s *BlacklistServiceImpl) Update(blacklist *models.Blacklist) error {
	if err := facades.Orm().Query().Save(blacklist); err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}
	return nil
}

// Delete 删除黑名单
func (s *BlacklistServiceImpl) Delete(blacklist *models.Blacklist) error {
	if _, err := facades.Orm().Query().Delete(blacklist); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
