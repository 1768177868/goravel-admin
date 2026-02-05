package main

import (
	"os"
	"os/signal"
	"syscall"

	"goravel/app/facades"
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

func main() {
	// Bootstrap the application
	app := bootstrap.Boot()

	// Start the application (HTTP server, schedule, queue workers via runners)
	app.Start()

	// Create a channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Listen for the OS signal
	<-quit
	shutdownApplication(app)
}

// shutdownApplication 优雅关闭应用程序（框架会依次关闭所有 runners：Route、Queue 等）
func shutdownApplication(app interface{ Shutdown() error }) {
	if err := app.Shutdown(); err != nil {
		facades.Log().Errorf("Application Shutdown error: %v", err)
		os.Exit(1)
	}
}
