package admin

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
)

type AuthController struct {
	authService                services.AuthService
	captchaService             services.CaptchaService
	googleAuthenticatorService services.GoogleAuthenticatorService
	treeService                services.TreeService
	lockoutService             services.LoginLockoutService
}

func NewAuthController() *AuthController {
	adminService := services.NewAdminServiceImpl()
	tokenService := services.NewTokenServiceImpl()
	authService := services.NewAuthServiceImpl(adminService, tokenService)
	return &AuthController{
		authService:                authService,
		captchaService:             services.NewCaptchaServiceImpl(),
		googleAuthenticatorService: services.NewGoogleAuthenticatorServiceImpl(),
		treeService:                services.NewTreeServiceImpl(),
		lockoutService:             services.NewLoginLockoutService(),
	}
}

// getLoginRequestData 获取登录请求数据（排除敏感信息）
func (r *AuthController) getLoginRequestData(ctx http.Context) string {
	inputs := make(map[string]any)
	allInputs := ctx.Request().All()
	for key, value := range allInputs {
		// 使用工具函数检查是否是敏感字段
		if utils.IsSensitiveField(key) {
			inputs[key] = "***"
		} else {
			inputs[key] = value
		}
	}
	if data, err := json.Marshal(inputs); err == nil {
		return string(data)
	}
	return ""
}

// Login 管理员登录
func (r *AuthController) Login(ctx http.Context) http.Response {
	var loginRequest admin.Login
	errors, err := ctx.Request().ValidateRequest(&loginRequest)
	if err != nil {
		requestData := r.getLoginRequestData(ctx)
		r.authService.RecordLoginLog(ctx, 0, loginRequest.Username, 0, "validation_failed", requestData)
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		requestData := r.getLoginRequestData(ctx)
		r.authService.RecordLoginLog(ctx, 0, loginRequest.Username, 0, "validation_failed", requestData)
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	requestData := r.getLoginRequestData(ctx)
	ip := helpers.GetRealIP(ctx)

	// ---- 登录失败锁定检查 ----
	if locked, _ := r.lockoutService.IsLocked(ip, loginRequest.Username); locked {
		lockMinutes := facades.Config().GetInt("login_security.lock_duration_minutes", 15)
		r.authService.RecordLoginLog(ctx, 0, loginRequest.Username, 0, "login_locked", requestData)
		return response.Error(ctx, http.StatusTooManyRequests,
			apperrors.ErrLoginLocked.WithParams(map[string]any{"minutes": lockMinutes}))
	}

	// 验证用户名是否存在
	exists, err := facades.Orm().Query().Model(&models.Admin{}).Where("username", loginRequest.Username).Exists()
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"username": loginRequest.Username,
		})
	}
	if !exists {
		r.lockoutService.RecordFailure(ip, loginRequest.Username)
		r.authService.RecordLoginLog(ctx, 0, loginRequest.Username, 0, "username_not_found", requestData)
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrUsernameOrPasswordErr.Code)
	}

	// 获取管理员信息
	var admin models.Admin
	if err := facades.Orm().Query().Where("username", loginRequest.Username).FirstOrFail(&admin); err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"username": loginRequest.Username,
		})
	}

	if admin.Status == 0 {
		r.authService.RecordLoginLog(ctx, admin.ID, loginRequest.Username, 0, "account_disabled", requestData)
		return response.Error(ctx, http.StatusForbidden, apperrors.ErrAccountDisabled.Code)
	}

	// 验证密码
	if !facades.Hash().Check(loginRequest.Password, admin.Password) {
		r.lockoutService.RecordFailure(ip, loginRequest.Username)
		r.authService.RecordLoginLog(ctx, admin.ID, loginRequest.Username, 0, "password_error", requestData)
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrUsernameOrPasswordErr.Code)
	}

	// 检查是否绑定了谷歌验证码
	isBound, err := r.googleAuthenticatorService.IsBound(admin.ID)
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	// 如果绑定了谷歌验证码，验证谷歌验证码
	if isBound {
		googleCode := loginRequest.GoogleCode
		if googleCode == "" {
			r.authService.RecordLoginLog(ctx, admin.ID, loginRequest.Username, 0, "google_code_required", requestData)
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrGoogleCodeRequired.Code)
		}

		secret, err := r.googleAuthenticatorService.GetSecret(admin.ID)
		if err != nil {
			return response.ErrorWithLog(ctx, "auth", err, map[string]any{
				"admin_id": admin.ID,
			})
		}

		if !r.googleAuthenticatorService.Verify(secret, googleCode) {
			r.lockoutService.RecordFailure(ip, loginRequest.Username)
			r.authService.RecordLoginLog(ctx, admin.ID, loginRequest.Username, 0, "google_code_error", requestData)
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrGoogleCodeInvalid.Code)
		}
	} else {
		if r.captchaService.Enabled() {
			captchaID := ctx.Request().Input("captcha_id")
			captchaAnswer := ctx.Request().Input("captcha_answer")
			if ok, messageKey := r.captchaService.Verify(captchaID, captchaAnswer); !ok {
				if messageKey == "" {
					messageKey = "captcha_invalid"
				}
				r.authService.RecordLoginLog(ctx, admin.ID, loginRequest.Username, 0, messageKey, requestData)
				return response.Error(ctx, http.StatusBadRequest, messageKey)
			}
		}
	}

	// 所有验证通过，清除失败计数
	r.lockoutService.ClearFailures(ip, loginRequest.Username)

	// 验证通过，生成token并完成登录
	browser, os := helpers.GetBrowserAndOS(ctx)

	// 生成token
	var expiresAt *time.Time
	ttl := facades.Config().GetInt("jwt.ttl", 60) // 默认60分钟
	if ttl > 0 {
		exp := time.Now().Add(time.Duration(ttl) * time.Minute)
		expiresAt = &exp
	}

	tokenService := services.NewTokenServiceImpl()
	plainToken, _, err := tokenService.CreateToken("admin", admin.ID, "admin-token", expiresAt, browser, ip, os, "")
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}
	token := plainToken

	// 记录登录成功日志
	r.authService.RecordLoginLog(ctx, admin.ID, loginRequest.Username, 1, "login_success", requestData)

	// 更新最后登录时间（ORM会自动更新UpdatedAt）
	facades.Orm().Query().Save(&admin)

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

