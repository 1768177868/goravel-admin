package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type AttachmentCategoryCreate struct {
	Name   string `form:"name" json:"name"`
	Status uint8  `form:"status" json:"status"`
	Sort   int    `form:"sort" json:"sort"`
	Remark string `form:"remark" json:"remark"`
}

func (r *AttachmentCategoryCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *AttachmentCategoryCreate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":   "required|max:50",
		"status": "in:0,1",
		"remark": "max:500",
	}
}

func (r *AttachmentCategoryCreate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"name":   trans.Get(ctx, "validation.attributes.name"),
		"status": trans.Get(ctx, "validation.attributes.status"),
		"remark": trans.Get(ctx, "validation.attributes.remark"),
	}
}

func (r *AttachmentCategoryCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
