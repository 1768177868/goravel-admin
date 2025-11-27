package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/utils/errorlog"
)

type PasswordController struct {
}

func NewPasswordController() *PasswordController {
	return &PasswordController{}
}

func (r *PasswordController) UpdatePassword(ctx http.Context) http.Response {
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	var admin models.Admin
	if adminVal, ok := adminValue.(models.Admin); ok {
		admin = adminVal
	} else if adminPtr, ok := adminValue.(*models.Admin); ok {
		admin = *adminPtr
	} else {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	if err := facades.Orm().Query().Where("id", admin.ID).First(&admin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "admin_not_found")
	}

	oldPassword := ctx.Request().Input("old_password")
	newPassword := ctx.Request().Input("new_password")
	confirmPassword := ctx.Request().Input("confirm_password")

	if oldPassword == "" || newPassword == "" || confirmPassword == "" {
		return response.Error(ctx, http.StatusBadRequest, "params_required")
	}

	if newPassword != confirmPassword {
		return response.Error(ctx, http.StatusBadRequest, "password_not_match")
	}

	if len(newPassword) < 6 {
		return response.Error(ctx, http.StatusBadRequest, "password_length_error")
	}

	if !facades.Hash().Check(oldPassword, admin.Password) {
		return response.Error(ctx, http.StatusBadRequest, "old_password_error")
	}

	hashedPassword, err := facades.Hash().Make(newPassword)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
	}

	admin.Password = hashedPassword
	if err := facades.Orm().Query().Save(&admin); err != nil {
		errorlog.RecordHTTP(ctx, "password", "Failed to update password", map[string]any{
			"error":    err.Error(),
			"admin_id": admin.ID,
		}, "Update password error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "password_update_failed")
	}

	return response.Success(ctx, "password_update_success")
}

// ResetPassword 重置密码（管理员操作）
func (r *PasswordController) ResetPassword(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	newPassword := ctx.Request().Input("password")

	if newPassword == "" {
		return response.Error(ctx, http.StatusBadRequest, "new_password_required")
	}

	if len(newPassword) < 6 {
		return response.Error(ctx, http.StatusBadRequest, "password_length_error")
	}

	var admin models.Admin
	if err := facades.Orm().Query().Where("id", id).First(&admin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "admin_not_found")
	}

	hashedPassword, err := facades.Hash().Make(newPassword)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
	}

	admin.Password = hashedPassword
	if err := facades.Orm().Query().Save(&admin); err != nil {
		errorlog.RecordHTTP(ctx, "password", "Failed to reset password", map[string]any{
			"error":    err.Error(),
			"admin_id": admin.ID,
		}, "Reset password error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "password_reset_failed")
	}

	return response.Success(ctx, "password_reset_success")
}
