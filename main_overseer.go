//go:build overseer
// +build overseer

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/goravel/framework/facades"
	"github.com/jpillora/overseer"
	"github.com/jpillora/overseer/fetcher"

	"goravel/bootstrap"
)

// 使用 overseer 实现零停机重启的启动文件
// 编译时使用: go build -tags overseer -o main .
// 或: go run -tags overseer .

func main() {
	// 从环境变量获取配置（避免在 main 中初始化框架）
	// 如果使用 .env 文件，可以通过环境变量传递，或使用默认值
	host := os.Getenv("APP_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}
	address := fmt.Sprintf("%s:%s", host, port)

	overseer.Run(overseer.Config{
		Program: prog,
		Address: address,
		Fetcher: &fetcher.File{
			Path: os.Args[0], // 使用当前可执行文件
		},
		// 可选：健康检查
		HealthCheck: func() error {
			// 可以在这里实现健康检查逻辑
			// 例如：检查数据库连接、Redis 连接等
			return nil
		},
		// 可选：升级前回调
		PreUpgrade: func() error {
			// 注意：此时 facades 可能还未初始化，使用标准日志
			fmt.Println("准备升级，正在完成当前请求...")
			return nil
		},
		// 可选：升级后回调
		PostUpgrade: func() error {
			fmt.Println("升级完成，新进程已启动")
			return nil
		},
	})
}

// prog 是实际运行的程序
func prog(state overseer.State) {
	// 初始化框架
	bootstrap.Boot()

	// 创建信号通道（用于优雅关闭）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动队列服务
	worker := facades.Queue().Worker()
	go func() {
		if err := worker.Run(); err != nil {
			facades.Log().Errorf("Queue run error: %v", err)
		}
	}()

	// 启动 HTTP 服务器
	go func() {
		if err := facades.Route().Run(); err != nil {
			facades.Log().Errorf("Route run error: %v", err)
		}
	}()

	// 处理优雅关闭
	go func() {
		<-quit
		facades.Log().Info("收到关闭信号，正在优雅关闭...")
		if err := facades.Route().Shutdown(); err != nil {
			facades.Log().Errorf("Route Shutdown error: %v", err)
		}
		if err := worker.Shutdown(); err != nil {
			facades.Log().Errorf("Queue Shutdown error: %v", err)
		}
		os.Exit(0)
	}()

	// 等待
	select {}
}
