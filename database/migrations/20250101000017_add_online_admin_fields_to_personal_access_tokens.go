package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000017AddOnlineAdminFieldsToPersonalAccessTokens struct {
}

func (r *M20250101000017AddOnlineAdminFieldsToPersonalAccessTokens) Signature() string {
	return "20250101000017_add_online_admin_fields_to_personal_access_tokens"
}

func (r *M20250101000017AddOnlineAdminFieldsToPersonalAccessTokens) Up() error {
	if !facades.Schema().HasTable("personal_access_tokens") {
		return nil
	}

	// 构建需要添加的列（用于在线管理员监控）
	columnsToAdd := []struct {
		name    string
		length  int
		comment string
	}{
		{"browser", 100, "浏览器"},
		{"ip", 45, "IP地址"},
		{"os", 100, "操作系统"},
		{"session_id", 64, "会话编号"},
	}

	hasNewColumns := false
	for _, col := range columnsToAdd {
		if !facades.Schema().HasColumn("personal_access_tokens", col.name) {
			hasNewColumns = true
			break
		}
	}

	// 如果有需要添加的列，则添加
	if hasNewColumns {
		return facades.Schema().Table("personal_access_tokens", func(table schema.Blueprint) {
			for _, col := range columnsToAdd {
				if !facades.Schema().HasColumn("personal_access_tokens", col.name) {
					table.String(col.name, col.length).Nullable().Comment(col.comment)
				}
			}
		})
	}

	return nil
}

func (r *M20250101000017AddOnlineAdminFieldsToPersonalAccessTokens) Down() error {
	if facades.Schema().HasTable("personal_access_tokens") {
		return facades.Schema().Table("personal_access_tokens", func(table schema.Blueprint) {
			table.DropColumn("browser", "ip", "os", "session_id")
		})
	}
	return nil
}
