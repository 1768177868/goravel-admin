package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000017AddOnlineUserFieldsToPersonalAccessTokens struct {
}

func (r *M20250101000017AddOnlineUserFieldsToPersonalAccessTokens) Signature() string {
	return "20250101000017_add_online_user_fields_to_personal_access_tokens"
}

func (r *M20250101000017AddOnlineUserFieldsToPersonalAccessTokens) Up() error {
	if facades.Schema().HasTable("personal_access_tokens") {
		return facades.Schema().Table("personal_access_tokens", func(table schema.Blueprint) {
			// 直接添加字段（如果字段已存在，迁移会失败，需要手动处理）
			table.String("browser", 100).Nullable().Comment("浏览器")
			table.String("ip", 45).Nullable().Comment("IP地址")
			table.String("os", 100).Nullable().Comment("操作系统")
			table.String("session_id", 64).Nullable().Comment("会话编号")
		})
	}
	return nil
}

func (r *M20250101000017AddOnlineUserFieldsToPersonalAccessTokens) Down() error {
	if facades.Schema().HasTable("personal_access_tokens") {
		return facades.Schema().Table("personal_access_tokens", func(table schema.Blueprint) {
			table.DropColumn("browser", "ip", "os", "session_id")
		})
	}
	return nil
}
