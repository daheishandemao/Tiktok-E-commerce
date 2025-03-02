//go:build !sonic

package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/hashicorp/consul/api"

	// "github.com/cloudwego/kitex/server"
	kitexServer "github.com/cloudwego/kitex/server"
	"github.com/daheishandemao/Tiktok-E-commerce/kitex_gen/order"
	"github.com/daheishandemao/Tiktok-E-commerce/kitex_gen/order/orderservice"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/config"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/dal"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/handlers"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/middleware"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/redis"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/registry"
	consul "github.com/kitex-contrib/registry-consul"
	"go.uber.org/zap"
)

type OrderServiceImpl struct{}

func (s *OrderServiceImpl) UpdateStatus(ctx context.Context, req *order.UpdateReq) (bool, error) {
	zap.L().Info("收到RPC订单状态更新请求",
		zap.String("order_id", req.OrderID),
		zap.String("status", req.Status))

	result := dal.DB.Model(&dal.Order{}).
		Where("order_id = ?", req.OrderID).
		Update("status", req.Status)

	if result.Error != nil {
		zap.L().Error("数据库更新失败",
			zap.String("order_id", req.OrderID),
			zap.Error(result.Error))
	}

	return result.RowsAffected > 0, result.Error
}

func main() {
	// 初始化基础组件
	hlog.Info("=== 订单服务初始化 ===")
	if err := config.Init(); err != nil {
		panic("配置加载失败: " + err.Error())
	}

	// 初始化数据库和Redis
	dal.InitDB()
	if err := redis.InitRedis(); err != nil {
		panic("Redis初始化失败: " + err.Error())
	}

	// 初始化认证中间件
	middleware.InitAuthMiddleware("config/auth.yaml")

	// 创建Consul注册中心
	consulRegister, err := consul.NewConsulRegister(
		config.Conf.Consul.Address,
		consul.WithCheck(&api.AgentServiceCheck{
			HTTP: fmt.Sprintf("http://%s:%d/health",
				config.Conf.Service.IP,
				config.Conf.Service.OrderHTTPPort),
			Interval: "10s",
			Timeout:  "5s",
			Status:   api.HealthPassing,
		}),
		// consul.WithRegistryIP(config.Conf.Service.IP),/
		// consul.WithRegistryPort(config.Conf.Service.RpcPort),
	)
	if err != nil {
		panic("Consul注册失败: " + err.Error())
	}
	// if _, err := registry.RegisterService("order-service", 8083); err != nil {
	// 	panic("服务注册失败: " + err.Error())
	// }

	// 启动RPC服务端
	go func() {
		// Kitex RPC服务配置
		rpcAddr, _ := net.ResolveTCPAddr("tcp", ":8883")
		svr := orderservice.NewServer(
			new(OrderServiceImpl),
			kitexServer.WithRegistry(consulRegister),
			kitexServer.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
				ServiceName: "order.service",
				Tags: map[string]string{
					"protocol": "kitex",
					"env":      "dev",
				},
			}),
			kitexServer.WithServiceAddr(&net.TCPAddr{
				IP:   net.ParseIP(config.Conf.Service.IP),
				Port: config.Conf.Service.OrderRpcPort,
			}),
		)

		zap.L().Info("启动RPC服务", zap.String("addr", rpcAddr.String()))
		if err := svr.Run(); err != nil {
			panic("RPC服务启动失败: " + err.Error())
		}
	}()

	if _, err := registry.RegisterService("order-service", config.Conf.Service.OrderHTTPPort); err != nil {
		panic("服务注册失败: " + err.Error())
	}
	// 初始化HTTP服务器
	orderHandler := handlers.NewOrderHandler(dal.DB, redis.Client)
	h := server.Default(
		server.WithHostPorts(":8083"),
		server.WithExitWaitTime(5*time.Second),
	)

	// 注册HTTP路由
	h.POST("/orders", middleware.JWTAuth(), orderHandler.CreateOrder)
	h.PUT("/order/status", updateOrderStatusHTTP)

	// 健康检查
	h.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		if err := dal.DB.Exec("SELECT 1").Error; err != nil {
			ctx.JSON(503, map[string]string{"status": "unhealthy"})
			return
		}
		ctx.JSON(200, map[string]string{"status": "ok"})
	})

	// 优雅关闭
	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("HTTP服务关闭中...")
		redis.Client.Close()
	})

	zap.L().Info("启动HTTP服务", zap.String("addr", ":8083"))
	h.Spin()
}

// HTTP接口的订单状态更新
func updateOrderStatusHTTP(c context.Context, ctx *app.RequestContext) {
	var req struct {
		OrderID string `json:"order_id"`
		Status  string `json:"status"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(400, map[string]string{"error": "invalid params"})
		return
	}

	result := dal.DB.Model(&dal.Order{}).
		Where("order_id = ?", req.OrderID).
		Update("status", req.Status)

	if result.Error != nil {
		zap.L().Error("HTTP订单状态更新失败",
			zap.String("order_id", req.OrderID),
			zap.Error(result.Error))
		ctx.JSON(500, map[string]string{"error": "update failed"})
		return
	}

	ctx.JSON(200, map[string]interface{}{"msg": "success"})
}
