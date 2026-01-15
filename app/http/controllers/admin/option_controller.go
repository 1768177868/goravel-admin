package admin

import (
	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/services"
	"goravel/app/services/option_providers"
)

type OptionController struct {
	providers         map[string]services.OptionProvider
	dictionaryService services.DictionaryService
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
	providers["payment_method"] = option_providers.NewPaymentMethodOptionProvider()
	// 在此处添加新的选项提供者，例如：
	// providers["new_type"] = option_providers.NewNewTypeOptionProvider()

	return &OptionController{
		providers:         providers,
		dictionaryService: services.NewDictionaryService(),
	}
}

// Index 获取选项列表
// 通过 type 参数指定选项类型，例如: /options?type=role
func (r *OptionController) Index(ctx http.Context) http.Response {
	optionType := ctx.Request().Query("type", "")

	if optionType == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrOptionTypeRequired.Code)
	}

	// 如果 type 是 "dictionary"，则需要进一步检查 dictionary_type 参数
	if optionType == "dictionary" {
		dictType := ctx.Request().Query("dictionary_type", "")
		if dictType == "" {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrOptionTypeRequired.Code)
		}
		
		dictionaries, err := r.dictionaryService.GetByType(dictType)
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
		}

		// 转换为选项格式，并处理多语言
		var options []map[string]any
		for _, dict := range dictionaries {
			label := dict.Label
			// 如果有 translation_key，优先使用翻译键
			if dict.TranslationKey != "" {
				translated := trans.Get(ctx, dict.TranslationKey)
				// 只有当翻译结果不同于 key 本身（表示找到了翻译）且不为空时才使用
				if translated != "" && translated != dict.TranslationKey {
					label = translated
				}
			}

			options = append(options, map[string]any{
				"label": label,
				"value": dict.Value,
			})
		}
		return response.Success(ctx, options)
	}

	provider, exists := r.providers[optionType]
	if !exists {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidOptionType.Code)
	}

	data, err := provider.GetOptions(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}

	return response.Success(ctx, data)
}

// RegisterProvider 注册新的选项提供者（可选，用于动态注册）
// 如果需要在运行时动态添加提供者，可以使用此方法
func (r *OptionController) RegisterProvider(optionType string, provider services.OptionProvider) {
	r.providers[optionType] = provider
}
