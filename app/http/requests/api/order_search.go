package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/http/trans"
)

// OrderSearch C 端订单关键词检索（ES demo）查询参数。
// 分页：控制器中 helpers.ValidatePaginationEx(req.Page, req.PageSize, helpers.PaginationLimits{})；GET 无表单时可用 helpers.PaginationFromQuery。
// 可选 created_from / created_to：YYYY-MM-DD 或 YYYY-MM-DD HH:MM:SS（与写入 ES 的 created_at 格式一致）。
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

func (r *OrderSearch) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"q":             "max_len:200",
		"created_from":  "max_len:32",
		"created_to":    "max_len:32",
	}
}

func (r *OrderSearch) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"q.max_len":            trans.Get(ctx, "validation_order_search_q_max"),
		"created_from.max_len": trans.Get(ctx, "validation_order_search_created_from_max"),
		"created_to.max_len":   trans.Get(ctx, "validation_order_search_created_to_max"),
	}
}

func (r *OrderSearch) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"q":             trans.Get(ctx, "attribute_order_search_q"),
		"created_from":  trans.Get(ctx, "attribute_order_search_created_from"),
		"created_to":    trans.Get(ctx, "attribute_order_search_created_to"),
	}
}

func (r *OrderSearch) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
