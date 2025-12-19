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
	// FirstOrCreate: 第一个参数是查询条件（只包含唯一字段），第二个参数是创建时的默认值
	if err := facades.Orm().Query().Where("name", "超级管理员").First(&superRole); err != nil {
		// 不存在则创建
		superRole = models.Role{
			Name:        "超级管理员",
			Slug:        "super-admin",
			Description: "拥有所有权限",
			Status:      1,
			Sort:        0,
		}
		facades.Orm().Query().Create(&superRole)
	} else {
		// 存在则更新其他字段（如果为空或需要更新）
		update := false
		if superRole.Slug == "" {
			superRole.Slug = "super-admin"
			update = true
		}
		if superRole.Description == "" {
			superRole.Description = "拥有所有权限"
			update = true
		}
		if superRole.Status == 0 {
			superRole.Status = 1
			update = true
		}
		if superRole.Sort == 0 {
			superRole.Sort = 0
			update = true
		}
		if update {
			facades.Orm().Query().Save(&superRole)
		}
	}

	var adminRole models.Role
	if err := facades.Orm().Query().Where("name", "管理员").First(&adminRole); err != nil {
		// 不存在则创建
		adminRole = models.Role{
			Name:        "管理员",
			Slug:        "admin",
			Description: "普通管理员",
			Status:      1,
			Sort:        1,
		}
		facades.Orm().Query().Create(&adminRole)
	} else {
		// 存在则更新其他字段（如果为空或需要更新）
		update := false
		if adminRole.Slug == "" {
			adminRole.Slug = "admin"
			update = true
		}
		if adminRole.Description == "" {
			adminRole.Description = "普通管理员"
			update = true
		}
		if adminRole.Status == 0 {
			adminRole.Status = 1
			update = true
		}
		if adminRole.Sort == 0 {
			adminRole.Sort = 1
			update = true
		}
		if update {
			facades.Orm().Query().Save(&adminRole)
		}
	}

	// 关联超级管理员和超级角色
	facades.Orm().Query().Model(&superAdmin).Association("Roles").Replace([]models.Role{superRole})

	// 给开发者管理员分配 super-admin 角色
	facades.Orm().Query().Model(&developerAdmin).Association("Roles").Replace([]models.Role{superRole})

	// super-admin 角色不需要分配权限和菜单，因为它在权限中间件中会跳过权限检查
	// 在获取用户信息时，会特殊处理 super-admin 角色，返回所有菜单用于前端显示

	// 创建演示账户角色（只允许查看，不允许编辑创建删除）
	var demoRole models.Role
	if err := facades.Orm().Query().Where("name", "演示账户").First(&demoRole); err != nil {
		// 不存在则创建
		demoRole = models.Role{
			Name:        "演示账户",
			Slug:        "demo",
			Description: "演示账户，只允许查看，不允许编辑、创建、删除",
			Status:      1,
			Sort:        2,
		}
		facades.Orm().Query().Create(&demoRole)
	} else {
		// 存在则更新其他字段（如果为空或需要更新）
		update := false
		if demoRole.Slug == "" {
			demoRole.Slug = "demo"
			update = true
		}
		if demoRole.Description == "" {
			demoRole.Description = "演示账户，只允许查看，不允许编辑、创建、删除"
			update = true
		}
		if demoRole.Status == 0 {
			demoRole.Status = 1
			update = true
		}
		if demoRole.Sort == 0 {
			demoRole.Sort = 2
			update = true
		}
		if update {
			facades.Orm().Query().Save(&demoRole)
		}
	}

	// 给演示角色分配所有查看权限（index 和 show）
	var viewPermissions []models.Permission
	if err := facades.Orm().Query().Where("slug LIKE ?", "%.index").OrWhere("slug LIKE ?", "%.show").Find(&viewPermissions); err == nil {
		facades.Orm().Query().Model(&demoRole).Association("Permissions").Replace(viewPermissions)
	}

	// 给演示角色分配所有菜单（用于前端显示）
	var allMenus []models.Menu
	if err := facades.Orm().Query().Where("status", 1).Find(&allMenus); err == nil {
		facades.Orm().Query().Model(&demoRole).Association("Menus").Replace(allMenus)
	}

	// 创建演示账户
	demoPassword, _ := facades.Hash().Make("demo123")
	demoAdmin := models.Admin{
		Username: "demo",
		Password: demoPassword,
		Nickname: "演示账户",
		Status:   1,
	}
	facades.Orm().Query().FirstOrCreate(&demoAdmin, models.Admin{Username: "demo"})

	// 给演示账户分配演示角色
	facades.Orm().Query().Model(&demoAdmin).Association("Roles").Replace([]models.Role{demoRole})

	return nil
}
