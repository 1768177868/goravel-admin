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

func (r *UserUpdate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname": "max_len:50",
		"email":    "email|max_len:100",
		"phone":    "max_len:20",
		"password": "min_len:6|max_len:50",
		"status":   "in:0,1",
	}
}

func (r *UserUpdate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname.max_len": trans.Get(ctx, "validation.max.nickname", map[string]string{"max": "50"}),
		"email.email":      trans.Get(ctx, "validation.email"),
		"email.max_len":    trans.Get(ctx, "validation.max.email", map[string]string{"max": "100"}),
		"phone.max_len":    trans.Get(ctx, "validation.max.phone", map[string]string{"max": "20"}),
		"password.min_len": trans.Get(ctx, "validation.min.password", map[string]string{"min": "6"}),
		"password.max_len": trans.Get(ctx, "validation.max.password", map[string]string{"max": "50"}),
		"status.in":        trans.Get(ctx, "validation.in.status", map[string]string{"values": "0,1"}),
	}
}

func (r *UserUpdate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
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
