package services

import (
	"slices"

	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type AdminService interface {
	// LoadRelations 加载管理员的关联数据（部门、角色）
	LoadRelations(admin *models.Admin) error
	// LoadRelationsWithPermissions 加载管理员的关联数据（包括权限和菜单）
	LoadRelationsWithPermissions(admin *models.Admin) error
	// LoadRelationsForList 批量加载管理员的关联数据
	LoadRelationsForList(admins []models.Admin) error
	// SyncRoles 同步管理员角色关联
	SyncRoles(admin *models.Admin, roleIDs []uint) error
}

type AdminServiceImpl struct {
}

func NewAdminServiceImpl() *AdminServiceImpl {
	return &AdminServiceImpl{}
}

// LoadRelations 加载管理员的关联数据（部门、角色）
func (s *AdminServiceImpl) LoadRelations(admin *models.Admin) error {
	if admin == nil {
		return nil
	}

	// 加载部门
	if admin.DepartmentID > 0 {
		var department models.Department
		if err := facades.Orm().Query().Where("id", admin.DepartmentID).First(&department); err == nil {
			admin.Department = department
		}
	}

	// 加载角色关联
	type AdminRole struct {
		AdminID uint `gorm:"column:admin_id"`
		RoleID  uint `gorm:"column:role_id"`
	}
	var adminRoles []AdminRole
	if err := facades.Orm().Query().Table("admin_role").Where("admin_id", admin.ID).Find(&adminRoles); err != nil {
		return err
	}

	var roleIDs []uint
	for _, ar := range adminRoles {
		if !contains(roleIDs, ar.RoleID) {
			roleIDs = append(roleIDs, ar.RoleID)
		}
	}

	admin.Roles = nil
	if len(roleIDs) > 0 {
		var roles []models.Role
		if err := facades.Orm().Query().Where("id IN ?", roleIDs).Find(&roles); err != nil {
			return err
		}
		admin.Roles = roles
	}

	return nil
}

// LoadRelationsWithPermissions 加载管理员的关联数据（包括权限和菜单）
func (s *AdminServiceImpl) LoadRelationsWithPermissions(admin *models.Admin) error {
	// 先加载基本关联
	if err := s.LoadRelations(admin); err != nil {
		return err
	}

	// 批量加载所有角色的权限和菜单，避免 N+1 查询
	if len(admin.Roles) > 0 {
		for i := range admin.Roles {
			admin.Roles[i].Permissions = nil
			admin.Roles[i].Menus = nil
		}

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

	// 将权限和菜单存储到 admin 的扩展字段（如果需要）
	// 这里可以根据实际需求调整

	return nil
}

// LoadRelationsForList 批量加载管理员的关联数据（优化版，避免 N+1 查询）
func (s *AdminServiceImpl) LoadRelationsForList(admins []models.Admin) error {
	if len(admins) == 0 {
		return nil
	}

	// 收集所有需要查询的 ID
	var departmentIDs []uint
	var adminIDs []uint
	for _, admin := range admins {
		if admin.DepartmentID > 0 {
			departmentIDs = append(departmentIDs, admin.DepartmentID)
		}
		adminIDs = append(adminIDs, admin.ID)
	}

	// 批量查询部门
	departmentsMap := make(map[uint]models.Department)
	if len(departmentIDs) > 0 {
		var departments []models.Department
		// 去重
		uniqueDeptIDs := make(map[uint]bool)
		var uniqueIDs []uint
		for _, id := range departmentIDs {
			if !uniqueDeptIDs[id] {
				uniqueIDs = append(uniqueIDs, id)
				uniqueDeptIDs[id] = true
			}
		}
		if err := facades.Orm().Query().Where("id IN ?", uniqueIDs).Find(&departments); err != nil {
			return err
		}
		for _, dept := range departments {
			departmentsMap[dept.ID] = dept
		}
	}

	// 批量查询所有管理员的角色关联
	// 查询中间表获取 admin_id 和 role_id 的映射
	type AdminRole struct {
		AdminID uint `gorm:"column:admin_id"`
		RoleID  uint `gorm:"column:role_id"`
	}
	var adminRoles []AdminRole
	if err := facades.Orm().Query().Table("admin_role").Where("admin_id IN ?", adminIDs).Find(&adminRoles); err != nil {
		return err
	}

	// 收集所有角色 ID（去重）
	var roleIDs []uint
	roleIDSet := make(map[uint]bool)
	adminRoleMap := make(map[uint][]uint)        // admin_id -> []role_id
	adminRoleSet := make(map[uint]map[uint]bool) // admin_id -> map[role_id]bool 用于去重
	for _, ar := range adminRoles {
		// 初始化管理员的角色集合
		if adminRoleSet[ar.AdminID] == nil {
			adminRoleSet[ar.AdminID] = make(map[uint]bool)
		}
		// 如果该角色ID还未添加，则添加
		if !adminRoleSet[ar.AdminID][ar.RoleID] {
			adminRoleSet[ar.AdminID][ar.RoleID] = true
			adminRoleMap[ar.AdminID] = append(adminRoleMap[ar.AdminID], ar.RoleID)
		}
		// 收集所有角色ID（去重）
		if !roleIDSet[ar.RoleID] {
			roleIDSet[ar.RoleID] = true
			roleIDs = append(roleIDs, ar.RoleID)
		}
	}

	// 批量查询所有角色
	rolesMap := make(map[uint]models.Role)
	if len(roleIDs) > 0 {
		var roles []models.Role
		if err := facades.Orm().Query().Where("id IN ?", roleIDs).Find(&roles); err != nil {
			return err
		}
		for _, role := range roles {
			rolesMap[role.ID] = role
		}
	}

	// 填充关联数据
	for i := range admins {
		// 填充部门
		if admins[i].DepartmentID > 0 {
			if dept, ok := departmentsMap[admins[i].DepartmentID]; ok {
				admins[i].Department = dept
			}
		}

		// 填充角色（去重）
		if roleIDs, ok := adminRoleMap[admins[i].ID]; ok {
			addedRoles := make(map[uint]bool) // 用于去重
			for _, roleID := range roleIDs {
				if role, ok := rolesMap[roleID]; ok {
					if !addedRoles[roleID] {
						addedRoles[roleID] = true
						admins[i].Roles = append(admins[i].Roles, role)
					}
				}
			}
		}
	}

	return nil
}

// contains 辅助函数，检查切片中是否包含某个值
func contains(slice []uint, val uint) bool {
	return slices.Contains(slice, val)
}

// SyncRoles 同步管理员角色关联
func (s *AdminServiceImpl) SyncRoles(admin *models.Admin, roleIDs []uint) error {
	// 去重角色ID，避免重复数据
	uniqueRoleIDs := make(map[uint]bool)
	var deduplicatedRoleIDs []uint
	for _, roleID := range roleIDs {
		if !uniqueRoleIDs[roleID] {
			uniqueRoleIDs[roleID] = true
			deduplicatedRoleIDs = append(deduplicatedRoleIDs, roleID)
		}
	}

	// 先清空该管理员的所有角色关联（包括重复的），确保彻底清理重复数据
	if err := facades.Orm().Query().Model(admin).Association("Roles").Clear(); err != nil {
		return err
	}

	// 如果有角色需要关联，则添加去重后的角色关联
	if len(deduplicatedRoleIDs) > 0 {
		var roles []models.Role
		if err := facades.Orm().Query().Where("id IN ?", deduplicatedRoleIDs).Find(&roles); err != nil {
			return err
		}

		// 对查询到的角色再次去重（双重保险）
		roleMap := make(map[uint]models.Role)
		for _, role := range roles {
			if _, exists := roleMap[role.ID]; !exists {
				roleMap[role.ID] = role
			}
		}

		// 将去重后的角色转换为切片
		deduplicatedRoles := make([]models.Role, 0, len(roleMap))
		for _, role := range roleMap {
			deduplicatedRoles = append(deduplicatedRoles, role)
		}

		// Replace 方法会先删除所有现有关联，然后添加新的关联
		if err := facades.Orm().Query().Model(admin).Association("Roles").Replace(deduplicatedRoles); err != nil {
			return err
		}
	}

	return nil
}
