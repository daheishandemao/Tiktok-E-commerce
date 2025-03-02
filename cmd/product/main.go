//go:build !sonic

package main

import (
	"context"
	"fmt"
	"net"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	kitexServer "github.com/cloudwego/kitex/server"
	"github.com/daheishandemao/Tiktok-E-commerce/kitex_gen/product"
	"github.com/daheishandemao/Tiktok-E-commerce/kitex_gen/product/productservice"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/config"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/dal"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/handlers"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/middleware"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/registry"
	"github.com/hashicorp/consul/api"
	consul "github.com/kitex-contrib/registry-consul"
	"go.uber.org/zap"
)

type ProductServiceImpl struct{}

// DecreaseStock implements product.ProductService.
func (p *ProductServiceImpl) DecreaseStock(ctx context.Context, req *product.DecreaseStockReq) (r bool, err error) {
	panic("unimplemented")
}

// GetProduct implements product.ProductService.
func (p *ProductServiceImpl) GetProduct(ctx context.Context, req *product.GetProductReq) (r *product.ProductInfo, err error) {
	panic("unimplemented")
}

func main() {

	hlog.Info("=== 初始化开始 ===")
	if err := config.Init(); err != nil {
		panic(err)
	}

	zap.L().Debug("配置加载结果",
		zap.Any("redis", config.Conf.Redis),
		zap.Any("mysql", config.Conf.MySQL))

	middleware.InitAuthMiddleware("config/auth.yaml") // 增加初始化调用

	// 创建RPCConsul注册中心
	consulRegister, err := consul.NewConsulRegister(
		config.Conf.Consul.Address,
		consul.WithCheck(&api.AgentServiceCheck{
			HTTP: fmt.Sprintf("http://%s:%d/health",
				config.Conf.Service.IP,
				config.Conf.Service.ProductHTTPPort),
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
	// 启动RPC服务端
	go func() {
		// Kitex RPC服务配置
		rpcAddr, _ := net.ResolveTCPAddr("tcp", ":8881")
		svr := productservice.NewServer(
			new(ProductServiceImpl),
			kitexServer.WithRegistry(consulRegister),
			kitexServer.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
				ServiceName: "product.service",
				Tags: map[string]string{
					"protocol": "kitex",
					"env":      "dev",
				},
			}),
			kitexServer.WithServiceAddr(&net.TCPAddr{
				IP:   net.ParseIP(config.Conf.Service.IP),
				Port: config.Conf.Service.PaymentRpcPort,
			}),
		)

		zap.L().Info("product启动RPC服务", zap.String("addr", rpcAddr.String()))
		if err := svr.Run(); err != nil {
			panic("product的RPC服务启动失败: " + err.Error())
		}
	}()

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
