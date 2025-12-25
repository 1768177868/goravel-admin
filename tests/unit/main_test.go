package unit

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// 单元测试不需要完整的应用初始化
	// 如果需要数据库等资源，请在 tests/feature 中编写功能测试

	// 执行测试
	exit := m.Run()

	os.Exit(exit)
}
