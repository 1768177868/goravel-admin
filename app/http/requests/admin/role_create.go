package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type RoleCreate struct {
	Name        string `form:"name" json:"name"`
	Slug        string `form:"slug" json:"slug"`
	Description string `form:"description" json:"description"`
	Status      uint8  `form:"status" json:"status"`
	Sort        int    `form:"sort" json:"sort"`
}

func (r *RoleCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *RoleCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        "required|max_len:50",
		"slug":        "required|max_len:50",
		"description": "max_len:255",
		"status":      "in:0,1",
	}
}

func (r *RoleCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":  trans.Get(ctx, "validation_name_required"),
		"name.max_len":   trans.Get(ctx, "validation_name_max"),
		"slug.required":  trans.Get(ctx, "validation_slug_required"),
		"slug.max_len":   trans.Get(ctx, "validation_slug_max"),
		"description.max_len": trans.Get(ctx, "validation_description_max"),
		"status.in":      trans.Get(ctx, "validation_status_in"),
	}
}

func (r *RoleCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        trans.Get(ctx, "validation_name"),
		"slug":        trans.Get(ctx, "validation_slug"),
		"description": trans.Get(ctx, "validation_description"),
		"status":      trans.Get(ctx, "validation_status"),
	}
}

func (r *RoleCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}

