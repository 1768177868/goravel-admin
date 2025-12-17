package admin

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/models"
	"goravel/app/services"
)

type AdminController struct {
	adminService               services.AdminService
	googleAuthenticatorService services.GoogleAuthenticatorService
}

// AdminExportRequest 导出管理员请求参数
type AdminExportRequest struct {
	Username     string `json:"username" form:"username" example:"admin"`                   // 用户名（模糊搜索）
	Status       string `json:"status" form:"status" example:"1"`                           // 状态：1-启用，0-禁用
	RoleID       string `json:"role_id" form:"role_id" example:"1"`                         // 角色ID
	DepartmentID string `json:"department_id" form:"department_id" example:"1"`             // 部门ID
	Is2FABound   string `json:"is_2fa_bound" form:"is_2fa_bound" example:"1"`               // 是否绑定2FA：1-已绑定，0-未绑定
	StartTime    string `json:"start_time" form:"start_time" example:"2024-01-01 00:00:00"` // 开始时间
	EndTime      string `json:"end_time" form:"end_time" example:"2024-12-31 23:59:59"`     // 结束时间
	OrderBy      string `json:"order_by" form:"order_by" example:"created_at:desc"`         // 排序
}

// AdminResponse 管理员响应数据
type AdminResponse struct {
	ID           uint             `json:"id" example:"1"`                           // 管理员ID
	Username     string           `json:"username" example:"admin"`                 // 用户名
	Nickname     string           `json:"nickname" example:"管理员"`                   // 昵称
	Avatar       string           `json:"avatar" example:""`                        // 头像
	Email        string           `json:"email" example:"admin@example.com"`        // 邮箱
	Phone        string           `json:"phone" example:"13800138000"`              // 手机号
	Status       uint8            `json:"status" example:"1"`                       // 状态：1-启用，0-禁用
	Is2FABound   bool             `json:"is_2fa_bound" example:"true"`              // 是否绑定2FA
	DepartmentID uint             `json:"department_id" example:"1"`                // 部门ID
	Department   map[string]any   `json:"department"`                               // 部门信息
	Roles        []map[string]any `json:"roles"`                                    // 角色列表
	CreatedAt    string           `json:"created_at" example:"2024-01-01 00:00:00"` // 创建时间
	UpdatedAt    string           `json:"updated_at" example:"2024-01-01 00:00:00"` // 更新时间
}

// PaginatedAdminResponse 分页管理员响应
type PaginatedAdminResponse struct {
	Code     int             `json:"code" example:"200"`                  // 状态码
	Message  string          `json:"message" example:"获取成功"`              // 消息
	Data     []AdminResponse `json:"data"`                                // 数据列表
	Total    int64           `json:"total" example:"100"`                 // 总数
	Page     int             `json:"page" example:"1"`                    // 当前页码
	PageSize int             `json:"page_size" example:"10"`              // 每页数量
	TraceID  string          `json:"trace_id,omitempty" example:"abc123"` // 追踪ID
}

// AdminDetailResponse 管理员详情响应
type AdminDetailResponse struct {
	Code    int           `json:"code" example:"200"`                  // 状态码
	Message string        `json:"message" example:"获取成功"`              // 消息
	Data    AdminResponse `json:"data"`                                // 管理员数据
	TraceID string        `json:"trace_id,omitempty" example:"abc123"` // 追踪ID
}

func NewAdminController() *AdminController {
	return &AdminController{
		adminService:               services.NewAdminServiceImpl(),
		googleAuthenticatorService: services.NewGoogleAuthenticatorServiceImpl(),
	}
}

// findAdminByID 根据ID查找管理员，如果不存在则返回错误响应
// withDepartment 为 true 时会预加载 Department 关联
// withRoles 为 true 时会预加载 Roles 关联
func (r *AdminController) findAdminByID(ctx http.Context, id uint, withDepartment bool, withRoles bool) (*models.Admin, http.Response) {
	var relations []string
	if withDepartment {
		relations = append(relations, "Department")
	}
	if withRoles {
		relations = append(relations, "Roles")
	}
	return response.FindByID[models.Admin](ctx, id, &response.FindByIDOptions{
		WithRelations:      relations,
		NotFoundMessageKey: "admin_not_found",
	})
}

