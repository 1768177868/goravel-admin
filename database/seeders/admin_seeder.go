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
	var superAdmin models.Admin
	exists, _ := facades.Orm().Query().Model(&models.Admin{}).Where("username", "admin").Exists()
	if !exists {
		hashedPassword, _ := facades.Hash().Make("admin123")
		superAdmin = models.Admin{
			Username: "admin",
			Password: hashedPassword,
			Nickname: "超级管理员",
			Status:   1,
		}
		facades.Orm().Query().Create(&superAdmin)
	} else {
		facades.Orm().Query().Where("username", "admin").First(&superAdmin)
	}

	// 创建开发者管理员（受保护，不显示在列表中）
	var developerAdmin models.Admin
	exists, _ = facades.Orm().Query().Model(&models.Admin{}).Where("username", "developer").Exists()
	if !exists {
		developerPassword, _ := facades.Hash().Make("developer123")
		developerAdmin = models.Admin{
			Username: "developer",
			Password: developerPassword,
			Nickname: "开发者管理员",
			Status:   1,
		}
		facades.Orm().Query().Create(&developerAdmin)
	} else {
		facades.Orm().Query().Where("username", "developer").First(&developerAdmin)
	}

	// 创建部门
	var rootDept models.Department
	exists, _ = facades.Orm().Query().Model(&models.Department{}).Where("code", "ROOT").Exists()
	if !exists {
		rootDept = models.Department{
			Name:   "总公司",
			Code:   "ROOT",
			Status: 1,
			Sort:   0,
		}
		facades.Orm().Query().Create(&rootDept)
	} else {
		facades.Orm().Query().Where("code", "ROOT").First(&rootDept)
	}

	var itDept models.Department
	exists, _ = facades.Orm().Query().Model(&models.Department{}).Where("code", "IT").Exists()
	if !exists {
		itDept = models.Department{
			ParentID: rootDept.ID,
			Name:     "技术部",
			Code:     "IT",
			Status:   1,
			Sort:     1,
		}
		facades.Orm().Query().Create(&itDept)
	} else {
		facades.Orm().Query().Where("code", "IT").First(&itDept)
	}

	// 创建角色
	var superRole models.Role
	exists, _ = facades.Orm().Query().Model(&models.Role{}).Where("slug", "super-admin").Exists()
	if !exists {
		superRole = models.Role{
			Name:        "超级管理员",
			Slug:        "super-admin",
			Description: "拥有所有权限",
			Status:      1,
			Sort:        0,
		}
		facades.Orm().Query().Create(&superRole)
	} else {
		facades.Orm().Query().Where("slug", "super-admin").First(&superRole)
	}

	var adminRole models.Role
	exists, _ = facades.Orm().Query().Model(&models.Role{}).Where("slug", "admin").Exists()
	if !exists {
		adminRole = models.Role{
			Name:        "管理员",
			Slug:        "admin",
			Description: "普通管理员",
			Status:      1,
			Sort:        1,
		}
		facades.Orm().Query().Create(&adminRole)
	} else {
		facades.Orm().Query().Where("slug", "admin").First(&adminRole)
	}

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

	return nil
}
