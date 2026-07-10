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

func (r *UserRegister) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"username": "required|min:3|max:50",
		"password": "required|min:6|max:50",
		"nickname": "max:50",
		"email":    "email|max:100",
		"phone":    "max:20",
	}
}

func (r *UserRegister) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"username": trans.Get(ctx, "validation.attributes.username"),
		"password": trans.Get(ctx, "validation.attributes.password"),
		"nickname": trans.Get(ctx, "validation.attributes.nickname"),
		"email":    trans.Get(ctx, "validation.attributes.email"),
		"phone":    trans.Get(ctx, "validation.attributes.phone"),
	}
}

func (r *UserRegister) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
