package admin

import (
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/apidoc"
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

// AdminResponse 管理员 JSON 字段（列表项含 2FA/超管；详情/写接口可能不含 is_2fa_bound）
type AdminResponse struct {
	ID           uint             `json:"id" example:"1"`                           // 管理员ID
	Username     string           `json:"username" example:"admin"`                 // 登录用户名
	Nickname     string           `json:"nickname" example:"管理员"`                   // 显示昵称
	Avatar       string           `json:"avatar" example:""`                        // 头像URL
	Email        string           `json:"email" example:"admin@example.com"`        // 联系邮箱
	Phone        string           `json:"phone" example:"13800138000"`              // 联系手机号
	Status       uint8            `json:"status" enums:"0,1" example:"1"`           // 账号状态（1启用，0禁用）
	Is2FABound   bool             `json:"is_2fa_bound,omitempty" example:"true"`    // 是否已绑定2FA
	IsSuperAdmin bool             `json:"is_super_admin,omitempty" example:"false"` // 是否为超级管理员
	DepartmentID uint             `json:"department_id" example:"1"`                // 所属部门ID
	Department   map[string]any   `json:"department"`                               // 部门信息对象
	PositionID   uint             `json:"position_id" example:"1"`                  // 所属岗位ID
	Position     map[string]any   `json:"position"`                                 // 岗位信息对象
	Roles        []map[string]any `json:"roles"`                                    // 角色信息数组
	CreatedAt    string           `json:"created_at" example:"2024-01-01 00:00:00"` // 创建时间
	UpdatedAt    string           `json:"updated_at" example:"2024-01-01 00:00:00"` // 更新时间
}

// AdminListData 列表成功时 data 内层
type AdminListData struct {
	List []AdminResponse `json:"list"`
	apidoc.Pagination
}

// AdminListResponse 管理员分页列表（命名约定：XxxListResponse = apidoc.Success + XxxListData）
type AdminListResponse struct {
	apidoc.Success
	Data AdminListData `json:"data"`
}

// AdminDetailData 详情/创建/更新时 data 内层
type AdminDetailData struct {
	Admin AdminResponse `json:"admin"`
}

// AdminDetailResponse 单条管理员（apidoc.Success + 业务 Data）
type AdminDetailResponse struct {
	apidoc.Success
	Data AdminDetailData `json:"data"`
}

// AdminExportData 导出成功时 data 内层
type AdminExportData struct {
	FilePath string `json:"file_path" example:"exports/admins_20260409_150000.csv"`         // 导出文件存储路径
	FileURL  string `json:"file_url" example:"/storage/exports/admins_20260409_150000.csv"` // 导出文件访问地址
}

// AdminExportResponse 管理员导出响应
type AdminExportResponse struct {
	apidoc.Success
	Data AdminExportData `json:"data"`
}

// UnbindGoogleAuthRequest 解绑管理员2FA请求
type UnbindGoogleAuthRequest struct {
	Code string `json:"code" example:"123456"` // 当前管理员的谷歌验证码（6位动态码）
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
	admin, err := r.adminService.GetByID(id, withDepartment, withRoles)
	if err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, apperrors.ErrAdminNotFound.Code)
	}
	return admin, nil
}

// buildFilters 构建查询过滤器（列表和导出共用）
// 同时支持查询参数（GET）和请求体参数（POST）
func (r *AdminController) buildFilters(ctx http.Context) services.AdminFilters {
	// 优先从请求体读取，如果没有则从查询参数读取（兼容 GET 和 POST）
	username := ctx.Request().Input("username", ctx.Request().Query("username", ""))
	status := ctx.Request().Input("status", ctx.Request().Query("status", ""))
	roleID := ctx.Request().Input("role_id", ctx.Request().Query("role_id", ""))
	departmentID := ctx.Request().Input("department_id", ctx.Request().Query("department_id", ""))
	positionID := ctx.Request().Input("position_id", ctx.Request().Query("position_id", ""))
	is2FABound := ctx.Request().Input("is_2fa_bound", ctx.Request().Query("is_2fa_bound", ""))
	orderBy := ctx.Request().Input("order_by", ctx.Request().Query("order_by", ""))
	// 时间参数同时支持从查询参数和请求体读取，并统一转换为 UTC
	startTime := getTimeInputOrQueryUTC(ctx, "start_time")
	endTime := getTimeInputOrQueryUTC(ctx, "end_time")

	return services.AdminFilters{
		Username:     username,
		Status:       status,
		RoleID:       roleID,
		DepartmentID: departmentID,
		PositionID:   positionID,
		Is2FABound:   is2FABound,
		StartTime:    startTime,
		EndTime:      endTime,
		OrderBy:      orderBy,
	}
}

