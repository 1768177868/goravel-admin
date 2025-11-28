package utils

import (
	"regexp"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// GetOperationTitle 根据请求方法和路径生成操作标题
// 返回标题的 key，用于多语言翻译
// 使用通用规则，无需为每个新模块添加映射
func GetOperationTitle(method, path string) string {
	return getTitleFromPath(method, path)
}

// GetOperationTitleFromContext 从 context 中获取路由信息并生成标题
// 目前主要使用路径解析，这种方式更通用，不需要依赖路由的具体实现细节
func GetOperationTitleFromContext(ctx http.Context) string {
	if ctx == nil {
		return "operation.unknown"
	}

	method := ctx.Request().Method()
	path := ctx.Request().Path()

	// 使用路径解析方式生成标题
	// 这种方式更通用，适用于所有路由，无需为每个新模块添加映射
	return getTitleFromPath(method, path)
}

// getTitleFromPath 根据路径解析生成标题
func getTitleFromPath(method, path string) string {
	// 移除路径前缀
	path = strings.TrimPrefix(path, "/api/admin")
	path = strings.TrimPrefix(path, "/admin")
	path = strings.TrimSuffix(path, "/")

	// 提取路径部分
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return "operation.unknown"
	}

	// 获取资源名称（第一个部分）
	resource := parts[0]

	// 检查是否是批量操作
	isBatch := len(parts) > 1 && (parts[1] == "batch-delete" || parts[1] == "batch-destroy" || parts[1] == "batch-kick-out")

	// 检查是否是特殊操作（如 password, profile, tokens 等）
	if len(parts) > 1 {
		secondPart := parts[1]
		// 处理特殊路径，如 /admins/{id}/password, /admins/{id}/tokens
		if secondPart == "password" || secondPart == "tokens" {
			resource = secondPart
		} else if isNumericID(secondPart) && len(parts) > 2 {
			// 如果第二部分是ID，第三部分是操作，使用第三部分作为资源
			resource = parts[2]
		}
	}

	// 转换资源名称格式（kebab-case 转 snake_case，复数转单数）
	resourceKey := normalizeResourceName(resource)

	// 根据HTTP方法确定操作类型
	var action string
	switch method {
	case "POST":
		if isBatch {
			action = "batch_create"
		} else {
			action = "create"
		}
	case "PUT", "PATCH":
		action = "update"
	case "DELETE":
		if isBatch {
			action = "batch_delete"
		} else {
			// 特殊处理：在线用户删除是"踢出"
			if resource == "online-users" || resource == "tokens" {
				action = "kick_out"
			} else {
				action = "delete"
			}
		}
	default:
		return "operation.unknown"
	}

	// 生成标题key：operation.{action}_{resource}
	return "operation." + action + "_" + resourceKey
}

// normalizeResourceName 规范化资源名称
// 将 kebab-case 转为 snake_case，复数转为单数
func normalizeResourceName(resource string) string {
	// 将 kebab-case 转为 snake_case
	resource = strings.ReplaceAll(resource, "-", "_")

	// 从配置中获取复数转单数映射
	pluralToSingularMap := getPluralToSingularMap()

	// 如果存在映射，使用映射值
	if singular, ok := pluralToSingularMap[resource]; ok {
		return singular
	}

	// 通用规则：如果以 s 结尾，去掉 s
	if strings.HasSuffix(resource, "s") && len(resource) > 1 {
		return resource[:len(resource)-1]
	}

	return resource
}

// getPluralToSingularMap 从配置中获取复数转单数映射
func getPluralToSingularMap() map[string]string {
	// 从配置中获取映射
	configMap := facades.Config().Get("operation_log.plural_to_singular", map[string]string{})

	// 类型断言
	if pluralToSingular, ok := configMap.(map[string]interface{}); ok {
		result := make(map[string]string)
		for k, v := range pluralToSingular {
			if str, ok := v.(string); ok {
				result[k] = str
			}
		}
		return result
	}

	// 如果类型断言失败，返回空映射
	return map[string]string{}
}

// isNumericID 判断字符串是否是数字ID
func isNumericID(s string) bool {
	matched, _ := regexp.MatchString(`^\d+$`, s)
	return matched
}
