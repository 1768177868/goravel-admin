package option_providers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"
	"github.com/samber/lo"
	"github.com/spf13/cast"

	"goravel/app/models"
)

type AdminOptionProvider struct{}

func NewAdminOptionProvider() *AdminOptionProvider {
	return &AdminOptionProvider{}
}

func (p *AdminOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	var admins []models.Admin

	// 排除开发者ID
	developerIDsStr := facades.Config().GetString("admin.developer_ids", "2")
	developerIDs := parseProtectedIDs(developerIDsStr)

	query := facades.Orm().Query().Where("status", 1)
	if len(developerIDs) > 0 {
		query = query.Where("id NOT IN ?", developerIDs)
	}

	if err := query.Order("id asc").Get(&admins); err != nil {
		return nil, err
	}

	options := lo.Map(admins, func(admin models.Admin, _ int) map[string]any {
		label := admin.Username
		if admin.Nickname != "" {
			label = admin.Nickname + " (" + admin.Username + ")"
		}
		return map[string]any{
			"label": label,
			"value": cast.ToString(admin.ID),
		}
	})

	return map[string]any{
		"options": options,
	}, nil
}

// parseProtectedIDs 解析受保护的管理员ID字符串（支持逗号分隔）
func parseProtectedIDs(idsStr string) []uint {
	var ids []uint
	if idsStr == "" {
		return ids
	}

	// 使用字符串分割
	parts := str.Of(idsStr).Split(",")
	for _, part := range parts {
		part = str.Of(part).Trim().String()
		if !str.Of(part).IsEmpty() {
			if id := cast.ToUint(part); id > 0 {
				ids = append(ids, id)
			}
		}
	}

	return ids
}
