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
		"nickname": trans.Get(ctx, "validation.attributes.nickname"),
		"email":    trans.Get(ctx, "validation.attributes.email"),
		"phone":    trans.Get(ctx, "validation.attributes.phone"),
		"password": trans.Get(ctx, "validation.attributes.password"),
		"status":   trans.Get(ctx, "validation.attributes.status"),
	}
}

func (r *UserUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
