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
)

type AdminController struct {
	adminService services.AdminService
}

func NewAdminController() *AdminController {
	return &AdminController{
		adminService: services.NewAdminServiceImpl(),
	}
}

// buildQuery 构建查询（列表和导出共用）
func (r *AdminController) buildQuery(ctx http.Context) orm.Query {
	username := ctx.Request().Query("username", "")
	status := ctx.Request().Query("status", "")
	orderBy := ctx.Request().Query("order_by", "")

	query := facades.Orm().Query().Model(&models.Admin{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if status != "" {
		query = query.Where("status", status)
	}

	// 应用排序（默认按创建时间倒序）
	query = helpers.ApplySort(query, orderBy, "created_at:desc")

	return query
}

// Index 管理员列表
func (r *AdminController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "10"))

	// 使用公共查询构建方法
	query := r.buildQuery(ctx)

	total, err := query.Count()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	var admins []models.Admin
	offset := (page - 1) * pageSize
	// 使用 With 预加载关联，避免 N+1 查询问题
	if err := query.With("Department").With("Roles").Offset(offset).Limit(pageSize).Get(&admins); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "get_success", admins, total, page, pageSize)
}

// Show 管理员详情
func (r *AdminController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var admin models.Admin
	// 使用 With 预加载关联
	if err := facades.Orm().Query().With("Department").With("Roles").Where("id", id).First(&admin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "admin_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"admin": admin,
	})
}

// Store 创建管理员
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
		"status":        status, // 明确设置 status，即使是 0 也会被保存
		"created_at":    now,
		"updated_at":    now,
	}

	if err := facades.Orm().Query().Table("admins").Create(adminData); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	var admin models.Admin
	if err := facades.Orm().Query().Where("username", username).First(&admin); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	// 处理角色关联
	if roleIds := ctx.Request().Input("role_ids"); roleIds != "" {
		var roleIDs []uint
		for _, idStr := range ctx.Request().InputArray("role_ids") {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				roleIDs = append(roleIDs, uint(id))
			}
		}
		if err := r.adminService.SyncRoles(&admin, roleIDs); err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "create_failed")
		}
	}

	return response.Success(ctx, "create_success", http.Json{
		"admin": admin,
	})
}

// Update 更新管理员
func (r *AdminController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var admin models.Admin
	if err := facades.Orm().Query().Where("id", id).First(&admin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "admin_not_found")
	}

	protectedIDsStr := facades.Config().GetString("admin.protected_ids", "1")
	protectedIDs := r.parseProtectedIDs(protectedIDsStr)
	isProtected := false
	for _, protectedID := range protectedIDs {
		if id == protectedID {
			isProtected = true
			break
		}
	}

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

	// 更新密码
	if password := ctx.Request().Input("password"); password != "" {
		hashedPassword, err := facades.Hash().Make(password)
		if err == nil {
			admin.Password = hashedPassword
		}
	}

	if err := facades.Orm().Query().Save(&admin); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	// 处理角色关联
	if roleIds := ctx.Request().Input("role_ids"); roleIds != "" {
		var roleIDs []uint
		for _, idStr := range ctx.Request().InputArray("role_ids") {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				roleIDs = append(roleIDs, uint(id))
			}
		}
		if err := r.adminService.SyncRoles(&admin, roleIDs); err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "update_failed")
		}
	}

	return response.Success(ctx, "update_success", http.Json{
		"admin": admin,
	})
}

// Destroy 删除管理员
func (r *AdminController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))

	// 检查是否是受保护的管理员ID
	protectedIDsStr := facades.Config().GetString("admin.protected_ids", "1")
	protectedIDs := r.parseProtectedIDs(protectedIDsStr)
	for _, protectedID := range protectedIDs {
		if id == protectedID {
			return response.Error(ctx, http.StatusForbidden, "admin_protected_cannot_delete")
		}
	}

	// 检查是否是当前登录的管理员自己
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

// Export 导出管理员列表
func (r *AdminController) Export(ctx http.Context) http.Response {
	// 使用公共查询构建方法
	query := r.buildQuery(ctx)

	var admins []models.Admin
	// 导出时获取所有数据，不分页
	if err := query.With("Department").With("Roles").Get(&admins); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	// 准备CSV表头（使用翻译键）
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

	// 准备数据行
	var data [][]string
	for _, admin := range admins {
		// 状态文本（使用翻译键）
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
