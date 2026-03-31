package option_providers

import (
	"github.com/goravel/framework/contracts/http"
)

type FormDemoOptionProvider struct{}

func NewFormDemoOptionProvider() *FormDemoOptionProvider {
	return &FormDemoOptionProvider{}
}

func (p *FormDemoOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	scene := ctx.Request().Query("scene", "default")

	switch scene {
	case "tree":
		return map[string]any{
			"options": []map[string]any{
				{
					"id":   1,
					"name": "研发中心",
					"children": []map[string]any{
						{"id": 11, "name": "后端组"},
						{"id": 12, "name": "前端组"},
					},
				},
				{
					"id":   2,
					"name": "运营中心",
					"children": []map[string]any{
						{"id": 21, "name": "活动组"},
						{"id": 22, "name": "内容组"},
					},
				},
			},
		}, nil
	case "cascader":
		return map[string]any{
			"options": []map[string]any{
				{
					"id":   100,
					"name": "华南",
					"children": []map[string]any{
						{
							"id":   110,
							"name": "广东",
							"children": []map[string]any{
								{"id": 111, "name": "广州"},
								{"id": 112, "name": "深圳"},
							},
						},
					},
				},
				{
					"id":   200,
					"name": "华东",
					"children": []map[string]any{
						{
							"id":   210,
							"name": "江苏",
							"children": []map[string]any{
								{"id": 211, "name": "南京"},
								{"id": 212, "name": "苏州"},
							},
						},
					},
				},
			},
		}, nil
	case "icon":
		return map[string]any{
			"options": []map[string]any{
				{"label": "User", "value": "User"},
				{"label": "Setting", "value": "Setting"},
				{"label": "Bell", "value": "Bell"},
				{"label": "EditPen", "value": "EditPen"},
				{"label": "MagicStick", "value": "MagicStick"},
			},
		}, nil
	default:
		return map[string]any{
			"options": []map[string]any{
				{"label": "测试选项A", "value": "A"},
				{"label": "测试选项B", "value": "B"},
				{"label": "测试选项C", "value": "C"},
			},
		}, nil
	}
}