// buildQuery 构建查询（列表和导出共用）
// 同时支持查询参数（GET）和请求体参数（POST）
func (r *AdminController) buildQuery(ctx http.Context) orm.Query {
	// 优先从请求体读取，如果没有则从查询参数读取（兼容 GET 和 POST）
	username := ctx.Request().Input("username", ctx.Request().Query("username", ""))
	status := ctx.Request().Input("status", ctx.Request().Query("status", ""))
	roleID := ctx.Request().Input("role_id", ctx.Request().Query("role_id", ""))
	departmentID := ctx.Request().Input("department_id", ctx.Request().Query("department_id", ""))
	is2FABound := ctx.Request().Input("is_2fa_bound", ctx.Request().Query("is_2fa_bound", ""))
	orderBy := ctx.Request().Input("order_by", ctx.Request().Query("order_by", ""))
	// 时间参数同时支持从请求体和查询参数读取，并转换为 UTC
	startTimeStr := ctx.Request().Input("start_time", ctx.Request().Query("start_time", ""))
	endTimeStr := ctx.Request().Input("end_time", ctx.Request().Query("end_time", ""))
	startTime := ""
	endTime := ""
	if startTimeStr != "" {
		startTime = helpers.ConvertTimeToUTC(ctx, startTimeStr)
	}
	if endTimeStr != "" {
		endTime = helpers.ConvertTimeToUTC(ctx, endTimeStr)
	}

	query := facades.Orm().Query().Model(&models.Admin{})

	developerIDsStr := facades.Config().GetString("admin.developer_ids", "2")
	developerIDs := r.parseProtectedIDs(developerIDsStr)
	if len(developerIDs) > 0 {
		query = query.Where("id NOT IN ?", developerIDs)
	}

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if status != "" {
		query = query.Where("status", status)
	}
	if roleID != "" {
		roleIDUint := cast.ToUint(roleID)
		if roleIDUint > 0 {
			query = query.Where("id IN (SELECT admin_id FROM admin_role WHERE role_id = ?)", roleIDUint)
		}
	}
	if departmentID != "" {
		departmentIDUint := cast.ToUint(departmentID)
		if departmentIDUint > 0 {
			departmentIDs := r.getDepartmentAndChildrenIDs(departmentIDUint)
			if len(departmentIDs) > 0 {
				idsAny := make([]any, len(departmentIDs))
				for i, id := range departmentIDs {
					idsAny[i] = id
				}
				query = query.WhereIn("department_id", idsAny)
			}
		}
	}
	if is2FABound != "" {
		switch is2FABound {
		case "1":
			// 已绑定：google_secret IS NOT NULL AND google_secret != ''
			query = query.Where("google_secret IS NOT NULL AND google_secret != ?", "")
		case "0":
			// 未绑定：google_secret IS NULL OR google_secret = ''
			query = query.Where("(google_secret IS NULL OR google_secret = ?)", "")
		}
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	query = helpers.ApplySort(query, orderBy, "created_at:desc")

	return query
}

// Index 管理员列表
// @Summary      获取管理员列表
// @Description  分页获取管理员列表，支持按用户名、状态、角色、部门等条件筛选
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        page          query     int     false  "页码" default(1)
// @Param        page_size     query     int     false  "每页数量" default(10)
// @Param        username      query     string  false  "用户名（模糊搜索）"
// @Param        status        query     string  false  "状态：1-启用，0-禁用"
// @Param        role_id       query     string  false  "角色ID"
// @Param        department_id query     string  false  "部门ID"
// @Param        start_time    query     string  false  "开始时间（格式：YYYY-MM-DD HH:mm:ss）"
// @Param        end_time      query     string  false  "结束时间（格式：YYYY-MM-DD HH:mm:ss）"
// @Param        order_by      query     string  false  "排序（格式：字段:asc/desc，如：created_at:desc）"
// @Success      200           {object}  PaginatedAdminResponse
// @Failure      400           {object}  map[string]any "参数错误"
// @Failure      401           {object}  map[string]any "未登录"
// @Failure      403           {object}  map[string]any "无权限"
// @Failure      500           {object}  map[string]any "服务器错误"
// @Router       /api/admin/admins [get]
// @Security     BearerAuth
func (r *AdminController) Index(ctx http.Context) http.Response {
	query := r.buildQuery(ctx)
	var admins []models.Admin

	// 获取超级管理员ID
	superAdminID := cast.ToUint(facades.Config().GetInt("admin.super_admin_id", 1))

	return response.PaginateQuery(ctx, query, &admins, &response.PaginateQueryOptions{
		WithRelations: []string{"Department", "Roles"},
		Transform: func(data any) any {
			admins := data.(*[]models.Admin)
			adminList := make([]http.Json, len(*admins))
			for i, admin := range *admins {
				isBound := admin.GoogleSecret != ""
				adminList[i] = http.Json{
					"id":             admin.ID,
					"username":       admin.Username,
					"nickname":       admin.Nickname,
					"avatar":         admin.Avatar,
					"email":          admin.Email,
					"phone":          admin.Phone,
					"status":         admin.Status,
					"is_2fa_bound":   isBound,
					"is_super_admin": admin.ID == superAdminID,
					"department_id":  admin.DepartmentID,
					"department":     admin.Department,
					"roles":          admin.Roles,
					"created_at":     admin.CreatedAt,
					"updated_at":     admin.UpdatedAt,
				}
			}
			return adminList
		},
	})
}

// Show 管理员详情
// @Summary      获取管理员详情
// @Description  根据ID获取管理员详细信息，包括部门、角色等关联信息
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        id   path     int  true  "管理员ID"
// @Success      200  {object} AdminDetailResponse
// @Failure      400  {object} map[string]any "参数错误"
// @Failure      401  {object} map[string]any "未登录"
// @Failure      403  {object} map[string]any "无权限"
// @Failure      404  {object} map[string]any "管理员不存在"
// @Failure      500  {object} map[string]any "服务器错误"
// @Router       /api/admin/admins/{id} [get]
// @Security     BearerAuth
func (r *AdminController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	admin, resp := r.findAdminByID(ctx, id, true, true) // 预加载 Department 和 Roles 关联
	if resp != nil {
		return resp
	}

	// 获取超级管理员ID
	superAdminID := cast.ToUint(facades.Config().GetInt("admin.super_admin_id", 1))

	return response.Success(ctx, http.Json{
		"admin": http.Json{
			"id":             admin.ID,
			"username":       admin.Username,
			"nickname":       admin.Nickname,
			"avatar":         admin.Avatar,
			"email":          admin.Email,
			"phone":          admin.Phone,
			"status":         admin.Status,
			"is_super_admin": admin.ID == superAdminID, // 标识是否是超级管理员
			"department_id":  admin.DepartmentID,
			"department":     admin.Department,
			"roles":          admin.Roles,
			"created_at":     admin.CreatedAt,
			"updated_at":     admin.UpdatedAt,
		},
	})
}

// Store 创建管理员
// @Summary      创建管理员
// @Description  创建新的管理员账号，支持设置部门、角色等信息
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        username      body     string  true  "用户名（必填）" example(admin)
// @Param        password      body     string  true  "密码（必填）" example(123456)
// @Param        nickname      body     string  false "昵称" example(管理员)
// @Param        email         body     string  false "邮箱" example(admin@example.com)
// @Param        phone         body     string  false "手机号" example(13800138000)
// @Param        department_id body     int     false "部门ID" example(1)
// @Param        status        body     int     false "状态：1-启用，0-禁用" example(1)
// @Param        role_ids      body     []int   false "角色ID列表" example([1,2])
// @Success      200           {object} AdminDetailResponse
// @Failure      400           {object} map[string]any "参数错误或用户名已存在"
// @Failure      401           {object} map[string]any "未登录"
// @Failure      403           {object} map[string]any "无权限"
// @Failure      500           {object} map[string]any "服务器错误"
// @Router       /api/admin/admins [post]
// @Security     BearerAuth
func (r *AdminController) Store(ctx http.Context) http.Response {
	// 使用请求验证
	var adminCreate adminrequests.AdminCreate
	errors, err := ctx.Request().ValidateRequest(&adminCreate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 检查用户名是否已存在
	exists, err := facades.Orm().Query().Model(&models.Admin{}).Where("username", adminCreate.Username).Exists()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}
	if exists {
		return response.Error(ctx, http.StatusBadRequest, "username_exists")
	}

	// 加密密码
	hashedPassword, err := facades.Hash().Make(adminCreate.Password)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
	}

	now := carbon.Now()
	adminData := map[string]any{
		"username":      adminCreate.Username,
		"password":      hashedPassword,
		"nickname":      adminCreate.Nickname,
		"avatar":        "",
		"email":         adminCreate.Email,
		"phone":         adminCreate.Phone,
		"department_id": adminCreate.DepartmentID,
		"status":        adminCreate.Status,
		"created_at":    now,
		"updated_at":    now,
	}

	if err := facades.Orm().Query().Table("admins").Create(adminData); err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"username": adminCreate.Username,
		})
	}

	var admin models.Admin
	if err := facades.Orm().Query().Where("username", adminCreate.Username).First(&admin); err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"username": adminCreate.Username,
		})
	}

	if len(adminCreate.RoleIDs) > 0 {
		if err := r.adminService.SyncRoles(&admin, adminCreate.RoleIDs); err != nil {
			return response.ErrorWithLog(ctx, "admin", err, map[string]any{
				"admin_id": admin.ID,
				"role_ids": adminCreate.RoleIDs,
			})
		}
	}

	return response.Success(ctx, http.Json{
		"admin": admin,
	})
}

