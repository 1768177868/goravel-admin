package admin

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/response"
	"goravel/app/services"
	"goravel/app/services/option_providers"
)

type OptionController struct {
	providers map[string]services.OptionProvider
}

func NewOptionController() *OptionController {

	// 注册所有选项提供者
	// 添加新的选项类型时，只需：
	// 1. 在 app/services/option_providers/ 目录下创建新的提供者文件
	// 2. 实现 services.OptionProvider 接口
	// 3. 在此处注册新的提供者
	providers := make(map[string]services.OptionProvider)
	providers["role"] = option_providers.NewRoleOptionProvider()
	providers["department"] = option_providers.NewDepartmentOptionProvider()
	providers["menu"] = option_providers.NewMenuOptionProvider(services.NewTreeServiceImpl())
	providers["status"] = option_providers.NewStatusOptionProvider()
	providers["method"] = option_providers.NewMethodOptionProvider()
	providers["yes_no"] = option_providers.NewYesNoOptionProvider()
	providers["admin"] = option_providers.NewAdminOptionProvider()
	// 在此处添加新的选项提供者，例如：
	// providers["new_type"] = option_providers.NewNewTypeOptionProvider()

	return &OptionController{
		providers: providers,
	}
}

// Index 获取选项列表
// 通过 type 参数指定选项类型，例如: /options?type=role
func (r *OptionController) Index(ctx http.Context) http.Response {
	optionType := ctx.Request().Query("type", "")

	if optionType == "" {
		return response.Error(ctx, http.StatusBadRequest, "option_type_required")
	}

	provider, exists := r.providers[optionType]
	if !exists {
		return response.Error(ctx, http.StatusBadRequest, "invalid_option_type")
	}

	data, err := provider.GetOptions(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Success(ctx, data)
}

// RegisterProvider 注册新的选项提供者（可选，用于动态注册）
// 如果需要在运行时动态添加提供者，可以使用此方法
func (r *OptionController) RegisterProvider(optionType string, provider services.OptionProvider) {
	r.providers[optionType] = provider
}
