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

	// 创建开发者管理员（受保护，不显示在列表中）
	developerPassword, _ := facades.Hash().Make("developer123")
	developerAdmin := models.Admin{
		Username: "developer",
		Password: developerPassword,
		Nickname: "开发者管理员",
		Status:   1,
	}
	facades.Orm().Query().FirstOrCreate(&developerAdmin, models.Admin{Username: "developer"})

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

	// 关联超级管理员和超级角色
	facades.Orm().Query().Model(&superAdmin).Association("Roles").Replace([]models.Role{superRole})

	// 给开发者管理员分配 super-admin 角色
	facades.Orm().Query().Model(&developerAdmin).Association("Roles").Replace([]models.Role{superRole})

	// 关联超级角色和所有权限
	var allPerms []models.Permission
	facades.Orm().Query().Find(&allPerms)
	facades.Orm().Query().Model(&superRole).Association("Permissions").Replace(allPerms)

	// 关联超级角色和所有菜单
	var allMenus []models.Menu
	facades.Orm().Query().Find(&allMenus)
	facades.Orm().Query().Model(&superRole).Association("Menus").Replace(allMenus)

	return nil
}
