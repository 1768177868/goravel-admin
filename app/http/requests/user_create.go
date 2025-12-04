package requests

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"
	"goravel/app/models"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type UserCreate struct {
	Username string           `form:"username" json:"username"`
	Password string           `form:"password" json:"password"`
	Name     string           `form:"name" json:"name"`
	Avatar   string           `form:"avatar" json:"avatar"`
	Alias    string           `form:"alias" json:"alias"`
	Mail     string           `form:"mail" json:"mail"`
	Status   uint8            `form:"status" json:"status"`
	Tags     []models.UserTag `form:"tags" json:"tags"`
}

func (r *UserCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "required|min_len:3|max_len:50",
		"password": "required|min_len:6",
		"name":     "required",
		"mail":     "email",
		"status":   "in:0,1",
	}
}

func (r *UserCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required": trans.Get(ctx, "validation_username_required"),
		"username.min_len":  trans.Get(ctx, "validation_username_min"),
		"username.max_len":  trans.Get(ctx, "validation_username_max"),
		"password.required": trans.Get(ctx, "validation_password_required"),
		"password.min_len":  trans.Get(ctx, "validation_password_min"),
		"name.required":     trans.Get(ctx, "validation_name_required"),
		"mail.email":        trans.Get(ctx, "validation_email_format"),
		"status.in":         trans.Get(ctx, "validation_status_in"),
	}
}

func (r *UserCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"username": trans.Get(ctx, "validation_username"),
		"password": trans.Get(ctx, "validation_password"),
		"name":     trans.Get(ctx, "validation_name"),
		"mail":     trans.Get(ctx, "validation_email"),
		"status":   trans.Get(ctx, "validation_status"),
	}
}

func (r *UserCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	// 将 status 字段转换为字符串，以便 in 规则能正确验证
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