// Update 更新管理员
// @Summary      更新管理员信息
// @Description  更新管理员的基本信息，包括昵称、邮箱、手机号、部门、状态、角色等。受保护的管理员不能禁用。
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        id            path     int     true  "管理员ID" example(1)
// @Param        nickname      body     string  false "昵称" example(管理员)
// @Param        email         body     string  false "邮箱" example(admin@example.com)
// @Param        phone         body     string  false "手机号" example(13800138000)
// @Param        department_id body     int     false "部门ID" example(1)
// @Param        status        body     string  false "状态：1-启用，0-禁用" example(1)
// @Param        password      body     string  false "密码（可选，不传则不更新）" example(123456)
// @Param        role_ids      body     []int   false "角色ID列表" example([1,2])
// @Success      200           {object} AdminDetailResponse
// @Failure      400           {object} map[string]any "参数错误"
// @Failure      401           {object} map[string]any "未登录"
// @Failure      403           {object} map[string]any "无权限或受保护管理员不能禁用"
// @Failure      404           {object} map[string]any "管理员不存在"
// @Failure      500           {object} map[string]any "服务器错误"
// @Router       /api/admin/admins/{id} [put]
// @Security     BearerAuth
func (r *AdminController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	// 加载管理员的当前角色，用于后续比较角色是否改变
	admin, resp := r.findAdminByID(ctx, id, false, true) // 预加载 Roles 关联
	if resp != nil {
		return resp
	}

	allProtectedIDs := r.getAllProtectedAdminIDs()
	isProtected := allProtectedIDs[id]

	// 使用请求验证
	var adminUpdate adminrequests.AdminUpdate
	errors, err := ctx.Request().ValidateRequest(&adminUpdate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	if adminUpdate.Nickname != "" {
		admin.Nickname = adminUpdate.Nickname
	}
	if adminUpdate.Email != "" {
		admin.Email = adminUpdate.Email
	}
	if adminUpdate.Phone != "" {
		admin.Phone = adminUpdate.Phone
	}
	if adminUpdate.DepartmentID > 0 {
		admin.DepartmentID = adminUpdate.DepartmentID
	}
	// 如果请求中提供了 status 字段，则更新状态
	// 使用 All() 方法检查 status 字段是否存在（包括值为 0 的情况）
	allInputs := ctx.Request().All()
	if _, exists := allInputs["status"]; exists {
		// 请求中提供了 status 字段，使用验证后的值
		// 检查是否是超级管理员或受保护的管理员
		superAdminID := cast.ToUint(facades.Config().GetInt("admin.super_admin_id", 1))
		isSuperAdmin := admin.ID == superAdminID
		if (isProtected || isSuperAdmin) && adminUpdate.Status == 0 {
			return response.Error(ctx, http.StatusForbidden, "admin_protected_cannot_disable")
		}
		admin.Status = adminUpdate.Status
	}

	if adminUpdate.Password != "" {
		hashedPassword, err := facades.Hash().Make(adminUpdate.Password)
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
		}
		admin.Password = hashedPassword
	}

	if err := facades.Orm().Query().Save(admin); err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	// 检查是否尝试修改 admin 用户的角色
	if _, exists := allInputs["role_ids"]; exists {
		// 获取当前管理员的角色ID列表（去重）
		currentRoleIDSet := make(map[uint]bool)
		var currentRoleIDs []uint
		for _, role := range admin.Roles {
			if !currentRoleIDSet[role.ID] {
				currentRoleIDSet[role.ID] = true
				currentRoleIDs = append(currentRoleIDs, role.ID)
			}
		}

		// 对传入的角色ID进行去重
		newRoleIDSet := make(map[uint]bool)
		var deduplicatedRoleIDs []uint
		for _, roleID := range adminUpdate.RoleIDs {
			if !newRoleIDSet[roleID] {
				newRoleIDSet[roleID] = true
				deduplicatedRoleIDs = append(deduplicatedRoleIDs, roleID)
			}
		}

		// 比较新的角色ID列表和当前的角色ID列表
		// 只有当角色ID真正改变时才阻止修改
		roleIDsChanged := false

		// 如果长度不同，肯定改变了
		if len(deduplicatedRoleIDs) != len(currentRoleIDs) {
			roleIDsChanged = true
		} else {
			// 长度相同，需要检查内容是否完全一致（忽略顺序）
			// 检查新的角色ID是否都在当前角色ID中
			for _, newRoleID := range deduplicatedRoleIDs {
				if !currentRoleIDSet[newRoleID] {
					roleIDsChanged = true
					break
				}
			}
			// 如果所有新角色ID都在当前角色ID中，且长度相同，说明没有改变
		}

		// 检查是否是超级管理员（通过配置的ID判断，不依赖用户名）
		superAdminID := cast.ToUint(facades.Config().GetInt("admin.super_admin_id", 1))
		isSuperAdmin := admin.ID == superAdminID

		// 只有当角色ID真正改变时才阻止修改
		// 如果角色ID没有改变，允许调用 SyncRoles 来清理重复数据
		if roleIDsChanged && isSuperAdmin {
			return response.Error(ctx, http.StatusForbidden, "admin_cannot_modify_roles")
		}

		// 即使角色ID没有改变，也调用 SyncRoles 来清理重复数据
		// 使用去重后的角色ID列表
		if err := r.adminService.SyncRoles(admin, deduplicatedRoleIDs); err != nil {
			return response.ErrorWithLog(ctx, "admin", err, map[string]any{
				"admin_id": admin.ID,
				"role_ids": deduplicatedRoleIDs,
			})
		}
	}

	return response.Success(ctx, http.Json{
		"admin": *admin,
	})
}

