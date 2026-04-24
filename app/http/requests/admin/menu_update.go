package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type MenuUpdate struct {
	ParentID   uint   `form:"parent_id" json:"parent_id"`
	Title      string `form:"title" json:"title"`
	Slug       string `form:"slug" json:"slug"`
	Icon       string `form:"icon" json:"icon"`
	Path       string `form:"path" json:"path"`
	Component  string `form:"component" json:"component"`
	Permission string `form:"permission" json:"permission"`
	Type       uint8  `form:"type" json:"type"`
	Status     uint8  `form:"status" json:"status"`
	Sort       int    `form:"sort" json:"sort"`
	IsHidden   uint8  `form:"is_hidden" json:"is_hidden"`
	LinkType   uint8  `form:"link_type" json:"link_type"`
	OpenType   uint8  `form:"open_type" json:"open_type"`
	NoCache    uint8  `form:"no_cache" json:"no_cache"`
}

func (r *MenuUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *MenuUpdate) Rules(ctx http.Context) map[string]string {
	rules := map[string]string{
		"title":      "max_len:50",
		"slug":       "max_len:50",
		"icon":       "max_len:50",
		"component":  "max_len:255",
		"permission": "max_len:100",
		"type":       "in:1,2,3",
		"status":     "in:0,1",
		"is_hidden":  "in:0,1",
		"link_type":  "in:1,2",
		"open_type":  "in:1,2",
		"no_cache":   "in:0,1",
	}

	// 根据 link_type 动态设置 path 的验证规则
	linkType := ctx.Request().Input("link_type")
	if linkType == "2" {
		// 外部链接：需要验证为完整的 URL
		rules["path"] = "max_len:1000|full_url"
	} else {
		// 内部页面：只需要长度验证
		rules["path"] = "max_len:1000"
		// 内部页面不验证 open_type
		delete(rules, "open_type")
	}

	return rules
}

func (r *MenuUpdate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"title.max_len":      trans.GetReplace(ctx, "validation.max.title", map[string]string{"max": "50"}),
		"slug.max_len":       trans.GetReplace(ctx, "validation.max.slug", map[string]string{"max": "50"}),
		"icon.max_len":       trans.GetReplace(ctx, "validation.max.icon", map[string]string{"max": "50"}),
		"path.max_len":       trans.GetReplace(ctx, "validation.max.path", map[string]string{"max": "0"}),
		"path.full_url":      trans.Get(ctx, "validation.path.url_invalid"),
		"component.max_len":  trans.GetReplace(ctx, "validation.max.component", map[string]string{"max": "255"}),
		"permission.max_len": trans.GetReplace(ctx, "validation.max.permission", map[string]string{"max": "100"}),
		"type.in":            trans.GetReplace(ctx, "validation.in.menu_type", map[string]string{"values": "1,2,3"}),
		"status.in":          trans.GetReplace(ctx, "validation.in.status", map[string]string{"values": "0,1"}),
		"is_hidden.in":       trans.GetReplace(ctx, "validation.in.status", map[string]string{"values": "0,1"}),
	}
}

func (r *MenuUpdate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"title":      trans.Get(ctx, "validation.attributes.title"),
		"slug":       trans.Get(ctx, "validation.attributes.slug"),
		"icon":       trans.Get(ctx, "validation.attributes.icon"),
		"path":       trans.Get(ctx, "validation.attributes.path"),
		"component":  trans.Get(ctx, "validation.attributes.component"),
		"permission": trans.Get(ctx, "validation.attributes.permission"),
		"type":       trans.Get(ctx, "validation.attributes.type"),
		"status":     trans.Get(ctx, "validation.attributes.status"),
		"is_hidden":  trans.Get(ctx, "validation.attributes.is_hidden"),
	}
}

func (r *MenuUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	// 将数字字段转换为字符串，以便 in 规则能正确验证
	if err := helpers.PrepareNumericFieldForValidation(data, "type"); err != nil {
		return err
	}
	if err := helpers.PrepareNumericFieldForValidation(data, "status"); err != nil {
		return err
	}
	if err := helpers.PrepareNumericFieldForValidation(data, "is_hidden"); err != nil {
		return err
	}
	if err := helpers.PrepareNumericFieldForValidation(data, "link_type"); err != nil {
		return err
	}
	if err := helpers.PrepareNumericFieldForValidation(data, "open_type"); err != nil {
		return err
	}
	if err := helpers.PrepareNumericFieldForValidation(data, "no_cache"); err != nil {
		return err
	}
	return nil
}
