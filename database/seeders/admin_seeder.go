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
	// 检查是否已经初始化过（通过检查超级管理员是否存在）
	var existingSuperAdmin models.Admin
	if err := facades.Orm().Query().Where("username", "admin").First(&existingSuperAdmin); err == nil {
		// 超级管理员已存在，说明已经初始化过，直接跳过
		return nil
	}

	// 创建超级管理员
	hashedPassword, _ := facades.Hash().Make("admin123")
	superAdmin := models.Admin{
		Username: "admin",
		Password: hashedPassword,
		Nickname: "超级管理员",
		Status:   1,
	}
	if err := facades.Orm().Query().Create(&superAdmin); err != nil {
		return err
	}

	// 创建开发者管理员（受保护，不显示在列表中）
	developerPassword, _ := facades.Hash().Make("developer123")
	developerAdmin := models.Admin{
		Username: "developer",
		Password: developerPassword,
		Nickname: "开发者管理员",
		Status:   1,
	}
	if err := facades.Orm().Query().Create(&developerAdmin); err != nil {
		return err
	}

	// 创建部门
	rootDept := models.Department{
		Name:   "总公司",
		Code:   "ROOT",
		Status: 1,
		Sort:   0,
	}
	if err := facades.Orm().Query().Create(&rootDept); err != nil {
		return err
	}

	itDept := models.Department{
		ParentID: rootDept.ID,
		Name:     "技术部",
		Code:     "IT",
		Status:   1,
		Sort:     1,
	}
	if err := facades.Orm().Query().Create(&itDept); err != nil {
		return err
	}

	// 创建角色
	superRole := models.Role{
		Name:        "超级管理员",
		Slug:        "super-admin",
		Description: "拥有所有权限",
		Status:      1,
		Sort:        0,
	}
	if err := facades.Orm().Query().Create(&superRole); err != nil {
		return err
	}

	adminRole := models.Role{
		Name:        "管理员",
		Slug:        "admin",
		Description: "普通管理员",
		Status:      1,
		Sort:        1,
	}
	if err := facades.Orm().Query().Create(&adminRole); err != nil {
		return err
	}

	// 关联超级管理员和超级角色
	if err := facades.Orm().Query().Model(&superAdmin).Association("Roles").Append([]models.Role{superRole}); err != nil {
		return err
	}

	// 给开发者管理员分配 super-admin 角色
	if err := facades.Orm().Query().Model(&developerAdmin).Association("Roles").Append([]models.Role{superRole}); err != nil {
		return err
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
	if err := facades.Orm().Query().Create(&demoRole); err != nil {
		return err
	}

	// 给演示角色分配所有查看权限（index 和 show）
	var viewPermissions []models.Permission
	if err := facades.Orm().Query().Where("slug LIKE ?", "%.index").OrWhere("slug LIKE ?", "%.show").OrWhere("slug", "dashboard.data").Find(&viewPermissions); err == nil && len(viewPermissions) > 0 {
		if err := facades.Orm().Query().Model(&demoRole).Association("Permissions").Append(viewPermissions); err != nil {
			return err
		}
	}

	// 给演示角色分配所有菜单（用于前端显示）
	var allMenus []models.Menu
	if err := facades.Orm().Query().Where("status", 1).Find(&allMenus); err == nil && len(allMenus) > 0 {
		if err := facades.Orm().Query().Model(&demoRole).Association("Menus").Append(allMenus); err != nil {
			return err
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
	if err := facades.Orm().Query().Create(&demoAdmin); err != nil {
		return err
	}

	// 给演示账户分配演示角色
	if err := facades.Orm().Query().Model(&demoAdmin).Association("Roles").Append([]models.Role{demoRole}); err != nil {
		return err
	}

	return nil
}
