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

func (r *PositionUpdate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":   "max_len:50",
		"code":   "max_len:50",
		"status": "in:0,1",
		"remark": "max_len:500",
	}
}

func (r *PositionUpdate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.max_len":   trans.Get(ctx, "validation_name_max"),
		"code.max_len":   trans.Get(ctx, "validation_code_max"),
		"status.in":      trans.Get(ctx, "validation_status_in"),
		"remark.max_len": trans.Get(ctx, "validation_remark_max"),
	}
}

func (r *PositionUpdate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"name":   trans.Get(ctx, "validation_name"),
		"code":   trans.Get(ctx, "validation_code"),
		"status": trans.Get(ctx, "validation_status"),
		"remark": trans.Get(ctx, "validation_remark"),
	}
}

func (r *PositionUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
