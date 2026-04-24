package trans

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/translation"
	"github.com/goravel/framework/facades"
)

// Get 获取翻译文本（支持多语言）
// 自动尝试 messages. 前缀的翻译键。
// 可选 replace 参数用于占位符替换，如 :max/:min。
func Get(ctx http.Context, key string, replace ...map[string]string) string {
	if len(replace) > 0 && len(replace[0]) > 0 {
		return resolve(ctx, key, translation.Option{Replace: replace[0]})
	}
	return resolve(ctx, key)
}

func resolve(ctx http.Context, key string, opts ...translation.Option) string {
	var opt translation.Option
	if len(opts) > 0 {
		opt = opts[0]
	}

	get := func(fullKey string) string {
		if len(opt.Replace) > 0 {
			return facades.Lang(ctx).Get(fullKey, opt)
		}
		return facades.Lang(ctx).Get(fullKey)
	}

	if len(key) > 8 && key[:8] == "messages." {
		message := get(key)
		if message != key && message != "" {
			return message
		}
		return key
	}

	messageKey := "messages." + key
	message := get(messageKey)
	if message != messageKey && message != "" {
		return message
	}

	message = get(key)
	if message != key && message != "" {
		return message
	}

	return key
}
