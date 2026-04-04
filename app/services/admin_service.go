package services

import (
	"slices"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"
	"github.com/samber/lo"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/models"
)

type AdminService interface {
	// GetByID 根据ID获取管理员
	GetByID(id uint, withDepartment bool, withRoles bool) (*models.Admin, error)
	// GetList 获取管理员列表
	GetList(filters AdminFilters, page, pageSize int) ([]models.Admin, int64, error)
	// GetAllAdminsForExport 获取所有管理员用于导出（不分页）
	GetAllAdminsForExport(filters AdminFilters) ([]models.Admin, error)
	// LoadRelations 加载管理员的关联数据（部门、角色）
	LoadRelations(admin *models.Admin) error
	// LoadRelationsWithPermissions 加载管理员的关联数据（包括权限和菜单）
	LoadRelationsWithPermissions(admin *models.Admin) error
	// LoadRelationsForList 批量加载管理员的关联数据
	LoadRelationsForList(admins []models.Admin) error
	// SyncRoles 同步管理员角色关联
	SyncRoles(admin *models.Admin, roleIDs []uint) error
	// GetProtectedAdminIDs 获取所有受保护的管理员ID
	GetProtectedAdminIDs() map[uint]bool
	// GetDepartmentAndChildrenIDs 获取部门及其子部门ID
	GetDepartmentAndChildrenIDs(departmentID uint) []uint
	// Update 更新管理员
	Update(admin *models.Admin) error
}

// AdminFilters 管理员查询过滤器
type AdminFilters struct {
	Username     string
	Status       string
	RoleID       string
	DepartmentID string
	PositionID   string
	Is2FABound   string
	StartTime    string
	EndTime      string
	OrderBy      string
}

type AdminServiceImpl struct {
}

func NewAdminServiceImpl() *AdminServiceImpl {
	return &AdminServiceImpl{}
}

// GetByID 根据ID获取管理员
func (s *AdminServiceImpl) GetByID(id uint, withDepartment bool, withRoles bool) (*models.Admin, error) {
	var admin models.Admin
	query := facades.Orm().Query().Where("id", id)

	// 预加载关联
	if withDepartment {
		query = query.With("Department").With("Position")
	}
	if withRoles {
		query = query.With("Roles")
	}

	if err := query.First(&admin); err != nil {
		return nil, apperrors.ErrAdminNotFound.WithError(err)
	}

	return &admin, nil
}

// buildQuery 构建查询（公共方法，用于列表和导出）
func (s *AdminServiceImpl) buildQuery(filters AdminFilters) orm.Query {
	query := facades.Orm().Query().Model(&models.Admin{})

	// 排除受保护的管理员
	protectedIDs := s.GetProtectedAdminIDs()
	if len(protectedIDs) > 0 {
		var ids []uint
		for id := range protectedIDs {
			ids = append(ids, id)
		}
		if len(ids) > 0 {
			query = query.Where("id NOT IN ?", ids)
		}
	}

	// 应用筛选条件
	if filters.Username != "" {
		query = query.Where("username LIKE ?", "%"+filters.Username+"%")
	}
	if filters.Status != "" {
		query = query.Where("status", filters.Status)
	}
	if filters.RoleID != "" {
		roleIDUint := cast.ToUint(filters.RoleID)
		if roleIDUint > 0 {
			query = query.Where("id IN (SELECT admin_id FROM admin_role WHERE role_id = ?)", roleIDUint)
		}
	}
	if filters.DepartmentID != "" {
		departmentIDUint := cast.ToUint(filters.DepartmentID)
		if departmentIDUint > 0 {
			departmentIDs := s.GetDepartmentAndChildrenIDs(departmentIDUint)
			if len(departmentIDs) > 0 {
				idsAny := make([]any, len(departmentIDs))
				for i, id := range departmentIDs {
					idsAny[i] = id
				}
				query = query.WhereIn("department_id", idsAny)
			} else {
				// 如果部门不存在，返回空结果
				query = query.Where("1 = 0")
			}
		}
	}
	if filters.PositionID != "" {
		if pid := cast.ToUint(filters.PositionID); pid > 0 {
			query = query.Where("position_id", pid)
		}
	}
	if filters.Is2FABound != "" {
		switch filters.Is2FABound {
		case "1":
			// 已绑定：google_secret IS NOT NULL AND google_secret != ''
			query = query.Where("google_secret IS NOT NULL AND google_secret != ?", "")
		case "0":
			// 未绑定：google_secret IS NULL OR google_secret = ''
			query = query.Where("(google_secret IS NULL OR google_secret = ?)", "")
		}
	}
	if filters.StartTime != "" {
		query = query.Where("created_at >= ?", filters.StartTime)
	}
	if filters.EndTime != "" {
		query = query.Where("created_at <= ?", filters.EndTime)
	}

	return query
}

