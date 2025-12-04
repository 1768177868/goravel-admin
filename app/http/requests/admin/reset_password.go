package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type ResetPassword struct {
	Password string `form:"password" json:"password"`
}

func (r *ResetPassword) Authorize(ctx http.Context) error {
	return nil
}

func (r *ResetPassword) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"password": "required|min_len:6",
	}
}

func (r *ResetPassword) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"password.required": trans.Get(ctx, "validation_password_required"),
		"password.min_len":  trans.Get(ctx, "validation_password_min"),
	}
}

func (r *ResetPassword) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"password": trans.Get(ctx, "validation_password"),
	}
}
