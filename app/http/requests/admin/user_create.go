package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type UserCreate struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
	Nickname string `form:"nickname" json:"nickname"`
	Email    string `form:"email" json:"email"`
	Phone    string `form:"phone" json:"phone"`
	Status   uint8  `form:"status" json:"status"`
}

func (r *UserCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "required|min_len:3|max_len:50|not_exists:users,username",
		"password": "required|min_len:6|max_len:50",
		"nickname": "max_len:50",
		"email":    "email|max_len:100|not_exists:users,email",
		"phone":    "max_len:20",
		"status":   "in:0,1",
	}
}

func (r *UserCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required":  trans.Get(ctx, "validation_username_required"),
		"username.min_len":   trans.Get(ctx, "validation_username_min"),
		"username.max_len":   trans.Get(ctx, "validation_username_max"),
		"username.not_exists": trans.Get(ctx, "username_exists"),
		"password.required":  trans.Get(ctx, "validation_password_required"),
		"password.min_len":   trans.Get(ctx, "validation_password_min"),
		"password.max_len":   trans.Get(ctx, "validation_password_max"),
		"nickname.max_len":   trans.Get(ctx, "validation_nickname_max"),
		"email.email":        trans.Get(ctx, "validation_email_format"),
		"email.max_len":      trans.Get(ctx, "validation_email_max"),
		"email.not_exists":   trans.Get(ctx, "email_already_exists"),
		"phone.max_len":      trans.Get(ctx, "validation_phone_max"),
		"status.in":          trans.Get(ctx, "validation_status_in"),
	}
}

func (r *UserCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"username": trans.Get(ctx, "attribute_username"),
		"password": trans.Get(ctx, "attribute_password"),
		"nickname": trans.Get(ctx, "attribute_nickname"),
		"email":    trans.Get(ctx, "attribute_email"),
		"phone":    trans.Get(ctx, "attribute_phone"),
		"status":   trans.Get(ctx, "attribute_status"),
	}
}

func (r *UserCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}

