//go:build !sonic

package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/config"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/dal"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/handlers"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/middleware"
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


	dal.InitDB() // 使用独立数据库配置

	h := server.Default(
		server.WithHostPorts(":8081"),                 // 不同端口
		server.WithTransport(standard.NewTransporter), // 使用标准网络库
	)

	// 注册服务
	if _, err := registry.RegisterService("product-service", 8081); err != nil {
		panic(err)
	}

	// 商品服务路由
	h.GET("/products/:id", handlers.GetProduct)
	h.POST("/products", handlers.CreateProduct)

	// 健康检查
	h.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(200, map[string]string{"status": "ok"})
	})

	h.Spin()
}
