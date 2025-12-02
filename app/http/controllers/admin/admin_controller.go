package admin

import (
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/errorlog"
)

type AdminController struct {
	adminService services.AdminService
}

// AdminExportRequest 导出管理员请求参数
type AdminExportRequest struct {
	Username     string `json:"username" form:"username" example:"admin"`                   // 用户名（模糊搜索）
	Status       string `json:"status" form:"status" example:"1"`                           // 状态：1-启用，0-禁用
	RoleID       string `json:"role_id" form:"role_id" example:"1"`                         // 角色ID
	DepartmentID string `json:"department_id" form:"department_id" example:"1"`             // 部门ID
	StartTime    string `json:"start_time" form:"start_time" example:"2024-01-01 00:00:00"` // 开始时间
	EndTime      string `json:"end_time" form:"end_time" example:"2024-12-31 23:59:59"`     // 结束时间
	OrderBy      string `json:"order_by" form:"order_by" example:"created_at:desc"`         // 排序
}

// AdminResponse 管理员响应数据
type AdminResponse struct {
	ID           uint                     `json:"id" example:"1"`                           // 管理员ID
	Username     string                   `json:"username" example:"admin"`                 // 用户名
	Nickname     string                   `json:"nickname" example:"管理员"`                   // 昵称
	Avatar       string                   `json:"avatar" example:""`                        // 头像
	Email        string                   `json:"email" example:"admin@example.com"`        // 邮箱
	Phone        string                   `json:"phone" example:"13800138000"`              // 手机号
	Status       uint8                    `json:"status" example:"1"`                       // 状态：1-启用，0-禁用
	DepartmentID uint                     `json:"department_id" example:"1"`                // 部门ID
	Department   map[string]interface{}   `json:"department"`                               // 部门信息
	Roles        []map[string]interface{} `json:"roles"`                                    // 角色列表
	CreatedAt    string                   `json:"created_at" example:"2024-01-01 00:00:00"` // 创建时间
	UpdatedAt    string                   `json:"updated_at" example:"2024-01-01 00:00:00"` // 更新时间
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
		adminService: services.NewAdminServiceImpl(),
	}
}

// buildQuery 构建查询（列表和导出共用）
// 同时支持查询参数（GET）和请求体参数（POST）
func (r *AdminController) buildQuery(ctx http.Context) orm.Query {
	// 优先从请求体读取，如果没有则从查询参数读取（兼容 GET 和 POST）
	username := ctx.Request().Input("username", ctx.Request().Query("username", ""))
	status := ctx.Request().Input("status", ctx.Request().Query("status", ""))
	roleID := ctx.Request().Input("role_id", ctx.Request().Query("role_id", ""))
	departmentID := ctx.Request().Input("department_id", ctx.Request().Query("department_id", ""))
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
// @Failure      400           {object}  map[string]interface{} "参数错误"
// @Failure      401           {object}  map[string]interface{} "未登录"
// @Failure      403           {object}  map[string]interface{} "无权限"
// @Failure      500           {object}  map[string]interface{} "服务器错误"
// @Router       /api/admin/admins [get]
// @Security     BearerAuth
func (r *AdminController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "10"))

	query := r.buildQuery(ctx)

	total, err := query.Count()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	var admins []models.Admin
	offset := (page - 1) * pageSize
	if err := query.With("Department").With("Roles").Offset(offset).Limit(pageSize).Get(&admins); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "get_success", admins, total, page, pageSize)
}

