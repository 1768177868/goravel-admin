package utils

import (
	"strings"

	"github.com/goravel/framework/contracts/http"
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
	// 分片上传相关（与权限配置中的 slug 保持一致）
	if strings.Contains(path, "/attachments/chunk") {
		if method == "POST" {
			// 权限配置中的 slug 是 attachment.chunk
			return "attachment.chunk"
		}
		if method == "GET" {
			// GET 请求用于获取进度，也使用相同的权限标识
			return "attachment.chunk"
		}
	}

	// 附件上传
	if strings.Contains(path, "/attachments/upload") && method == "POST" {
		return "attachment.upload"
	}

	// 附件删除
	if strings.Contains(path, "/attachments/") && strings.HasSuffix(path, "/batch-delete") && method == "POST" {
		return "attachment.batch_delete"
	}
	if strings.Contains(path, "/attachments/") && method == "DELETE" {
		return "attachment.destroy"
	}

	// 附件更新显示名称
	if strings.Contains(path, "/attachments/") && strings.HasSuffix(path, "/display-name") && method == "PUT" {
		return "attachment.update_display_name"
	}

	// 导出下载
	if strings.Contains(path, "/exports/") && strings.HasSuffix(path, "/download") && method == "GET" {
		return "export.download"
	}

	// 管理员解绑谷歌验证码
	if strings.Contains(path, "/admins/") && strings.HasSuffix(path, "/unbind-google-auth") && method == "POST" {
		return "admin.unbind_google_auth"
	}

	// 批量删除（通用模式）
	if strings.HasSuffix(path, "/batch-delete") && method == "POST" {
		// 提取模块名（直接使用路径中的原始模块名，不做任何转换）
		parts := strings.Split(strings.TrimPrefix(path, "/api/admin/"), "/")
		if len(parts) > 0 {
			module := parts[0]
			// 将连字符转换为下划线（如 operation-logs -> operation_logs）
			module = strings.ReplaceAll(module, "-", "_")
			return module + ".batch_delete"
		}
	}

	// 清理操作（通用模式）
	if strings.HasSuffix(path, "/clean") && method == "POST" {
		parts := strings.Split(strings.TrimPrefix(path, "/api/admin/"), "/")
		if len(parts) > 0 {
			module := parts[0]
			// 将连字符转换为下划线（如 operation-logs -> operation_logs）
			module = strings.ReplaceAll(module, "-", "_")
			return module + ".clean"
		}
	}

	// 标准 CRUD 操作（通用模式）
	parts := strings.Split(strings.TrimPrefix(path, "/api/admin/"), "/")
	if len(parts) >= 1 {
		module := parts[0]
		// 将连字符转换为下划线（如 operation-logs -> operation_logs）
		module = strings.ReplaceAll(module, "-", "_")
		switch method {
		case "POST":
			// 创建操作
			if len(parts) == 1 || (len(parts) == 2 && parts[1] != "batch-delete" && parts[1] != "clean") {
				return module + ".store"
			}
		case "PUT", "PATCH":
			// 更新操作
			if len(parts) >= 2 {
				return module + ".update"
			}
		case "DELETE":
			// 删除操作
			if len(parts) >= 2 {
				return module + ".destroy"
			}
		}
	}

	return ""
}
