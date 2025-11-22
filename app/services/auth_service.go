package services

import (
	"errors"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/trans"
	"goravel/app/models"
)

type AuthService interface {
	// Login 管理员登录
	Login(ctx http.Context, username, password string) (*models.Admin, string, error)
	// GetAdminInfo 获取管理员完整信息（包括权限和菜单）
	GetAdminInfo(ctx http.Context) (*models.Admin, []models.Permission, []models.Menu, error)
	// RecordLoginLog 记录登录日志
	RecordLoginLog(ctx http.Context, adminID uint, username string, status uint8, message string) error
}

type AuthServiceImpl struct {
	adminService AdminService
}

func NewAuthServiceImpl(adminService AdminService) *AuthServiceImpl {
	return &AuthServiceImpl{
		adminService: adminService,
	}
}

// Login 管理员登录
func (s *AuthServiceImpl) Login(ctx http.Context, username, password string) (*models.Admin, string, error) {
	var admin models.Admin
	if err := facades.Orm().Query().Where("username", username).First(&admin); err != nil {
		return nil, "", err
	}

	if admin.Status == 0 {
		return nil, "", errors.New("account_disabled")
	}

	// 验证密码
	if !facades.Hash().Check(password, admin.Password) {
		// 记录登录失败日志
		s.RecordLoginLog(ctx, 0, username, 0, trans.Get(ctx, "password_error"))
		return nil, "", errors.New("password_error")
	}

	// 生成JWT token
	token, err := facades.Auth(ctx).Guard("admin").Login(admin)
	if err != nil {
		return nil, "", err
	}

	// 记录登录成功日志
	s.RecordLoginLog(ctx, admin.ID, username, 1, trans.Get(ctx, "login_success"))

	// 更新最后登录时间（ORM会自动更新UpdatedAt）
	facades.Orm().Query().Save(&admin)

	return &admin, token, nil
}

// GetAdminInfo 获取管理员完整信息（包括权限和菜单）
func (s *AuthServiceImpl) GetAdminInfo(ctx http.Context) (*models.Admin, []models.Permission, []models.Menu, error) {
	var admin models.Admin
	if err := facades.Auth(ctx).Guard("admin").User(&admin); err != nil {
		return nil, nil, nil, err
	}

	// 加载基本关联
	if err := facades.Orm().Query().Load(&admin, "Department", "Roles"); err != nil {
		return nil, nil, nil, err
	}

	// 批量加载所有角色的权限和菜单，避免 N+1 查询
	if len(admin.Roles) > 0 {
		var roleIDs []uint
		for _, role := range admin.Roles {
			roleIDs = append(roleIDs, role.ID)
		}

		// 批量加载权限
		type RolePermission struct {
			RoleID       uint `gorm:"column:role_id"`
			PermissionID uint `gorm:"column:permission_id"`
		}
		var rolePermissions []RolePermission
		if err := facades.Orm().Query().Table("role_permission").Where("role_id IN ?", roleIDs).Find(&rolePermissions); err == nil {
			var permissionIDs []uint
			rolePermissionMap := make(map[uint][]uint)
			for _, rp := range rolePermissions {
				rolePermissionMap[rp.RoleID] = append(rolePermissionMap[rp.RoleID], rp.PermissionID)
				permissionIDs = append(permissionIDs, rp.PermissionID)
			}
			if len(permissionIDs) > 0 {
				var permissions []models.Permission
				if err := facades.Orm().Query().Where("id IN ?", permissionIDs).Find(&permissions); err == nil {
					permissionMap := make(map[uint]models.Permission)
					for _, perm := range permissions {
						permissionMap[perm.ID] = perm
					}
					for i := range admin.Roles {
						if permIDs, ok := rolePermissionMap[admin.Roles[i].ID]; ok {
							for _, permID := range permIDs {
								if perm, ok := permissionMap[permID]; ok {
									admin.Roles[i].Permissions = append(admin.Roles[i].Permissions, perm)
								}
							}
						}
					}
				}
			}
		}

		// 批量加载菜单
		type RoleMenu struct {
			RoleID uint `gorm:"column:role_id"`
			MenuID uint `gorm:"column:menu_id"`
		}
		var roleMenus []RoleMenu
		if err := facades.Orm().Query().Table("role_menu").Where("role_id IN ?", roleIDs).Find(&roleMenus); err == nil {
			var menuIDs []uint
			roleMenuMap := make(map[uint][]uint)
			for _, rm := range roleMenus {
				roleMenuMap[rm.RoleID] = append(roleMenuMap[rm.RoleID], rm.MenuID)
				menuIDs = append(menuIDs, rm.MenuID)
			}
			if len(menuIDs) > 0 {
				var menus []models.Menu
				if err := facades.Orm().Query().Where("id IN ?", menuIDs).Find(&menus); err == nil {
					menuMap := make(map[uint]models.Menu)
					for _, menu := range menus {
						menuMap[menu.ID] = menu
					}
					for i := range admin.Roles {
						if mIDs, ok := roleMenuMap[admin.Roles[i].ID]; ok {
							for _, menuID := range mIDs {
								if menu, ok := menuMap[menuID]; ok {
									admin.Roles[i].Menus = append(admin.Roles[i].Menus, menu)
								}
							}
						}
					}
				}
			}
		}
	}

	// 收集所有角色的权限和菜单（去重）
	permissionMap := make(map[uint]models.Permission)
	menuMap := make(map[uint]models.Menu)
	
	for _, role := range admin.Roles {
		for _, perm := range role.Permissions {
			permissionMap[perm.ID] = perm
		}
		for _, menu := range role.Menus {
			menuMap[menu.ID] = menu
		}
	}

	// 转换为切片
	var permissions []models.Permission
	var menus []models.Menu
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}
	for _, menu := range menuMap {
		menus = append(menus, menu)
	}

	return &admin, permissions, menus, nil
}

// RecordLoginLog 记录登录日志
func (s *AuthServiceImpl) RecordLoginLog(ctx http.Context, adminID uint, username string, status uint8, message string) error {
	loginLog := models.LoginLog{
		AdminID:   adminID,
		Username:  username,
		IP:        ctx.Request().Ip(),
		UserAgent: ctx.Request().Header("User-Agent", ""),
		Status:    status,
		Message:   message,
	}
	return facades.Orm().Query().Create(&loginLog)
}
