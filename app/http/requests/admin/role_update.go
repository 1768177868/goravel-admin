package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type RoleUpdate struct {
	Name        *string `form:"name" json:"name"`
	Slug        *string `form:"slug" json:"slug"`
	Description *string `form:"description" json:"description"`
	Status      *uint8  `form:"status" json:"status"`
	Sort        *int    `form:"sort" json:"sort"`
}

func (r *RoleUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *RoleUpdate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":        "max:50",
		"slug":        "max:50",
		"description": "max:255",
		"status":      "in:0,1",
	}
}

func (r *RoleUpdate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"name":        trans.Get(ctx, "validation.attributes.name"),
		"slug":        trans.Get(ctx, "validation.attributes.slug"),
		"description": trans.Get(ctx, "validation.attributes.description"),
		"status":      trans.Get(ctx, "validation.attributes.status"),
	}
}

func (r *RoleUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
