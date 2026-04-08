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
		{Type: "status", Label: "启用", Value: "1", Description: "启用状态", Status: 1, Sort: 1, TranslationKey: "enabled"},
		{Type: "status", Label: "禁用", Value: "0", Description: "禁用状态", Status: 1, Sort: 2, TranslationKey: "disabled"},
		{Type: "menu_type", Label: "目录", Value: "1", Description: "目录类型", Status: 1, Sort: 1, TranslationKey: "directory"},
		{Type: "menu_type", Label: "菜单", Value: "2", Description: "菜单类型", Status: 1, Sort: 2, TranslationKey: "menu"},
		// {Type: "menu_type", Label: "按钮", Value: "3", Description: "按钮类型", Status: 1, Sort: 3, TranslationKey: "button"},
	}

	for _, dict := range dictionaries {
		// 使用 type + label 作为幂等键：存在则更新，不存在才创建
		query := facades.Orm().Query().Model(&models.Dictionary{}).Where("type", dict.Type).Where("label", dict.Label)
		count, err := query.Count()
		if err != nil {
			return err
		}

		if count == 0 {
			if err := facades.Orm().Query().Create(&dict); err != nil {
				return err
			}
			continue
		}

		var existing models.Dictionary
		if err := facades.Orm().Query().Model(&models.Dictionary{}).Where("type", dict.Type).Where("label", dict.Label).First(&existing); err != nil {
			return err
		}

		existing.Value = dict.Value
		existing.Description = dict.Description
		existing.Status = dict.Status
		existing.Sort = dict.Sort
		existing.TranslationKey = dict.TranslationKey
		if err := facades.Orm().Query().Save(&existing); err != nil {
			return err
		}
	}

	return nil
}
