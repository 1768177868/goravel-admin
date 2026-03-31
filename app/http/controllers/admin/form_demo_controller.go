package admin

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/response"
)

type FormDemoController struct{}

func NewFormDemoController() *FormDemoController {
	return &FormDemoController{}
}

// GetData 返回 FormDemo 测试回显数据
func (c *FormDemoController) GetData(ctx http.Context) http.Response {
	data := map[string]any{
		"username":                    "api_demo_admin",
		"password":                    "123456",
		"intro":                       "Loaded from API test endpoint.",
		"role":                        "editor",
		"role_remote":                 "B",
		"department_id":               12,
		"department_remote_id":        12,
		"gender":                      "female",
		"gender_remote":               "B",
		"interests":                   []string{"reading"},
		"interests_remote":            []string{"B"},
		"hobbies":                     []string{"sports", "music"},
		"hobbies_remote":              []string{"A", "C"},
		"birthday":                    "2026-03-15",
		"meeting_at":                  "2026-03-15 14:30:00",
		"active_days":                 []string{"2026-03-01", "2026-03-20"},
		"active_period":               []string{"2026-03-01 09:00:00", "2026-03-20 19:00:00"},
		"transfer_permissions":        []string{"user.edit", "menu.manage"},
		"transfer_permissions_remote": []string{"A", "B"},
		"score":                       77,
		"score_input_number":          188,
		"score_input_number_suffix":   256,
		"satisfaction":                5,
		"volume":                      80,
		"enabled":                     true,
		"icon_name":                   "User",
		"icon_name_remote":            "Bell",
		"avatar":                      "/api/admin/public/images/70",
		"avatar_list": []string{
			"/api/admin/public/images/70",
			"/api/admin/public/images/71",
		},
		"color":         "#E6A23C",
		"region":        []string{"china", "guangdong", "guangzhou"},
		"region_remote": []int{100, 110, 112},
		"custom_note":   "This value is loaded from test API.",
	}

	return response.Success(ctx, data)
}
