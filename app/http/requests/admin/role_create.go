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
		"name.required":       trans.Get(ctx, "validation.required.name"),
		"name.max_len":        trans.Get(ctx, "validation.max.name", map[string]string{"max": "50"}),
		"slug.required":       trans.Get(ctx, "validation.required.slug"),
		"slug.max_len":        trans.Get(ctx, "validation.max.slug", map[string]string{"max": "50"}),
		"description.max_len": trans.Get(ctx, "validation.max.description", map[string]string{"max": "255"}),
		"status.in":           trans.Get(ctx, "validation.in.status", map[string]string{"values": "0,1"}),
	}
}

func (r *RoleCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"name":        trans.Get(ctx, "validation.attributes.name"),
		"slug":        trans.Get(ctx, "validation.attributes.slug"),
		"description": trans.Get(ctx, "validation.attributes.description"),
		"status":      trans.Get(ctx, "validation.attributes.status"),
	}
}

func (r *RoleCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
