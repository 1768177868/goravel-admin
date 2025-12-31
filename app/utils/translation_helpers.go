package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goravel/framework/facades"
)

// loadLangMessages 加载语言文件的 messages 对象（内部辅助函数）
// 返回 messages map 和是否成功加载
func loadLangMessages(lang string) (map[string]any, bool) {
	// 获取语言文件路径（使用框架配置的路径）
	langPath := facades.Config().GetString("app.lang_path", "lang")
	langFile := filepath.Join(langPath, fmt.Sprintf("%s.json", lang))

	// 尝试读取语言文件
	langData, err := os.ReadFile(langFile)
	if err != nil {
		facades.Log().Debugf("读取语言文件失败: %s, error=%v", langFile, err)
		return nil, false
	}

	// 解析 JSON
	var langMap map[string]any
	if err := json.Unmarshal(langData, &langMap); err != nil {
		facades.Log().Debugf("解析语言文件失败: %s, error=%v", langFile, err)
		return nil, false
	}

	// 获取 messages 对象（使用类型辅助函数）
	messages, ok := GetMap(langMap, "messages")
	if !ok {
		facades.Log().Debugf("语言文件中没有 messages 对象: %s", langFile)
		return nil, false
	}

	return messages, true
}

// TranslateHeaders 翻译表头（直接读取语言文件）
// 这是一个通用函数，可以被任何需要翻译表头的 job 使用
//
// 参数:
//   - headerKeys: 需要翻译的键列表（如 ["export_header_id", "export_header_order_no"]）
//   - lang: 语言代码（如 "cn" 或 "en"）
//
// 返回:
//   - []string: 翻译后的表头列表，如果翻译失败则返回原始键
//
// 示例:
//
//	headerKeys := []string{"export_header_id", "export_header_order_no"}
//	headers := utils.TranslateHeaders(headerKeys, "cn")
func TranslateHeaders(headerKeys []string, lang string) []string {
	headers := make([]string, len(headerKeys))

	// 加载语言文件的 messages 对象
	messages, ok := loadLangMessages(lang)
	if !ok {
		// 如果读取失败，使用原始键
		facades.Log().Warningf("加载语言文件失败: lang=%s, 使用原始键", lang)
		return append([]string(nil), headerKeys...)
	}

	// 翻译每个键（使用泛型辅助函数）
	for i, key := range headerKeys {
		fullKey := "messages." + key
		if value, ok := GetString(messages, key); ok && value != "" {
			headers[i] = value
		} else {
			// 如果翻译失败，使用原始键
			headers[i] = key
			facades.Log().Debugf("翻译键未找到: %s (语言: %s)", fullKey, lang)
		}
	}

	return headers
}

// TranslateKey 翻译单个键（直接读取语言文件）
// 这是一个通用函数，可以被任何需要翻译单个键的地方使用
//
// 参数:
//   - key: 需要翻译的键（如 "export_order_status_pending"）
//   - lang: 语言代码（如 "cn" 或 "en"）
//   - defaultValue: 如果翻译失败时返回的默认值（通常为原始键）
//
// 返回:
//   - string: 翻译后的文本，如果翻译失败则返回 defaultValue
//
// 示例:
//
//	statusText := utils.TranslateKey("export_order_status_pending", "cn", "pending")
func TranslateKey(key, lang, defaultValue string) string {
	// 加载语言文件的 messages 对象
	messages, ok := loadLangMessages(lang)
	if !ok {
		return defaultValue
	}

	// 尝试获取翻译值
	if value, ok := GetString(messages, key); ok && value != "" {
		return value
	}

	// 如果翻译失败，返回默认值
	return defaultValue
}
