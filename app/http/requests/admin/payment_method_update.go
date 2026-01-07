package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type PaymentMethodUpdate struct {
	Name        string         `form:"name" json:"name"`
	Config      map[string]any `form:"config" json:"config"`
	IsActive    bool           `form:"is_active" json:"is_active"`
	Sort        int            `form:"sort" json:"sort"`
	Description string         `form:"description" json:"description"`
}

func (r *PaymentMethodUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *PaymentMethodUpdate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":      "required|max_len:50",
		"is_active": "boolean",
		"sort":      "min:0",
	}
}

func (r *PaymentMethodUpdate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":     trans.Get(ctx, "validation_name_required"),
		"name.max_len":      trans.Get(ctx, "validation_name_max"),
		"is_active.boolean": trans.Get(ctx, "validation_boolean"),
		"sort.min":          trans.Get(ctx, "validation_min"),
	}
}

func (r *PaymentMethodUpdate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"name":      trans.Get(ctx, "validation_name"),
		"is_active": trans.Get(ctx, "validation_is_active"),
		"sort":      trans.Get(ctx, "validation_sort"),
	}
}

// func (r *PaymentMethodUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
// 	// return helpers.PrepareNumericFieldForValidation(data, "sort")
// 	if val, exist := data.Get("sort"); !exist || val == nil || val == "" {
// 		return data.Set("sort", 0)
// 	}
// 	return nil
// }
