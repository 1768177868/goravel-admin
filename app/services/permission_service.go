package services

import (
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/models"
)

type PermissionService interface {
	// GetByID 根据ID获取权限
	GetByID(id uint, withMenu bool) (*models.Permission, error)
	// GetList 获取权限列表
	GetList(filters PermissionFilters, page, pageSize int) ([]models.Permission, int64, error)
	// Create 创建权限
	Create(name, slug, method, path, description string, status uint8, sort int, menuID uint) (*models.Permission, error)
	// Update 更新权限
	Update(permission *models.Permission) error
	// Delete 删除权限
	Delete(permission *models.Permission) error
}

// PermissionFilters 权限查询过滤器
type PermissionFilters struct {
	Name      string
	Slug      string
	Method    string
	Path      string
	Status    string
	MenuID    string
	StartTime string
	EndTime   string
	OrderBy   string
}

type PermissionServiceImpl struct {
	treeService TreeService
}

func NewPermissionService() PermissionService {
	return &PermissionServiceImpl{
		treeService: NewTreeServiceImpl(),
	}
}

// GetByID 根据ID获取权限
func (s *PermissionServiceImpl) GetByID(id uint, withMenu bool) (*models.Permission, error) {
	var permission models.Permission
	query := facades.Orm().Query().Where("id", id)

	// 预加载关联
	if withMenu {
		query = query.With("Menu")
	}

	if err := query.FirstOrFail(&permission); err != nil {
		return nil, apperrors.ErrPermissionNotFound.WithError(err)
	}

	return &permission, nil
}

// GetList 获取权限列表
func (s *PermissionServiceImpl) GetList(filters PermissionFilters, page, pageSize int) ([]models.Permission, int64, error) {
	query := facades.Orm().Query().Model(&models.Permission{})

	// 应用筛选条件
	if filters.Name != "" {
		query = query.Where("name LIKE ?", "%"+filters.Name+"%")
	}
	if filters.Slug != "" {
		query = query.Where("slug LIKE ?", "%"+filters.Slug+"%")
	}
	if filters.Path != "" {
		query = query.Where("path LIKE ?", "%"+filters.Path+"%")
	}
	if filters.Method != "" {
		query = query.Where("method", filters.Method)
	}
	if filters.Status != "" {
		query = query.Where("status", filters.Status)
	}
	if filters.MenuID != "" {
		// 获取菜单及其所有子菜单的ID列表
		menuIDUint := cast.ToUint(filters.MenuID)
		if menuIDUint > 0 {
			menuIDs, err := s.treeService.GetMenuChildrenIDs(menuIDUint)
			if err == nil && len(menuIDs) > 0 {
				idsAny := helpers.ConvertUintSliceToAny(menuIDs)
				query = query.WhereIn("menu_id", idsAny)
			} else {
				// 如果获取菜单ID失败，返回空查询
				query = query.Where("1 = 0")
			}
		}
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
	var permissions []models.Permission
	var total int64
	if err := query.With("Menu").Paginate(page, pageSize, &permissions, &total); err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

// Create 创建权限
func (s *PermissionServiceImpl) Create(name, slug, method, path, description string, status uint8, sort int, menuID uint) (*models.Permission, error) {
	permission := &models.Permission{}
	createData := map[string]any{
		"name":        name,
		"slug":        slug,
		"method":      method,
		"path":        path,
		"description": description,
		"status":      status,
		"sort":        sort,
		"menu_id":     menuID,
	}

	if err := facades.Orm().Query().Model(permission).Create(createData); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	return permission, nil
}

// Update 更新权限
func (s *PermissionServiceImpl) Update(permission *models.Permission) error {
	if err := facades.Orm().Query().Save(permission); err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}
	return nil
}

// Delete 删除权限
func (s *PermissionServiceImpl) Delete(permission *models.Permission) error {
	if _, err := facades.Orm().Query().Delete(permission); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
