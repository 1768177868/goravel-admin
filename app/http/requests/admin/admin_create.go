package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type AdminCreate struct {
	Username     string `form:"username" json:"username" binding:"required" example:"admin"`        // 登录用户名（必填，3-50字符，系统内唯一）
	Password     string `form:"password" json:"password" binding:"required" example:"123456"`       // 登录密码（必填，6-50字符）
	Nickname     string `form:"nickname" json:"nickname" example:"管理员"`                             // 显示昵称（可选，最大50字符）
	Email        string `form:"email" json:"email" example:"admin@example.com"`                     // 联系邮箱（可选，需符合邮箱格式）
	Phone        string `form:"phone" json:"phone" example:"13800138000"`                           // 联系手机号（可选，最大20字符）
	DepartmentID uint   `form:"department_id" json:"department_id" example:"1"`                     // 所属部门ID（可选，关联部门表主键）
	PositionID   uint   `form:"position_id" json:"position_id" example:"1"`                         // 所属岗位ID（可选，关联岗位表主键）
	Status       uint8  `form:"status" json:"status" enums:"0,1" example:"1"`                       // 账号状态（可选，1-启用，0-禁用）
	RoleIDs      []uint `form:"role_ids" json:"role_ids" swaggertype:"array,integer" example:"1,2"` // 角色ID数组（可选，关联角色表主键）
}

func (r *AdminCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *AdminCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "required|min_len:3|max_len:50",
		"password": "required|min_len:6|max_len:50",
		"nickname": "max_len:50",
		"email":    "email|max_len:100",
		"phone":    "max_len:20",
		"status":   "in:0,1",
	}
}

func (r *AdminCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required": trans.Get(ctx, "validation.required.username"),
		"username.min_len":  trans.Get(ctx, "validation.min.username", map[string]string{"min": "3"}),
		"username.max_len":  trans.Get(ctx, "validation.max.username", map[string]string{"max": "50"}),
		"password.required": trans.Get(ctx, "validation.required.password"),
		"password.min_len":  trans.Get(ctx, "validation.min.password", map[string]string{"min": "6"}),
		"password.max_len":  trans.Get(ctx, "validation.max.password", map[string]string{"max": "50"}),
		"nickname.max_len":  trans.Get(ctx, "validation.max.nickname", map[string]string{"max": "50"}),
		"email.email":       trans.Get(ctx, "validation.email"),
		"email.max_len":     trans.Get(ctx, "validation.max.email", map[string]string{"max": "100"}),
		"phone.max_len":     trans.Get(ctx, "validation.max.phone", map[string]string{"max": "20"}),
		"status.in":         trans.Get(ctx, "validation.in.status", map[string]string{"values": "0,1"}),
	}
}

func (r *AdminCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"username": trans.Get(ctx, "validation.attributes.username"),
		"password": trans.Get(ctx, "validation.attributes.password"),
		"nickname": trans.Get(ctx, "validation.attributes.nickname"),
		"email":    trans.Get(ctx, "validation.attributes.email"),
		"phone":    trans.Get(ctx, "validation.attributes.phone"),
		"status":   trans.Get(ctx, "validation.attributes.status"),
	}
}

func (r *AdminCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	// 将 status 字段转换为字符串，以便 in 规则能正确验证
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
