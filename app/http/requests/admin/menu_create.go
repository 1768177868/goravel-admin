package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type MenuCreate struct {
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
}

func (r *MenuCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *MenuCreate) Rules(ctx http.Context) map[string]string {
	rules := map[string]string{
		"title":      "required|max_len:50",
		"slug":       "required|max_len:50",
		"icon":       "max_len:50",
		"component":  "max_len:255",
		"permission": "max_len:100",
		"type":       "in:1,2,3",
		"status":     "in:0,1",
		"is_hidden":  "in:0,1",
		"link_type":  "in:1,2",
		"open_type":  "in:1,2",
	}

	// 根据 link_type 动态设置 path 的验证规则
	linkType := ctx.Request().Input("link_type")
	if linkType == "2" {
		// 外部链接：需要验证为完整的 URL
		rules["path"] = "required|max_len:1000|full_url"
	} else {
		// 内部页面：只需要必填和长度验证
		rules["path"] = "required|max_len:1000"
	}

	return rules
}

func (r *MenuCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"title.required":     trans.Get(ctx, "menu_title_required"),
		"title.max_len":      trans.Get(ctx, "validation_title_max"),
		"slug.required":      trans.Get(ctx, "menu_slug_required"),
		"slug.max_len":       trans.Get(ctx, "validation_slug_max"),
		"icon.max_len":       trans.Get(ctx, "validation_icon_max"),
		"path.required":      trans.Get(ctx, "menu_path_required"),
		"path.max_len":       trans.Get(ctx, "validation_path_max"),
		"path.full_url":      trans.Get(ctx, "validation_path_url_invalid"),
		"component.max_len":  trans.Get(ctx, "validation_component_max"),
		"permission.max_len": trans.Get(ctx, "validation_permission_max"),
		"type.in":            trans.Get(ctx, "validation_menu_type_in"),
		"status.in":          trans.Get(ctx, "validation_status_in"),
		"is_hidden.in":       trans.Get(ctx, "validation_status_in"),
	}
}

func (r *MenuCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"title":      trans.Get(ctx, "validation_title"),
		"slug":       trans.Get(ctx, "validation_slug"),
		"icon":       trans.Get(ctx, "validation_icon"),
		"path":       trans.Get(ctx, "validation_path"),
		"component":  trans.Get(ctx, "validation_component"),
		"permission": trans.Get(ctx, "validation_permission"),
		"type":       trans.Get(ctx, "validation_type"),
		"status":     trans.Get(ctx, "validation_status"),
		"is_hidden":  trans.Get(ctx, "validation_is_hidden"),
	}
}

func (r *MenuCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
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
	return nil
}