// Destroy 删除管理员
// @Summary      删除管理员
// @Description  删除指定的管理员账号。受保护的管理员和当前登录的管理员不能删除。
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        id   path     int  true  "管理员ID"
// @Success      200  {object} map[string]any "删除成功"
// @Failure      400  {object} map[string]any "参数错误"
// @Failure      401  {object} map[string]any "未登录"
// @Failure      403  {object} map[string]any "无权限、受保护管理员不能删除或不能删除自己"
// @Failure      404  {object} map[string]any "管理员不存在"
// @Failure      500  {object} map[string]any "服务器错误"
// @Router       /api/admin/admins/{id} [delete]
// @Security     BearerAuth
func (r *AdminController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))

	allProtectedIDs := r.getAllProtectedAdminIDs()
	if allProtectedIDs[id] {
		return response.Error(ctx, http.StatusForbidden, "admin_protected_cannot_delete")
	}

	adminValue := ctx.Value("admin")
	if adminValue != nil {
		var currentAdmin models.Admin
		if admin, ok := adminValue.(models.Admin); ok {
			currentAdmin = admin
		} else if adminPtr, ok := adminValue.(*models.Admin); ok {
			currentAdmin = *adminPtr
		}

		if currentAdmin.ID > 0 && currentAdmin.ID == id {
			return response.Error(ctx, http.StatusForbidden, "admin_cannot_delete_self")
		}
	}

	admin, resp := r.findAdminByID(ctx, id, false, false)
	if resp != nil {
		return resp
	}

	if _, err := facades.Orm().Query().Delete(admin); err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	return response.Success(ctx)
}

