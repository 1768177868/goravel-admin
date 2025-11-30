package main

import (
	"os"
	"os/signal"
	"syscall"

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

func main() {
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

	// Start grpc server by facades.Grpc().
	// go func() {
	// 	if err := facades.Grpc().Run(); err != nil {
	// 		facades.Log().Errorf("Grpc run error: %v", err)
	// 	}
	// }()

	// Start queue server by facades.Queue().
	worker := facades.Queue().Worker()
	go func() {
		if err := worker.Run(); err != nil {
			facades.Log().Errorf("Queue run error: %v", err)
		}
	}()

	// Listen for the OS signal
	go func() {
		<-quit
		if err := facades.Route().Shutdown(); err != nil {
			facades.Log().Errorf("Route Shutdown error: %v", err)
		}
		// if err := facades.Grpc().Shutdown(); err != nil {
		// 	facades.Log().Errorf("Grpc Shutdown error: %v", err)
		// }
		if err := worker.Shutdown(); err != nil {
			facades.Log().Errorf("Queue Shutdown error: %v", err)
		}

		os.Exit(0)
	}()

	select {}
}
