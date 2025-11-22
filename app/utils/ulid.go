package utils

import (
	"time"

	"github.com/goravel/framework/support/str"
	"github.com/oklog/ulid/v2"
)

// GenerateULID 生成 ULID（默认使用当前时间，与 Laravel 行为一致）
// 返回: ULID 字符串
func GenerateULID() string {
	entropy := ulid.DefaultEntropy()
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

// GenerateULIDWithTime 使用指定时间生成 ULID
// t: 指定的时间
// 返回: ULID 字符串
func GenerateULIDWithTime(t time.Time) string {
	entropy := ulid.DefaultEntropy()
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

// ParseULID 解析 ULID 字符串
// ulidStr: ULID 字符串
// 返回: ULID 对象和错误
func ParseULID(ulidStr string) (ulid.ULID, error) {
	return ulid.Parse(ulidStr)
}

// ParseULIDTime 从 ULID 解析出时间
// ulidStr: ULID 字符串
// 返回: 时间对象和错误
func ParseULIDTime(ulidStr string) (time.Time, error) {
	id, err := ulid.Parse(ulidStr)
	if err != nil {
		return time.Time{}, err
	}
	return ulid.Time(id.Time()), nil
}

// ParseULIDTimeString 从 ULID 解析出时间字符串
// ulidStr: ULID 字符串
// format: 时间格式，如 "2006-01-02 15:04:05"
// 返回: 格式化的时间字符串和错误
func ParseULIDTimeString(ulidStr string, format string) (string, error) {
	t, err := ParseULIDTime(ulidStr)
	if err != nil {
		return "", err
	}
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return t.Format(format), nil
}

// IsValidULID 验证 ULID 字符串是否有效
// 使用 Goravel 框架的字符串库进行验证（与框架的 IsUlid 方法一致）
// 参考: https://www.goravel.dev/zh_CN/digging-deeper/strings.html#isulid
// ulidStr: ULID 字符串
// 返回: 是否有效
func IsValidULID(ulidStr string) bool {
	// 使用 Goravel 框架的字符串库验证（框架只提供验证功能，不提供生成功能）
	return str.Of(ulidStr).IsUlid()
}

// GetULIDTimestamp 获取 ULID 的时间戳（毫秒）
// ulidStr: ULID 字符串
// 返回: Unix 时间戳（毫秒）和错误
func GetULIDTimestamp(ulidStr string) (int64, error) {
	id, err := ulid.Parse(ulidStr)
	if err != nil {
		return 0, err
	}
	return int64(id.Time()), nil
}