// UnbindGoogleAuthenticator 管理员解绑其他管理员的谷歌验证码
// @Summary      解绑管理员的谷歌验证码
// @Description  管理员可以解绑其他管理员的谷歌验证码，需要当前管理员已绑定谷歌验证码并输入验证码确认
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        id   path     int  true  "要解绑的管理员ID"
// @Param        code body     string true "当前管理员的谷歌验证码"
// @Success      200  {object} map[string]any "解绑成功"
// @Failure      400  {object} map[string]any "参数错误或验证码错误"
// @Failure      401  {object} map[string]any "未登录"
// @Failure      403  {object} map[string]any "无权限或当前管理员未绑定谷歌验证码"
// @Failure      404  {object} map[string]any "管理员不存在"
// @Failure      500  {object} map[string]any "服务器错误"
// @Router       /api/admin/admins/{id}/unbind-google-auth [post]
// @Security     BearerAuth
func (r *AdminController) UnbindGoogleAuthenticator(ctx http.Context) http.Response {
	// 获取要解绑的管理员ID
	targetAdminID := cast.ToUint(ctx.Request().Route("id"))
	if targetAdminID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	// 检查目标管理员是否存在
	var targetAdmin models.Admin
	if err := facades.Orm().Query().Where("id", targetAdminID).First(&targetAdmin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "admin_not_found")
	}

	// 从context中获取当前管理员信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	var currentAdmin models.Admin
	if adminVal, ok := adminValue.(models.Admin); ok {
		currentAdmin = adminVal
	} else if adminPtr, ok := adminValue.(*models.Admin); ok {
		currentAdmin = *adminPtr
	} else {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	// 检查当前管理员是否已绑定谷歌验证码
	isBound, err := r.googleAuthenticatorService.IsBound(currentAdmin.ID)
	if err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"admin_id": currentAdmin.ID,
		})
	}

	if !isBound {
		return response.Error(ctx, http.StatusForbidden, "google_authenticator_not_bound")
	}

	// 需要验证码确认
	code := ctx.Request().Input("code")
	if code == "" {
		return response.Error(ctx, http.StatusBadRequest, "code_required")
	}

	// 获取当前管理员的密钥
	secret, err := r.googleAuthenticatorService.GetSecret(currentAdmin.ID)
	if err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"admin_id": currentAdmin.ID,
		})
	}

	if secret == "" {
		return response.Error(ctx, http.StatusBadRequest, "google_authenticator_not_bound")
	}

	// 验证当前管理员的验证码
	if !r.googleAuthenticatorService.Verify(secret, code) {
		return response.Error(ctx, http.StatusBadRequest, "google_code_invalid")
	}

	// 检查目标管理员是否已绑定
	targetIsBound, err := r.googleAuthenticatorService.IsBound(targetAdminID)
	if err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"target_admin_id": targetAdminID,
		})
	}

	if !targetIsBound {
		return response.Error(ctx, http.StatusBadRequest, "google_authenticator_not_bound")
	}

	// 解绑目标管理员的谷歌验证码
	if err := r.googleAuthenticatorService.Unbind(targetAdminID); err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"target_admin_id":  targetAdminID,
			"current_admin_id": currentAdmin.ID,
		})
	}

	return response.Success(ctx, "unbind_success")
}

