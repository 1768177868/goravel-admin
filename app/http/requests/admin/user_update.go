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
		"nickname.max_len": trans.Get(ctx, "validation_nickname_max"),
		"email.email":      trans.Get(ctx, "validation_email_format"),
		"email.max_len":    trans.Get(ctx, "validation_email_max"),
		"phone.max_len":    trans.Get(ctx, "validation_phone_max"),
		"password.min_len": trans.Get(ctx, "validation_password_min"),
		"password.max_len": trans.Get(ctx, "validation_password_max"),
		"status.in":        trans.Get(ctx, "validation_status_in"),
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


