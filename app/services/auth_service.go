package services

import (
	"errors"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/trans"
	"goravel/app/models"
	"goravel/app/utils/logger"
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
	tokenService TokenService
}

func NewAuthServiceImpl(adminService AdminService, tokenService TokenService) *AuthServiceImpl {
	return &AuthServiceImpl{
		adminService: adminService,
		tokenService: tokenService,
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

	// 生成token并存入数据库（类似Laravel Sanctum）
	// 按配置的过期时间生成token，如果需要永久token，可以在创建token时设置 expiresAt 为 nil
	var expiresAt *time.Time
	ttl := facades.Config().GetInt("jwt.ttl", 60) // 默认60分钟
	if ttl > 0 {
		// 如果配置了过期时间，设置过期时间
		exp := time.Now().Add(time.Duration(ttl) * time.Minute)
		expiresAt = &exp
	}
	// 如果 ttl 为 0 或负数，expiresAt 为 nil，表示永不过期

	plainToken, _, err := s.tokenService.CreateToken("admin", admin.ID, "admin-token", expiresAt)
	if err != nil {
		return nil, "", err
	}
	token := plainToken

	// 记录登录成功日志
	s.RecordLoginLog(ctx, admin.ID, username, 1, trans.Get(ctx, "login_success"))

	// 更新最后登录时间（ORM会自动更新UpdatedAt）
	facades.Orm().Query().Save(&admin)

	return &admin, token, nil
}

// GetAdminInfo 获取管理员完整信息（包括权限和菜单）
func (s *AuthServiceImpl) GetAdminInfo(ctx http.Context) (*models.Admin, []models.Permission, []models.Menu, error) {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		logger.ErrorfHTTP(ctx, "GetAdminInfo: admin value is nil in context")
		return nil, nil, nil, errors.New("not_logged_in")
	}

	var admin models.Admin
	// 尝试值类型
	if adminVal, ok := adminValue.(models.Admin); ok {
		admin = adminVal
	} else if adminPtr, ok := adminValue.(*models.Admin); ok {
		// 尝试指针类型
		admin = *adminPtr
	} else {
		logger.ErrorfHTTP(ctx, "GetAdminInfo: admin value type assertion failed, type: %T, value: %+v", adminValue, adminValue)
		return nil, nil, nil, errors.New("not_logged_in")
	}
	
	facades.Log().Debugf("GetAdminInfo: admin found, ID: %d, Username: %s", admin.ID, admin.Username)

	// 重新查询admin并加载关联（避免使用已存在的admin对象，可能导致关联加载问题）
	var adminWithRelations models.Admin
	if err := facades.Orm().Query().With("Department").With("Roles").Where("id", admin.ID).First(&adminWithRelations); err != nil {
		logger.ErrorfHTTP(ctx, "GetAdminInfo: failed to load admin with relations, error: %v", err)
		return nil, nil, nil, err
	}
	admin = adminWithRelations

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
