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

func (r *Login) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"username": "required",
		"password": "required|min:6",
	}
}

func (r *Login) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"username": trans.Get(ctx, "validation.attributes.username"),
		"password": trans.Get(ctx, "validation.attributes.password"),
	}
}
