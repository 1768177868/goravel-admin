package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type DictionaryCreate struct {
	Type           string `form:"type" json:"type" example:"order_status"`                            // 字典类型（必填）
	Label          string `form:"label" json:"label" example:"已支付"`                                   // 字典标签（必填）
	Value          string `form:"value" json:"value" example:"paid"`                                  // 字典值（必填）
	TranslationKey string `form:"translation_key" json:"translation_key" example:"order.status.paid"` // 多语言翻译Key（可选）
	Description    string `form:"description" json:"description" example:"订单支付成功状态"`                  // 字典描述（可选）
	Status         uint8  `form:"status" json:"status" enums:"0,1" example:"1"`                       // 状态（1-启用，0-禁用）
	Sort           int    `form:"sort" json:"sort" example:"10"`                                      // 排序值（越小越靠前）
	Remark         string `form:"remark" json:"remark" example:"系统默认值"`                               // 备注（可选）
}

func (r *DictionaryCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *DictionaryCreate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"type":            "required|max:50",
		"label":           "required|max:50",
		"value":           "required|max:100",
		"translation_key": "max:255",
		"description":     "max:255",
		"status":          "in:0,1",
		"remark":          "max:500",
	}
}

func (r *DictionaryCreate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"type":            trans.Get(ctx, "validation.attributes.type"),
		"label":           trans.Get(ctx, "validation.attributes.label"),
		"value":           trans.Get(ctx, "validation.attributes.value"),
		"translation_key": trans.Get(ctx, "validation.attributes.translation_key"),
		"description":     trans.Get(ctx, "validation.attributes.description"),
		"status":          trans.Get(ctx, "validation.attributes.status"),
		"remark":          trans.Get(ctx, "validation.attributes.remark"),
	}
}

func (r *DictionaryCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
