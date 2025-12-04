package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type AdminUpdate struct {
	Nickname     string `form:"nickname" json:"nickname"`
	Email        string `form:"email" json:"email"`
	Phone        string `form:"phone" json:"phone"`
	Password     string `form:"password" json:"password"`
	DepartmentID uint   `form:"department_id" json:"department_id"`
	Status       uint8  `form:"status" json:"status"`
	RoleIDs      []uint `form:"role_ids" json:"role_ids"`
}

func (r *AdminUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *AdminUpdate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname": "max_len:50",
		"email":    "email|max_len:100",
		"phone":    "max_len:20",
		"password": "min_len:6|max_len:50",
		"status":   "in:0,1",
	}
}

func (r *AdminUpdate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname.max_len": trans.Get(ctx, "validation_nickname_max"),
		"email.email":      trans.Get(ctx, "validation_email_format"),
		"email.max_len":    trans.Get(ctx, "validation_email_max"),
		"phone.max_len":    trans.Get(ctx, "validation_phone_max"),
		"password.min_len": trans.Get(ctx, "validation_password_min"),
		"password.max_len": trans.Get(ctx, "validation_password_max"),
		"status.in":        trans.Get(ctx, "validation_status_in"),
	}
}

func (r *AdminUpdate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname": trans.Get(ctx, "validation_nickname"),
		"email":    trans.Get(ctx, "validation_email"),
		"phone":    trans.Get(ctx, "validation_phone"),
		"password": trans.Get(ctx, "validation_password"),
		"status":   trans.Get(ctx, "validation_status"),
	}
}

func (r *AdminUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	// 将 status 字段转换为字符串，以便 in 规则能正确验证
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
