package providers

import (
	frameworkauth "github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"
)

type AuthServiceProvider struct {
}

func (receiver *AuthServiceProvider) Register(app foundation.Application) {

}

func (receiver *AuthServiceProvider) Boot(app foundation.Application) {
	facades.Auth().Extend("another-jwt", frameworkauth.NewJwtGuard)
	facades.Auth().Provider("another-orm", frameworkauth.NewOrmUserProvider)

	// 注册永久token的JWT guard（使用admin-permanent-jwt配置）
	facades.Auth().Extend("admin-permanent-jwt", frameworkauth.NewJwtGuard)
}
