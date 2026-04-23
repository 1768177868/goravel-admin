package middleware

import (
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/response"
	"goravel/app/models"
)

func DevelopmentOrDeveloperOnly() http.Middleware {
	return func(ctx http.Context) {
		env := facades.Config().GetString("app.env", "production")
		if env == "local" || env == "development" {
			ctx.Request().Next()
			return
		}

		adminID := currentAdminID(ctx)
		if isDeveloperAdmin(adminID, facades.Config().GetString("admin.developer_ids", "2")) {
			ctx.Request().Next()
			return
		}

		response.Error(ctx, http.StatusForbidden, "development_only")
		ctx.Request().Abort()
	}
}

func currentAdminID(ctx http.Context) uint {
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return 0
	}

	if admin, ok := adminValue.(models.Admin); ok {
		return admin.ID
	}
	if adminPtr, ok := adminValue.(*models.Admin); ok && adminPtr != nil {
		return adminPtr.ID
	}

	return 0
}

func isDeveloperAdmin(adminID uint, developerIDsStr string) bool {
	if adminID == 0 || developerIDsStr == "" {
		return false
	}

	parts := strings.Split(developerIDsStr, ",")
	for _, part := range parts {
		id := cast.ToUint(strings.TrimSpace(part))
		if id > 0 && id == adminID {
			return true
		}
	}

	return false
}
