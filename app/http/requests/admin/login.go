package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type Login struct {
	Username      string `form:"username" json:"username"`
	Password      string `form:"password" json:"password"`
	CaptchaID     string `form:"captcha_id" json:"captcha_id"`
	CaptchaAnswer string `form:"captcha_answer" json:"captcha_answer"`
	GoogleCode    string `form:"google_code" json:"google_code"` // 谷歌验证码
}

func (r *Login) Authorize(ctx http.Context) error {
	return nil
}

func (r *Login) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "required",
		"password": "required|min_len:6",
	}
}

func (r *Login) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required": trans.Get(ctx, "validation_username_required"),
		"password.required": trans.Get(ctx, "validation_password_required"),
		"password.min_len":  trans.Get(ctx, "validation_password_min"),
	}
}

func (r *Login) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"username": trans.Get(ctx, "validation_username"),
		"password": trans.Get(ctx, "validation_password"),
	}
}
