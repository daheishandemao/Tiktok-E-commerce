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

	// "github.com/daheishandemao/Tiktok-E-commerce/pkg/logger" // 新增日志包
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/middleware"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/redis"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/registry"
	"go.uber.org/zap"
)

func main() {
	// 1. 初始化日志系统（必须新增）
	// logger.InitLogger()
	// defer logger.Sync()

	hlog.Info("=== 初始化开始 ===")
	if err := config.Init(); err != nil {
		panic(err)
	}

	// 2. 增强配置校验
	if config.Conf.Redis.Addr == "" || config.Conf.MySQL.DSN == "" {
		panic("配置文件不完整")
	}

	zap.L().Debug("配置加载结果",
		zap.Any("redis", config.Conf.Redis),
		zap.Any("mysql", config.Conf.MySQL))

	middleware.InitAuthMiddleware("config/auth.yaml") // 增加初始化调用
	// 3. 初始化顺序调整（先DB后Redis）
	dal.InitDB() // 必须放在Redis之前
	if err := redis.InitRedis(); err != nil {
		panic("Redis初始化失败: " + err.Error())
	}

	// 4. 创建订单处理器实例（关键修改）
	orderHandler := handlers.NewOrderHandler(
		dal.DB,       // 传递GORM实例
		redis.Client, // 传递Redis客户端
	)

	h := server.Default(
		server.WithHostPorts(":8083"),
		server.WithExitWaitTime(5*time.Second), // 优雅停机
	)

	// 5. 服务注册增强
	if _, err := registry.RegisterService("order-service", 8083); err != nil {
		panic("服务注册失败: " + err.Error())
	}

	// 6. 路由定义修正（使用handler方法）
	h.POST("/orders",
		middleware.JWTAuth(),     // 修正中间件名称
		orderHandler.CreateOrder) // 直接传递方法

	// 7. 移除测试路由（生产环境不需要）
	// h.POST("/order/order-test", ...)

	// 8. 增强健康检查
	h.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		// 添加数据库健康检查
		if err := dal.DB.Exec("SELECT 1").Error; err != nil {
			ctx.JSON(503, map[string]string{"status": "unhealthy"})
			return
		}
		ctx.JSON(200, map[string]string{"status": "ok"})
	})

	// 9. 添加优雅终止处理
	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("服务正在关闭...")
		redis.Client.Close()
	})

	h.Spin()
}