// Captcha 获取登录验证码
func (r *AuthController) Captcha(ctx http.Context) http.Response {
	enabled := r.captchaService.Enabled()
	captchaData := http.Json{
		"enabled": enabled,
	}

	if enabled {
		captchaID, image, err := r.captchaService.Generate()
		if err != nil {
			return response.ErrorWithLog(ctx, "captcha", err)
		}
		captchaData["captcha_id"] = captchaID
		captchaData["captcha_image"] = image
		// captchaData["captcha_image"] = "data:image/png;base64," + image
	}

	return response.Success(ctx, http.Json{
		"captcha": captchaData,
	})
}

// Info 获取当前登录管理员信息
func (r *AuthController) Info(ctx http.Context) http.Response {
	admin, permissions, menus, err := r.authService.GetAdminInfo(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	// 获取配置：是否显示无权限的按钮
	showButtonsWithoutPermission := facades.Config().GetBool("admin.show_buttons_without_permission", false)

	// 检查是否是超级管理员
	const SuperAdminRoleSlug = "super-admin"
	isSuperAdmin := false
	for _, role := range admin.Roles {
		if role.Slug == SuperAdminRoleSlug && role.Status == 1 {
			isSuperAdmin = true
			break
		}
	}

	// 将扁平菜单数组构建为树形结构，然后转换为前端格式
	// 可见菜单 = 角色勾选的菜单 + 权限关联的菜单（只要勾选了某目录下任一权限，该目录就显示）+ 这些菜单的所有祖先
	var menuTree []models.Menu
	menuIDSet := make(map[uint]bool)
	for _, menu := range menus {
		menuIDSet[menu.ID] = true
	}
	for _, perm := range permissions {
		if perm.MenuID > 0 {
			menuIDSet[perm.MenuID] = true
		}
	}
	var menuIDs []uint
	for id := range menuIDSet {
		menuIDs = append(menuIDs, id)
	}
	if len(menuIDs) > 0 {
		expandedIDs, err := r.treeService.GetMenuIDsWithAncestors(menuIDs)
		if err == nil && len(expandedIDs) > 0 {
			menuIDSet = make(map[uint]bool)
			for _, id := range expandedIDs {
				menuIDSet[id] = true
			}
		}
		// 先构建完整树形结构
		menuTree, _ = r.treeService.BuildMenuTree(0)
		// 递归过滤树形结构，只保留有权限的菜单（含因权限而显示的目录）
		var filterMenuTree func([]models.Menu) []models.Menu
		filterMenuTree = func(menuList []models.Menu) []models.Menu {
			var result []models.Menu
			for _, menu := range menuList {
				if menuIDSet[menu.ID] {
					menu.Children = filterMenuTree(menu.Children)
					result = append(result, menu)
				}
			}
			return result
		}
		menuTree = filterMenuTree(menuTree)
	}

	// 转换为前端格式
	menuTreeData := utils.ConvertMenuTree(menuTree)

	return response.Success(ctx, http.Json{
		"admin": http.Json{
			"id":             admin.ID,
			"username":       admin.Username,
			"nickname":       admin.Nickname,
			"avatar":         admin.Avatar,
			"email":          admin.Email,
			"phone":          admin.Phone,
			"department_id":  admin.DepartmentID,
			"department":     admin.Department,
			"position_id":    admin.PositionID,
			"position":       admin.Position,
			"roles":          admin.Roles,
			"permissions":    permissions,
			"menus":          menuTreeData, // 返回树形结构
			"is_super_admin": isSuperAdmin,
		},
		"config": http.Json{
			"show_buttons_without_permission": showButtonsWithoutPermission,
			"monitor_hidden":                  facades.Config().GetString("admin.monitor_hidden", ""),
		},
	})
}

// UpdateProfile 更新个人信息
func (r *AuthController) UpdateProfile(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	var admin models.Admin
	// 尝试值类型
	if adminVal, ok := adminValue.(models.Admin); ok {
		admin = adminVal
	} else if adminPtr, ok := adminValue.(*models.Admin); ok {
		// 尝试指针类型
		if adminPtr == nil {
			return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
		}
		admin = *adminPtr
	} else {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	// 重新查询admin以确保获取最新数据
	if err := facades.Orm().Query().Where("id", admin.ID).FirstOrFail(&admin); err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrAdminNotFound.Code)
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
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	// 重新加载关联数据（确保部门和角色被正确加载）
	var adminWithRelations models.Admin
	if err := facades.Orm().Query().With("Department").With("Position").With("Roles").Where("id", admin.ID).FirstOrFail(&adminWithRelations); err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}
	admin = adminWithRelations

	return response.Success(ctx, http.Json{
		"admin": http.Json{
			"id":            admin.ID,
			"username":      admin.Username,
			"nickname":      admin.Nickname,
			"avatar":        admin.Avatar,
			"email":         admin.Email,
			"phone":         admin.Phone,
			"department_id": admin.DepartmentID,
			"department":    admin.Department,
			"position_id":   admin.PositionID,
			"position":      admin.Position,
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
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code)
	}

	// 移除Bearer前缀
	token = str.Of(token).ChopStart("Bearer ").Trim().String()

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
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrTokenRefreshFailed.Code)
	}

	// 刷新成功，返回新token
	return response.SuccessWithHeader(ctx, "token_refresh_success", "Authorization", "Bearer "+newToken, http.Json{
		"token": newToken,
	})
}

