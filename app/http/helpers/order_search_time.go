package helpers

import (
	"fmt"
	"strings"
	"time"
)

const orderESDateLayout = "2006-01-02"
const orderESDateTimeLayout = "2006-01-02 15:04:05"

// ParseOrderCreatedAtRangeForES 解析订单 ES 检索可选时间范围；空字符串表示不限制。
// 日期支持 YYYY-MM-DD（起：当天 00:00:00，止：当天 23:59:59）或 YYYY-MM-DD HH:MM:SS（本地时区，与写入 ES 的 ToDateTimeString 一致）。
// 若解析失败，返回 errField + errMsgKey（i18n 键）；errMsgKey 为空表示成功。
func ParseOrderCreatedAtRangeForES(createdFrom, createdTo string) (gte, lte *string, errField, errMsgKey string) {
	fromS := strings.TrimSpace(createdFrom)
	toS := strings.TrimSpace(createdTo)
	if fromS == "" && toS == "" {
		return nil, nil, "", ""
	}

	var fromT, toT time.Time
	var hasFrom, hasTo bool

	if fromS != "" {
		t, err := parseOrderCreatedBound(fromS, true)
		if err != nil {
			return nil, nil, "created_from", "validation_order_search_created_from_invalid"
		}
		fromT = t
		hasFrom = true
		s := fromT.Format(orderESDateTimeLayout)
		gte = &s
	}
	if toS != "" {
		t, err := parseOrderCreatedBound(toS, false)
		if err != nil {
			return nil, nil, "created_to", "validation_order_search_created_to_invalid"
		}
		toT = t
		hasTo = true
		s := toT.Format(orderESDateTimeLayout)
		lte = &s
	}
	if hasFrom && hasTo && fromT.After(toT) {
		return nil, nil, "created_to", "validation_order_search_time_range_inverted"
	}
	return gte, lte, "", ""
}

func parseOrderCreatedBound(s string, isStartBound bool) (time.Time, error) {
	s = strings.TrimSpace(s)
	loc := time.Local
	if len(s) == len(orderESDateLayout) {
		t, err := time.ParseInLocation(orderESDateLayout, s, loc)
		if err != nil {
			return time.Time{}, err
		}
		if isStartBound {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc), nil
		}
		return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, loc), nil
	}
	if len(s) > len(orderESDateLayout) {
		return time.ParseInLocation(orderESDateTimeLayout, s, loc)
	}
	return time.Time{}, fmt.Errorf("invalid datetime length")
}
