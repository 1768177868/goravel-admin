package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/trans"
	"goravel/app/models"
)

// Permission 权限验证中间件
func Permission() http.Middleware {
	return func(ctx http.Context) {
		var admin models.Admin
		if err := facades.Auth(ctx).Guard("admin").User(&admin); err != nil {
			_ = ctx.Response().Json(http.StatusUnauthorized, http.Json{
				"code":    401,
				"message": trans.Get(ctx, "not_logged_in"),
			}).Abort()
			return
		}

		// 加载管理员的角色
		if err := facades.Orm().Query().Load(&admin, "Roles"); err != nil {
			_ = ctx.Response().Json(http.StatusInternalServerError, http.Json{
				"code":    500,
				"message": trans.Get(ctx, "load_permissions_failed"),
			}).Abort()
			return
		}

		// 批量加载所有角色的权限，避免 N+1 查询
		if len(admin.Roles) > 0 {
			var roleIDs []uint
			for _, role := range admin.Roles {
				roleIDs = append(roleIDs, role.ID)
			}
			// 查询角色权限中间表
			type RolePermission struct {
				RoleID       uint `gorm:"column:role_id"`
				PermissionID uint `gorm:"column:permission_id"`
			}
			var rolePermissions []RolePermission
			if err := facades.Orm().Query().Table("role_permission").Where("role_id IN ?", roleIDs).Find(&rolePermissions); err == nil {
				// 收集所有权限 ID
				var permissionIDs []uint
				rolePermissionMap := make(map[uint][]uint) // role_id -> []permission_id
				for _, rp := range rolePermissions {
					rolePermissionMap[rp.RoleID] = append(rolePermissionMap[rp.RoleID], rp.PermissionID)
					permissionIDs = append(permissionIDs, rp.PermissionID)
				}
				// 批量查询权限
				if len(permissionIDs) > 0 {
					var permissions []models.Permission
					if err := facades.Orm().Query().Where("id IN ?", permissionIDs).Find(&permissions); err == nil {
						permissionMap := make(map[uint]models.Permission)
						for _, perm := range permissions {
							permissionMap[perm.ID] = perm
						}
						// 填充到角色中
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
		}

		// 检查是否是超级管理员（跳过权限检查）
		for _, role := range admin.Roles {
			if role.Slug == "super-admin" && role.Status == 1 {
				ctx.Request().Next()
				return
			}
		}

		// 获取当前请求的方法和路径
		method := ctx.Request().Method()
		path := ctx.Request().Path()

		// 收集所有角色的权限（已通过预加载获取）
		var allPermissions []models.Permission
		for _, role := range admin.Roles {
			if role.Status == 1 {
				allPermissions = append(allPermissions, role.Permissions...)
			}
		}

		// 检查是否有权限
		hasPermission := false
		for _, perm := range allPermissions {
			if perm.Status == 1 {
				// 检查方法匹配
				if perm.Method == "" || perm.Method == method {
					// 检查路径匹配（支持通配符）
					if perm.Path == "" || perm.Path == path || matchPath(perm.Path, path) {
						hasPermission = true
						break
					}
				}
			}
		}

		if !hasPermission {
			_ = ctx.Response().Json(http.StatusForbidden, http.Json{
				"code":    403,
				"message": trans.Get(ctx, "no_permission"),
			}).Abort()
			return
		}

		ctx.Request().Next()
	}
}

// matchPath 简单的路径匹配，支持通配符
func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}
	// 简单的通配符匹配，如 /admin/admins/* 匹配 /admin/admins/1
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

