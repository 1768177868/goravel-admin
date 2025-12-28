package services

import (
	"fmt"

	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
	"goravel/app/models"
)

type BlacklistService interface {
	// GetByID 根据ID获取黑名单
	GetByID(id uint) (*models.Blacklist, error)
	// GetList 获取黑名单列表
	GetList(filters BlacklistFilters, page, pageSize int) ([]models.Blacklist, int64, error)
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
	if err := facades.Orm().Query().Where("id", id).First(&blacklist); err != nil {
		return nil, fmt.Errorf("黑名单不存在: %v", err)
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

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	var blacklists []models.Blacklist
	err = query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&blacklists)
	if err != nil {
		return nil, 0, err
	}

	return blacklists, total, nil
}
