package services

import (
	"fmt"

	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
	"goravel/app/models"
)

type DictionaryService interface {
	// GetByID 根据ID获取字典
	GetByID(id uint) (*models.Dictionary, error)
	// GetList 获取字典列表
	GetList(filters DictionaryFilters, page, pageSize int) ([]models.Dictionary, int64, error)
	// GetByType 根据类型获取字典列表
	GetByType(dictType string) ([]models.Dictionary, error)
}

// DictionaryFilters 字典查询过滤器
type DictionaryFilters struct {
	Type      string
	Status    string
	StartTime string
	EndTime   string
	OrderBy   string
}

type DictionaryServiceImpl struct {
}

func NewDictionaryService() DictionaryService {
	return &DictionaryServiceImpl{}
}

// GetByID 根据ID获取字典
func (s *DictionaryServiceImpl) GetByID(id uint) (*models.Dictionary, error) {
	var dictionary models.Dictionary
	if err := facades.Orm().Query().Where("id", id).First(&dictionary); err != nil {
		return nil, fmt.Errorf("字典不存在: %v", err)
	}
	return &dictionary, nil
}

// GetList 获取字典列表
func (s *DictionaryServiceImpl) GetList(filters DictionaryFilters, page, pageSize int) ([]models.Dictionary, int64, error) {
	query := facades.Orm().Query().Model(&models.Dictionary{})

	// 应用筛选条件
	if filters.Type != "" {
		query = query.Where("type LIKE ?", "%"+filters.Type+"%")
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
		orderBy = "sort:asc,id:desc"
	}
	query = helpers.ApplySort(query, orderBy, "sort:asc,id:desc")

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	var dictionaries []models.Dictionary
	err = query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&dictionaries)
	if err != nil {
		return nil, 0, err
	}

	return dictionaries, total, nil
}

// GetByType 根据类型获取字典列表
func (s *DictionaryServiceImpl) GetByType(dictType string) ([]models.Dictionary, error) {
	var dictionaries []models.Dictionary
	if err := facades.Orm().Query().
		Where("type", dictType).
		Where("status", 1).
		Order("sort asc, id asc").
		Find(&dictionaries); err != nil {
		return nil, err
	}
	return dictionaries, nil
}
