package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type DictionaryUpdate struct {
	Type           string `form:"type" json:"type" example:"order_status"`                            // 字典类型（可选）
	Label          string `form:"label" json:"label" example:"已支付"`                                   // 字典标签（可选）
	Value          string `form:"value" json:"value" example:"paid"`                                  // 字典值（可选）
	TranslationKey string `form:"translation_key" json:"translation_key" example:"order.status.paid"` // 多语言翻译Key（可选）
	Description    string `form:"description" json:"description" example:"订单支付成功状态"`                  // 字典描述（可选）
	Status         uint8  `form:"status" json:"status" enums:"0,1" example:"1"`                       // 状态（1-启用，0-禁用）
	Sort           int    `form:"sort" json:"sort" example:"10"`                                      // 排序值（越小越靠前）
	Remark         string `form:"remark" json:"remark" example:"系统默认值"`                               // 备注（可选）
}

func (r *DictionaryUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *DictionaryUpdate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"type":            "max_len:50",
		"label":           "max_len:50",
		"value":           "max_len:100",
		"translation_key": "max_len:255",
		"description":     "max_len:255",
		"status":          "in:0,1",
		"remark":          "max_len:500",
	}
}

func (r *DictionaryUpdate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"type.max_len":            trans.Get(ctx, "validation.max.type", map[string]string{"max": "50"}),
		"label.max_len":           trans.Get(ctx, "validation.max.label", map[string]string{"max": "50"}),
		"value.max_len":           trans.Get(ctx, "validation.max.value", map[string]string{"max": "100"}),
		"translation_key.max_len": trans.Get(ctx, "validation.max.translation_key", map[string]string{"max": "255"}),
		"description.max_len":     trans.Get(ctx, "validation.max.description", map[string]string{"max": "255"}),
		"status.in":               trans.Get(ctx, "validation.in.status", map[string]string{"values": "0,1"}),
		"remark.max_len":          trans.Get(ctx, "validation.max.remark", map[string]string{"max": "500"}),
	}
}

func (r *DictionaryUpdate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"type":            trans.Get(ctx, "validation.attributes.type"),
		"label":           trans.Get(ctx, "validation.attributes.label"),
		"value":           trans.Get(ctx, "validation.attributes.value"),
		"translation_key": trans.Get(ctx, "validation.attributes.translation_key"),
		"description":     trans.Get(ctx, "validation.attributes.description"),
		"status":          trans.Get(ctx, "validation.attributes.status"),
		"remark":          trans.Get(ctx, "validation.attributes.remark"),
	}
}

func (r *DictionaryUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
