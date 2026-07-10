package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type PositionUpdate struct {
	Name   string `form:"name" json:"name"`
	Code   string `form:"code" json:"code"`
	Status uint8  `form:"status" json:"status"`
	Sort   int    `form:"sort" json:"sort"`
	Remark string `form:"remark" json:"remark"`
}

func (r *PositionUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *PositionUpdate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":   "max:50",
		"code":   "max:50",
		"status": "in:0,1",
		"remark": "max:500",
	}
}

func (r *PositionUpdate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"name":   trans.Get(ctx, "validation.attributes.name"),
		"code":   trans.Get(ctx, "validation.attributes.code"),
		"status": trans.Get(ctx, "validation.attributes.status"),
		"remark": trans.Get(ctx, "validation.attributes.remark"),
	}
}

func (r *PositionUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
