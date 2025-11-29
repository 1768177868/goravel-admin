package helpers

import (
	"errors"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/models"
)

// GetAdminFromContext 从 context 中获取 admin 对象
// 如果 context 中没有 admin 或类型不匹配，返回错误
func GetAdminFromContext(ctx http.Context) (*models.Admin, error) {
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return nil, errors.New("admin not found in context")
	}

	// 尝试值类型
	if admin, ok := adminValue.(models.Admin); ok {
		return &admin, nil
	}

	// 尝试指针类型
	if adminPtr, ok := adminValue.(*models.Admin); ok {
		return adminPtr, nil
	}

	return nil, errors.New("invalid admin type in context")
}

// GetAdminIDFromContext 从 context 中获取 admin ID
// 如果 context 中没有 admin 或类型不匹配，返回 0 和错误
func GetAdminIDFromContext(ctx http.Context) (uint, error) {
	admin, err := GetAdminFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return admin.ID, nil
}
