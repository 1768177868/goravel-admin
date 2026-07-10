package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type PositionCreate struct {
	Name   string `form:"name" json:"name"`
	Code   string `form:"code" json:"code"`
	Status uint8  `form:"status" json:"status"`
	Sort   int    `form:"sort" json:"sort"`
	Remark string `form:"remark" json:"remark"`
}

func (r *PositionCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *PositionCreate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":   "required|max:50",
		"code":   "max:50",
		"status": "in:0,1",
		"remark": "max:500",
	}
}

func (r *PositionCreate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"name":   trans.Get(ctx, "validation.attributes.name"),
		"code":   trans.Get(ctx, "validation.attributes.code"),
		"status": trans.Get(ctx, "validation.attributes.status"),
		"remark": trans.Get(ctx, "validation.attributes.remark"),
	}
}

func (r *PositionCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
