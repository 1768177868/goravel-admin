package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/models"
	"goravel/app/services"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController() *AuthController {
	adminService := services.NewAdminServiceImpl()
	authService := services.NewAuthServiceImpl(adminService)
	return &AuthController{
		authService: authService,
	}
}

// Login 管理员登录
func (r *AuthController) Login(ctx http.Context) http.Response {
	var loginRequest admin.Login
	errors, err := ctx.Request().ValidateRequest(&loginRequest)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	admin, token, err := r.authService.Login(ctx, loginRequest.Username, loginRequest.Password)
	if err != nil {
		// 检查是否是账号被禁用
		if err.Error() == "account_disabled" {
			return response.Error(ctx, http.StatusForbidden, "account_disabled")
		}
		// 检查是否是用户名或密码错误
		if err.Error() == "password_error" || err.Error() == "record not found" {
			return response.Error(ctx, http.StatusUnauthorized, "username_or_password_error")
		}
		// 其他错误（登录失败）
		return response.Error(ctx, http.StatusInternalServerError, "login_failed")
	}

	return response.SuccessWithHeader(ctx, "login_success", "Authorization", "Bearer "+token, http.Json{
		"token": token,
		"admin": http.Json{
			"id":       admin.ID,
			"username": admin.Username,
			"nickname": admin.Nickname,
			"avatar":   admin.Avatar,
		},
	})
}

// Info 获取当前登录管理员信息
func (r *AuthController) Info(ctx http.Context) http.Response {
	admin, permissions, menus, err := r.authService.GetAdminInfo(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	return response.Success(ctx, "get_success", http.Json{
		"admin": http.Json{
			"id":            admin.ID,
			"username":      admin.Username,
			"nickname":      admin.Nickname,
			"avatar":        admin.Avatar,
			"email":         admin.Email,
			"phone":         admin.Phone,
			"department_id": admin.DepartmentID,
			"department":    admin.Department,
			"roles":         admin.Roles,
			"permissions":   permissions,
			"menus":         menus,
		},
	})
}

// Logout 退出登录
func (r *AuthController) Logout(ctx http.Context) http.Response {
	id, err := facades.Auth(ctx).Guard("admin").ID()
	if err == nil {
		// 记录退出日志
		var admin models.Admin
		if err := facades.Auth(ctx).Guard("admin").User(&admin); err == nil {
			r.authService.RecordLoginLog(ctx, admin.ID, admin.Username, 1, trans.Get(ctx, "logout_success"))
		} else {
			r.authService.RecordLoginLog(ctx, cast.ToUint(id), "", 1, trans.Get(ctx, "logout_success"))
		}
	}

	return response.Success(ctx, "logout_success")
}