// Heartbeat 心跳接口，用于更新用户的最后活跃时间
// JWT中间件会自动更新 last_used_at，这个接口只是确保用户在线状态
func (r *AuthController) Heartbeat(ctx http.Context) http.Response {
	// JWT中间件已经更新了 last_used_at，这里只需要返回成功即可
	return response.Success(ctx, "heartbeat_success")
}

// Logout 退出登录
func (r *AuthController) Logout(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue != nil {
		var admin models.Admin
		if adminVal, ok := adminValue.(models.Admin); ok {
			admin = adminVal
		} else if adminPtr, ok := adminValue.(*models.Admin); ok {
			if adminPtr != nil {
				admin = *adminPtr
			}
		}

		if admin.ID > 0 {
			// 获取token
			token := ctx.Request().Header("Authorization", "")
			token = str.Of(token).ChopStart("Bearer ").Trim().String()

			if token != "" {
				// 删除token
				tokenService := services.NewTokenServiceImpl()
				_ = tokenService.DeleteToken(token)
			}

			// 记录退出日志
			logoutRequestData := r.getLoginRequestData(ctx)
			r.authService.RecordLoginLog(ctx, admin.ID, admin.Username, 1, "logout_success", logoutRequestData)
		}
	}

	return response.Success(ctx, "logout_success")
}

