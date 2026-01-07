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

func (r *PaymentMethodCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":      "required|max_len:50",
		"code":      "required|max_len:20",
		"type":      "required|max_len:20",
		"config":    "required",
		"is_active": "boolean",
		"sort":      "min:0",
	}
}

func (r *PaymentMethodCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":     trans.Get(ctx, "validation_name_required"),
		"name.max_len":      trans.Get(ctx, "validation_name_max"),
		"code.required":     trans.Get(ctx, "validation_code_required"),
		"code.max_len":      trans.Get(ctx, "validation_code_max"),
		"type.required":     trans.Get(ctx, "validation_type_required"),
		"type.max_len":      trans.Get(ctx, "validation_type_max"),
		"config.required":   trans.Get(ctx, "validation_config_required"),
		"is_active.boolean": trans.Get(ctx, "validation_boolean"),
		"sort.min":          trans.Get(ctx, "validation_min"),
	}
}

func (r *PaymentMethodCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"name":      trans.Get(ctx, "validation_name"),
		"code":      trans.Get(ctx, "validation_code"),
		"type":      trans.Get(ctx, "validation_type"),
		"config":    trans.Get(ctx, "validation_config"),
		"is_active": trans.Get(ctx, "validation_is_active"),
		"sort":      trans.Get(ctx, "validation_sort"),
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