func hasInput(allInputs map[string]any, key string) bool {
	_, exists := allInputs[key]
	return exists
}

func (r *AdminController) applyAdminUpdatableFields(admin *models.Admin, adminUpdate adminrequests.AdminUpdate, allInputs map[string]any) {
	if hasInput(allInputs, "nickname") {
		admin.Nickname = adminUpdate.Nickname
	}
	if hasInput(allInputs, "email") {
		admin.Email = adminUpdate.Email
	}
	if hasInput(allInputs, "phone") {
		admin.Phone = adminUpdate.Phone
	}
	if hasInput(allInputs, "department_id") {
		admin.DepartmentID = adminUpdate.DepartmentID
	}
	if hasInput(allInputs, "position_id") {
		admin.PositionID = adminUpdate.PositionID
	}
}

// Index 管理员列表
// @Summary      获取管理员列表
// @Description  分页获取管理员列表
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        page          query     int     false  "页码（从1开始）" default(1)
// @Param        page_size     query     int     false  "每页数量（建议 10-100）" default(10)
// @Param        username      query     string  false  "登录用户名（模糊匹配）"
// @Param        status        query     string  false  "账号状态（1-启用，0-禁用）" Enums(0,1)
// @Param        role_id       query     string  false  "角色ID（精确匹配）"
// @Param        department_id query     string  false  "部门ID（精确匹配）"
// @Param        position_id   query     string  false  "岗位ID（精确匹配）"
// @Param        is_2fa_bound  query     string  false  "是否已绑定2FA（1-已绑定，0-未绑定）" Enums(0,1)
// @Param        start_time    query     string  false  "创建时间开始（格式：YYYY-MM-DD HH:mm:ss）"
// @Param        end_time      query     string  false  "创建时间结束（格式：YYYY-MM-DD HH:mm:ss）"
// @Param        order_by      query     string  false  "排序字段（格式：字段:asc/desc，例如：created_at:desc）"
// @Success      200           {object}  AdminListResponse
// @Failure      500           {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/admins [get]
// @Security     BearerAuth
func (r *AdminController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	filters := r.buildFilters(ctx)

	admins, total, err := r.adminService.GetList(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"action": "list_admins",
		})
	}

	// 获取超级管理员ID
	superAdminID := cast.ToUint(facades.Config().GetInt("admin.super_admin_id", 1))

	// 转换数据格式
	adminList := make([]http.Json, len(admins))
	for i, admin := range admins {
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
			"position_id":    admin.PositionID,
			"position":       admin.Position,
			"roles":          admin.Roles,
			"created_at":     admin.CreatedAt,
			"updated_at":     admin.UpdatedAt,
		}
	}

	return response.Success(ctx, http.Json{
		"list":      adminList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show 管理员详情
// @Summary      获取管理员详情
// @Description  根据ID获取管理员详细信息
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        id   path     int  true  "管理员ID"
// @Success      200  {object} AdminDetailResponse
// @Failure      404  {object} apidoc.Error "管理员不存在"
// @Router       /api/admin/admins/{id} [get]
// @Security     BearerAuth
func (r *AdminController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
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
			"position_id":    admin.PositionID,
			"position":       admin.Position,
			"roles":          admin.Roles,
			"created_at":     admin.CreatedAt,
			"updated_at":     admin.UpdatedAt,
		},
	})
}

