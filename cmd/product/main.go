package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/dal"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/handlers"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/registry"
)

func main() {
	dal.InitDB() // 使用独立数据库配置

	h := server.Default(
		server.WithHostPorts(":8081"), // 不同端口
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
