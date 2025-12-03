package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000024AddGoogleSecretToAdmins struct {
}

func (r *M20250101000024AddGoogleSecretToAdmins) Signature() string {
	return "20250101000024_add_google_secret_to_admins"
}

func (r *M20250101000024AddGoogleSecretToAdmins) Up() error {
	if !facades.Schema().HasTable("admins") {
		return nil
	}

	// 检查列是否已存在
	columns, err := facades.Schema().GetColumns("admins")
	if err != nil {
		return err
	}

	hasGoogleSecret := false
	for _, column := range columns {
		if column.Name == "google_secret" {
			hasGoogleSecret = true
			break
		}
	}

	// 如果列不存在，则添加
	if !hasGoogleSecret {
		return facades.Schema().Table("admins", func(table schema.Blueprint) {
			table.String("google_secret", 255).Nullable().Comment("谷歌验证码密钥")
		})
	}

	return nil
}

func (r *M20250101000024AddGoogleSecretToAdmins) Down() error {
	if facades.Schema().HasTable("admins") {
		return facades.Schema().Table("admins", func(table schema.Blueprint) {
			table.DropColumn("google_secret")
		})
	}
	return nil
}