// GetList 获取管理员列表
func (s *AdminServiceImpl) GetList(filters AdminFilters, page, pageSize int) ([]models.Admin, int64, error) {
	query := s.buildQuery(filters)

	// 应用排序
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "created_at:desc"
	}
	query = helpers.ApplySort(query, orderBy, "created_at:desc")

	// 分页查询
	var admins []models.Admin
	var total int64
	if err := query.With("Department").With("Position").With("Roles").Paginate(page, pageSize, &admins, &total); err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

// GetAllAdminsForExport 获取所有管理员用于导出（不分页）
func (s *AdminServiceImpl) GetAllAdminsForExport(filters AdminFilters) ([]models.Admin, error) {
	query := s.buildQuery(filters)

	// 应用排序
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "created_at:desc"
	}
	query = helpers.ApplySort(query, orderBy, "created_at:desc")

	// 不分页，获取所有数据
	var admins []models.Admin
	if err := query.With("Department").With("Position").With("Roles").Find(&admins); err != nil {
		return nil, apperrors.ErrQueryFailed.WithError(err)
	}

	return admins, nil
}

// Update 更新管理员
func (s *AdminServiceImpl) Update(admin *models.Admin) error {
	if err := facades.Orm().Query().Save(admin); err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}
	return nil
}

// GetProtectedAdminIDs 获取所有受保护的管理员ID
func (s *AdminServiceImpl) GetProtectedAdminIDs() map[uint]bool {
	allProtectedIDs := make(map[uint]bool)
	// 添加超级管理员ID（从配置读取，默认1）
	superAdminID := cast.ToUint(facades.Config().GetInt("admin.super_admin_id", 1))
	allProtectedIDs[superAdminID] = true
	// 添加开发者管理员ID
	developerIDsStr := facades.Config().GetString("admin.developer_ids", "2")
	developerIDs := s.parseProtectedIDs(developerIDsStr)
	for _, did := range developerIDs {
		allProtectedIDs[did] = true
	}
	return allProtectedIDs
}

// parseProtectedIDs 解析受保护的管理员ID字符串（支持逗号分隔）
func (s *AdminServiceImpl) parseProtectedIDs(idsStr string) []uint {
	var ids []uint
	if idsStr == "" {
		return ids
	}

	// 使用字符串分割
	parts := str.Of(idsStr).Split(",")
	for _, part := range parts {
		part = str.Of(part).Trim().String()
		if !str.Of(part).IsEmpty() {
			if id := cast.ToUint(part); id > 0 {
				ids = append(ids, id)
			}
		}
	}

	return ids
}

// GetDepartmentAndChildrenIDs 获取部门及其子部门ID
func (s *AdminServiceImpl) GetDepartmentAndChildrenIDs(departmentID uint) []uint {
	var departmentIDs []uint
	departmentIDs = append(departmentIDs, departmentID)
	s.getChildrenDepartmentIDs(departmentID, &departmentIDs)
	return departmentIDs
}