// Store 创建管理员
// @Summary      创建管理员
// @Description  创建新的管理员账号（status：1-启用，0-禁用）
// @Description  字段说明：username-登录用户名（必填）；password-登录密码（必填）；nickname-显示昵称；email-联系邮箱；phone-联系手机号；department_id-所属部门ID；position_id-所属岗位ID；role_ids-角色ID数组；status-账号状态（1启用/0禁用）
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        request       body     adminrequests.AdminCreate  true  "创建参数（必填：username、password；可选：nickname、email、phone、department_id、position_id、status、role_ids）"
// @Success      200           {object} AdminDetailResponse
// @Failure      500           {object} apidoc.Error "服务器错误"
// @Router       /api/admin/admins [post]
// @Security     BearerAuth
func (r *AdminController) Store(ctx http.Context) http.Response {
	// 使用请求验证
	var adminCreate adminrequests.AdminCreate
	validationErrors, err := ctx.Request().ValidateRequest(&adminCreate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if validationErrors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", validationErrors.All())
	}

	admin, err := r.adminService.CreateAdmin(services.CreateAdminInput{
		Username:     adminCreate.Username,
		Password:     adminCreate.Password,
		Nickname:     adminCreate.Nickname,
		Email:        adminCreate.Email,
		Phone:        adminCreate.Phone,
		DepartmentID: adminCreate.DepartmentID,
		PositionID:   adminCreate.PositionID,
		Status:       adminCreate.Status,
		RoleIDs:      adminCreate.RoleIDs,
	})
	if err != nil {
		var businessErr *apperrors.BusinessError
		if stderrors.As(err, &businessErr) && businessErr.Code == apperrors.ErrUsernameExists.Code {
			return response.Error(ctx, http.StatusBadRequest, businessErr)
		}
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"action":   "create_admin",
			"username": adminCreate.Username,
		})
	}

	return response.Success(ctx, http.Json{
		"admin": admin,
	})
}

