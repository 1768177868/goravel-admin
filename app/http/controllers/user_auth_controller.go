package controllers

import (
	stderrors "errors"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/str"

	apperrors "goravel/app/errors"
	"goravel/app/http/requests/user"
	"goravel/app/http/response"
	"goravel/app/services"
)

type UserAuthController struct {
	userAuthService services.UserAuthService
}

func NewUserAuthController() *UserAuthController {
	tokenService := services.NewTokenServiceImpl()
	userAuthService := services.NewUserAuthServiceImpl(tokenService)
	return &UserAuthController{
		userAuthService: userAuthService,
	}
}

// Login 用户登录
func (r *UserAuthController) Login(ctx http.Context) http.Response {
	var loginRequest user.Login
	errors, err := ctx.Request().ValidateRequest(&loginRequest)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	user, token, err := r.userAuthService.Login(ctx, loginRequest.Username, loginRequest.Password)
	if err != nil {
		// 检查是否是业务错误
		if be, ok := apperrors.GetBusinessError(err); ok {
			switch be.Code {
			case "account_disabled":
				return response.Error(ctx, http.StatusForbidden, "account_disabled")
			case "password_error":
				return response.Error(ctx, http.StatusUnauthorized, "username_or_password_error")
			case "not_logged_in":
				return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
			default:
				return response.Error(ctx, http.StatusInternalServerError, "login_failed")
			}
		}

		// 检查是否是标准错误
		if stderrors.Is(err, apperrors.ErrAccountDisabled) {
			return response.Error(ctx, http.StatusForbidden, "account_disabled")
		}
		if stderrors.Is(err, apperrors.ErrPasswordError) {
			return response.Error(ctx, http.StatusUnauthorized, "username_or_password_error")
		}
		if stderrors.Is(err, apperrors.ErrNotLoggedIn) {
			return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
		}

		// 检查是否是记录不存在错误（通过错误消息判断）
		if strings.Contains(err.Error(), "record not found") {
			return response.Error(ctx, http.StatusUnauthorized, "username_or_password_error")
		}

		// 默认返回登录失败
		return response.Error(ctx, http.StatusInternalServerError, "login_failed")
	}

	return response.SuccessWithHeader(ctx, "login_success", "Authorization", "Bearer "+token, http.Json{
		"token": token,
		"user": http.Json{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
			"avatar":   user.Avatar,
			"mail":     user.Mail,
		},
	})
}

// Info 获取当前登录用户信息
func (r *UserAuthController) Info(ctx http.Context) http.Response {
	user, err := r.userAuthService.GetUserInfo(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	return response.Success(ctx, http.Json{
		"user": http.Json{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
			"avatar":   user.Avatar,
			"alias":    user.Alias,
			"mail":     user.Mail,
			"status":   user.Status,
		},
	})
}

// Logout 退出登录
func (r *UserAuthController) Logout(ctx http.Context) http.Response {
	// 从context中获取user信息（由JWT中间件设置）
	userValue := ctx.Value("user")
	if userValue != nil {
		// 获取token
		token := ctx.Request().Header("Authorization", "")
		token = str.Of(token).ChopStart("Bearer ").Trim().String()

		if token != "" {
			// 删除token
			tokenService := services.NewTokenServiceImpl()
			_ = tokenService.DeleteToken(token)
		}
	}

	return response.Success(ctx, "logout_success")
}
