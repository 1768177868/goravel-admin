package services

import (
	"context"
	appfacades "goravel/app/facades"
	"strconv"

	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/models"
)

type RoleService interface {
	// GetByID 根据ID获取角色
	GetByID(id uint, withRelations bool) (*models.Role, error)
	// GetList 获取角色列表
	GetList(filters RoleFilters, page, pageSize int) ([]models.Role, int64, error)
	// LoadRelations 加载角色的关联数据（权限、菜单）
	LoadRelations(role *models.Role) error
	// SyncPermissions 同步角色权限关联
	SyncPermissions(role *models.Role, permissionIDs []uint) error
	// SyncMenus 同步角色菜单关联
	SyncMenus(role *models.Role, menuIDs []uint) error
	// ParseIDsFromRequest 从请求中解析ID数组
	ParseIDsFromRequest(ctx http.Context, key string) []uint
	// Create 创建角色
	Create(name, slug, description string, status uint8, sort int) (*models.Role, error)
	// Update 更新角色
	Update(role *models.Role) error
	// Delete 删除角色
	Delete(role *models.Role) error
}

// RoleFilters 角色查询过滤器
type RoleFilters struct {
	Name      string
	Status    string
	StartTime string
	EndTime   string
	OrderBy   string
}

type RoleServiceImpl struct {
	ctx context.Context
}

func NewRoleServiceImpl(ctx context.Context) *RoleServiceImpl {
	return &RoleServiceImpl{
		ctx: ctx}
}

// GetByID 根据ID获取角色
func (s *RoleServiceImpl) GetByID(id uint, withRelations bool) (*models.Role, error) {
	var role models.Role
	query := appfacades.OrmQuery(s.ctx).Where("id", id)

	// 预加载关联
	if withRelations {
		query = query.With("Permissions").With("Menus")
	}

	if err := query.FirstOrFail(&role); err != nil {
		return nil, apperrors.ErrRoleNotFound.WithError(err)
	}

	return &role, nil
}

// GetList 获取角色列表
func (s *RoleServiceImpl) GetList(filters RoleFilters, page, pageSize int) ([]models.Role, int64, error) {
	query := appfacades.OrmQuery(s.ctx).Model(&models.Role{})

	// 应用筛选条件
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

	// 应用排序
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "sort:asc,created_at:desc"
	}
	query = helpers.ApplySort(query, orderBy, "sort:asc,created_at:desc")

	// 分页查询
	var roles []models.Role
	var total int64
	if err := query.Paginate(page, pageSize, &roles, &total); err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// LoadRelations 加载角色的关联数据（权限、菜单）
func (s *RoleServiceImpl) LoadRelations(role *models.Role) error {
	// 加载权限
	if err := appfacades.OrmQuery(s.ctx).Model(role).Association("Permissions").Find(&role.Permissions); err != nil {
		return err
	}

	// 加载菜单
	if err := appfacades.OrmQuery(s.ctx).Model(role).Association("Menus").Find(&role.Menus); err != nil {
		return err
	}

	return nil
}

// SyncPermissions 同步角色权限关联
func (s *RoleServiceImpl) SyncPermissions(role *models.Role, permissionIDs []uint) error {
	var permissions []models.Permission
	if len(permissionIDs) > 0 {
		if err := appfacades.OrmQuery(s.ctx).Where("id IN ?", permissionIDs).Find(&permissions); err != nil {
			return err
		}
	}
	return appfacades.OrmQuery(s.ctx).Model(role).Association("Permissions").Replace(permissions)
}

// SyncMenus 同步角色菜单关联
func (s *RoleServiceImpl) SyncMenus(role *models.Role, menuIDs []uint) error {
	var menus []models.Menu
	if len(menuIDs) > 0 {
		if err := appfacades.OrmQuery(s.ctx).Where("id IN ?", menuIDs).Find(&menus); err != nil {
			return err
		}
	}
	return appfacades.OrmQuery(s.ctx).Model(role).Association("Menus").Replace(menus)
}

// ParseIDsFromRequest 从请求中解析ID数组
func (s *RoleServiceImpl) ParseIDsFromRequest(ctx http.Context, key string) []uint {
	var ids []uint
	if idsStr := ctx.Request().Input(key); idsStr != "" {
		for _, idStr := range ctx.Request().InputArray(key) {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				ids = append(ids, uint(id))
			}
		}
	}
	return ids
}

// Create 创建角色
func (s *RoleServiceImpl) Create(name, slug, description string, status uint8, sort int) (*models.Role, error) {
	// 使用结构体创建，这样数据库生成的主键 ID 会回填到 role 上
	role := &models.Role{
		Name:        name,
		Slug:        slug,
		Description: description,
		Status:      status,
		Sort:        sort,
	}

	if err := appfacades.OrmQuery(s.ctx).Create(role); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	return role, nil
}

// Update 更新角色
func (s *RoleServiceImpl) Update(role *models.Role) error {
	if err := appfacades.OrmQuery(s.ctx).Save(role); err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}
	return nil
}

// Delete 删除角色
func (s *RoleServiceImpl) Delete(role *models.Role) error {
	if _, err := appfacades.OrmQuery(s.ctx).Delete(role); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
