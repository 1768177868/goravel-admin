package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type UpdatePassword struct {
	OldPassword     string `form:"old_password" json:"old_password"`
	NewPassword     string `form:"new_password" json:"new_password"`
	ConfirmPassword string `form:"confirm_password" json:"confirm_password"`
}

func (r *UpdatePassword) Authorize(ctx http.Context) error {
	return nil
}

func (r *UpdatePassword) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"old_password":     "required",
		"new_password":     "required|min_len:6",
		"confirm_password": "required|same:new_password",
	}
}

func (r *UpdatePassword) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"old_password.required":     trans.Get(ctx, "validation_old_password_required"),
		"new_password.required":     trans.Get(ctx, "validation_new_password_required"),
		"new_password.min_len":      trans.Get(ctx, "validation_password_min"),
		"confirm_password.required": trans.Get(ctx, "validation_confirm_password_required"),
		"confirm_password.same":     trans.Get(ctx, "validation_password_not_match"),
	}
}

func (r *UpdatePassword) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"old_password":     trans.Get(ctx, "validation_old_password"),
		"new_password":     trans.Get(ctx, "validation_new_password"),
		"confirm_password": trans.Get(ctx, "validation_confirm_password"),
	}
}

