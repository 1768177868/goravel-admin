package services

import (
	"github.com/dromara/carbon/v2"
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
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
	// GetAllTypes 获取所有字典类型
	GetAllTypes() ([]string, error)
	// Create 创建字典
	Create(dictType, label, value, translationKey, description, remark string, status uint8, sort int) (*models.Dictionary, error)
	// Update 更新字典
	Update(dictionary *models.Dictionary) error
	// Delete 删除字典
	Delete(dictionary *models.Dictionary) error
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
	if err := facades.Orm().Query().Where("id", id).FirstOrFail(&dictionary); err != nil {
		return nil, apperrors.ErrDictionaryNotFound.WithError(err)
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

	// 分页查询
	var dictionaries []models.Dictionary
	var total int64
	if err := query.Paginate(page, pageSize, &dictionaries, &total); err != nil {
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

// GetAllTypes 获取所有字典类型
func (s *DictionaryServiceImpl) GetAllTypes() ([]string, error) {
	var types []string
	if err := facades.Orm().Query().Model(&models.Dictionary{}).Distinct("type").Pluck("type", &types); err != nil {
		return []string{}, nil
	}
	return types, nil
}

// Create 创建字典
func (s *DictionaryServiceImpl) Create(dictType, label, value, translationKey, description, remark string, status uint8, sort int) (*models.Dictionary, error) {
	dictionary := &models.Dictionary{}
	createData := map[string]any{
		"type":            dictType,
		"label":           label,
		"value":           value,
		"translation_key": translationKey,
		"description":     description,
		"remark":          remark,
		"status":          status,
		"sort":            sort,
		"created_at":      carbon.Now(),
		"updated_at":      carbon.Now(),
	}

	if err := facades.Orm().Query().Model(dictionary).Create(createData); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	return dictionary, nil
}

// Update 更新字典
func (s *DictionaryServiceImpl) Update(dictionary *models.Dictionary) error {
	if err := facades.Orm().Query().Save(dictionary); err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}
	return nil
}

// Delete 删除字典
func (s *DictionaryServiceImpl) Delete(dictionary *models.Dictionary) error {
	if _, err := facades.Orm().Query().Delete(dictionary); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
