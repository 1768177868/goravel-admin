package seeders

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type PermissionSeeder struct {
}

func (s *PermissionSeeder) Signature() string {
	return "PermissionSeeder"
}

func (s *PermissionSeeder) Run() error {
	// 获取菜单（权限需要关联菜单）
	var adminMenu, roleMenu, permissionMenu, menuMenu, departmentMenu, dictionaryMenu, configMenu, blacklistMenu, onlineUserMenu models.Menu
	var operationLogMenu, loginLogMenu, systemLogMenu, monitorMenu, profileMenu, exportMenu, attachmentMenu models.Menu

	facades.Orm().Query().Where("slug", "admin").First(&adminMenu)
	facades.Orm().Query().Where("slug", "role").First(&roleMenu)
	facades.Orm().Query().Where("slug", "permission").First(&permissionMenu)
	facades.Orm().Query().Where("slug", "menu").First(&menuMenu)
	facades.Orm().Query().Where("slug", "department").First(&departmentMenu)
	facades.Orm().Query().Where("slug", "dictionary").First(&dictionaryMenu)
	facades.Orm().Query().Where("slug", "config").First(&configMenu)
	facades.Orm().Query().Where("slug", "blacklist").First(&blacklistMenu)
	facades.Orm().Query().Where("slug", "online-user").First(&onlineUserMenu)
	facades.Orm().Query().Where("slug", "operation-log").First(&operationLogMenu)
	facades.Orm().Query().Where("slug", "login-log").First(&loginLogMenu)
	facades.Orm().Query().Where("slug", "system-log").First(&systemLogMenu)
	facades.Orm().Query().Where("slug", "monitor").First(&monitorMenu)
	facades.Orm().Query().Where("slug", "profile").First(&profileMenu)
	facades.Orm().Query().Where("slug", "export").First(&exportMenu)
	facades.Orm().Query().Where("slug", "attachment").First(&attachmentMenu)

	// 创建权限（关联菜单ID）
	permissions := []models.Permission{
		// 管理员管理
		{Name: "管理员列表", Slug: "admin.index", Method: "GET", Path: "/api/admin/admins", Description: "查看管理员列表", Status: 1, Sort: 1, MenuID: adminMenu.ID},
		{Name: "管理员详情", Slug: "admin.show", Method: "GET", Path: "/api/admin/admins/*", Description: "查看管理员详情", Status: 1, Sort: 2, MenuID: adminMenu.ID},
		{Name: "管理员创建", Slug: "admin.store", Method: "POST", Path: "/api/admin/admins", Description: "创建管理员", Status: 1, Sort: 3, MenuID: adminMenu.ID},
		{Name: "管理员更新", Slug: "admin.update", Method: "PUT", Path: "/api/admin/admins/*", Description: "更新管理员", Status: 1, Sort: 4, MenuID: adminMenu.ID},
		{Name: "管理员删除", Slug: "admin.destroy", Method: "DELETE", Path: "/api/admin/admins/*", Description: "删除管理员", Status: 1, Sort: 5, MenuID: adminMenu.ID},
		{Name: "管理员导出", Slug: "admin.export", Method: "POST", Path: "/api/admin/admins/export", Description: "导出管理员列表", Status: 1, Sort: 6, MenuID: adminMenu.ID},
		{Name: "重置密码", Slug: "admin.password", Method: "PUT", Path: "/api/admin/admins/*/password", Description: "重置管理员密码", Status: 1, Sort: 7, MenuID: adminMenu.ID},
		{Name: "踢出用户", Slug: "admin.kick_out", Method: "DELETE", Path: "/api/admin/admins/*/tokens", Description: "踢出指定用户的所有token", Status: 1, Sort: 8, MenuID: adminMenu.ID},
		// 角色管理
		{Name: "角色列表", Slug: "role.index", Method: "GET", Path: "/api/admin/roles", Description: "查看角色列表", Status: 1, Sort: 1, MenuID: roleMenu.ID},
		{Name: "角色详情", Slug: "role.show", Method: "GET", Path: "/api/admin/roles/*", Description: "查看角色详情", Status: 1, Sort: 2, MenuID: roleMenu.ID},
		{Name: "角色创建", Slug: "role.store", Method: "POST", Path: "/api/admin/roles", Description: "创建角色", Status: 1, Sort: 3, MenuID: roleMenu.ID},
		{Name: "角色更新", Slug: "role.update", Method: "PUT", Path: "/api/admin/roles/*", Description: "更新角色", Status: 1, Sort: 4, MenuID: roleMenu.ID},
		{Name: "角色删除", Slug: "role.destroy", Method: "DELETE", Path: "/api/admin/roles/*", Description: "删除角色", Status: 1, Sort: 5, MenuID: roleMenu.ID},
		// 权限管理
		{Name: "权限列表", Slug: "permission.index", Method: "GET", Path: "/api/admin/permissions", Description: "查看权限列表", Status: 1, Sort: 1, MenuID: permissionMenu.ID},
		{Name: "权限详情", Slug: "permission.show", Method: "GET", Path: "/api/admin/permissions/*", Description: "查看权限详情", Status: 1, Sort: 2, MenuID: permissionMenu.ID},
		{Name: "权限创建", Slug: "permission.store", Method: "POST", Path: "/api/admin/permissions", Description: "创建权限", Status: 1, Sort: 3, MenuID: permissionMenu.ID},
		{Name: "权限更新", Slug: "permission.update", Method: "PUT", Path: "/api/admin/permissions/*", Description: "更新权限", Status: 1, Sort: 4, MenuID: permissionMenu.ID},
		{Name: "权限删除", Slug: "permission.destroy", Method: "DELETE", Path: "/api/admin/permissions/*", Description: "删除权限", Status: 1, Sort: 5, MenuID: permissionMenu.ID},
		// 菜单管理
		{Name: "菜单列表", Slug: "menu.index", Method: "GET", Path: "/api/admin/menus", Description: "查看菜单列表", Status: 1, Sort: 1, MenuID: menuMenu.ID},
		{Name: "菜单详情", Slug: "menu.show", Method: "GET", Path: "/api/admin/menus/*", Description: "查看菜单详情", Status: 1, Sort: 2, MenuID: menuMenu.ID},
		{Name: "菜单创建", Slug: "menu.store", Method: "POST", Path: "/api/admin/menus", Description: "创建菜单", Status: 1, Sort: 3, MenuID: menuMenu.ID},
		{Name: "菜单更新", Slug: "menu.update", Method: "PUT", Path: "/api/admin/menus/*", Description: "更新菜单", Status: 1, Sort: 4, MenuID: menuMenu.ID},
		{Name: "菜单删除", Slug: "menu.destroy", Method: "DELETE", Path: "/api/admin/menus/*", Description: "删除菜单", Status: 1, Sort: 5, MenuID: menuMenu.ID},
		// 部门管理
		{Name: "部门列表", Slug: "department.index", Method: "GET", Path: "/api/admin/departments", Description: "查看部门列表", Status: 1, Sort: 1, MenuID: departmentMenu.ID},
		{Name: "部门详情", Slug: "department.show", Method: "GET", Path: "/api/admin/departments/*", Description: "查看部门详情", Status: 1, Sort: 2, MenuID: departmentMenu.ID},
		{Name: "部门创建", Slug: "department.store", Method: "POST", Path: "/api/admin/departments", Description: "创建部门", Status: 1, Sort: 3, MenuID: departmentMenu.ID},
		{Name: "部门更新", Slug: "department.update", Method: "PUT", Path: "/api/admin/departments/*", Description: "更新部门", Status: 1, Sort: 4, MenuID: departmentMenu.ID},
		{Name: "部门删除", Slug: "department.destroy", Method: "DELETE", Path: "/api/admin/departments/*", Description: "删除部门", Status: 1, Sort: 5, MenuID: departmentMenu.ID},
		// 字典管理
		{Name: "字典列表", Slug: "dictionary.index", Method: "GET", Path: "/api/admin/dictionaries", Description: "查看字典列表", Status: 1, Sort: 1, MenuID: dictionaryMenu.ID},
		{Name: "字典详情", Slug: "dictionary.show", Method: "GET", Path: "/api/admin/dictionaries/*", Description: "查看字典详情", Status: 1, Sort: 2, MenuID: dictionaryMenu.ID},
		{Name: "字典创建", Slug: "dictionary.store", Method: "POST", Path: "/api/admin/dictionaries", Description: "创建字典", Status: 1, Sort: 3, MenuID: dictionaryMenu.ID},
		{Name: "字典更新", Slug: "dictionary.update", Method: "PUT", Path: "/api/admin/dictionaries/*", Description: "更新字典", Status: 1, Sort: 4, MenuID: dictionaryMenu.ID},
		{Name: "字典删除", Slug: "dictionary.destroy", Method: "DELETE", Path: "/api/admin/dictionaries/*", Description: "删除字典", Status: 1, Sort: 5, MenuID: dictionaryMenu.ID},
		{Name: "字典查询", Slug: "dictionary.type", Method: "GET", Path: "/api/admin/dictionaries/type/*", Description: "根据类型查询字典", Status: 1, Sort: 6, MenuID: dictionaryMenu.ID},
		// 配置管理
		{Name: "获取配置", Slug: "config.group", Method: "GET", Path: "/api/admin/configs/group/*", Description: "根据分组获取配置", Status: 1, Sort: 1, MenuID: configMenu.ID},
		{Name: "保存配置", Slug: "config.save", Method: "POST", Path: "/api/admin/configs/save", Description: "保存配置", Status: 1, Sort: 2, MenuID: configMenu.ID},
		{Name: "测试邮箱", Slug: "config.test_email", Method: "POST", Path: "/api/admin/configs/test-email", Description: "测试邮箱配置", Status: 1, Sort: 3, MenuID: configMenu.ID},
		// 黑名单管理
		{Name: "黑名单列表", Slug: "blacklist.index", Method: "GET", Path: "/api/admin/blacklists", Description: "查看黑名单列表", Status: 1, Sort: 1, MenuID: blacklistMenu.ID},
		{Name: "黑名单详情", Slug: "blacklist.show", Method: "GET", Path: "/api/admin/blacklists/*", Description: "查看黑名单详情", Status: 1, Sort: 2, MenuID: blacklistMenu.ID},
		{Name: "黑名单创建", Slug: "blacklist.store", Method: "POST", Path: "/api/admin/blacklists", Description: "创建黑名单", Status: 1, Sort: 3, MenuID: blacklistMenu.ID},
		{Name: "黑名单更新", Slug: "blacklist.update", Method: "PUT", Path: "/api/admin/blacklists/*", Description: "更新黑名单", Status: 1, Sort: 4, MenuID: blacklistMenu.ID},
		{Name: "黑名单删除", Slug: "blacklist.destroy", Method: "DELETE", Path: "/api/admin/blacklists/*", Description: "删除黑名单", Status: 1, Sort: 5, MenuID: blacklistMenu.ID},

		// 在线用户管理
		{Name: "在线用户列表", Slug: "online-user.index", Method: "GET", Path: "/api/admin/online-users", Description: "查看在线用户列表", Status: 1, Sort: 1, MenuID: onlineUserMenu.ID},
		{Name: "踢下线", Slug: "online-user.kick-out", Method: "DELETE", Path: "/api/admin/online-users/*", Description: "踢下线用户", Status: 1, Sort: 2, MenuID: onlineUserMenu.ID},
		{Name: "批量踢下线", Slug: "online-user.batch-kick-out", Method: "POST", Path: "/api/admin/online-users/batch-kick-out", Description: "批量踢下线用户", Status: 1, Sort: 3, MenuID: onlineUserMenu.ID},
		// 操作日志
		{Name: "操作日志列表", Slug: "operation_log.index", Method: "GET", Path: "/api/admin/operation-logs", Description: "查看操作日志列表", Status: 1, Sort: 1, MenuID: operationLogMenu.ID},
		{Name: "操作日志详情", Slug: "operation_log.show", Method: "GET", Path: "/api/admin/operation-logs/*", Description: "查看操作日志详情", Status: 1, Sort: 2, MenuID: operationLogMenu.ID},
		{Name: "操作日志删除", Slug: "operation_log.destroy", Method: "DELETE", Path: "/api/admin/operation-logs/*", Description: "删除操作日志", Status: 1, Sort: 3, MenuID: operationLogMenu.ID},
		{Name: "操作日志批量删除", Slug: "operation_log.batch_delete", Method: "POST", Path: "/api/admin/operation-logs/batch-delete", Description: "批量删除操作日志", Status: 1, Sort: 4, MenuID: operationLogMenu.ID},
		{Name: "操作日志清理", Slug: "operation_log.clean", Method: "POST", Path: "/api/admin/operation-logs/clean", Description: "清理操作日志", Status: 1, Sort: 5, MenuID: operationLogMenu.ID},
		// 登录日志
		{Name: "登录日志列表", Slug: "login_log.index", Method: "GET", Path: "/api/admin/login-logs", Description: "查看登录日志列表", Status: 1, Sort: 1, MenuID: loginLogMenu.ID},
		{Name: "登录日志详情", Slug: "login_log.show", Method: "GET", Path: "/api/admin/login-logs/*", Description: "查看登录日志详情", Status: 1, Sort: 2, MenuID: loginLogMenu.ID},
		{Name: "登录日志删除", Slug: "login_log.destroy", Method: "DELETE", Path: "/api/admin/login-logs/*", Description: "删除登录日志", Status: 1, Sort: 3, MenuID: loginLogMenu.ID},
		{Name: "登录日志批量删除", Slug: "login_log.batch_delete", Method: "POST", Path: "/api/admin/login-logs/batch-delete", Description: "批量删除登录日志", Status: 1, Sort: 4, MenuID: loginLogMenu.ID},
		{Name: "登录日志清理", Slug: "login_log.clean", Method: "POST", Path: "/api/admin/login-logs/clean", Description: "清理登录日志", Status: 1, Sort: 5, MenuID: loginLogMenu.ID},
		// 系统日志
		{Name: "系统日志列表", Slug: "system_log.index", Method: "GET", Path: "/api/admin/system-logs", Description: "查看系统日志列表", Status: 1, Sort: 1, MenuID: systemLogMenu.ID},
		{Name: "系统日志详情", Slug: "system_log.show", Method: "GET", Path: "/api/admin/system-logs/*", Description: "查看系统日志详情", Status: 1, Sort: 2, MenuID: systemLogMenu.ID},
		{Name: "系统日志删除", Slug: "system_log.destroy", Method: "DELETE", Path: "/api/admin/system-logs/*", Description: "删除系统日志", Status: 1, Sort: 3, MenuID: systemLogMenu.ID},
		{Name: "系统日志批量删除", Slug: "system_log.batch_delete", Method: "POST", Path: "/api/admin/system-logs/batch-delete", Description: "批量删除系统日志", Status: 1, Sort: 4, MenuID: systemLogMenu.ID},
		{Name: "系统日志清理", Slug: "system_log.clean", Method: "POST", Path: "/api/admin/system-logs/clean", Description: "清理系统日志", Status: 1, Sort: 5, MenuID: systemLogMenu.ID},
		// 服务监控
		{Name: "系统监控", Slug: "monitor.system_info", Method: "GET", Path: "/api/admin/monitor/system-info", Description: "查看系统监控信息", Status: 1, Sort: 1, MenuID: monitorMenu.ID},
		// 个人中心
		{Name: "修改资料", Slug: "profile.update", Method: "PUT", Path: "/api/admin/profile", Description: "修改当前登录管理员资料", Status: 1, Sort: 1, MenuID: profileMenu.ID},
		{Name: "修改密码", Slug: "password.update", Method: "PUT", Path: "/api/admin/password", Description: "修改当前登录管理员密码", Status: 1, Sort: 2, MenuID: profileMenu.ID},
		// 导出管理
		{Name: "导出列表", Slug: "export.index", Method: "GET", Path: "/api/admin/exports", Description: "查看导出记录列表", Status: 1, Sort: 1, MenuID: exportMenu.ID},
		{Name: "删除导出", Slug: "export.destroy", Method: "DELETE", Path: "/api/admin/exports/*", Description: "删除导出记录及源文件", Status: 1, Sort: 2, MenuID: exportMenu.ID},
		{Name: "导出批量删除", Slug: "export.batch_delete", Method: "POST", Path: "/api/admin/exports/batch-delete", Description: "批量删除导出记录", Status: 1, Sort: 3, MenuID: exportMenu.ID},
		// 附件管理
		{Name: "附件列表", Slug: "attachment.index", Method: "GET", Path: "/api/admin/attachments", Description: "查看附件列表", Status: 1, Sort: 1, MenuID: attachmentMenu.ID},
		{Name: "附件上传", Slug: "attachment.upload", Method: "POST", Path: "/api/admin/attachments/upload", Description: "上传附件", Status: 1, Sort: 2, MenuID: attachmentMenu.ID},
		{Name: "大文件分片上传", Slug: "attachment.chunk", Method: "POST", Path: "/api/admin/attachments/chunk", Description: "大文件分片上传（包含初始化、上传分片、合并分片、获取进度）", Status: 1, Sort: 3, MenuID: attachmentMenu.ID},
		// {Name: "附件预览", Slug: "attachment.preview", Method: "GET", Path: "/api/admin/attachments/*/preview", Description: "预览附件", Status: 1, Sort: 4, MenuID: attachmentMenu.ID},
		{Name: "附件下载", Slug: "attachment.download", Method: "GET", Path: "/api/admin/attachments/*/download", Description: "下载附件", Status: 1, Sort: 5, MenuID: attachmentMenu.ID},
		{Name: "附件更新显示名称", Slug: "attachment.update_display_name", Method: "PUT", Path: "/api/admin/attachments/*/display-name", Description: "更新附件显示名称", Status: 1, Sort: 6, MenuID: attachmentMenu.ID},
		{Name: "附件删除", Slug: "attachment.destroy", Method: "DELETE", Path: "/api/admin/attachments/*", Description: "删除附件", Status: 1, Sort: 7, MenuID: attachmentMenu.ID},
		{Name: "附件批量删除", Slug: "attachment.batch_delete", Method: "POST", Path: "/api/admin/attachments/batch-delete", Description: "批量删除附件", Status: 1, Sort: 8, MenuID: attachmentMenu.ID},
	}

	for _, perm := range permissions {
		// 检查 slug 是否有效
		if perm.Slug == "" {
			facades.Log().Errorf("Permission has empty slug, skipping")
			continue
		}

		// 检查菜单ID是否有效
		if perm.MenuID == 0 {
			facades.Log().Errorf("Permission %s has MenuID 0, skipping", perm.Slug)
			continue
		}

		var existPerm models.Permission
		if err := facades.Orm().Query().Where("slug", perm.Slug).First(&existPerm); err != nil {
			// 不存在则创建
			if err := facades.Orm().Query().Create(&perm); err != nil {
				facades.Log().Errorf("Failed to create permission %s: %v", perm.Slug, err)
			} else {
				facades.Log().Infof("Created permission: %s (MenuID: %d)", perm.Slug, perm.MenuID)
			}
		} else {
			// 存在则更新 MenuID 和其他字段（保留 slug）
			existPerm.MenuID = perm.MenuID
			existPerm.Name = perm.Name
			existPerm.Slug = perm.Slug // 确保 slug 不被清空
			existPerm.Method = perm.Method
			existPerm.Path = perm.Path
			existPerm.Description = perm.Description
			existPerm.Status = perm.Status
			existPerm.Sort = perm.Sort
			if err := facades.Orm().Query().Save(&existPerm); err != nil {
				facades.Log().Errorf("Failed to update permission %s: %v", perm.Slug, err)
			} else {
				facades.Log().Infof("Updated permission: %s (MenuID: %d)", perm.Slug, perm.MenuID)
			}
		}
	}

	return nil
}
