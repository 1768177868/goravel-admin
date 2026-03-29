// Package utils 中的操作日志变更审计：默认规则见注册顺序，扩展方式如下。
//
// 路径带数字 ID 的 REST（PUT/DELETE …/admins/1）：在 audit_diff.auditModelFactories 登记模型即可。
//
// 路径无 ID、靠请求体区分批次（如 POST …/configs/save、body 内含 group + configs）：
// 使用 RegisterAuditHandler，Before 按 body 查库得到 key→旧值，Diff 用 ComputeDiffAgainstNestedMap 与 body 中嵌套字段对比。
//
// 路径无 ID、但修改的是当前登录用户（如 PUT …/profile）：Before 会收到 operatorAdminID（JWT），用其加载 admins 快照。
//
// 其它表同理；若需优先匹配，用 RegisterAuditHandlerFirst。
//
// 对比逻辑非嵌套 map 时：RegisterAuditHandler 自定义 Before + Diff。
package utils

import (
	"encoding/json"
	"strings"
)

// AuditBeforeFunc 在 Next 之前加载快照；operatorAdminID 为当前登录管理员 ID（JWT），无则 0。
type AuditBeforeFunc func(method, path, requestBody string, operatorAdminID uint) map[string]any

// AuditHandler 自定义审计规则：匹配请求 → 加载修改前快照 → 与请求体对比生成 changes JSON。
type AuditHandler struct {
	Name   string
	Match  func(method, path string) bool
	Before AuditBeforeFunc
	Diff   func(method, path string, before map[string]any, requestBody string) string
}

var auditHandlers []AuditHandler

// RegisterAuditHandler 注册审计处理器（按注册顺序匹配，先注册者优先）。
func RegisterAuditHandler(h AuditHandler) {
	auditHandlers = append(auditHandlers, h)
}

// RegisterAuditHandlerFirst 将规则插到最前，优先于内置配置保存 / REST PUT / REST DELETE。
// 无路径 ID 的自定义接口若与通用规则路径相似，应使用本函数注册，避免被其它规则先匹配。
func RegisterAuditHandlerFirst(h AuditHandler) {
	auditHandlers = append([]AuditHandler{h}, auditHandlers...)
}

// PrepareAuditChanges 在 Next 之前调用；返回的闭包在 Next 之后执行以生成 changes JSON。
// operatorAdminID 为 JWT 中的管理员 ID（个人资料等无路径 ID 的接口依赖此值加载快照）。
// 无匹配规则时返回 nil。
func PrepareAuditChanges(method, path, requestBody string, operatorAdminID uint) func() string {
	for _, h := range auditHandlers {
		if h.Match == nil || !h.Match(method, path) {
			continue
		}
		before := safeBefore(h.Before, method, path, requestBody, operatorAdminID)
		h := h
		return func() string {
			if h.Diff == nil {
				return ""
			}
			return h.Diff(method, path, before, requestBody)
		}
	}
	return nil
}

func safeBefore(before AuditBeforeFunc, method, path, requestBody string, operatorAdminID uint) map[string]any {
	if before == nil {
		return nil
	}
	return before(method, path, requestBody, operatorAdminID)
}

func init() {
	registerDefaultAuditHandlers()
}

func registerDefaultAuditHandlers() {
	// 配置批量保存：POST …/configs/save，body.group + body.configs
	RegisterAuditHandler(AuditHandler{
		Name:  "config_batch_save",
		Match: matchConfigSavePath,
		Before: func(method, path, requestBody string, _ uint) map[string]any {
			g := ExtractStringFromJSON(requestBody, "group")
			return LoadConfigSnapshotByGroup(g)
		},
		Diff: func(method, path string, before map[string]any, requestBody string) string {
			return ComputeDiffAgainstNestedMap(before, requestBody, "configs")
		},
	})

	// 个人资料：PUT /api/admin/profile（无路径 ID，用当前登录管理员 ID 加载 admins 快照）
	RegisterAuditHandler(AuditHandler{
		Name: "profile_put",
		Match: func(method, path string) bool {
			return method == "PUT" && path == "/api/admin/profile"
		},
		Before: func(method, path, requestBody string, operatorAdminID uint) map[string]any {
			if operatorAdminID == 0 {
				return nil
			}
			return LoadModelSnapshot("admins", operatorAdminID)
		},
		Diff: func(method, path string, before map[string]any, requestBody string) string {
			return ComputeDiffFromRequest(before, requestBody)
		},
	})

	// REST：PUT，路径 .../resource/id
	RegisterAuditHandler(AuditHandler{
		Name: "rest_put",
		Match: func(method, path string) bool {
			if method != "PUT" {
				return false
			}
			_, id := ParseResourcePath(path)
			return id > 0
		},
		Before: func(method, path, requestBody string, _ uint) map[string]any {
			table, id := ParseResourcePath(path)
			return LoadModelSnapshot(table, id)
		},
		Diff: func(method, path string, before map[string]any, requestBody string) string {
			return ComputeDiffFromRequest(before, requestBody)
		},
	})

	// REST：DELETE，路径 .../resource/id
	RegisterAuditHandler(AuditHandler{
		Name: "rest_delete",
		Match: func(method, path string) bool {
			if method != "DELETE" {
				return false
			}
			_, id := ParseResourcePath(path)
			return id > 0
		},
		Before: func(method, path, requestBody string, _ uint) map[string]any {
			table, id := ParseResourcePath(path)
			return LoadModelSnapshot(table, id)
		},
		Diff: func(method, path string, before map[string]any, requestBody string) string {
			return FormatDeleteSnapshot(before)
		},
	})
}

func matchConfigSavePath(method, path string) bool {
	if method != "POST" {
		return false
	}
	return path == "/api/admin/configs/save"
}

// ExtractStringFromJSON 从 JSON 请求体字符串读取顶层字符串字段（供无 ID 接口的 loadBefore 解析 group、code 等）。
func ExtractStringFromJSON(jsonStr, key string) string {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return ""
	}
	s, _ := m[key].(string)
	return s
}
