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
	if !facades.Schema().HasTable("personal_access_tokens") {
		return nil
	}

	// 检查列是否已存在
	columns, err := facades.Schema().GetColumns("personal_access_tokens")
	if err != nil {
		return err
	}

	// 检查哪些列需要添加
	columnMap := make(map[string]bool)
	for _, column := range columns {
		columnMap[column.Name] = true
	}

	// 构建需要添加的列
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
		if !columnMap[col.name] {
			hasNewColumns = true
			break
		}
	}

	// 如果有需要添加的列，则添加
	if hasNewColumns {
		return facades.Schema().Table("personal_access_tokens", func(table schema.Blueprint) {
			for _, col := range columnsToAdd {
				if !columnMap[col.name] {
					table.String(col.name, col.length).Nullable().Comment(col.comment)
				}
			}
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
