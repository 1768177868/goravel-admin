package routes

import (
	"net/http/pprof"
	"runtime"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// Pprof 注册 pprof 性能分析路由
// 仅在开发/调试模式下启用
func Pprof() {
	// 只在开发模式或调试模式下启用 pprof
	if !facades.Config().GetBool("app.debug", false) {
		return
	}

	// 获取底层 HTTP 服务器（Gin 或 Fiber）
	// 由于 Goravel 框架封装，我们需要通过反射或直接访问底层路由
	// 这里我们创建一个包装器来处理 pprof 路由

	// CPU 性能分析主页
	facades.Route().Get("/debug/pprof/", func(ctx http.Context) http.Response {
		pprof.Index(ctx.Response().Writer(), ctx.Request().Origin())
		return nil // pprof 已直接写入响应，返回 nil
	})

	// CPU 性能分析（30秒采样）
	facades.Route().Get("/debug/pprof/profile", func(ctx http.Context) http.Response {
		pprof.Profile(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})

	// CPU 性能分析（自定义采样时间，通过 seconds 参数）
	facades.Route().Get("/debug/pprof/profile/{seconds}", func(ctx http.Context) http.Response {
		pprof.Profile(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})

	// 堆内存分析
	facades.Route().Get("/debug/pprof/heap", func(ctx http.Context) http.Response {
		pprof.Handler("heap").ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})

	// 协程分析
	facades.Route().Get("/debug/pprof/goroutine", func(ctx http.Context) http.Response {
		pprof.Handler("goroutine").ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})

	// 阻塞分析
	facades.Route().Get("/debug/pprof/block", func(ctx http.Context) http.Response {
		pprof.Handler("block").ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})

	// 互斥锁分析
	facades.Route().Get("/debug/pprof/mutex", func(ctx http.Context) http.Response {
		pprof.Handler("mutex").ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})

	// 内存分配分析
	facades.Route().Get("/debug/pprof/allocs", func(ctx http.Context) http.Response {
		pprof.Handler("allocs").ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})

	// 命令行工具
	facades.Route().Get("/debug/pprof/cmdline", func(ctx http.Context) http.Response {
		pprof.Cmdline(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})

	// 符号表
	facades.Route().Get("/debug/pprof/symbol", func(ctx http.Context) http.Response {
		pprof.Symbol(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})

	// Trace 追踪
	facades.Route().Get("/debug/pprof/trace", func(ctx http.Context) http.Response {
		pprof.Trace(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})

	// 运行时统计信息
	facades.Route().Get("/debug/pprof/runtime", func(ctx http.Context) http.Response {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return ctx.Response().Json(http.StatusOK, http.Json{
			"goroutines": runtime.NumGoroutine(),
			"memory": map[string]any{
				"alloc":       m.Alloc,
				"total_alloc": m.TotalAlloc,
				"sys":         m.Sys,
				"lookups":     m.Lookups,
				"mallocs":     m.Mallocs,
				"frees":       m.Frees,
			},
			"gc": map[string]any{
				"num_gc":      m.NumGC,
				"pause_total": m.PauseTotalNs,
			},
		})
	})
}
