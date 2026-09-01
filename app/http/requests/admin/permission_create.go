package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type PermissionCreate struct {
	Name        string `form:"name" json:"name"`
	Slug        string `form:"slug" json:"slug"`
	Method      string `form:"method" json:"method"`
	Path        string `form:"path" json:"path"`
	HTTPMethod  string `form:"http_method" json:"http_method"`
	HTTPPath    string `form:"http_path" json:"http_path"`
	Description string `form:"description" json:"description"`
	Status      uint8  `form:"status" json:"status"`
	Sort        int    `form:"sort" json:"sort"`
	MenuID      uint   `form:"menu_id" json:"menu_id"`
}

func (r *PermissionCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *PermissionCreate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name": "required|max:50",
		"slug": "required|max:100",
	}
}

func (r *PermissionCreate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"name": trans.Get(ctx, "validation.attributes.name"),
		"slug": trans.Get(ctx, "validation.attributes.slug"),
	}
}

func (r *PermissionCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	if err := helpers.PrepareNumericFieldForValidation(data, "status"); err != nil {
		return err
	}
	return helpers.PrepareNumericFieldForValidation(data, "menu_id")
}