// Update 更新管理员
// @Summary      更新管理员信息
// @Description  更新管理员的基本信息（status：1-启用，0-禁用）
// @Description  字段说明：nickname-显示昵称；email-联系邮箱；phone-联系手机号；password-登录密码；department_id-所属部门ID；position_id-所属岗位ID；role_ids-角色ID数组；status-账号状态（1启用/0禁用）
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        id            path     int                       true  "管理员ID"
// @Param        request       body     adminrequests.AdminUpdate true  "更新参数（可按需提交任意字段）"
// @Success      200           {object} AdminDetailResponse
// @Failure      403           {object} apidoc.Error "无权限或受保护管理员不能禁用"
// @Failure      404           {object} apidoc.Error "管理员不存在"
// @Router       /api/admin/admins/{id} [put]
// @Security     BearerAuth
func (r *AdminController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	admin, resp := r.findAdminByID(ctx, id, false, true)
	if resp != nil {
		return resp
	}

	// 使用请求验证
	var adminUpdate adminrequests.AdminUpdate
	errors, err := ctx.Request().ValidateRequest(&adminUpdate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 使用 All() 方法检查字段是否存在
	allInputs := ctx.Request().All()

	r.applyAdminUpdatableFields(admin, adminUpdate, allInputs)

	if hasInput(allInputs, "status") {
		if err := r.adminService.ValidateStatusChange(admin.ID, adminUpdate.Status); err != nil {
			return response.Error(ctx, http.StatusForbidden, err)
		}
		admin.Status = adminUpdate.Status
	}

	if adminUpdate.Password != "" {
		hashedPassword, err := facades.Hash().Make(adminUpdate.Password)
		if err != nil {
			return response.ErrorWithLog(ctx, "admin", err, map[string]any{
				"action":   "encrypt_password",
				"admin_id": admin.ID,
			})
		}
		admin.Password = hashedPassword
	}

	if err := r.adminService.Update(admin); err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	// 检查是否尝试修改 admin 用户的角色
	if hasInput(allInputs, "role_ids") {
		deduplicatedRoleIDs := r.adminService.NormalizeRoleIDs(adminUpdate.RoleIDs)
		if err := r.adminService.ValidateRoleChange(admin.ID, admin.Roles, deduplicatedRoleIDs); err != nil {
			return response.Error(ctx, http.StatusForbidden, err)
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
// @Description  删除指定的管理员账号
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        id   path     int  true  "管理员ID"
// @Success      200  {object} apidoc.Success "删除成功"
// @Failure      403  {object} apidoc.Error "无权限、受保护管理员不能删除或不能删除自己"
// @Failure      404  {object} apidoc.Error "管理员不存在"
// @Router       /api/admin/admins/{id} [delete]
// @Security     BearerAuth
func (r *AdminController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")

	if r.adminService.IsProtectedAdmin(id) {
		return response.Error(ctx, http.StatusForbidden, apperrors.ErrAdminProtectedCannotDelete.Code)
	}

	currentAdmin, resp := r.currentAdminFromContext(ctx)
	if resp != nil {
		return resp
	}
	if currentAdmin != nil && currentAdmin.ID == id {
		return response.Error(ctx, http.StatusForbidden, apperrors.ErrAdminCannotDeleteSelf.Code)
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
// @Description  管理员解绑其他管理员的谷歌验证码
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        id       path     int                     true  "要解绑的管理员ID"
// @Param        request  body     UnbindGoogleAuthRequest true  "验证码确认参数"
// @Success      200      {object} apidoc.Success "解绑成功"
// @Failure      400      {object} apidoc.Error "参数错误、验证码错误或目标管理员未绑定2FA"
// @Failure      401      {object} apidoc.Error "未登录"
// @Failure      403      {object} apidoc.Error "当前管理员未绑定2FA，禁止执行"
// @Failure      404      {object} apidoc.Error "管理员不存在"
// @Failure      500      {object} apidoc.Error "服务器错误"
// @Router       /api/admin/admins/{id}/unbind-google-auth [post]
// @Security     BearerAuth
func (r *AdminController) UnbindGoogleAuthenticator(ctx http.Context) http.Response {
	// 获取要解绑的管理员ID
	targetAdminID := helpers.GetUintRoute(ctx, "id")
	if targetAdminID == 0 {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDRequired.Code)
	}

	// 检查目标管理员是否存在
	if _, resp := r.findAdminByID(ctx, targetAdminID, false, false); resp != nil {
		return resp
	}

	// 从 context 中获取当前管理员信息（由 JWT 中间件设置）
	currentAdmin, resp := r.currentAdminFromContext(ctx)
	if resp != nil {
		return resp
	}
	if currentAdmin == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	// 检查当前管理员是否已绑定谷歌验证码
	isBound, err := r.googleAuthenticatorService.IsBound(currentAdmin.ID)
	if err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"admin_id": currentAdmin.ID,
		})
	}

	if !isBound {
		return response.Error(ctx, http.StatusForbidden, apperrors.ErrGoogleAuthenticatorNotBound.Code)
	}

	// 需要验证码确认
	code := ctx.Request().Input("code")
	if code == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrCodeRequired.Code)
	}

	// 获取当前管理员的密钥
	secret, err := r.googleAuthenticatorService.GetSecret(currentAdmin.ID)
	if err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"admin_id": currentAdmin.ID,
		})
	}

	if secret == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrGoogleAuthenticatorNotBound.Code)
	}

	// 验证当前管理员的验证码
	if !r.googleAuthenticatorService.Verify(secret, code) {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrGoogleCodeInvalid.Code)
	}

	// 检查目标管理员是否已绑定
	targetIsBound, err := r.googleAuthenticatorService.IsBound(targetAdminID)
	if err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"target_admin_id": targetAdminID,
		})
	}

	if !targetIsBound {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrGoogleAuthenticatorNotBound.Code)
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

// ResetGoogleAuthenticator 重置管理员的谷歌验证码（强制清除绑定，无需验证码，用于管理员丢失手机等场景）
// @Summary      重置管理员的谷歌验证码
// @Description  强制清除指定管理员的谷歌验证码绑定
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        id   path     int true "要重置的管理员ID"
// @Success      200  {object} apidoc.Success "重置成功"
// @Failure      400  {object} apidoc.Error "参数错误或目标管理员未绑定2FA"
// @Failure      401  {object} apidoc.Error "未登录"
// @Failure      403  {object} apidoc.Error "无权限或不可操作受保护管理员"
// @Failure      404  {object} apidoc.Error "管理员不存在"
// @Failure      500  {object} apidoc.Error "服务器错误"
// @Router       /api/admin/admins/{id}/reset-google-auth [post]
// @Security     BearerAuth
func (r *AdminController) ResetGoogleAuthenticator(ctx http.Context) http.Response {
	targetAdminID := helpers.GetUintRoute(ctx, "id")
	if targetAdminID == 0 {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDRequired.Code)
	}

	if _, resp := r.findAdminByID(ctx, targetAdminID, false, false); resp != nil {
		return resp
	}

	if r.adminService.IsProtectedAdmin(targetAdminID) {
		return response.Error(ctx, http.StatusForbidden, apperrors.ErrProtectedAdmin.Code)
	}

	targetIsBound, err := r.googleAuthenticatorService.IsBound(targetAdminID)
	if err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"target_admin_id": targetAdminID,
		})
	}
	if !targetIsBound {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrGoogleAuthenticatorNotBound.Code)
	}

	if err := r.googleAuthenticatorService.Unbind(targetAdminID); err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"target_admin_id": targetAdminID,
		})
	}

	return response.Success(ctx, "reset_success")
}

