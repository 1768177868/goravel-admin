package seeders

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type AdminSeeder struct {
}

func (s *AdminSeeder) Signature() string {
	return "AdminSeeder"
}

func (s *AdminSeeder) Run() error {
	// 创建超级管理员
	hashedPassword, _ := facades.Hash().Make("admin123")
	superAdmin := models.Admin{
		Username: "admin",
		Password: hashedPassword,
		Nickname: "超级管理员",
		Status:   1,
	}
	facades.Orm().Query().FirstOrCreate(&superAdmin, models.Admin{Username: "admin"})

	// 创建部门
	var rootDept models.Department
	facades.Orm().Query().FirstOrCreate(&rootDept, models.Department{
		Name:   "总公司",
		Code:   "ROOT",
		Status: 1,
		Sort:   0,
	})

	var itDept models.Department
	facades.Orm().Query().FirstOrCreate(&itDept, models.Department{
		ParentID: rootDept.ID,
		Name:     "技术部",
		Code:     "IT",
		Status:   1,
		Sort:     1,
	})

	// 创建角色
	var superRole models.Role
	facades.Orm().Query().FirstOrCreate(&superRole, models.Role{
		Name:        "超级管理员",
		Slug:        "super-admin",
		Description: "拥有所有权限",
		Status:      1,
		Sort:        0,
	})

	var adminRole models.Role
	facades.Orm().Query().FirstOrCreate(&adminRole, models.Role{
		Name:        "管理员",
		Slug:        "admin",
		Description: "普通管理员",
		Status:      1,
		Sort:        1,
	})

	// 创建权限
	permissions := []models.Permission{
		// 管理员管理
		{Name: "管理员列表", Slug: "admin.index", Method: "GET", Path: "/api/admin/admins", Description: "查看管理员列表", Status: 1, Sort: 1},
		{Name: "管理员详情", Slug: "admin.show", Method: "GET", Path: "/api/admin/admins/*", Description: "查看管理员详情", Status: 1, Sort: 2},
		{Name: "管理员创建", Slug: "admin.store", Method: "POST", Path: "/api/admin/admins", Description: "创建管理员", Status: 1, Sort: 3},
		{Name: "管理员更新", Slug: "admin.update", Method: "PUT", Path: "/api/admin/admins/*", Description: "更新管理员", Status: 1, Sort: 4},
		{Name: "管理员删除", Slug: "admin.destroy", Method: "DELETE", Path: "/api/admin/admins/*", Description: "删除管理员", Status: 1, Sort: 5},
		{Name: "重置密码", Slug: "admin.password", Method: "PUT", Path: "/api/admin/admins/*/password", Description: "重置管理员密码", Status: 1, Sort: 6},
		// 角色管理
		{Name: "角色列表", Slug: "role.index", Method: "GET", Path: "/api/admin/roles", Description: "查看角色列表", Status: 1, Sort: 7},
		{Name: "角色详情", Slug: "role.show", Method: "GET", Path: "/api/admin/roles/*", Description: "查看角色详情", Status: 1, Sort: 8},
		{Name: "角色创建", Slug: "role.store", Method: "POST", Path: "/api/admin/roles", Description: "创建角色", Status: 1, Sort: 9},
		{Name: "角色更新", Slug: "role.update", Method: "PUT", Path: "/api/admin/roles/*", Description: "更新角色", Status: 1, Sort: 10},
		{Name: "角色删除", Slug: "role.destroy", Method: "DELETE", Path: "/api/admin/roles/*", Description: "删除角色", Status: 1, Sort: 11},
		// 权限管理
		{Name: "权限列表", Slug: "permission.index", Method: "GET", Path: "/api/admin/permissions", Description: "查看权限列表", Status: 1, Sort: 12},
		{Name: "权限详情", Slug: "permission.show", Method: "GET", Path: "/api/admin/permissions/*", Description: "查看权限详情", Status: 1, Sort: 13},
		{Name: "权限创建", Slug: "permission.store", Method: "POST", Path: "/api/admin/permissions", Description: "创建权限", Status: 1, Sort: 14},
		{Name: "权限更新", Slug: "permission.update", Method: "PUT", Path: "/api/admin/permissions/*", Description: "更新权限", Status: 1, Sort: 15},
		{Name: "权限删除", Slug: "permission.destroy", Method: "DELETE", Path: "/api/admin/permissions/*", Description: "删除权限", Status: 1, Sort: 16},
		// 菜单管理
		{Name: "菜单列表", Slug: "menu.index", Method: "GET", Path: "/api/admin/menus", Description: "查看菜单列表", Status: 1, Sort: 17},
		{Name: "菜单详情", Slug: "menu.show", Method: "GET", Path: "/api/admin/menus/*", Description: "查看菜单详情", Status: 1, Sort: 18},
		{Name: "菜单创建", Slug: "menu.store", Method: "POST", Path: "/api/admin/menus", Description: "创建菜单", Status: 1, Sort: 19},
		{Name: "菜单更新", Slug: "menu.update", Method: "PUT", Path: "/api/admin/menus/*", Description: "更新菜单", Status: 1, Sort: 20},
		{Name: "菜单删除", Slug: "menu.destroy", Method: "DELETE", Path: "/api/admin/menus/*", Description: "删除菜单", Status: 1, Sort: 21},
		// 部门管理
		{Name: "部门列表", Slug: "department.index", Method: "GET", Path: "/api/admin/departments", Description: "查看部门列表", Status: 1, Sort: 22},
		{Name: "部门详情", Slug: "department.show", Method: "GET", Path: "/api/admin/departments/*", Description: "查看部门详情", Status: 1, Sort: 23},
		{Name: "部门创建", Slug: "department.store", Method: "POST", Path: "/api/admin/departments", Description: "创建部门", Status: 1, Sort: 24},
		{Name: "部门更新", Slug: "department.update", Method: "PUT", Path: "/api/admin/departments/*", Description: "更新部门", Status: 1, Sort: 25},
		{Name: "部门删除", Slug: "department.destroy", Method: "DELETE", Path: "/api/admin/departments/*", Description: "删除部门", Status: 1, Sort: 26},
		// 字典管理
		{Name: "字典列表", Slug: "dictionary.index", Method: "GET", Path: "/api/admin/dictionaries", Description: "查看字典列表", Status: 1, Sort: 27},
		{Name: "字典详情", Slug: "dictionary.show", Method: "GET", Path: "/api/admin/dictionaries/*", Description: "查看字典详情", Status: 1, Sort: 28},
		{Name: "字典创建", Slug: "dictionary.store", Method: "POST", Path: "/api/admin/dictionaries", Description: "创建字典", Status: 1, Sort: 29},
		{Name: "字典更新", Slug: "dictionary.update", Method: "PUT", Path: "/api/admin/dictionaries/*", Description: "更新字典", Status: 1, Sort: 30},
		{Name: "字典删除", Slug: "dictionary.destroy", Method: "DELETE", Path: "/api/admin/dictionaries/*", Description: "删除字典", Status: 1, Sort: 31},
		{Name: "字典查询", Slug: "dictionary.type", Method: "GET", Path: "/api/admin/dictionaries/type/*", Description: "根据类型查询字典", Status: 1, Sort: 32},
		// 操作日志
		{Name: "操作日志列表", Slug: "operation_log.index", Method: "GET", Path: "/api/admin/operation-logs", Description: "查看操作日志列表", Status: 1, Sort: 33},
		{Name: "操作日志详情", Slug: "operation_log.show", Method: "GET", Path: "/api/admin/operation-logs/*", Description: "查看操作日志详情", Status: 1, Sort: 34},
		{Name: "操作日志删除", Slug: "operation_log.destroy", Method: "DELETE", Path: "/api/admin/operation-logs/*", Description: "删除操作日志", Status: 1, Sort: 35},
		{Name: "操作日志清理", Slug: "operation_log.clean", Method: "POST", Path: "/api/admin/operation-logs/clean", Description: "清理操作日志", Status: 1, Sort: 36},
		// 登录日志
		{Name: "登录日志列表", Slug: "login_log.index", Method: "GET", Path: "/api/admin/login-logs", Description: "查看登录日志列表", Status: 1, Sort: 37},
		{Name: "登录日志详情", Slug: "login_log.show", Method: "GET", Path: "/api/admin/login-logs/*", Description: "查看登录日志详情", Status: 1, Sort: 38},
		{Name: "登录日志删除", Slug: "login_log.destroy", Method: "DELETE", Path: "/api/admin/login-logs/*", Description: "删除登录日志", Status: 1, Sort: 39},
		{Name: "登录日志清理", Slug: "login_log.clean", Method: "POST", Path: "/api/admin/login-logs/clean", Description: "清理登录日志", Status: 1, Sort: 40},
		// 系统日志
		{Name: "系统日志列表", Slug: "system_log.index", Method: "GET", Path: "/api/admin/system-logs", Description: "查看系统日志列表", Status: 1, Sort: 41},
		{Name: "系统日志详情", Slug: "system_log.show", Method: "GET", Path: "/api/admin/system-logs/*", Description: "查看系统日志详情", Status: 1, Sort: 42},
		{Name: "系统日志删除", Slug: "system_log.destroy", Method: "DELETE", Path: "/api/admin/system-logs/*", Description: "删除系统日志", Status: 1, Sort: 43},
		{Name: "系统日志清理", Slug: "system_log.clean", Method: "POST", Path: "/api/admin/system-logs/clean", Description: "清理系统日志", Status: 1, Sort: 44},
		// 密码管理
		{Name: "修改密码", Slug: "password.update", Method: "PUT", Path: "/api/admin/password", Description: "修改当前登录管理员密码", Status: 1, Sort: 45},
		// Token管理
		{Name: "Token列表", Slug: "token.index", Method: "GET", Path: "/api/admin/tokens", Description: "查看当前用户的token列表", Status: 1, Sort: 46},
		{Name: "删除Token", Slug: "token.destroy", Method: "DELETE", Path: "/api/admin/tokens/*", Description: "删除指定的token", Status: 1, Sort: 47},
		{Name: "删除所有Token", Slug: "token.destroy_all", Method: "DELETE", Path: "/api/admin/tokens", Description: "删除当前用户的所有token", Status: 1, Sort: 48},
		{Name: "踢出用户", Slug: "admin.kick_out", Method: "DELETE", Path: "/api/admin/admins/*/tokens", Description: "踢出指定用户的所有token", Status: 1, Sort: 49},
	}

	for _, perm := range permissions {
		facades.Orm().Query().FirstOrCreate(&perm, models.Permission{Slug: perm.Slug})
	}

	// 创建菜单
	var systemMenu models.Menu
	facades.Orm().Query().FirstOrCreate(&systemMenu, models.Menu{
		ParentID:  0,
		Title:     "系统管理",
		Icon:      "Setting",
		Path:      "/system",
		Component: "Layout",
		Type:      1,
		Status:    1,
		Sort:      1,
		IsHidden:  0,
	})

	var adminMenu models.Menu
	facades.Orm().Query().FirstOrCreate(&adminMenu, models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "管理员管理",
		Icon:      "User",
		Path:      "/system/admin",
		Component: "system/admin/index",
		Type:      2,
		Status:    1,
		Sort:      1,
		IsHidden:  0,
	})

	var roleMenu models.Menu
	facades.Orm().Query().FirstOrCreate(&roleMenu, models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "角色管理",
		Icon:      "UserFilled",
		Path:      "/system/role",
		Component: "system/role/index",
		Type:      2,
		Status:    1,
		Sort:      2,
		IsHidden:  0,
	})

	// 关联超级管理员和超级角色
	facades.Orm().Query().Model(&superAdmin).Association("Roles").Replace([]models.Role{superRole})

	// 关联超级角色和所有权限
	var allPerms []models.Permission
	facades.Orm().Query().Find(&allPerms)
	facades.Orm().Query().Model(&superRole).Association("Permissions").Replace(allPerms)

	// 关联超级角色和所有菜单
	var allMenus []models.Menu
	facades.Orm().Query().Find(&allMenus)
	facades.Orm().Query().Model(&superRole).Association("Menus").Replace(allMenus)

	// 创建字典数据
	dictionaries := []models.Dictionary{
		{Type: "status", Label: "启用", Value: "1", Description: "启用状态", Status: 1, Sort: 1},
		{Type: "status", Label: "禁用", Value: "0", Description: "禁用状态", Status: 1, Sort: 2},
		{Type: "menu_type", Label: "目录", Value: "1", Description: "目录类型", Status: 1, Sort: 1},
		{Type: "menu_type", Label: "菜单", Value: "2", Description: "菜单类型", Status: 1, Sort: 2},
		{Type: "menu_type", Label: "按钮", Value: "3", Description: "按钮类型", Status: 1, Sort: 3},
	}

	for _, dict := range dictionaries {
		facades.Orm().Query().FirstOrCreate(&dict, models.Dictionary{
			Type:  dict.Type,
			Value: dict.Value,
		})
	}

	return nil
}
