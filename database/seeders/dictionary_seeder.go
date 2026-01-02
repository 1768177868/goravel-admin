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
		// 检查是否已存在，只创建不存在的（增量添加模式）
		var existingDict models.Dictionary
		if err := facades.Orm().Query().Where("type", dict.Type).Where("value", dict.Value).First(&existingDict); err != nil {
			// 不存在则创建
			if err := facades.Orm().Query().Create(&dict); err != nil {
				facades.Log().Errorf("Failed to create dictionary type=%s value=%s: %v", dict.Type, dict.Value, err)
			} else {
				facades.Log().Infof("Created dictionary: type=%s value=%s", dict.Type, dict.Value)
			}
		} else {
			// 已存在则跳过，不修改
			facades.Log().Infof("Dictionary type=%s value=%s already exists, skipping", dict.Type, dict.Value)
		}
	}

	return nil
}

