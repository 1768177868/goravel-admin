package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type PaymentMethodCreate struct {
	Name        string         `form:"name" json:"name"`
	Code        string         `form:"code" json:"code"`
	Type        string         `form:"type" json:"type"`
	Config      map[string]any `form:"config" json:"config"`
	IsActive    bool           `form:"is_active" json:"is_active"`
	Sort        int            `form:"sort" json:"sort"`
	Description string         `form:"description" json:"description"`
}

func (r *PaymentMethodCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *PaymentMethodCreate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":      "required|max:50",
		"code":      "required|max:20",
		"type":      "required|max:20",
		"config":    "required",
		"is_active": "boolean",
		"sort":      "min:0",
	}
}

func (r *PaymentMethodCreate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"name":      trans.Get(ctx, "validation.attributes.name"),
		"code":      trans.Get(ctx, "validation.attributes.code"),
		"type":      trans.Get(ctx, "validation.attributes.type"),
		"config":    trans.Get(ctx, "validation.attributes.config"),
		"is_active": trans.Get(ctx, "validation.attributes.is_active"),
		"sort":      trans.Get(ctx, "validation.attributes.sort"),
	}
}

// func (r *PaymentMethodCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
// 	// sort 字段使用 integer|min:0 规则，不需要转换为字符串
// 	// 如果 sort 为空或不存在，设置为默认值 0
// 	if val, exist := data.Get("sort"); !exist || val == nil || val == "" {
// 		return data.Set("sort", 0)
// 	}
// 	return nil
// }
