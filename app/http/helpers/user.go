package helpers

import (
	"errors"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/str"

	"goravel/app/models"
)

// GetUserFromContext 从 context 中获取 user 对象
// 如果 context 中没有 user 或类型不匹配，返回错误
func GetUserFromContext(ctx http.Context) (*models.User, error) {
	userValue := ctx.Value("user")
	if userValue == nil {
		return nil, errors.New("user not found in context")
	}

	// 尝试值类型
	if user, ok := userValue.(models.User); ok {
		return &user, nil
	}

	// 尝试指针类型
	if userPtr, ok := userValue.(*models.User); ok {
		return userPtr, nil
	}

	return nil, errors.New("invalid user type in context")
}

// GetUserIDFromContext 从 context 中获取 user ID
// 如果 context 中没有 user 或类型不匹配，返回 0 和错误
func GetUserIDFromContext(ctx http.Context) (uint, error) {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

// GetTokenFromHeader 从请求头中获取token
// 支持从Authorization header或URL参数中获取
func GetTokenFromHeader(ctx http.Context) string {
	token := ctx.Request().Header("Authorization", "")

	// 如果 Header 中没有 token，尝试从 URL 参数中获取
	if str.Of(token).IsEmpty() {
		token = ctx.Request().Query("_token", "")
	}

	// 移除Bearer前缀（如果有）
	token = str.Of(token).ChopStart("Bearer ").Trim().String()

	return token
}