// parseProtectedIDs 解析受保护的管理员ID字符串（支持逗号分隔）
func (r *AdminController) parseProtectedIDs(idsStr string) []uint {
	var ids []uint
	if idsStr == "" {
		return ids
	}

	// 使用字符串分割
	parts := str.Of(idsStr).Split(",")
	for _, part := range parts {
		part = str.Of(part).Trim().String()
		if !str.Of(part).IsEmpty() {
			if id := cast.ToUint(part); id > 0 {
				ids = append(ids, id)
			}
		}
	}

	return ids
}

func (r *AdminController) getAllProtectedAdminIDs() map[uint]bool {
	allProtectedIDs := make(map[uint]bool)
	// 添加超级管理员ID（从配置读取，默认1）
	superAdminID := cast.ToUint(facades.Config().GetInt("admin.super_admin_id", 1))
	allProtectedIDs[superAdminID] = true
	// 添加开发者管理员ID
	developerIDsStr := facades.Config().GetString("admin.developer_ids", "2")
	developerIDs := r.parseProtectedIDs(developerIDsStr)
	for _, did := range developerIDs {
		allProtectedIDs[did] = true
	}
	return allProtectedIDs
}

func (r *AdminController) getDepartmentAndChildrenIDs(departmentID uint) []uint {
	var departmentIDs []uint
	departmentIDs = append(departmentIDs, departmentID)
	r.getChildrenDepartmentIDs(departmentID, &departmentIDs)
	return departmentIDs
}

