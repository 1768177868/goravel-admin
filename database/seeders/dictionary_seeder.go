package seeders

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type DictionarySeeder struct {
}

func (s *DictionarySeeder) Signature() string {
	return "DictionarySeeder"
}

func (s *DictionarySeeder) Run() error {
	// 创建字典数据
	dictionaries := []models.Dictionary{
		{Type: "status", Label: "启用", Value: "1", Description: "启用状态", Status: 1, Sort: 1},
		{Type: "status", Label: "禁用", Value: "0", Description: "禁用状态", Status: 1, Sort: 2},
		{Type: "menu_type", Label: "目录", Value: "1", Description: "目录类型", Status: 1, Sort: 1},
		{Type: "menu_type", Label: "菜单", Value: "2", Description: "菜单类型", Status: 1, Sort: 2},
		{Type: "menu_type", Label: "按钮", Value: "3", Description: "按钮类型", Status: 1, Sort: 3},
	}

	for _, dict := range dictionaries {
		facades.Orm().Query().FirstOrCreate(&dict, models.Dictionary{
			Type:  dict.Type,
			Value: dict.Value,
		})
	}

	return nil
}

