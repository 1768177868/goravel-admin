package admin

import (
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
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

	// 检查用户名是否已存在（排除软删除的记录）
	// 注意：数据库层面使用了 username + deleted_at 的联合唯一索引
	// 这样可以允许软删除的用户名被重新使用，但同一时间只能有一个未删除的用户名
	// GORM 默认会排除软删除的记录，使用 Count() 方法检查未删除的记录
	count, err := facades.Orm().Query().Model(&models.Admin{}).Where("username", username).Count()
	if err != nil {
		// 查询出错，记录日志但不阻止创建（可能是数据库问题）
		facades.Log().Errorf("Check username exists error: %v", err)
	} else if count > 0 {
		return response.Error(ctx, http.StatusBadRequest, "username_exists")
	}
	// 如果没有找到记录（包括软删除的记录），用户名可用，继续创建
	// 数据库层面的联合唯一索引会确保唯一性约束

	// 加密密码
	hashedPassword, err := facades.Hash().Make(password)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
	}

	// 使用 map 方式创建，确保零值字段（status=0）也能被正确保存
	// GORM 在使用结构体创建时，可能会忽略零值字段而使用数据库默认值
	// 使用 map 可以明确指定所有字段的值，包括零值
	adminData := map[string]interface{}{
		"username":      username,
		"password":      hashedPassword,
		"nickname":      nickname,
		"avatar":        "",
		"email":         email,
		"phone":         phone,
		"department_id": departmentID,
		"status":        status, // 明确设置 status，即使是 0 也会被保存
	}

	// 使用 Table().Create() 创建记录，这样可以确保所有字段（包括零值）都被保存
	if err := facades.Orm().Query().Table("admins").Create(adminData); err != nil {
		facades.Log().Errorf("Create admin error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	// 获取创建后的记录
	var admin models.Admin
	if err := facades.Orm().Query().Where("username", username).First(&admin); err != nil {
		facades.Log().Errorf("Get created admin error: %v", err)
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
		admin.Status = cast.ToUint8(status)
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
