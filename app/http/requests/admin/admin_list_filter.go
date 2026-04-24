package admin

// AdminListFilter 管理员列表与导出共用筛选（GET query / POST JSON body 字段一致，见 buildFilters）
type AdminListFilter struct {
	Username     string `json:"username" form:"username" example:"admin"`                   // 登录用户名（模糊匹配）
	Status       string `json:"status" form:"status" enums:"0,1" example:"1"`               // 账号状态（1-启用，0-禁用）
	RoleID       string `json:"role_id" form:"role_id" example:"2"`                         // 角色ID（精确匹配）
	DepartmentID string `json:"department_id" form:"department_id" example:"1"`             // 部门ID（精确匹配）
	PositionID   string `json:"position_id" form:"position_id" example:"3"`                 // 岗位ID（精确匹配）
	Is2FABound   string `json:"is_2fa_bound" form:"is_2fa_bound" enums:"0,1" example:"1"`   // 是否已绑定2FA（1-已绑定，0-未绑定）
	StartTime    string `json:"start_time" form:"start_time" example:"2024-01-01 00:00:00"` // 创建时间开始（格式：YYYY-MM-DD HH:mm:ss）
	EndTime      string `json:"end_time" form:"end_time" example:"2024-12-31 23:59:59"`     // 创建时间结束（格式：YYYY-MM-DD HH:mm:ss）
	OrderBy      string `json:"order_by" form:"order_by" example:"created_at:desc"`         // 排序字段（格式：字段:asc/desc）
}