// Show 管理员详情
// @Summary      获取管理员详情
// @Description  根据ID获取管理员详细信息，包括部门、角色等关联信息
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        id   path     int  true  "管理员ID"
// @Success      200  {object} AdminDetailResponse
// @Failure      400  {object} map[string]interface{} "参数错误"
// @Failure      401  {object} map[string]interface{} "未登录"
// @Failure      403  {object} map[string]interface{} "无权限"
// @Failure      404  {object} map[string]interface{} "管理员不存在"
// @Failure      500  {object} map[string]interface{} "服务器错误"
// @Router       /api/admin/admins/{id} [get]
// @Security     BearerAuth
func (r *AdminController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var admin models.Admin
	if err := facades.Orm().Query().With("Department").With("Roles").Where("id", id).First(&admin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "admin_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"admin": admin,
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
// @Failure      400           {object} map[string]interface{} "参数错误或用户名已存在"
// @Failure      401           {object} map[string]interface{} "未登录"
// @Failure      403           {object} map[string]interface{} "无权限"
// @Failure      500           {object} map[string]interface{} "服务器错误"
// @Router       /api/admin/admins [post]
// @Security     BearerAuth
func (r *AdminController) Store(ctx http.Context) http.Response {
	username := ctx.Request().Input("username")
	password := ctx.Request().Input("password")
	nickname := ctx.Request().Input("nickname")
	email := ctx.Request().Input("email")
	phone := ctx.Request().Input("phone")
	departmentID := cast.ToUint(ctx.Request().Input("department_id", "0"))
	status := cast.ToUint8(ctx.Request().Input("status", "0"))

	if username == "" || password == "" {
		return response.Error(ctx, http.StatusBadRequest, "username_and_password_required")
	}

	exists, err := facades.Orm().Query().Model(&models.Admin{}).Where("username", username).Exists()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}
	if exists {
		return response.Error(ctx, http.StatusBadRequest, "username_exists")
	}

	// 加密密码
	hashedPassword, err := facades.Hash().Make(password)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
	}

	now := carbon.Now()
	adminData := map[string]any{
		"username":      username,
		"password":      hashedPassword,
		"nickname":      nickname,
		"avatar":        "",
		"email":         email,
		"phone":         phone,
		"department_id": departmentID,
		"status":        status,
		"created_at":    now,
		"updated_at":    now,
	}

	if err := facades.Orm().Query().Table("admins").Create(adminData); err != nil {
		errorlog.RecordHTTP(ctx, "admin", "Failed to create admin", map[string]any{
			"error":    err.Error(),
			"username": username,
		}, "Create admin error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	var admin models.Admin
	if err := facades.Orm().Query().Where("username", username).First(&admin); err != nil {
		errorlog.RecordHTTP(ctx, "admin", "Failed to query created admin", map[string]any{
			"error":    err.Error(),
			"username": username,
		}, "Query created admin error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	if roleIds := ctx.Request().Input("role_ids"); roleIds != "" {
		var roleIDs []uint
		for _, idStr := range ctx.Request().InputArray("role_ids") {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				roleIDs = append(roleIDs, uint(id))
			}
		}
		if err := r.adminService.SyncRoles(&admin, roleIDs); err != nil {
			errorlog.RecordHTTP(ctx, "admin", "Failed to sync admin roles", map[string]any{
				"error":    err.Error(),
				"admin_id": admin.ID,
				"role_ids": roleIDs,
			}, "Sync admin roles error: %v", err)
			return response.Error(ctx, http.StatusInternalServerError, "create_failed")
		}
	}

	return response.Success(ctx, "create_success", http.Json{
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
// @Failure      400           {object} map[string]interface{} "参数错误"
// @Failure      401           {object} map[string]interface{} "未登录"
// @Failure      403           {object} map[string]interface{} "无权限或受保护管理员不能禁用"
// @Failure      404           {object} map[string]interface{} "管理员不存在"
// @Failure      500           {object} map[string]interface{} "服务器错误"
// @Router       /api/admin/admins/{id} [put]
// @Security     BearerAuth
func (r *AdminController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var admin models.Admin
	if err := facades.Orm().Query().Where("id", id).First(&admin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "admin_not_found")
	}

	allProtectedIDs := r.getAllProtectedAdminIDs()
	isProtected := allProtectedIDs[id]

	nickname := ctx.Request().Input("nickname")
	email := ctx.Request().Input("email")
	phone := ctx.Request().Input("phone")
	departmentID := cast.ToUint(ctx.Request().Input("department_id", "0"))
	status := ctx.Request().Input("status", "")

	if nickname != "" {
		admin.Nickname = nickname
	}
	if email != "" {
		admin.Email = email
	}
	if phone != "" {
		admin.Phone = phone
	}
	if departmentID > 0 {
		admin.DepartmentID = departmentID
	}
	if status != "" {
		newStatus := cast.ToUint8(status)
		if isProtected && newStatus == 0 {
			return response.Error(ctx, http.StatusForbidden, "admin_protected_cannot_disable")
		}
		admin.Status = newStatus
	}

	if password := ctx.Request().Input("password"); password != "" {
		hashedPassword, err := facades.Hash().Make(password)
		if err == nil {
			admin.Password = hashedPassword
		}
	}

	if err := facades.Orm().Query().Save(&admin); err != nil {
		errorlog.RecordHTTP(ctx, "admin", "Failed to update admin", map[string]any{
			"error":    err.Error(),
			"admin_id": admin.ID,
		}, "Update admin error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	if roleIds := ctx.Request().Input("role_ids"); roleIds != "" {
		var roleIDs []uint
		for _, idStr := range ctx.Request().InputArray("role_ids") {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				roleIDs = append(roleIDs, uint(id))
			}
		}
		if err := r.adminService.SyncRoles(&admin, roleIDs); err != nil {
			errorlog.RecordHTTP(ctx, "admin", "Failed to sync admin roles in update", map[string]any{
				"error":    err.Error(),
				"admin_id": admin.ID,
				"role_ids": roleIDs,
			}, "Sync admin roles in update error: %v", err)
			return response.Error(ctx, http.StatusInternalServerError, "update_failed")
		}
	}

	return response.Success(ctx, "update_success", http.Json{
		"admin": admin,
	})
}

// Destroy 删除管理员
// @Summary      删除管理员
// @Description  删除指定的管理员账号。受保护的管理员和当前登录的管理员不能删除。
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        id   path     int  true  "管理员ID"
// @Success      200  {object} map[string]interface{} "删除成功"
// @Failure      400  {object} map[string]interface{} "参数错误"
// @Failure      401  {object} map[string]interface{} "未登录"
// @Failure      403  {object} map[string]interface{} "无权限、受保护管理员不能删除或不能删除自己"
// @Failure      404  {object} map[string]interface{} "管理员不存在"
// @Failure      500  {object} map[string]interface{} "服务器错误"
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

	var admin models.Admin
	if err := facades.Orm().Query().Where("id", id).First(&admin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "admin_not_found")
	}

	if _, err := facades.Orm().Query().Delete(&admin); err != nil {
		errorlog.RecordHTTP(ctx, "admin", "Failed to delete admin", map[string]any{
			"error":    err.Error(),
			"admin_id": admin.ID,
		}, "Delete admin error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

// parseProtectedIDs 解析受保护的管理员ID字符串（支持逗号分隔）
func (r *AdminController) parseProtectedIDs(idsStr string) []uint {
	var ids []uint
	if idsStr == "" {
		return ids
	}

	// 使用字符串分割
	parts := strings.Split(idsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			if id := cast.ToUint(part); id > 0 {
				ids = append(ids, id)
			}
		}
	}

	return ids
}

func (r *AdminController) getAllProtectedAdminIDs() map[uint]bool {
	allProtectedIDs := make(map[uint]bool)
	allProtectedIDs[1] = true
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
// @Success      200     {object} map[string]interface{} "导出成功，返回文件下载信息"
// @Failure      400     {object} map[string]interface{} "参数错误"
// @Failure      401     {object} map[string]interface{} "未登录"
// @Failure      403     {object} map[string]interface{} "无权限"
// @Failure      500     {object} map[string]interface{} "服务器错误"
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
