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
