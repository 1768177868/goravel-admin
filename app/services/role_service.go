package services

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type RoleService interface {
	// LoadRelations 加载角色的关联数据（权限、菜单）
	LoadRelations(role *models.Role) error
	// SyncPermissions 同步角色权限关联
	SyncPermissions(role *models.Role, permissionIDs []uint) error
	// SyncMenus 同步角色菜单关联
	SyncMenus(role *models.Role, menuIDs []uint) error
	// ParseIDsFromRequest 从请求中解析ID数组
	ParseIDsFromRequest(ctx http.Context, key string) []uint
}

type RoleServiceImpl struct {
}

func NewRoleServiceImpl() *RoleServiceImpl {
	return &RoleServiceImpl{}
}

// LoadRelations 加载角色的关联数据（权限、菜单）
func (s *RoleServiceImpl) LoadRelations(role *models.Role) error {
	// 加载权限
	if err := facades.Orm().Query().Model(role).Association("Permissions").Find(&role.Permissions); err != nil {
		return err
	}

	// 加载菜单
	if err := facades.Orm().Query().Model(role).Association("Menus").Find(&role.Menus); err != nil {
		return err
	}

	return nil
}

// SyncPermissions 同步角色权限关联
func (s *RoleServiceImpl) SyncPermissions(role *models.Role, permissionIDs []uint) error {
	var permissions []models.Permission
	if len(permissionIDs) > 0 {
		if err := facades.Orm().Query().Where("id IN ?", permissionIDs).Find(&permissions); err != nil {
			return err
		}
	}
	return facades.Orm().Query().Model(role).Association("Permissions").Replace(permissions)
}

// SyncMenus 同步角色菜单关联
func (s *RoleServiceImpl) SyncMenus(role *models.Role, menuIDs []uint) error {
	var menus []models.Menu
	if len(menuIDs) > 0 {
		if err := facades.Orm().Query().Where("id IN ?", menuIDs).Find(&menus); err != nil {
			return err
		}
	}
	return facades.Orm().Query().Model(role).Association("Menus").Replace(menus)
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

