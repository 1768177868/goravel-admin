package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type UserUpdate struct {
	Nickname string `form:"nickname" json:"nickname"`
	Email    string `form:"email" json:"email"`
	Phone    string `form:"phone" json:"phone"`
	Password string `form:"password" json:"password"`
	Status   uint8  `form:"status" json:"status"`
}

func (r *UserUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserUpdate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"nickname": "max:50",
		"email":    "email|max:100",
		"phone":    "max:20",
		"password": "min:6|max:50",
		"status":   "in:0,1",
	}
}

func (r *UserUpdate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"nickname": trans.Get(ctx, "attribute_nickname"),
		"email":    trans.Get(ctx, "attribute_email"),
		"phone":    trans.Get(ctx, "attribute_phone"),
		"password": trans.Get(ctx, "attribute_password"),
		"status":   trans.Get(ctx, "attribute_status"),
	}
}

func (r *UserUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
