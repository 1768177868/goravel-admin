package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"

	"goravel/bootstrap"
)

// @title           Goravel Admin API
// @version         1.0
// @description     这是一个基于 Goravel 框架的后台管理系统 API 文档
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:3000
// @BasePath  /api/admin

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                JWT 认证，格式：Bearer {token}

// runApplication 启动应用程序的核心逻辑
func runApplication() {
	// This bootstraps the framework and gets it ready for use.
	bootstrap.Boot()

	// Create a channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start http server by facades.Route().
	go func() {
		if err := facades.Route().Run(); err != nil {
			facades.Log().Errorf("Route run error: %v", err)
		}
	}()

	// Start queue server by facades.Queue().
	// 从配置文件读取重试次数和并发数（支持环境变量）
	// 重试次数是上限，实际重试次数由每个 Job 的 ShouldRetry 方法决定
	tries := facades.Config().GetInt("queue.tries", 5)
	concurrent := facades.Config().GetInt("queue.concurrent", 1)

	worker := facades.Queue().Worker(queue.Args{
		Connection: "",         // 使用默认连接
		Queue:      "",         // 使用默认队列
		Concurrent: concurrent, // 并发数（从配置读取，支持环境变量 QUEUE_CONCURRENT）
		Tries:      tries,      // 最大重试次数上限（从配置读取，支持环境变量 QUEUE_TRIES）
	})
	facades.Log().Infof("队列工作进程启动 - 并发数: %d, 最大重试次数: %d", concurrent, tries)
	go func() {
		if err := worker.Run(); err != nil {
			facades.Log().Errorf("Queue run error: %v", err)
		}
	}()

	// Listen for the OS signal
	go func() {
		<-quit
		shutdownApplication()
		os.Exit(0)
	}()

	select {}
}

// shutdownApplication 优雅关闭应用程序
func shutdownApplication() {
	if err := facades.Route().Shutdown(); err != nil {
		facades.Log().Errorf("Route Shutdown error: %v", err)
	}
	worker := facades.Queue().Worker()
	if err := worker.Shutdown(); err != nil {
		facades.Log().Errorf("Queue Shutdown error: %v", err)
	}
}

func main() {
	runApplication()
}
