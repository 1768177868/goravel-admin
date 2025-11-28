package services

import (
	"errors"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
	"goravel/app/models"
)

type UserAuthService interface {
	// Login 用户登录
	Login(ctx http.Context, username, password string) (*models.User, string, error)
	// GetUserInfo 获取用户信息
	GetUserInfo(ctx http.Context) (*models.User, error)
}

type UserAuthServiceImpl struct {
	tokenService TokenService
}

func NewUserAuthServiceImpl(tokenService TokenService) *UserAuthServiceImpl {
	return &UserAuthServiceImpl{
		tokenService: tokenService,
	}
}

// Login 用户登录
func (s *UserAuthServiceImpl) Login(ctx http.Context, username, password string) (*models.User, string, error) {
	var user models.User
	if err := facades.Orm().Query().Where("username", username).First(&user); err != nil {
		return nil, "", err
	}

	if user.Status == 0 {
		return nil, "", errors.New("account_disabled")
	}

	// 验证密码
	if !facades.Hash().Check(password, user.Password) {
		return nil, "", errors.New("password_error")
	}

	// 生成token并存入数据库（类似Laravel Sanctum）
	var expiresAt *time.Time
	ttl := facades.Config().GetInt("jwt.ttl", 60) // 默认60分钟
	if ttl > 0 {
		exp := time.Now().Add(time.Duration(ttl) * time.Minute)
		expiresAt = &exp
	}

	// 获取浏览器和操作系统信息
	browser, os := helpers.GetBrowserAndOS(ctx)
	// 获取真实IP地址
	ip := helpers.GetRealIP(ctx)

	plainToken, _, err := s.tokenService.CreateToken("user", user.ID, "user-token", expiresAt, browser, ip, os, "")
	if err != nil {
		return nil, "", err
	}

	return &user, plainToken, nil
}

// GetUserInfo 获取用户信息
func (s *UserAuthServiceImpl) GetUserInfo(ctx http.Context) (*models.User, error) {
	// 从context中获取user信息（由JWT中间件设置）
	userValue := ctx.Value("user")
	if userValue == nil {
		return nil, errors.New("not_logged_in")
	}

	user, ok := userValue.(models.User)
	if !ok {
		return nil, errors.New("not_logged_in")
	}

	return &user, nil
}
