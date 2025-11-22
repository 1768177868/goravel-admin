package admin

import (
	"github.com/goravel/framework/contracts/http"
)

type AdminCreate struct {
	Username     string `form:"username" json:"username"`
	Password     string `form:"password" json:"password"`
	Nickname     string `form:"nickname" json:"nickname"`
	Email        string `form:"email" json:"email"`
	Phone        string `form:"phone" json:"phone"`
	DepartmentID uint   `form:"department_id" json:"department_id"`
	Status       uint8  `form:"status" json:"status"`
	RoleIDs      []uint `form:"role_ids" json:"role_ids"`
}

func (r *AdminCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *AdminCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "required|min:3|max:50",
		"password": "required|min:6|max:50",
		"nickname": "max:50",
		"email":    "email|max:100",
		"phone":    "max:20",
		"status":   "in:0,1",
	}
}

func (r *AdminCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required": "用户名不能为空",
		"username.min":      "用户名长度不能少于3位",
		"username.max":      "用户名长度不能超过50位",
		"password.required": "密码不能为空",
		"password.min":      "密码长度不能少于6位",
		"password.max":      "密码长度不能超过50位",
		"email.email":       "邮箱格式不正确",
	}
}

func (r *AdminCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "用户名",
		"password": "密码",
		"nickname": "昵称",
		"email":    "邮箱",
		"phone":    "手机号",
	}
}

