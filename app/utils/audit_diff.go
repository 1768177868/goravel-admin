package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	appfacades "goravel/app/facades"

	"goravel/app/models"
)

// FieldChange 记录单个字段的变更。
type FieldChange struct {
	Field string `json:"field"`
	Old   any    `json:"old"`
	New   any    `json:"new"`
}

// auditModelFactories 仅用于「路径中带 ID」的 REST 更新/删除（见 audit_handlers 默认规则）。
// 无路径 ID、按 body 批量更新的接口勿在此登记，应使用 RegisterAuditHandler（Diff 可用 ComputeDiffAgainstNestedMap 等）。
var auditModelFactories = map[string]func() any{
	"admins":      func() any { return &models.Admin{} },
	"roles":       func() any { return &models.Role{} },
	"menus":       func() any { return &models.Menu{} },
	"departments": func() any { return &models.Department{} },
	"positions":   func() any { return &models.Position{} },
	"blacklists":  func() any { return &models.Blacklist{} },
}

// ParseResourcePath 从 REST API 路径中提取资源表名和记录 ID。
// 例："/api/admin/admins/1" → ("admins", 1)
func ParseResourcePath(path string) (string, uint) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return "", 0
	}
	idStr := parts[len(parts)-1]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return "", 0
	}
	resource := parts[len(parts)-2]
	return strings.ReplaceAll(resource, "-", "_"), uint(id)
}

// LoadModelSnapshot 根据请求路径中的表名和 ID，用 ORM 查询修改前的记录并转为 map。
// 返回 nil 表示该表不需要审计或记录不存在。
func LoadModelSnapshot(ctx context.Context, tableName string, id uint) map[string]any {
	factory, ok := auditModelFactories[tableName]
	if !ok || id == 0 {
		return nil
	}
	model := factory()
	if err := appfacades.OrmQuery(ctx).Find(model, id); err != nil {
		return nil
	}
	snapshot := ModelToMap(model)
	if snapshot == nil || snapshot["id"] == nil || snapshot["id"] == float64(0) {
		return nil
	}
	return snapshot
}

// ModelToMap 通过 JSON 序列化将 model struct 转为 map。
// 所有 key 统一转为 snake_case（兼容有/无 json tag 的模型），自动过滤敏感字段。
func ModelToMap(model any) map[string]any {
	data, err := json.Marshal(model)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	result := make(map[string]any, len(m))
	for key, val := range m {
		snakeKey := toSnakeCase(key)
		if !IsSensitiveField(snakeKey) {
			result[snakeKey] = val
		}
	}
	return result
}

// ComputeDiffFromRequest 对比模型快照与请求体，返回变更 JSON。
// 仅对比请求体中包含且快照中存在的非敏感字段。
func ComputeDiffFromRequest(before map[string]any, requestBody string) string {
	var request map[string]any
	if err := json.Unmarshal([]byte(requestBody), &request); err != nil {
		return ""
	}
	var changes []FieldChange
	for key, newVal := range request {
		if IsSensitiveField(key) {
			continue
		}
		oldVal, exists := before[key]
		if !exists {
			continue
		}
		if !reflect.DeepEqual(oldVal, newVal) {
			changes = append(changes, FieldChange{Field: key, Old: oldVal, New: newVal})
		}
	}
	if len(changes) == 0 {
		return ""
	}
	data, err := json.Marshal(changes)
	if err != nil {
		return ""
	}
	return string(data)
}

// FormatDeleteSnapshot 将删除操作的记录快照格式化为变更 JSON。
func FormatDeleteSnapshot(before map[string]any) string {
	changes := []FieldChange{{Field: "_action", Old: before, New: nil}}
	data, err := json.Marshal(changes)
	if err != nil {
		return ""
	}
	return string(data)
}

// LoadConfigSnapshotByGroup 加载某分组下全部配置项，返回 key → value（用于 configs/save 等批量保存审计）。
func LoadConfigSnapshotByGroup(ctx context.Context, group string) map[string]any {
	if group == "" {
		return nil
	}
	var configs []models.Config
	if err := appfacades.OrmQuery(ctx).Where("group", group).Get(&configs); err != nil || len(configs) == 0 {
		return nil
	}
	snapshot := make(map[string]any, len(configs))
	for _, c := range configs {
		if IsSensitiveField(c.Key) {
			continue
		}
		snapshot[c.Key] = c.Value
	}
	return snapshot
}

// ComputeDiffAgainstNestedMap 将请求体中 nestedKey 对应的对象（如 configs）与扁平快照 before（key→旧值）对比。
// 用于无 REST 路径 ID、但在 body 中带嵌套 map 的保存接口。
func ComputeDiffAgainstNestedMap(before map[string]any, requestBody, nestedKey string) string {
	var body map[string]any
	if err := json.Unmarshal([]byte(requestBody), &body); err != nil {
		return ""
	}
	nestedRaw, ok := body[nestedKey]
	if !ok {
		return ""
	}
	nested, ok := nestedRaw.(map[string]any)
	if !ok {
		return ""
	}
	if before == nil {
		before = map[string]any{}
	}
	var changes []FieldChange
	for key, newVal := range nested {
		if IsSensitiveField(key) {
			continue
		}
		oldVal, exists := before[key]
		if !exists {
			changes = append(changes, FieldChange{Field: key, Old: nil, New: newVal})
			continue
		}
		if !configValuesEqual(oldVal, newVal) {
			changes = append(changes, FieldChange{Field: key, Old: oldVal, New: newVal})
		}
	}
	if len(changes) == 0 {
		return ""
	}
	data, err := json.Marshal(changes)
	if err != nil {
		return ""
	}
	return string(data)
}

// configValuesEqual 对比配置项旧值（库中多为字符串）与新值（JSON 可能为 bool/数字）。
func configValuesEqual(oldVal, newVal any) bool {
	return normalizeConfigValue(oldVal) == normalizeConfigValue(newVal)
}

func normalizeConfigValue(v any) string {
	switch x := v.(type) {
	case bool:
		if x {
			return "1"
		}
		return "0"
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(x))
	}
}
