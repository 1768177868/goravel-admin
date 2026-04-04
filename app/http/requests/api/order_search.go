package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/http/trans"
)

// OrderSearch C 端订单关键词检索（ES demo）查询参数；分页在控制器中规范化。
type OrderSearch struct {
	Q        string `form:"q" json:"q"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
}

func (r *OrderSearch) Authorize(ctx http.Context) error {
	return nil
}

func (r *OrderSearch) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"q": "max_len:200",
	}
}

func (r *OrderSearch) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"q.max_len": trans.Get(ctx, "validation_order_search_q_max"),
	}
}

func (r *OrderSearch) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"q": trans.Get(ctx, "attribute_order_search_q"),
	}
}

func (r *OrderSearch) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