// Tokens 获取当前用户的所有token列表
func (r *AuthController) Tokens(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	admin, ok := adminValue.(models.Admin)
	if !ok {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	// 获取用户的所有token
	tokenService := services.NewTokenServiceImpl()
	tokens, err := tokenService.GetTokensByUser("admin", admin.ID)
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	// 获取当前使用的token
	currentTokenValue := ctx.Value("token")
	var currentTokenID uint
	if currentTokenValue != nil {
		if currentToken, ok := currentTokenValue.(models.PersonalAccessToken); ok {
			currentTokenID = currentToken.ID
		} else if currentTokenPtr, ok := currentTokenValue.(*models.PersonalAccessToken); ok {
			if currentTokenPtr != nil {
				currentTokenID = currentTokenPtr.ID
			}
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

	return response.Success(ctx, http.Json{
		"tokens": tokenList,
	})
}

// RevokeToken 删除指定的token（踢出指定设备）
func (r *AuthController) RevokeToken(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	admin, ok := adminValue.(models.Admin)
	if !ok {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	// 获取要删除的token ID
	tokenIDStr := ctx.Request().Route("id")
	if tokenIDStr == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrTokenIDRequired.Code)
	}

	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidTokenID.Code)
	}

	// 查询token是否存在且属于当前用户
	var token models.PersonalAccessToken
	if err := facades.Orm().Query().
		Where("id", tokenID).
		Where("tokenable_type", "admin").
		Where("tokenable_id", admin.ID).
		First(&token); err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrTokenNotFound.Code)
	}

	// 删除token（直接通过ID删除，因为数据库中存储的是hash值，无法获取原始token）
	_, err = facades.Orm().Query().Delete(&token)
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"token_id": token.ID,
			"admin_id": admin.ID,
		})
	}

	return response.Success(ctx, "revoke_success")
}

// RevokeAllTokens 删除当前用户的所有token（踢出所有设备）
func (r *AuthController) RevokeAllTokens(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	admin, ok := adminValue.(models.Admin)
	if !ok {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	// 删除用户的所有token
	tokenService := services.NewTokenServiceImpl()
	if err := tokenService.DeleteTokensByUser("admin", admin.ID); err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	return response.Success(ctx, "revoke_all_success")
}

// KickOutUser 踢出指定用户的所有token（管理员操作）
func (r *AuthController) KickOutUser(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	var admin models.Admin
	if adminVal, ok := adminValue.(models.Admin); ok {
		admin = adminVal
	} else if adminPtr, ok := adminValue.(*models.Admin); ok {
		admin = *adminPtr
	} else {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	// 获取要踢出的用户ID
	userIDStr := ctx.Request().Route("id")
	if userIDStr == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrUserIDRequired.Code)
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidUserID.Code)
	}

	// 查询用户是否存在
	var targetAdmin models.Admin
	if err := facades.Orm().Query().Where("id", userID).FirstOrFail(&targetAdmin); err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrUserNotFound.Code)
	}

	// 删除用户的所有token
	tokenService := services.NewTokenServiceImpl()
	if err := tokenService.DeleteTokensByUser("admin", targetAdmin.ID); err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"target_user_id": targetAdmin.ID,
			"operator_id":    admin.ID,
		})
	}

	return response.Success(ctx, "kick_out_success")
}