func (r *AdminController) getChildrenDepartmentIDs(parentID uint, departmentIDs *[]uint) {
	var children []models.Department
	if err := facades.Orm().Query().Where("parent_id", parentID).Get(&children); err == nil {
		for _, child := range children {
			*departmentIDs = append(*departmentIDs, child.ID)
			r.getChildrenDepartmentIDs(child.ID, departmentIDs)
		}
	}
}

// Export 导出管理员列表
// @Summary      导出管理员列表
// @Description  根据筛选条件导出管理员列表为CSV文件，支持与列表查询相同的筛选条件
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        request body     AdminExportRequest false "导出筛选条件（可选）"
// @Success      200     {object} map[string]any "导出成功，返回文件下载信息"
// @Failure      400     {object} map[string]any "参数错误"
// @Failure      401     {object} map[string]any "未登录"
// @Failure      403     {object} map[string]any "无权限"
// @Failure      500     {object} map[string]any "服务器错误"
// @Router       /api/admin/admins/export [post]
// @Security     BearerAuth
func (r *AdminController) Export(ctx http.Context) http.Response {
	query := r.buildQuery(ctx)

	var admins []models.Admin
	if err := query.With("Department").With("Roles").Get(&admins); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	headers := []string{
		"export_header_id",
		"export_header_username",
		"export_header_nickname",
		"export_header_email",
		"export_header_phone",
		"export_header_status",
		"export_header_department",
		"export_header_roles",
		"export_header_created_at",
		"export_header_updated_at",
	}

	var data [][]string
	for _, admin := range admins {
		statusText := trans.Get(ctx, "export_status_disabled")
		if admin.Status == 1 {
			statusText = trans.Get(ctx, "export_status_enabled")
		}

		// 部门名称
		departmentName := ""
		if admin.Department.ID > 0 {
			departmentName = admin.Department.Name
		}

		// 角色名称（多个角色用逗号分隔）
		roleNames := ""
		if len(admin.Roles) > 0 {
			for i, role := range admin.Roles {
				if i > 0 {
					roleNames += ", "
				}
				roleNames += role.Name
			}
		}

		// 时间格式化
		createdAt := ""
		updatedAt := ""
		if !admin.CreatedAt.IsZero() {
			createdAt = admin.CreatedAt.Format("2006-01-02 15:04:05")
		}
		if !admin.UpdatedAt.IsZero() {
			updatedAt = admin.UpdatedAt.Format("2006-01-02 15:04:05")
		}

		row := []string{
			cast.ToString(admin.ID),
			admin.Username,
			admin.Nickname,
			admin.Email,
			admin.Phone,
			statusText,
			departmentName,
			roleNames,
			createdAt,
			updatedAt,
		}
		data = append(data, row)
	}

	return response.Export(ctx, "export_success", headers, data, "admins")
}
