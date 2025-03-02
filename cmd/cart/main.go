//go:build !sonic

package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/config"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/dal"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/handlers"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/middleware"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/redis"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/registry"
	"go.uber.org/zap"
)

func main() {

	hlog.Info("=== 初始化开始 ===")
	if err := config.Init(); err != nil {
		panic(err)
	}

	zap.L().Debug("配置加载结果",
		zap.Any("redis", config.Conf.Redis),
		zap.Any("mysql", config.Conf.MySQL))

	middleware.InitAuthMiddleware("config/auth.yaml") // 增加初始化调用

	// 初始化Redis（带错误处理）
	if err := redis.InitRedis(); err != nil {
		panic("Redis初始化失败: " + err.Error())
	}
	dal.InitDB() // 使用独立数据库配置

	h := server.Default(
		server.WithHostPorts(":8082"), // 不同端口
	)

	// 注册服务
	if _, err := registry.RegisterService("cart-service", 8082); err != nil {
		panic(err)
	}

	// 购物车服务路由
	h.POST("/cart/add", middleware.JWTAuth(), handlers.AddToCart)
	h.DELETE("/cart/delete", middleware.JWTAuth(), handlers.ClearCart)
	//测试
	h.POST("/cart/redis-test", middleware.JWTAuth(), handlers.TestRedis)

	// 健康检查
	h.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(200, map[string]string{"status": "ok"})
	})

	h.Spin()
}
