package utils

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/str"
)

// GetOperationTitleFromContext 从 context 中获取操作标题
// 优先使用权限标识（permission_slug），如果没有权限标识，则根据路径和方法生成默认标题
func GetOperationTitleFromContext(ctx http.Context) string {
	if ctx == nil {
		return "operation.unknown"
	}

	// 优先从 context 中获取权限标识（由权限中间件设置）
	permissionSlugValue := ctx.Value("permission_slug")
	if permissionSlugValue != nil {
		if permissionSlug, ok := permissionSlugValue.(string); ok && permissionSlug != "" {
			// 直接返回权限标识，前端多语言文件中已有对应翻译
			return permissionSlug
		}
	}

	// 如果没有权限标识，根据路径和方法生成默认标题
	method := ctx.Request().Method()
	path := ctx.Request().Path()

	// 根据路径生成默认标题
	defaultTitle := generateDefaultTitle(method, path)
	if defaultTitle != "" {
		return defaultTitle
	}

	// 无法生成标题时，返回未知操作
	return "operation.unknown"
}

// generateDefaultTitle 根据方法和路径生成默认操作标题
func generateDefaultTitle(method, path string) string {
	pathStr := str.Of(path)

	// 分片上传相关（与权限配置中的 slug 保持一致）
	if pathStr.Contains("/attachments/chunk") {
		if method == "POST" || method == "GET" {
			// 权限配置中的 slug 是 attachment.chunk
			return "attachment.chunk"
		}
	}

	// 附件上传
	if pathStr.Contains("/attachments/upload") && method == "POST" {
		return "attachment.upload"
	}

	// 附件删除
	if pathStr.Contains("/attachments/") && pathStr.EndsWith("/batch-delete") && method == "POST" {
		return "attachment.batch_delete"
	}
	if pathStr.Contains("/attachments/") && method == "DELETE" {
		return "attachment.destroy"
	}

	// 附件更新显示名称
	if pathStr.Contains("/attachments/") && pathStr.EndsWith("/display-name") && method == "PUT" {
		return "attachment.update_display_name"
	}

	// 导出下载
	if pathStr.Contains("/exports/") && pathStr.EndsWith("/download") && method == "GET" {
		return "export.download"
	}

	// 管理员解绑谷歌验证码
	if pathStr.Contains("/admins/") && pathStr.EndsWith("/unbind-google-auth") && method == "POST" {
		return "admin.unbind_google_auth"
	}

	// 批量删除（通用模式）
	if pathStr.EndsWith("/batch-delete") && method == "POST" {
		parts := pathStr.ChopStart("/api/admin/").Split("/")
		if len(parts) > 0 {
			module := str.Of(parts[0]).Replace("-", "_").String()
			return str.Of(module).Append(".batch_delete").String()
		}
	}

	// 清理操作（通用模式）
	if pathStr.EndsWith("/clean") && method == "POST" {
		parts := pathStr.ChopStart("/api/admin/").Split("/")
		if len(parts) > 0 {
			module := str.Of(parts[0]).Replace("-", "_").String()
			return str.Of(module).Append(".clean").String()
		}
	}

	// 标准 CRUD 操作（通用模式）
	parts := pathStr.ChopStart("/api/admin/").Split("/")
	if len(parts) >= 1 {
		module := str.Of(parts[0]).Replace("-", "_").String()
		switch method {
		case "POST":
			// 创建操作
			if len(parts) == 1 || (len(parts) == 2 && parts[1] != "batch-delete" && parts[1] != "clean") {
				return str.Of(module).Append(".store").String()
			}
		case "PUT", "PATCH":
			// 更新操作
			if len(parts) >= 2 {
				return str.Of(module).Append(".update").String()
			}
		case "DELETE":
			// 删除操作
			if len(parts) >= 2 {
				return str.Of(module).Append(".destroy").String()
			}
		}
	}

	return ""
}
