package api

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"

	apperrors "goravel/app/errors"
	apirequests "goravel/app/http/requests/api"
	"goravel/app/http/response"
	"goravel/app/models"
)

type AuthController struct {
}

func NewAuthController() *AuthController {
	return &AuthController{}
}

// Register 用户注册
func (r *AuthController) Register(ctx http.Context) http.Response {
	var registerRequest apirequests.UserRegister
	errors, err := ctx.Request().ValidateRequest(&registerRequest)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 检查用户名是否已存在
	exists, err := facades.Orm().Query().Model(&models.User{}).Where("username", registerRequest.Username).Exists()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrCreateFailed.Code)
	}
	if exists {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrUsernameExists.Code)
	}

	// 检查邮箱是否已存在（如果提供了邮箱）
	if registerRequest.Email != "" {
		exists, err := facades.Orm().Query().Model(&models.User{}).Where("email", registerRequest.Email).Exists()
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrCreateFailed.Code)
		}
		if exists {
			return response.Error(ctx, http.StatusBadRequest, "email_already_exists")
		}
	}

	// 检查手机号是否已存在（如果提供了手机号）
	if registerRequest.Phone != "" {
		exists, err := facades.Orm().Query().Model(&models.User{}).Where("phone", registerRequest.Phone).Exists()
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrCreateFailed.Code)
		}
		if exists {
			return response.Error(ctx, http.StatusBadRequest, "phone_already_exists")
		}
	}

	// 密码加密
	hashedPassword, err := facades.Hash().Make(registerRequest.Password)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
	}

	// 如果未设置货币ID，默认使用人民币
	var currencyID uint
	var cnyCurrency models.Currency
	if err := facades.Orm().Query().Where("code", "CNY").First(&cnyCurrency); err == nil {
		currencyID = cnyCurrency.ID
	}

	now := carbon.Now()
	userData := map[string]any{
		"username":    registerRequest.Username,
		"password":    hashedPassword,
		"nickname":    registerRequest.Nickname,
		"avatar":      "",
		"email":       registerRequest.Email,
		"phone":       registerRequest.Phone,
		"balance":     0,
		"currency_id": currencyID,
		"status":      1, // 默认启用
		"created_at":  now,
		"updated_at":  now,
	}

	var user models.User
	if err := facades.Orm().Query().Table("users").Create(userData); err != nil {
		return response.ErrorWithLog(ctx, "user", err, map[string]any{
			"username": registerRequest.Username,
		})
	}

	// 查询创建后的用户
	if err := facades.Orm().Query().Where("username", registerRequest.Username).First(&user); err != nil {
		return response.ErrorWithLog(ctx, "user", err, map[string]any{
			"username": registerRequest.Username,
		})
	}

	// 注册成功后自动登录，使用Goravel标准Auth生成token
	token, err := facades.Auth(ctx).Guard("user").Login(&user)
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"user_id": user.ID,
		})
	}

	return response.SuccessWithHeader(ctx, "register_success", "Authorization", "Bearer "+token, http.Json{
		"token": token,
		"user": http.Json{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"email":    user.Email,
			"phone":    user.Phone,
			"balance":  user.Balance,
			"status":   user.Status,
		},
	})
}

// Login 用户登录
func (r *AuthController) Login(ctx http.Context) http.Response {
	var loginRequest apirequests.UserLogin
	errors, err := ctx.Request().ValidateRequest(&loginRequest)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 验证用户名和密码
	var user models.User
	if err := facades.Orm().Query().Where("username", loginRequest.Username).First(&user); err != nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrUsernameOrPasswordErr.Code)
	}

	if user.ID == 0 {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrUsernameOrPasswordErr.Code)
	}

	if user.Status == 0 {
		return response.Error(ctx, http.StatusForbidden, apperrors.ErrAccountDisabled.Code)
	}

	// 验证密码
	if !facades.Hash().Check(loginRequest.Password, user.Password) {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrUsernameOrPasswordErr.Code)
	}

	// 使用Goravel标准Auth生成token
	token, err := facades.Auth(ctx).Guard("user").Login(&user)
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"user_id": user.ID,
		})
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	facades.Orm().Query().Save(&user)

	return response.SuccessWithHeader(ctx, "login_success", "Authorization", "Bearer "+token, http.Json{
		"token": token,
		"user": http.Json{
			"id":            user.ID,
			"username":      user.Username,
			"nickname":      user.Nickname,
			"avatar":        user.Avatar,
			"email":         user.Email,
			"phone":         user.Phone,
			"balance":       user.Balance,
			"status":        user.Status,
			"last_login_at": user.LastLoginAt,
		},
	})
}

// Info 获取当前用户信息
func (r *AuthController) Info(ctx http.Context) http.Response {
	// 使用Goravel标准Auth获取用户
	var user models.User
	if err := facades.Auth(ctx).Guard("user").User(&user); err != nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	// 重新查询用户以确保获取最新数据（包括关联的货币信息）
	if err := facades.Orm().Query().With("Currency").Where("id", user.ID).First(&user); err != nil {
		return response.Error(ctx, http.StatusNotFound, "user_not_found")
	}

	return response.Success(ctx, http.Json{
		"user": http.Json{
			"id":            user.ID,
			"username":      user.Username,
			"nickname":      user.Nickname,
			"avatar":        user.Avatar,
			"email":         user.Email,
			"phone":         user.Phone,
			"balance":       user.Balance,
			"currency_id":   user.CurrencyID,
			"currency":      user.Currency,
			"status":        user.Status,
			"last_login_at": user.LastLoginAt,
			"created_at":    user.CreatedAt,
			"updated_at":    user.UpdatedAt,
		},
	})
}

// Logout 用户登出
func (r *AuthController) Logout(ctx http.Context) http.Response {
	// 使用Goravel标准Auth登出
	if err := facades.Auth(ctx).Guard("user").Logout(); err != nil {
		// 即使登出失败也返回成功，因为token可能已经过期
		facades.Log().Warningf("Failed to logout: %v", err)
	}

	return response.Success(ctx, "logout_success", http.Json{})
}
