package utils

import (
	"fmt"
	"reflect"
)

// GetValue 从 map[string]any 中安全地获取指定类型的值
// 支持多种类型转换，如果转换失败返回零值和 false
func GetValue[T any](m map[string]any, key string) (T, bool) {
	var zero T
	val, ok := m[key]
	if !ok {
		return zero, false
	}

	// 尝试直接类型断言
	if v, ok := val.(T); ok {
		return v, true
	}

	// 对于数字类型，尝试从其他数字类型转换
	return convertNumeric[T](val)
}

// GetUint 从 map[string]any 中获取 uint 值（支持多种数字类型转换）
func GetUint(m map[string]any, key string) (uint, bool) {
	return GetValue[uint](m, key)
}

// GetFloat64 从 map[string]any 中获取 float64 值（支持多种数字类型转换）
func GetFloat64(m map[string]any, key string) (float64, bool) {
	return GetValue[float64](m, key)
}

// GetString 从 map[string]any 中获取 string 值
func GetString(m map[string]any, key string) (string, bool) {
	val, ok := m[key]
	if !ok {
		return "", false
	}
	if v, ok := val.(string); ok {
		return v, true
	}
	return "", false
}

// GetMap 从 map[string]any 中获取 map[string]any 值
func GetMap(m map[string]any, key string) (map[string]any, bool) {
	val, ok := m[key]
	if !ok {
		return nil, false
	}
	if v, ok := val.(map[string]any); ok {
		return v, true
	}
	return nil, false
}

// convertNumeric 将值转换为数字类型（支持多种数字类型）
func convertNumeric[T any](val any) (T, bool) {
	var zero T
	switch v := val.(type) {
	case float64:
		return convertFromFloat64[T](v)
	case int:
		return convertFromInt[T](v)
	case uint:
		return convertFromUint[T](v)
	case int64:
		return convertFromInt64[T](v)
	case uint64:
		return convertFromUint64[T](v)
	default:
		return zero, false
	}
}

// convertFromFloat64 从 float64 转换
func convertFromFloat64[T any](v float64) (T, bool) {
	var zero T
	switch any(zero).(type) {
	case uint:
		return any(uint(v)).(T), true
	case int:
		return any(int(v)).(T), true
	case float64:
		return any(v).(T), true
	default:
		return zero, false
	}
}

// convertFromInt 从 int 转换
func convertFromInt[T any](v int) (T, bool) {
	var zero T
	switch any(zero).(type) {
	case uint:
		return any(uint(v)).(T), true
	case int:
		return any(v).(T), true
	case float64:
		return any(float64(v)).(T), true
	default:
		return zero, false
	}
}

// convertFromUint 从 uint 转换
func convertFromUint[T any](v uint) (T, bool) {
	var zero T
	switch any(zero).(type) {
	case uint:
		return any(v).(T), true
	case int:
		return any(int(v)).(T), true
	case float64:
		return any(float64(v)).(T), true
	default:
		return zero, false
	}
}

// convertFromInt64 从 int64 转换
func convertFromInt64[T any](v int64) (T, bool) {
	var zero T
	switch any(zero).(type) {
	case uint:
		return any(uint(v)).(T), true
	case int:
		return any(int(v)).(T), true
	case float64:
		return any(float64(v)).(T), true
	default:
		return zero, false
	}
}

// convertFromUint64 从 uint64 转换
func convertFromUint64[T any](v uint64) (T, bool) {
	var zero T
	switch any(zero).(type) {
	case uint:
		return any(uint(v)).(T), true
	case int:
		return any(int(v)).(T), true
	case float64:
		return any(float64(v)).(T), true
	default:
		return zero, false
	}
}

// MustGetValue 从 map[string]any 中获取值，如果不存在或类型不匹配则 panic
// 仅在确定值存在且类型正确时使用
func MustGetValue[T any](m map[string]any, key string) T {
	val, ok := GetValue[T](m, key)
	if !ok {
		panic(fmt.Sprintf("key %s not found or type mismatch in map", key))
	}
	return val
}

// FillFiltersFromMap 从 map[string]any 填充 Filters 结构体
// 支持 string, uint, float64 类型，使用字段名的 snake_case 作为 map 的 key
// 示例：
//
//	filters := services.OrderFilters{}
//	utils.FillFiltersFromMap(m, &filters)
func FillFiltersFromMap(m map[string]any, filtersPtr any) {
	v := reflect.ValueOf(filtersPtr)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		structField := t.Field(i)

		// 获取 json tag 或使用 snake_case 字段名
		key := structField.Tag.Get("json")
		if key == "" || key == "-" {
			key = toSnakeCase(structField.Name)
		}

		switch field.Kind() {
		case reflect.String:
			if val, ok := GetString(m, key); ok {
				field.SetString(val)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if val, ok := GetUint(m, key); ok {
				field.SetUint(uint64(val))
			}
		case reflect.Float64, reflect.Float32:
			if val, ok := GetFloat64(m, key); ok {
				field.SetFloat(val)
			}
		}
	}
}

// toSnakeCase 将 PascalCase/camelCase 转换为 snake_case
func toSnakeCase(s string) string {
	var result []byte
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, byte(c+'a'-'A'))
		} else {
			result = append(result, byte(c))
		}
	}
	return string(result)
}
