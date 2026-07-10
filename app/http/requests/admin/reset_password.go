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

func (r *ResetPassword) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"password": "required|min:6",
	}
}

func (r *ResetPassword) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"password": trans.Get(ctx, "validation.attributes.password"),
	}
}
