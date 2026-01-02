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
	facades.Orm().Query().Where("username", "admin").FirstOrCreate(&superAdmin, superAdmin)

	// 创建开发者管理员（受保护，不显示在列表中）
	developerPassword, _ := facades.Hash().Make("developer123")
	developerAdmin := models.Admin{
		Username: "developer",
		Password: developerPassword,
		Nickname: "开发者管理员",
		Status:   1,
	}
	facades.Orm().Query().Where("username", "developer").FirstOrCreate(&developerAdmin, developerAdmin)

	// 创建部门
	rootDept := models.Department{
		Name:   "总公司",
		Code:   "ROOT",
		Status: 1,
		Sort:   0,
	}
	facades.Orm().Query().Where("code", "ROOT").FirstOrCreate(&rootDept, rootDept)

	itDept := models.Department{
		ParentID: rootDept.ID,
		Name:     "技术部",
		Code:     "IT",
		Status:   1,
		Sort:     1,
	}
	facades.Orm().Query().Where("code", "IT").FirstOrCreate(&itDept, itDept)

	// 创建角色
	superRole := models.Role{
		Name:        "超级管理员",
		Slug:        "super-admin",
		Description: "拥有所有权限",
		Status:      1,
		Sort:        0,
	}
	facades.Orm().Query().Where("slug", "super-admin").FirstOrCreate(&superRole, superRole)

	adminRole := models.Role{
		Name:        "管理员",
		Slug:        "admin",
		Description: "普通管理员",
		Status:      1,
		Sort:        1,
	}
	facades.Orm().Query().Where("slug", "admin").FirstOrCreate(&adminRole, adminRole)

	// 关联超级管理员和超级角色（增量添加，不覆盖已有角色）
	var existingRoles []models.Role
	facades.Orm().Query().Model(&superAdmin).Association("Roles").Find(&existingRoles)
	hasSuperRole := false
	for _, r := range existingRoles {
		if r.Slug == "super-admin" {
			hasSuperRole = true
			break
		}
	}
	if !hasSuperRole {
		facades.Orm().Query().Model(&superAdmin).Association("Roles").Append([]models.Role{superRole})
	}

	// 给开发者管理员分配 super-admin 角色（增量添加）
	existingRoles = []models.Role{}
	facades.Orm().Query().Model(&developerAdmin).Association("Roles").Find(&existingRoles)
	hasSuperRole = false
	for _, r := range existingRoles {
		if r.Slug == "super-admin" {
			hasSuperRole = true
			break
		}
	}
	if !hasSuperRole {
		facades.Orm().Query().Model(&developerAdmin).Association("Roles").Append([]models.Role{superRole})
	}

	// super-admin 角色不需要分配权限和菜单，因为它在权限中间件中会跳过权限检查
	// 在获取用户信息时，会特殊处理 super-admin 角色，返回所有菜单用于前端显示

	// 创建演示账户角色（只允许查看，不允许编辑创建删除）
	demoRole := models.Role{
		Name:        "演示账户",
		Slug:        "demo",
		Description: "演示账户，只允许查看，不允许编辑、创建、删除",
		Status:      1,
		Sort:        2,
	}
	facades.Orm().Query().Where("slug", "demo").FirstOrCreate(&demoRole, demoRole)

	// 给演示角色分配所有查看权限（index 和 show）（增量添加）
	var viewPermissions []models.Permission
	facades.Orm().Query().Where("slug LIKE ?", "%.index").OrWhere("slug LIKE ?", "%.show").OrWhere("slug", "dashboard.data").Find(&viewPermissions)
	if len(viewPermissions) > 0 {
		var existingPerms []models.Permission
		facades.Orm().Query().Model(&demoRole).Association("Permissions").Find(&existingPerms)
		existingPermMap := make(map[uint]bool)
		for _, p := range existingPerms {
			existingPermMap[p.ID] = true
		}
		var newPerms []models.Permission
		for _, p := range viewPermissions {
			if !existingPermMap[p.ID] {
				newPerms = append(newPerms, p)
			}
		}
		if len(newPerms) > 0 {
			facades.Orm().Query().Model(&demoRole).Association("Permissions").Append(newPerms)
		}
	}

	// 给演示角色分配所有菜单（用于前端显示）（增量添加）
	var allMenus []models.Menu
	facades.Orm().Query().Where("status", 1).Find(&allMenus)
	if len(allMenus) > 0 {
		var existingMenus []models.Menu
		facades.Orm().Query().Model(&demoRole).Association("Menus").Find(&existingMenus)
		existingMenuMap := make(map[uint]bool)
		for _, m := range existingMenus {
			existingMenuMap[m.ID] = true
		}
		var newMenus []models.Menu
		for _, m := range allMenus {
			if !existingMenuMap[m.ID] {
				newMenus = append(newMenus, m)
			}
		}
		if len(newMenus) > 0 {
			facades.Orm().Query().Model(&demoRole).Association("Menus").Append(newMenus)
		}
	}

	// 创建演示账户
	demoPassword, _ := facades.Hash().Make("demo123")
	demoAdmin := models.Admin{
		Username: "demo",
		Password: demoPassword,
		Nickname: "演示账户",
		Status:   1,
	}
	facades.Orm().Query().Where("username", "demo").FirstOrCreate(&demoAdmin, demoAdmin)

	// 给演示账户分配演示角色（增量添加）
	var existingDemoRoles []models.Role
	facades.Orm().Query().Model(&demoAdmin).Association("Roles").Find(&existingDemoRoles)
	hasDemoRole := false
	for _, r := range existingDemoRoles {
		if r.Slug == "demo" {
			hasDemoRole = true
			break
		}
	}
	if !hasDemoRole {
		facades.Orm().Query().Model(&demoAdmin).Association("Roles").Append([]models.Role{demoRole})
	}

	return nil
}
