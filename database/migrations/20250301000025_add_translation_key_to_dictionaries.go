package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250301000025AddTranslationKeyToDictionaries struct {
}

func (r *M20250301000025AddTranslationKeyToDictionaries) Signature() string {
	return "20250301000025_add_translation_key_to_dictionaries"
}

func (r *M20250301000025AddTranslationKeyToDictionaries) Up() error {
	if !facades.Schema().HasTable("dictionaries") {
		return nil
	}

	// 检查列是否已存在
	columns, err := facades.Schema().GetColumns("dictionaries")
	if err != nil {
		return err
	}

	for _, column := range columns {
		if column.Name == "translation_key" {
			return nil
		}
	}

	return facades.Schema().Table("dictionaries", func(table schema.Blueprint) {
		table.String("translation_key").Nullable().Comment("多语言Key")
	})
}

func (r *M20250301000025AddTranslationKeyToDictionaries) Down() error {
	if facades.Schema().HasTable("dictionaries") {
		return facades.Schema().Table("dictionaries", func(table schema.Blueprint) {
			table.DropColumn("translation_key")
		})
	}
	return nil
}
