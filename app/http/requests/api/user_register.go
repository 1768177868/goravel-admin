package api

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type UserRegister struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
	Nickname string `form:"nickname" json:"nickname"`
	Email    string `form:"email" json:"email"`
	Phone    string `form:"phone" json:"phone"`
}

func (r *UserRegister) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserRegister) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "required|min_len:3|max_len:50|not_exists:users,username",
		"password": "required|min_len:6|max_len:50",
		"nickname": "max_len:50",
		"email":    "email|max_len:100|not_exists:users,email",
		"phone":    "max_len:20",
	}
}

func (r *UserRegister) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required":   trans.Get(ctx, "validation.required.username"),
		"username.min_len":    trans.GetReplace(ctx, "validation.min.username", map[string]string{"min": "3"}),
		"username.max_len":    trans.GetReplace(ctx, "validation.max.username", map[string]string{"max": "50"}),
		"username.not_exists": trans.Get(ctx, "username_exists"),
		"password.required":   trans.Get(ctx, "validation.required.password"),
		"password.min_len":    trans.GetReplace(ctx, "validation.min.password", map[string]string{"min": "6"}),
		"password.max_len":    trans.GetReplace(ctx, "validation.max.password", map[string]string{"max": "50"}),
		"nickname.max_len":    trans.GetReplace(ctx, "validation.max.nickname", map[string]string{"max": "50"}),
		"email.email":         trans.Get(ctx, "validation.email"),
		"email.max_len":       trans.GetReplace(ctx, "validation.max.email", map[string]string{"max": "100"}),
		"email.not_exists":    trans.Get(ctx, "email_already_exists"),
		"phone.max_len":       trans.GetReplace(ctx, "validation.max.phone", map[string]string{"max": "20"}),
	}
}

func (r *UserRegister) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"username": trans.Get(ctx, "attribute_username"),
		"password": trans.Get(ctx, "attribute_password"),
		"nickname": trans.Get(ctx, "attribute_nickname"),
		"email":    trans.Get(ctx, "attribute_email"),
		"phone":    trans.Get(ctx, "attribute_phone"),
	}
}

func (r *UserRegister) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
