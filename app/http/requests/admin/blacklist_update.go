package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type BlacklistUpdate struct {
	IP     *string `form:"ip" json:"ip" example:"192.168.1.1"`
	Remark *string `form:"remark" json:"remark" example:"测试IP"`
	Status *uint8  `form:"status" json:"status" enums:"0,1" example:"1"`
}

func (r *BlacklistUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *BlacklistUpdate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"status": "in:0,1",
	}
}

func (r *BlacklistUpdate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"status": trans.Get(ctx, "validation.attributes.status"),
	}
}

func (r *BlacklistUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
