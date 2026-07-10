package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/http/trans"
)

// OrderSearch C 端「我的订单」搜索查询参数。
// 分页：helpers.ValidatePaginationEx 或 GET 场景 helpers.PaginationFromQuery。
// 可选 created_from / created_to：YYYY-MM-DD 或 YYYY-MM-DD HH:MM:SS；数据库模式下无时间参数时默认近 3 个月，ES 模式下无参数则不在 ES 侧按时间过滤。
type OrderSearch struct {
	Q           string `form:"q" json:"q"`
	Page        int    `form:"page" json:"page"`
	PageSize    int    `form:"page_size" json:"page_size"`
	CreatedFrom string `form:"created_from" json:"created_from"`
	CreatedTo   string `form:"created_to" json:"created_to"`
}

func (r *OrderSearch) Authorize(ctx http.Context) error {
	return nil
}

func (r *OrderSearch) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"q":            "max:200",
		"created_from": "max:32",
		"created_to":   "max:32",
	}
}

func (r *OrderSearch) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"q":            trans.Get(ctx, "validation.attributes.keyword"),
		"created_from": trans.Get(ctx, "validation.attributes.start_time"),
		"created_to":   trans.Get(ctx, "validation.attributes.end_time"),
	}
}

func (r *OrderSearch) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
