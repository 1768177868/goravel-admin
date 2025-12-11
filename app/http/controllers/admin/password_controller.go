package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/models"
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

	// 使用请求验证
	var updatePasswordRequest adminrequests.UpdatePassword
	errors, err := ctx.Request().ValidateRequest(&updatePasswordRequest)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 验证旧密码是否正确
	if !facades.Hash().Check(updatePasswordRequest.OldPassword, admin.Password) {
		return response.Error(ctx, http.StatusBadRequest, "old_password_error")
	}

	hashedPassword, err := facades.Hash().Make(updatePasswordRequest.NewPassword)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
	}

	admin.Password = hashedPassword
	if err := facades.Orm().Query().Save(&admin); err != nil {
		return response.ErrorWithLog(ctx, "password", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	return response.Success(ctx, "password_update_success")
}

// ResetPassword 重置密码（管理员操作）
func (r *PasswordController) ResetPassword(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))

	// 使用请求验证
	var resetPasswordRequest adminrequests.ResetPassword
	errors, err := ctx.Request().ValidateRequest(&resetPasswordRequest)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	var admin models.Admin
	if err := facades.Orm().Query().Where("id", id).First(&admin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "admin_not_found")
	}

	hashedPassword, err := facades.Hash().Make(resetPasswordRequest.Password)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
	}

	admin.Password = hashedPassword
	if err := facades.Orm().Query().Save(&admin); err != nil {
		return response.ErrorWithLog(ctx, "password", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	return response.Success(ctx, "password_reset_success")
}
