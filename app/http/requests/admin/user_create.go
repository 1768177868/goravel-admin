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

func (r *UserCreate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"username": "required|min:3|max:50|unique:users,username",
		"password": "required|min:6|max:50",
		"nickname": "max:50",
		"email":    "email|max:100|unique:users,email",
		"phone":    "max:20",
		"status":   "in:0,1",
	}
}

func (r *UserCreate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
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
