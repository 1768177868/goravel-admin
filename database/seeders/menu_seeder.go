package seeders

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type MenuSeeder struct {
}

func (s *MenuSeeder) Signature() string {
	return "MenuSeeder"
}

func (s *MenuSeeder) Run() error {
	// 辅助函数：根据Slug查找或创建菜单，如果存在则更新（保留用户修改的图标）
	createOrUpdateMenu := func(menuData models.Menu) models.Menu {
		// 如果 slug 为空，跳过（不应该发生，但作为保护）
		if menuData.Slug == "" {
			facades.Log().Errorf("Menu slug is empty for title: %s", menuData.Title)
			return menuData
		}

		var existingMenu models.Menu
		// 先尝试通过 slug 查找
		if err := facades.Orm().Query().Where("slug", menuData.Slug).First(&existingMenu); err == nil {
			// 菜单已存在，更新除图标和排序外的其他字段（保留用户可能修改的图标和排序）
			existingMenu.ParentID = menuData.ParentID
			existingMenu.Title = menuData.Title
			existingMenu.Slug = menuData.Slug // 确保 slug 也被更新
			// 如果现有菜单的图标为空，才更新图标；否则保留用户修改的图标
			if existingMenu.Icon == "" {
				existingMenu.Icon = menuData.Icon
			}
			existingMenu.Path = menuData.Path
			existingMenu.Component = menuData.Component
			existingMenu.Permission = menuData.Permission
			existingMenu.Type = menuData.Type
			existingMenu.Status = menuData.Status
			// 如果现有菜单的排序为0，则更新为填充数据中的排序值；否则保留用户手动调整的排序值
			if existingMenu.Sort == 0 {
				existingMenu.Sort = menuData.Sort
			}
			existingMenu.IsHidden = menuData.IsHidden
			facades.Orm().Query().Save(&existingMenu)
			// 重新查询确保获取最新数据
			facades.Orm().Query().Where("id", existingMenu.ID).First(&existingMenu)
			return existingMenu
		}

		// 如果通过 slug 找不到，尝试通过 path 和 title 查找（兼容旧数据，可能 slug 为空）
		if menuData.Path != "" {
			var existingByPath models.Menu
			if err := facades.Orm().Query().Where("path", menuData.Path).Where("title", menuData.Title).First(&existingByPath); err == nil {
				// 找到旧菜单（可能 slug 为空），更新它
				existingByPath.ParentID = menuData.ParentID
				existingByPath.Title = menuData.Title
				existingByPath.Slug = menuData.Slug // 更新 slug
				if existingByPath.Icon == "" {
					existingByPath.Icon = menuData.Icon
				}
				existingByPath.Path = menuData.Path
				existingByPath.Component = menuData.Component
				existingByPath.Permission = menuData.Permission
				existingByPath.Type = menuData.Type
				existingByPath.Status = menuData.Status
				// 如果现有菜单的排序为0，则更新为填充数据中的排序值；否则保留用户手动调整的排序值
				if existingByPath.Sort == 0 {
					existingByPath.Sort = menuData.Sort
				}
				existingByPath.IsHidden = menuData.IsHidden
				facades.Orm().Query().Save(&existingByPath)
				// 重新查询确保获取最新数据
				facades.Orm().Query().Where("id", existingByPath.ID).First(&existingByPath)
				return existingByPath
			}
		}

		// 菜单不存在，创建新菜单
		if err := facades.Orm().Query().Create(&menuData); err != nil {
			facades.Log().Errorf("Failed to create menu with slug %s: %v", menuData.Slug, err)
			return menuData
		}
		// 创建后重新查询获取完整的菜单信息（包括ID）
		var createdMenu models.Menu
		if err := facades.Orm().Query().Where("slug", menuData.Slug).First(&createdMenu); err == nil {
			return createdMenu
		}
		return menuData
	}

	// 创建菜单
	systemMenu := createOrUpdateMenu(models.Menu{
		ParentID:  0,
		Title:     "系统管理",
		Slug:      "system",
		Icon:      "Setting",
		Path:      "/system",
		Component: "Layout",
		Type:      1,
		Status:    1,
		Sort:      1,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "管理员管理",
		Slug:      "admin",
		Icon:      "User",
		Path:      "/admins",
		Component: "admin/index",
		Type:      2,
		Status:    1,
		Sort:      1,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "角色管理",
		Slug:      "role",
		Icon:      "UserFilled",
		Path:      "/roles",
		Component: "role/index",
		Type:      2,
		Status:    1,
		Sort:      2,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "权限管理",
		Slug:      "permission",
		Icon:      "Lock",
		Path:      "/permissions",
		Component: "permission/index",
		Type:      2,
		Status:    1,
		Sort:      3,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "菜单管理",
		Slug:      "menu",
		Icon:      "Menu",
		Path:      "/menus",
		Component: "menu/index",
		Type:      2,
		Status:    1,
		Sort:      4,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "部门管理",
		Slug:      "department",
		Icon:      "OfficeBuilding",
		Path:      "/departments",
		Component: "department/index",
		Type:      2,
		Status:    1,
		Sort:      5,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "在线用户",
		Slug:      "online-user",
		Icon:      "User",
		Path:      "/online-users",
		Component: "onlineUser/index",
		Type:      2,
		Status:    1,
		Sort:      6,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "字典管理",
		Slug:      "dictionary",
		Icon:      "Document",
		Path:      "/dictionaries",
		Component: "dictionary/index",
		Type:      2,
		Status:    1,
		Sort:      7,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "配置管理",
		Slug:      "config",
		Icon:      "Setting",
		Path:      "/configs",
		Component: "config/index",
		Type:      2,
		Status:    1,
		Sort:      8,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "导出管理",
		Slug:      "export",
		Icon:      "Document",
		Path:      "/exports",
		Component: "export/index",
		Type:      2,
		Status:    1,
		Sort:      9,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "附件管理",
		Slug:      "attachment",
		Icon:      "Folder",
		Path:      "/attachments",
		Component: "attachment/index",
		Type:      2,
		Status:    1,
		Sort:      10,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  systemMenu.ID,
		Title:     "IP黑名单",
		Slug:      "blacklist",
		Icon:      "Warning",
		Path:      "/blacklists",
		Component: "blacklist/index",
		Type:      2,
		Status:    1,
		Sort:      11,
		IsHidden:  0,
	})

	// 创建日志管理父菜单
	logMenu := createOrUpdateMenu(models.Menu{
		ParentID:  0,
		Title:     "日志管理",
		Slug:      "log",
		Icon:      "Document",
		Path:      "/logs",
		Component: "Layout",
		Type:      1,
		Status:    1,
		Sort:      2,
		IsHidden:  0,
	})

	// 创建日志管理子菜单
	createOrUpdateMenu(models.Menu{
		ParentID:  logMenu.ID,
		Title:     "操作日志",
		Slug:      "operation-log",
		Icon:      "Document",
		Path:      "/operation-logs",
		Component: "log/operation/index",
		Type:      2,
		Status:    1,
		Sort:      1,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  logMenu.ID,
		Title:     "登录日志",
		Slug:      "login-log",
		Icon:      "Document",
		Path:      "/login-logs",
		Component: "log/login/index",
		Type:      2,
		Status:    1,
		Sort:      2,
		IsHidden:  0,
	})

	createOrUpdateMenu(models.Menu{
		ParentID:  logMenu.ID,
		Title:     "系统日志",
		Slug:      "system-log",
		Icon:      "Document",
		Path:      "/system-logs",
		Component: "log/system/index",
		Type:      2,
		Status:    1,
		Sort:      3,
		IsHidden:  0,
	})

	// 创建服务监控菜单
	createOrUpdateMenu(models.Menu{
		ParentID:  0,
		Title:     "服务监控",
		Slug:      "monitor",
		Icon:      "Monitor",
		Path:      "/monitor",
		Component: "monitor/index",
		Type:      2,
		Status:    1,
		Sort:      3,
		IsHidden:  0,
	})

	// 创建个人中心菜单
	createOrUpdateMenu(models.Menu{
		ParentID:  0,
		Title:     "个人中心",
		Slug:      "profile",
		Icon:      "User",
		Path:      "/profile",
		Component: "profile/index",
		Type:      2,
		Status:    1,
		Sort:      4,
		IsHidden:  1,
	})

	return nil
}
