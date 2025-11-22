package trans

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// Get 获取翻译文本（支持多语言）
// 自动尝试 messages. 前缀的翻译键
func Get(ctx http.Context, key string) string {
	// 如果key已经包含 messages. 前缀，直接使用
	if len(key) > 8 && key[:8] == "messages." {
		message := facades.Lang(ctx).Get(key)
		// 如果返回的键和输入的键相同或为空，说明没找到
		if message != key && message != "" {
			return message
		}
		return key
	}

	// 尝试使用 messages. 前缀的key（这是语言文件中的实际格式）
	messageKey := "messages." + key
	message := facades.Lang(ctx).Get(messageKey)
	// 如果返回的键和输入的键不同且不为空，说明找到了
	if message != messageKey && message != "" {
		return message
	}

	// 如果带前缀的找不到，尝试直接获取（某些键可能不在messages下）
	message = facades.Lang(ctx).Get(key)
	if message != key && message != "" {
		return message
	}

	// 如果还是不存在，返回原始key
	return key
}

