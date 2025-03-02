//go:build !sonic

package main

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/config"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/dal"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/handlers"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/middleware"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/registry"
	"go.uber.org/zap"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			hlog.Errorf("服务崩溃: %v", r)
		}
	}()

	hlog.Info("=== 初始化开始 ===")
	if err := config.Init(); err != nil {
		panic(err)
	}

	zap.L().Debug("配置加载结果",
		zap.Any("redis", config.Conf.Redis),
		zap.Any("mysql", config.Conf.MySQL))

	middleware.InitAuthMiddleware("config/auth.yaml") // 增加初始化调用

	// 初始化数据库连接
	dal.InitDB()

	// 初始化Hertz（必须显式指定端口,端口8080）
	h := server.Default( //创建sever default实例
		server.WithHostPorts(":8080"),
		server.WithExitWaitTime(30*time.Second),
	)

	// 注册健康检查端点（必须最先执行） 注册路由
	h.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(200, map[string]string{"status": "ok"})
	})
	// registry.AddHealthCheck(h, "user-service")

	// 路由配置（重点区域）------------------------
	// 开放接口（无需认证）
	h.POST("/register", handlers.Register)
	h.POST("/login", handlers.Login)

	// 受保护接口（需要认证）
	h.GET("/userinfo",
		middleware.JWTAuth(), // 认证中间件
		handlers.GetUserInfo, // 业务处理函数
	)
	// -----------------------------------------

	// 服务注册（需在路由注册后执行）
	hlog.Info("正在注册Consul服务") //调用RegisterService注册服务
	if _, err := registry.RegisterService("user-service", 8080); err != nil {
		hlog.Fatal("Consul注册失败: ", err)
	}

	// 添加生命周期钩子
	h.OnRun = append(h.OnRun, func(ctx context.Context) error {
		hlog.Info("=== HTTP服务已启动 ===")
		return nil
	})

	// 添加优雅关闭处理
	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		registry.DeregisterService("user-service")
		hlog.Info("服务已优雅关闭")
	})

	hlog.Info("=== 启动服务监听 ===")
	h.Spin()
}
