package services

import (
	"github.com/goravel/framework/contracts/http"
)

// OptionProvider 选项提供者接口
// 所有选项提供者都需要实现此接口
type OptionProvider interface {
	// GetOptions 获取选项列表
	// 返回的 map 应该包含 "options" 键，值为选项数组
	// 可以包含其他额外的数据，如 "list" 等
	GetOptions(ctx http.Context) (map[string]any, error)
}