// GetGoogleAuthenticatorQRCode 获取谷歌验证码二维码（用于绑定）
func (r *AuthController) GetGoogleAuthenticatorQRCode(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	var admin models.Admin
	if adminVal, ok := adminValue.(models.Admin); ok {
		admin = adminVal
	} else if adminPtr, ok := adminValue.(*models.Admin); ok {
		admin = *adminPtr
	} else {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	// 检查是否已经绑定
	isBound, err := r.googleAuthenticatorService.IsBound(admin.ID)
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	if isBound {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrGoogleAuthenticatorAlreadyBound.Code)
	}

	// 生成密钥和二维码
	accountName := admin.Username
	if admin.Email != "" {
		accountName = admin.Email
	}
	secret, qrCodeURL, err := r.googleAuthenticatorService.GenerateSecret(accountName)
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	// 生成二维码图片
	qrCodeImage, err := r.googleAuthenticatorService.GenerateQRCodeImage(accountName, secret)
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	return response.Success(ctx, http.Json{
		"secret":        secret,
		"qr_code_url":   qrCodeURL,
		"qr_code_image": qrCodeImage,
	})
}

// BindGoogleAuthenticator 绑定谷歌验证码
func (r *AuthController) BindGoogleAuthenticator(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	var admin models.Admin
	if adminVal, ok := adminValue.(models.Admin); ok {
		admin = adminVal
	} else if adminPtr, ok := adminValue.(*models.Admin); ok {
		admin = *adminPtr
	} else {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	secret := ctx.Request().Input("secret")
	code := ctx.Request().Input("code")

	if secret == "" || code == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrSecretAndCodeRequired.Code)
	}

	// 绑定谷歌验证码
	if err := r.googleAuthenticatorService.Bind(admin.ID, secret, code); err != nil {
		if err.Error() == "invalid_code" {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrGoogleCodeInvalid.Code)
		}
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	return response.Success(ctx, "bind_success")
}

// UnbindGoogleAuthenticator 解绑谷歌验证码
func (r *AuthController) UnbindGoogleAuthenticator(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	var admin models.Admin
	if adminVal, ok := adminValue.(models.Admin); ok {
		admin = adminVal
	} else if adminPtr, ok := adminValue.(*models.Admin); ok {
		admin = *adminPtr
	} else {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	// 需要验证码确认
	code := ctx.Request().Input("code")
	if code == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrCodeRequired.Code)
	}

	// 获取管理员的密钥
	secret, err := r.googleAuthenticatorService.GetSecret(admin.ID)
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	if secret == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrGoogleAuthenticatorNotBound.Code)
	}

	// 验证验证码
	if !r.googleAuthenticatorService.Verify(secret, code) {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrGoogleCodeInvalid.Code)
	}

	// 解绑谷歌验证码
	if err := r.googleAuthenticatorService.Unbind(admin.ID); err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	return response.Success(ctx, "unbind_success")
}

// GetGoogleAuthenticatorStatus 获取谷歌验证码绑定状态
func (r *AuthController) GetGoogleAuthenticatorStatus(ctx http.Context) http.Response {
	// 从context中获取admin信息（由JWT中间件设置）
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	var admin models.Admin
	if adminVal, ok := adminValue.(models.Admin); ok {
		admin = adminVal
	} else if adminPtr, ok := adminValue.(*models.Admin); ok {
		admin = *adminPtr
	} else {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	// 检查是否绑定
	isBound, err := r.googleAuthenticatorService.IsBound(admin.ID)
	if err != nil {
		return response.ErrorWithLog(ctx, "auth", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	return response.Success(ctx, http.Json{
		"is_bound": isBound,
	})
}
