package admin

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/response"
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

// Index 管理员列表
func (r *AdminController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "10"))
	username := ctx.Request().Query("username", "")
	status := ctx.Request().Query("status", "")

	query := facades.Orm().Query().Model(&models.Admin{})

	if username != "" {
		query = query.Where("username", "like", "%"+username+"%")
	}
	if status != "" {
		query = query.Where("status", status)
	}

	total, err := query.Count()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	var admins []models.Admin
	offset := (page - 1) * pageSize
	// 使用 With 预加载关联，避免 N+1 查询问题
	if err := query.With("Department").With("Roles").Offset(offset).Limit(pageSize).Order("id desc").Get(&admins); err != nil {
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
	status := cast.ToUint8(ctx.Request().Input("status", "1"))

	if username == "" || password == "" {
		return response.Error(ctx, http.StatusBadRequest, "username_and_password_required")
	}

	// 检查用户名是否已存在
	var existAdmin models.Admin
	if err := facades.Orm().Query().Where("username", username).First(&existAdmin); err == nil {
		return response.Error(ctx, http.StatusBadRequest, "username_exists")
	}

	// 加密密码
	hashedPassword, err := facades.Hash().Make(password)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
	}

	admin := models.Admin{
		Username:     username,
		Password:     hashedPassword,
		Nickname:     nickname,
		Email:        email,
		Phone:        phone,
		DepartmentID: departmentID,
		Status:       status,
	}

	if err := facades.Orm().Query().Create(&admin); err != nil {
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
	var admin models.Admin
	if err := facades.Orm().Query().Where("id", id).First(&admin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "admin_not_found")
	}

	if _, err := facades.Orm().Query().Delete(&admin); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}
