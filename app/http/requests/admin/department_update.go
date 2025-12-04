package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type DepartmentUpdate struct {
	ParentID uint   `form:"parent_id" json:"parent_id"`
	Name     string `form:"name" json:"name"`
	Code     string `form:"code" json:"code"`
	Leader   string `form:"leader" json:"leader"`
	Phone    string `form:"phone" json:"phone"`
	Email    string `form:"email" json:"email"`
	Status   uint8  `form:"status" json:"status"`
	Sort     int    `form:"sort" json:"sort"`
	Remark   string `form:"remark" json:"remark"`
}

func (r *DepartmentUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *DepartmentUpdate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":   "max_len:50",
		"code":   "max_len:50",
		"leader": "max_len:50",
		"phone":  "max_len:20",
		"email":  "email|max_len:100",
		"status": "in:0,1",
		"remark": "max_len:500",
	}
}

func (r *DepartmentUpdate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.max_len":   trans.Get(ctx, "validation_name_max"),
		"code.max_len":   trans.Get(ctx, "validation_code_max"),
		"leader.max_len": trans.Get(ctx, "validation_leader_max"),
		"phone.max_len":  trans.Get(ctx, "validation_phone_max"),
		"email.email":    trans.Get(ctx, "validation_email_format"),
		"email.max_len":  trans.Get(ctx, "validation_email_max"),
		"status.in":      trans.Get(ctx, "validation_status_in"),
		"remark.max_len": trans.Get(ctx, "validation_remark_max"),
	}
}

func (r *DepartmentUpdate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"name":   trans.Get(ctx, "validation_name"),
		"code":   trans.Get(ctx, "validation_code"),
		"leader": trans.Get(ctx, "validation_leader"),
		"phone":  trans.Get(ctx, "validation_phone"),
		"email":  trans.Get(ctx, "validation_email"),
		"status": trans.Get(ctx, "validation_status"),
		"remark": trans.Get(ctx, "validation_remark"),
	}
}

func (r *DepartmentUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}

