package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type DepartmentCreate struct {
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

func (r *DepartmentCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *DepartmentCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":   "required|max_len:50",
		"code":   "max_len:50",
		"leader": "max_len:50",
		"phone":  "max_len:20",
		"email":  "email|max_len:100",
		"status": "in:0,1",
		"remark": "max_len:500",
	}
}

func (r *DepartmentCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":  trans.Get(ctx, "department_name_required"),
		"name.max_len":   trans.GetReplace(ctx, "validation.max.name", map[string]string{"max": "50"}),
		"code.max_len":   trans.GetReplace(ctx, "validation.max.code", map[string]string{"max": "50"}),
		"leader.max_len": trans.GetReplace(ctx, "validation.max.leader", map[string]string{"max": "50"}),
		"phone.max_len":  trans.GetReplace(ctx, "validation.max.phone", map[string]string{"max": "20"}),
		"email.email":    trans.Get(ctx, "validation.email"),
		"email.max_len":  trans.GetReplace(ctx, "validation.max.email", map[string]string{"max": "100"}),
		"status.in":      trans.GetReplace(ctx, "validation.in.status", map[string]string{"values": "0,1"}),
		"remark.max_len": trans.GetReplace(ctx, "validation.max.remark", map[string]string{"max": "500"}),
	}
}

func (r *DepartmentCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"name":   trans.Get(ctx, "validation.attributes.name"),
		"code":   trans.Get(ctx, "validation.attributes.code"),
		"leader": trans.Get(ctx, "validation.attributes.leader"),
		"phone":  trans.Get(ctx, "validation.attributes.phone"),
		"email":  trans.Get(ctx, "validation.attributes.email"),
		"status": trans.Get(ctx, "validation.attributes.status"),
		"remark": trans.Get(ctx, "validation.attributes.remark"),
	}
}

func (r *DepartmentCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
