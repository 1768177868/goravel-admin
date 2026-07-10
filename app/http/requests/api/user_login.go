package api

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type UserLogin struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

func (r *UserLogin) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserLogin) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"username": "required",
		"password": "required",
	}
}

func (r *UserLogin) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"username": trans.Get(ctx, "attribute_username"),
		"password": trans.Get(ctx, "attribute_password"),
	}
}

func (r *UserLogin) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
