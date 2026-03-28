package utils

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// FieldChange 记录单个字段的变更。
type FieldChange struct {
	Field string `json:"field"`
	Old   any    `json:"old"`
	New   any    `json:"new"`
}

// auditModelFactories 审计变更对比的模型工厂。
// 新增审计表只需在此追加一行即可。
var auditModelFactories = map[string]func() any{
	"admins":      func() any { return &models.Admin{} },
	"roles":       func() any { return &models.Role{} },
	"menus":       func() any { return &models.Menu{} },
	"departments": func() any { return &models.Department{} },
	"blacklists":  func() any { return &models.Blacklist{} },
	"configs":     func() any { return &models.Config{} },
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
func LoadModelSnapshot(tableName string, id uint) map[string]any {
	factory, ok := auditModelFactories[tableName]
	if !ok || id == 0 {
		return nil
	}
	model := factory()
	if err := facades.Orm().Query().Find(model, id); err != nil {
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
