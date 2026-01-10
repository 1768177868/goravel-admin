package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

// ShardingNoConfig 分表单号配置
type ShardingNoConfig struct {
	Prefix     string // 单号前缀，如 "ORD"、"PAY"
	DateFormat string // 日期格式，如 "200601"（年月）、"20060102"（年月日）
}

// 预定义配置
var (
	// OrderNoConfig 订单号配置（ORD + YYYYMM + ULID）
	OrderNoConfig = ShardingNoConfig{
		Prefix:     "ORD",
		DateFormat: "200601", // YYYYMM
	}

	// PaymentNoConfig 支付单号配置（PAY + YYYYMMDD + ULID）
	PaymentNoConfig = ShardingNoConfig{
		Prefix:     "PAY",
		DateFormat: "20060102", // YYYYMMDD
	}
)

// GenerateShardingNo 生成分表单号
// 格式：前缀 + 日期 + ULID
// 例如：ORD20250101ARZ3S0K5M2X9P4Q6R8T1V3W5Y7Z9
func GenerateShardingNo(config ShardingNoConfig) string {
	now := time.Now().UTC()
	dateStr := now.Format(config.DateFormat)
	ulidStr := ulid.Make().String()
	return fmt.Sprintf("%s%s%s", config.Prefix, dateStr, ulidStr)
}

// ParseShardingNoDate 从分表单号解析日期
// 返回：解析的时间和是否成功
func ParseShardingNoDate(no string, config ShardingNoConfig) (time.Time, bool) {
	prefixLen := len(config.Prefix)
	dateLen := len(config.DateFormat)
	minLen := prefixLen + dateLen + 1 // 前缀 + 日期 + 至少1位ULID

	// 验证长度和前缀
	if len(no) < minLen || !strings.HasPrefix(no, config.Prefix) {
		return time.Time{}, false
	}

	// 提取日期部分
	dateStr := no[prefixLen : prefixLen+dateLen]

	// 验证日期格式（简单验证：都是数字）
	for _, c := range dateStr {
		if c < '0' || c > '9' {
			return time.Time{}, false
		}
	}

	// 解析日期
	parsedTime, err := time.Parse(config.DateFormat, dateStr)
	if err != nil {
		return time.Time{}, false
	}

	return parsedTime, true
}

// ParseShardingNoYearMonth 从分表单号解析年月字符串（用于分表定位）
// 返回：年月字符串（如 "202501"）和是否成功
func ParseShardingNoYearMonth(no string, config ShardingNoConfig) (string, bool) {
	parsedTime, ok := ParseShardingNoDate(no, config)
	if !ok {
		return "", false
	}
	return parsedTime.Format("200601"), true
}
