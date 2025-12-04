package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type BlacklistUpdate struct {
	IP     string `form:"ip" json:"ip"`
	Remark string `form:"remark" json:"remark"`
	Status uint8  `form:"status" json:"status"`
}

func (r *BlacklistUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *BlacklistUpdate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"status": "in:0,1",
	}
}

func (r *BlacklistUpdate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"status.in": trans.Get(ctx, "validation_status_in"),
	}
}

func (r *BlacklistUpdate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"status": trans.Get(ctx, "validation_status"),
	}
}

func (r *BlacklistUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}

