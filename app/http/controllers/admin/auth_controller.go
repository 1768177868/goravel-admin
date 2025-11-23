package admin

import (
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

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
	tokenService := services.NewTokenServiceImpl()
	authService := services.NewAuthServiceImpl(adminService, tokenService)
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

// UpdateProfile 更新个人信息
func (r *AuthController) UpdateProfile(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	var admin models.Admin
	// 尝试值类型
	if adminVal, ok := adminValue.(models.Admin); ok {
		admin = adminVal
	} else if adminPtr, ok := adminValue.(*models.Admin); ok {
		// 尝试指针类型
		admin = *adminPtr
	} else {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	// 重新查询admin以确保获取最新数据
	if err := facades.Orm().Query().Where("id", admin.ID).First(&admin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "admin_not_found")
	}

	nickname := ctx.Request().Input("nickname")
	email := ctx.Request().Input("email")
	phone := ctx.Request().Input("phone")
	avatar := ctx.Request().Input("avatar")

	if nickname != "" {
		admin.Nickname = nickname
	}
	if email != "" {
		admin.Email = email
	}
	if phone != "" {
		admin.Phone = phone
	}
	if avatar != "" {
		admin.Avatar = avatar
	}

	if err := facades.Orm().Query().Save(&admin); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	// 重新加载关联数据（确保部门和角色被正确加载）
	var adminWithRelations models.Admin
	if err := facades.Orm().Query().With("Department").With("Roles").Where("id", admin.ID).First(&adminWithRelations); err != nil {
		facades.Log().Errorf("UpdateProfile: failed to load admin with relations, error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}
	admin = adminWithRelations

	facades.Log().Debugf("UpdateProfile: admin ID: %d, Department: %+v, Roles count: %d", admin.ID, admin.Department, len(admin.Roles))

	return response.Success(ctx, "update_success", http.Json{
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
		},
	})
}

// Refresh 刷新Token
// 注意：此接口需要在JWT中间件之前调用，或者使用特殊的中间件处理
// 因为Refresh方法需要token过期但仍在刷新窗口内才能工作
func (r *AuthController) Refresh(ctx http.Context) http.Response {
	// 从请求头获取token
	token := ctx.Request().Header("Authorization", "")
	if token == "" {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	// 移除Bearer前缀
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	// 先尝试解析token，如果token有效，直接重新生成（滑动过期）
	if _, err := facades.Auth(ctx).Guard("admin").Parse(token); err == nil {
		// Token有效，重新生成新token（延长过期时间）
		if userID, err := facades.Auth(ctx).Guard("admin").ID(); err == nil {
			if newToken, err := facades.Auth(ctx).Guard("admin").LoginUsingID(userID); err == nil {
				return response.SuccessWithHeader(ctx, "token_refresh_success", "Authorization", "Bearer "+newToken, http.Json{
					"token": newToken,
				})
			}
		}
	}

	// 如果token已过期，尝试刷新（需要在刷新窗口内）
	newToken, err := facades.Auth(ctx).Guard("admin").Refresh()
	if err != nil {
		// 刷新失败，返回错误
		return response.Error(ctx, http.StatusUnauthorized, "token_refresh_failed")
	}

	// 刷新成功，返回新token
	return response.SuccessWithHeader(ctx, "token_refresh_success", "Authorization", "Bearer "+newToken, http.Json{
		"token": newToken,
	})
}

// Logout 退出登录
func (r *AuthController) Logout(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue != nil {
		if admin, ok := adminValue.(models.Admin); ok {
			// 获取token
			token := ctx.Request().Header("Authorization", "")
			token = strings.TrimPrefix(token, "Bearer ")
			token = strings.TrimSpace(token)

			// 删除token
			tokenService := services.NewTokenServiceImpl()
			_ = tokenService.DeleteToken(token)

			// 记录退出日志
			r.authService.RecordLoginLog(ctx, admin.ID, admin.Username, 1, trans.Get(ctx, "logout_success"))
		}
	}

	return response.Success(ctx, "logout_success")
}

// Tokens 获取当前用户的所有token列表
func (r *AuthController) Tokens(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	admin, ok := adminValue.(models.Admin)
	if !ok {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	// 获取用户的所有token
	tokenService := services.NewTokenServiceImpl()
	tokens, err := tokenService.GetTokensByUser("admin", admin.ID)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	// 获取当前使用的token
	currentTokenValue := ctx.Value("token")
	var currentTokenID uint
	if currentTokenValue != nil {
		if currentToken, ok := currentTokenValue.(models.PersonalAccessToken); ok {
			currentTokenID = currentToken.ID
		}
	}

	// 格式化token列表
	var tokenList []http.Json
	for _, token := range tokens {
		tokenData := http.Json{
			"id":           token.ID,
			"name":         token.Name,
			"last_used_at": token.LastUsedAt,
			"expires_at":   token.ExpiresAt,
			"created_at":   token.CreatedAt,
			"is_current":   token.ID == currentTokenID,
		}
		tokenList = append(tokenList, tokenData)
	}

	return response.Success(ctx, "get_success", http.Json{
		"tokens": tokenList,
	})
}

// RevokeToken 删除指定的token（踢出指定设备）
func (r *AuthController) RevokeToken(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	admin, ok := adminValue.(models.Admin)
	if !ok {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	// 获取要删除的token ID
	tokenIDStr := ctx.Request().Route("id")
	if tokenIDStr == "" {
		return response.Error(ctx, http.StatusBadRequest, "token_id_required")
	}

	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_token_id")
	}

	// 查询token是否存在且属于当前用户
	var token models.PersonalAccessToken
	if err := facades.Orm().Query().
		Where("id", tokenID).
		Where("tokenable_type", "admin").
		Where("tokenable_id", admin.ID).
		First(&token); err != nil {
		return response.Error(ctx, http.StatusNotFound, "token_not_found")
	}

	// 删除token（直接通过ID删除，因为数据库中存储的是hash值，无法获取原始token）
	_, err = facades.Orm().Query().Delete(&token)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "revoke_success")
}

// RevokeAllTokens 删除当前用户的所有token（踢出所有设备）
func (r *AuthController) RevokeAllTokens(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	admin, ok := adminValue.(models.Admin)
	if !ok {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	// 删除用户的所有token
	tokenService := services.NewTokenServiceImpl()
	if err := tokenService.DeleteTokensByUser("admin", admin.ID); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "revoke_all_success")
}

// KickOutUser 踢出指定用户的所有token（管理员操作）
func (r *AuthController) KickOutUser(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	// 获取要踢出的用户ID
	userIDStr := ctx.Request().Route("id")
	if userIDStr == "" {
		return response.Error(ctx, http.StatusBadRequest, "user_id_required")
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_user_id")
	}

	// 查询用户是否存在
	var targetAdmin models.Admin
	if err := facades.Orm().Query().Where("id", userID).First(&targetAdmin); err != nil {
		return response.Error(ctx, http.StatusNotFound, "user_not_found")
	}

	// 删除用户的所有token
	tokenService := services.NewTokenServiceImpl()
	if err := tokenService.DeleteTokensByUser("admin", targetAdmin.ID); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "kick_out_success")
}
