package helpers

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/carbon"
)

// ConvertTimesInData 递归转换数据中的时间字段到对应时区
// 使用 JSON 序列化和反序列化来确保正确处理所有类型
func ConvertTimesInData(ctx http.Context, data any) any {
	if data == nil {
		return nil
	}

	// 获取请求的时区
	timezone := GetCurrentTimezone(ctx)

	// 检查是否传了时区请求头
	hasTimezoneHeader := ctx.Request().Header("X-Timezone", "") != "" ||
		ctx.Request().Header("Timezone", "") != "" ||
		ctx.Request().Input("timezone") != ""

	// 如果没有传时区请求头，且时区是 UTC，直接返回原数据（不做转换）
	if !hasTimezoneHeader && (timezone == carbon.UTC || timezone == "UTC") {
		return data
	}

	// 先序列化为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		// 如果序列化失败，尝试使用反射方法
		return convertTimesInValue(reflect.ValueOf(data), timezone)
	}

	// 反序列化为 map[string]any
	var result any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		// 如果反序列化失败，返回原数据
		return data
	}

	// 转换时间字段
	converted := convertTimesInMap(result, timezone)

	return converted
}

// convertTimesInMap 递归处理 map 或 slice 中的时间字段
func convertTimesInMap(data any, timezone string) any {
	if data == nil {
		return nil
	}

	switch v := data.(type) {
	case map[string]any:
		result := make(map[string]any)
		for key, value := range v {
			// 检查是否是时间字段
			if isTimeField(key) {
				// 尝试解析时间字符串并转换
				if timeStr, ok := value.(string); ok && timeStr != "" {
					// 如果时区是 UTC，直接返回原时间字符串（不做转换）
					if timezone == carbon.UTC || timezone == "UTC" {
						result[key] = timeStr
						continue
					}
					// 否则进行时区转换
					converted := convertTimeString(timeStr, timezone)
					if converted != nil && converted != "" {
						result[key] = converted
						continue
					}
					// 如果转换失败，保留原值
					result[key] = timeStr
					continue
				}
			}
			// 递归处理嵌套数据
			result[key] = convertTimesInMap(value, timezone)
		}
		return result

	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = convertTimesInMap(item, timezone)
		}
		return result

	default:
		return data
	}
}

// convertTimeString 转换时间字符串到指定时区
// 假设数据库存储的时间是 UTC 时区（如：2025-11-22 06:21:25）
func convertTimeString(timeStr string, timezone string) any {
	if timeStr == "" || timeStr == "null" {
		return nil
	}

	// 如果目标时区是 UTC，直接返回原时间字符串
	if timezone == carbon.UTC || timezone == "UTC" {
		return timeStr
	}

	// 加载时区
	utcLoc, _ := time.LoadLocation("UTC")
	targetLoc, err := time.LoadLocation(timezone)
	if err != nil {
		return timeStr
	}

	// 解析时间字符串为 UTC（数据库存储格式）
	t, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, utcLoc)
	if err != nil {
		// 如果标准格式失败，尝试其他格式
		t, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return timeStr
		}
		// RFC3339 格式可能带时区，转换为 UTC
		t = time.Unix(t.Unix(), 0).In(utcLoc)
	}

	// 转换到目标时区并格式化
	return t.In(targetLoc).Format("2006-01-02 15:04:05")
}

// convertTimesInValue 使用反射方法处理值（作为备用方案）
func convertTimesInValue(v reflect.Value, timezone string) any {
	if !v.IsValid() {
		return nil
	}

	// 处理指针
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		return convertTimesInValue(v.Elem(), timezone)
	}

	// 处理时间类型
	if v.Type() == reflect.TypeOf((*carbon.DateTime)(nil)).Elem() {
		dt := v.Interface().(carbon.DateTime)
		return dt.SetTimezone(timezone).ToDateTimeString()
	}

	// 处理 *carbon.DateTime
	if v.Type() == reflect.TypeOf((*carbon.DateTime)(nil)) {
		if v.IsNil() {
			return nil
		}
		dt := v.Interface().(*carbon.DateTime)
		if dt == nil {
			return nil
		}
		return dt.SetTimezone(timezone).ToDateTimeString()
	}

	// 处理 time.Time
	if v.Type() == reflect.TypeOf(time.Time{}) {
		t := v.Interface().(time.Time)
		dt := carbon.NewDateTime(carbon.Parse(t.Format("2006-01-02 15:04:05")))
		return dt.SetTimezone(timezone).ToDateTimeString()
	}

	// 处理 *time.Time
	if v.Type() == reflect.TypeOf((*time.Time)(nil)) {
		if v.IsNil() {
			return nil
		}
		t := v.Interface().(*time.Time)
		if t == nil {
			return nil
		}
		dt := carbon.NewDateTime(carbon.Parse(t.Format("2006-01-02 15:04:05")))
		return dt.SetTimezone(timezone).ToDateTimeString()
	}

	// 处理切片
	if v.Kind() == reflect.Slice {
		if v.IsNil() {
			return nil
		}
		result := make([]any, v.Len())
		for i := range result {
			result[i] = convertTimesInValue(v.Index(i), timezone)
		}
		return result
	}

	// 处理数组
	if v.Kind() == reflect.Array {
		result := make([]any, v.Len())
		for i := range result {
			result[i] = convertTimesInValue(v.Index(i), timezone)
		}
		return result
	}

	// 处理 map
	if v.Kind() == reflect.Map {
		if v.IsNil() {
			return nil
		}
		result := make(map[string]any)
		for _, key := range v.MapKeys() {
			keyStr := key.String()
			if key.Kind() == reflect.Interface {
				keyStr = reflect.ValueOf(key.Interface()).String()
			}
			result[keyStr] = convertTimesInValue(v.MapIndex(key), timezone)
		}
		return result
	}

	// 处理结构体
	if v.Kind() == reflect.Struct {
		result := make(map[string]any)
		t := v.Type()
		for i := range make([]int, v.NumField()) {
			field := t.Field(i)
			fieldValue := v.Field(i)

			// 跳过未导出字段
			if !fieldValue.CanInterface() {
				continue
			}

			fieldName := field.Name
			// 检查 json tag
			if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
				// 解析 json tag（处理 "name,omitempty" 格式）
				parts := strings.Split(jsonTag, ",")
				if len(parts) > 0 && parts[0] != "" {
					fieldName = parts[0]
				}
			}

			// 只处理时间相关字段
			if isTimeField(fieldName) || isTimeType(fieldValue.Type()) {
				result[fieldName] = convertTimesInValue(fieldValue, timezone)
			} else {
				// 递归处理嵌套结构
				if fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Ptr || fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Map {
					result[fieldName] = convertTimesInValue(fieldValue, timezone)
				} else {
					result[fieldName] = fieldValue.Interface()
				}
			}
		}
		return result
	}

	// 其他类型直接返回
	return v.Interface()
}

// isTimeField 检查字段名是否是时间字段
func isTimeField(fieldName string) bool {
	return fieldName == "created_at" || fieldName == "updated_at" || fieldName == "deleted_at" ||
		fieldName == "CreatedAt" || fieldName == "UpdatedAt" || fieldName == "DeletedAt"
}

// isTimeType 检查类型是否是时间类型
func isTimeType(t reflect.Type) bool {
	if t == reflect.TypeOf((*carbon.DateTime)(nil)).Elem() ||
		t == reflect.TypeOf((*carbon.DateTime)(nil)) ||
		t == reflect.TypeOf(time.Time{}) ||
		t == reflect.TypeOf((*time.Time)(nil)) {
		return true
	}
	return false
}