// currentAdminFromContext 从 context 读取当前管理员
// 若 context 中 admin 字段类型非法，按未登录处理
func (r *AdminController) currentAdminFromContext(ctx http.Context) (*models.Admin, http.Response) {
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return nil, nil
	}

	if admin, ok := adminValue.(models.Admin); ok {
		return &admin, nil
	}
	if adminPtr, ok := adminValue.(*models.Admin); ok {
		return adminPtr, nil
	}

	return nil, response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
}

// Export 导出管理员列表
// @Summary      导出管理员列表
// @Description  根据筛选条件导出管理员列表为CSV文件
// @Tags         管理员管理
// @Accept       json
// @Produce      json
// @Param        request body     adminrequests.AdminListFilter false "筛选（字段同列表 GET query）"
// @Success      200     {object} AdminExportResponse "导出成功，返回文件下载信息"
// @Failure      500     {object} apidoc.Error "服务器错误"
// @Router       /api/admin/admins/export [post]
// @Security     BearerAuth
func (r *AdminController) Export(ctx http.Context) http.Response {
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code)
	}

	// 防重复点击：使用框架自带的原子锁（锁会在10秒后自动过期，防止短时间内重复请求）
	lockKey := fmt.Sprintf("export:admins:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10*time.Second)

	// 尝试获取锁，如果获取失败则返回错误
	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, apperrors.ErrGetLockFailed.Code)
	}
	// 同步导出：锁会在 Redis 中自动过期（10秒），不需要手动释放

	filters := r.buildFilters(ctx)

	// 导出时获取所有数据，不分页
	admins, err := r.adminService.GetAllAdminsForExport(filters)
	if err != nil {
		return response.ErrorWithLog(ctx, "admin", err, map[string]any{
			"action":   "export_admins",
			"admin_id": adminID,
		})
	}

	headers := []string{
		"id",
		"username",
		"nickname",
		"email",
		"phone",
		"status",
		"department",
		"position",
		"roles",
		"created_at",
		"updated_at",
	}

	timezone := helpers.GetCurrentTimezone(ctx)
	var data [][]string
	for _, admin := range admins {
		statusText := trans.Get(ctx, "disabled")
		if admin.Status == 1 {
			statusText = trans.Get(ctx, "enabled")
		}

		// 部门名称
		departmentName := ""
		if admin.Department.ID > 0 {
			departmentName = admin.Department.Name
		}

		positionName := ""
		if admin.Position.ID > 0 {
			positionName = admin.Position.Name
		}

		// 角色名称（多个角色用逗号分隔）
		roleNameParts := make([]string, 0, len(admin.Roles))
		for _, role := range admin.Roles {
			roleNameParts = append(roleNameParts, role.Name)
		}
		roleNames := strings.Join(roleNameParts, ", ")

		// 时间格式化
		createdAt := helpers.FormatCarbonWithTimezone(admin.CreatedAt, timezone)
		updatedAt := helpers.FormatCarbonWithTimezone(admin.UpdatedAt, timezone)

		row := []string{
			cast.ToString(admin.ID),
			admin.Username,
			admin.Nickname,
			admin.Email,
			admin.Phone,
			statusText,
			departmentName,
			positionName,
			roleNames,
			createdAt,
			updatedAt,
		}
		data = append(data, row)
	}

	// 在 context 中设置导出类型，供 ExportService 使用
	ctx.WithValue("export_type", models.ExportTypeAdmins)

	return response.Export(ctx, "exported", headers, data, "admins")
}