// getChildrenDepartmentIDs 递归获取子部门ID
func (s *AdminServiceImpl) getChildrenDepartmentIDs(parentID uint, departmentIDs *[]uint) {
	var children []models.Department
	if err := facades.Orm().Query().Where("parent_id", parentID).Get(&children); err == nil {
		for _, child := range children {
			*departmentIDs = append(*departmentIDs, child.ID)
			s.getChildrenDepartmentIDs(child.ID, departmentIDs)
		}
	}
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
	if admin.PositionID > 0 {
		var position models.Position
		if err := facades.Orm().Query().Where("id", admin.PositionID).First(&position); err == nil {
			admin.Position = position
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
				// 预加载菜单关联，用于检查菜单状态
				if err := facades.Orm().Query().With("Menu").Where("id IN ?", permissionIDs).Find(&permissions); err == nil {
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
	var positionIDs []uint
	var adminIDs []uint
	for _, admin := range admins {
		if admin.DepartmentID > 0 {
			departmentIDs = append(departmentIDs, admin.DepartmentID)
		}
		if admin.PositionID > 0 {
			positionIDs = append(positionIDs, admin.PositionID)
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

	positionsMap := make(map[uint]models.Position)
	if len(positionIDs) > 0 {
		uniquePosIDs := make(map[uint]bool)
		var uniqueIDs []uint
		for _, id := range positionIDs {
			if !uniquePosIDs[id] {
				uniqueIDs = append(uniqueIDs, id)
				uniquePosIDs[id] = true
			}
		}
		var positions []models.Position
		if err := facades.Orm().Query().Where("id IN ?", uniqueIDs).Find(&positions); err != nil {
			return err
		}
		for _, pos := range positions {
			positionsMap[pos.ID] = pos
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

	// 按管理员ID分组角色关联
	adminRoleMap := lo.GroupBy(adminRoles, func(ar AdminRole) uint {
		return ar.AdminID
	})

	// 提取所有角色ID并去重
	allRoleIDs := lo.Map(adminRoles, func(ar AdminRole, _ int) uint {
		return ar.RoleID
	})
	roleIDs := lo.Uniq(allRoleIDs)

	// 批量查询所有角色
	rolesMap := make(map[uint]models.Role)
	if len(roleIDs) > 0 {
		var roles []models.Role
		if err := facades.Orm().Query().Where("id IN ?", roleIDs).Find(&roles); err != nil {
			return err
		}
		rolesMap = lo.SliceToMap(roles, func(role models.Role) (uint, models.Role) {
			return role.ID, role
		})
	}

	// 填充关联数据
	for i := range admins {
		// 填充部门
		if admins[i].DepartmentID > 0 {
			if dept, ok := departmentsMap[admins[i].DepartmentID]; ok {
				admins[i].Department = dept
			}
		}
		if admins[i].PositionID > 0 {
			if pos, ok := positionsMap[admins[i].PositionID]; ok {
				admins[i].Position = pos
			}
		}

		// 填充角色（去重）
		if adminRoleList, ok := adminRoleMap[admins[i].ID]; ok {
			roleIDs := lo.Uniq(lo.Map(adminRoleList, func(ar AdminRole, _ int) uint {
				return ar.RoleID
			}))
			admins[i].Roles = lo.FilterMap(roleIDs, func(roleID uint, _ int) (models.Role, bool) {
				role, ok := rolesMap[roleID]
				return role, ok
			})
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
	deduplicatedRoleIDs := lo.Uniq(roleIDs)

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
		roleMap := lo.SliceToMap(roles, func(role models.Role) (uint, models.Role) {
			return role.ID, role
		})

		// 将去重后的角色转换为切片
		deduplicatedRoles := lo.Values(roleMap)

		// Replace 方法会先删除所有现有关联，然后添加新的关联
		if err := facades.Orm().Query().Model(admin).Association("Roles").Replace(deduplicatedRoles); err != nil {
			return err
		}
	}

	return nil
}
