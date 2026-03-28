// Package utils 中的操作日志变更审计：默认规则见注册顺序，扩展方式如下。
//
// 路径带数字 ID 的 REST（PUT/DELETE …/admins/1）：在 audit_diff.auditModelFactories 登记模型即可。
//
// 路径无 ID、靠请求体区分批次（如 POST …/configs/save、body 内含 group + configs）：
// 使用 RegisterNestedMapBatchAudit，在 loadBefore 里按 body 条件查库，返回 key→旧值，再与 body 中嵌套字段对比。
// 其它表同理；若需优先匹配，用 RegisterAuditHandlerFirst。
//
// 对比逻辑非嵌套 map 时：RegisterAuditHandler 自定义 Before + Diff。
package utils

import (
	"encoding/json"
	"strings"
)

// AuditHandler 自定义审计规则：匹配请求 → 加载修改前快照 → 与请求体对比生成 changes JSON。
type AuditHandler struct {
	Name   string
	Match  func(method, path string) bool
	Before func(method, path, requestBody string) map[string]any
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

// RegisterNestedMapBatchAudit 通用模板：无 REST 路径 ID、请求体为「若干条件字段 + 嵌套 map」的批量保存。
// loadBefore 中根据 requestBody 查库，返回「业务键 → 修改前值」的扁平 map；再与 body[nestedKey] 逐项对比（见 ComputeDiffAgainstNestedMap）。
// 典型：configs 按 group 加载；其它表可按 slug、tenant_id 等在 loadBefore 里自行 ORM 查询。
func RegisterNestedMapBatchAudit(name string, match func(method, path string) bool, nestedKey string, loadBefore func(method, path, requestBody string) map[string]any) {
	RegisterAuditHandler(AuditHandler{
		Name:   name,
		Match:  match,
		Before: loadBefore,
		Diff: func(method, path string, before map[string]any, requestBody string) string {
			return ComputeDiffAgainstNestedMap(before, requestBody, nestedKey)
		},
	})
}

// PrepareAuditChanges 在 Next 之前调用；返回的闭包在 Next 之后执行以生成 changes JSON。
// 无匹配规则时返回 nil。
func PrepareAuditChanges(method, path, requestBody string) func() string {
	for _, h := range auditHandlers {
		if h.Match == nil || !h.Match(method, path) {
			continue
		}
		before := safeBefore(h.Before, method, path, requestBody)
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

func safeBefore(before func(string, string, string) map[string]any, method, path, requestBody string) map[string]any {
	if before == nil {
		return nil
	}
	return before(method, path, requestBody)
}

func init() {
	registerDefaultAuditHandlers()
}

func registerDefaultAuditHandlers() {
	// 1) 配置批量保存（嵌套 map 模板，路径可在 config operation_log.audit_config_save_path 修改）
	RegisterNestedMapBatchAudit("config_batch_save", matchConfigSavePath, "configs", func(method, path, requestBody string) map[string]any {
		g := ExtractStringFromJSON(requestBody, "group")
		return LoadConfigSnapshotByGroup(g)
	})

	// 2) REST：PUT，路径 .../resource/id
	RegisterAuditHandler(AuditHandler{
		Name: "rest_put",
		Match: func(method, path string) bool {
			if method != "PUT" {
				return false
			}
			_, id := ParseResourcePath(path)
			return id > 0
		},
		Before: func(method, path, requestBody string) map[string]any {
			table, id := ParseResourcePath(path)
			return LoadModelSnapshot(table, id)
		},
		Diff: func(method, path string, before map[string]any, requestBody string) string {
			return ComputeDiffFromRequest(before, requestBody)
		},
	})

	// 3) REST：DELETE，路径 .../resource/id
	RegisterAuditHandler(AuditHandler{
		Name: "rest_delete",
		Match: func(method, path string) bool {
			if method != "DELETE" {
				return false
			}
			_, id := ParseResourcePath(path)
			return id > 0
		},
		Before: func(method, path, requestBody string) map[string]any {
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
