//go:build !sonic

package main

import (
	"context"
	"fmt"
	// "net"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	hclient "github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	// "github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/daheishandemao/Tiktok-E-commerce/kitex_gen/order"
	"github.com/daheishandemao/Tiktok-E-commerce/kitex_gen/order/orderservice"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/config"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/dal"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/middleware"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/redis"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/registry"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentRecord struct {
	gorm.Model
	OrderID   string `gorm:"uniqueIndex"`
	PaymentID string
	Amount    float64
	Status    string // pending/success/failed
	UserID    uint
}

var (
	httpClient  *hclient.Client
	orderClient orderservice.Client
)

func initOrderClient() {
	r, err := consul.NewConsulResolver(config.Conf.Consul.Address)
	if err != nil {
		panic(fmt.Sprintf("Consul初始化失败: %v", err))
	}

	orderClient, err = orderservice.NewClient(
		"order.service",
		client.WithResolver(r),
		client.WithRPCTimeout(3*time.Second),
		client.WithHostPorts("localhost:8080"), // 与order服务端口一致
	)
	if err != nil {
		panic(fmt.Sprintf("订单服务客户端初始化失败: %v", err))
	}
}

func main() {
	// 初始化基础组件
	hlog.Info("=== 支付服务初始化 ===")
	if err := config.Init(); err != nil {
		panic("配置加载失败: " + err.Error())
	}

	// 初始化基础设施
	dal.InitDB()
	if err := redis.InitRedis(); err != nil {
		panic("Redis初始化失败: " + err.Error())
	}

	// 初始化服务客户端
	middleware.InitAuthMiddleware("config/auth.yaml")
	initOrderClient()

	// 创建HTTP服务器
	h := server.Default(
		server.WithHostPorts(":8084"),
		server.WithExitWaitTime(5*time.Second),
	)

	// 服务注册
	if _, err := registry.RegisterService("payment-service", 8084); err != nil {
		panic("服务注册失败: " + err.Error())
	}

	// 路由配置
	registerRoutes(h)

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
		zap.L().Info("支付服务关闭中...")
		redis.Client.Close()
	})

	h.Spin()
}

func registerRoutes(h *server.Hertz) {
	// 支付回调接口
	h.POST("/payment/callback", func(c context.Context, ctx *app.RequestContext) {
		orderID := ctx.Query("order_id")
		if err := UpdateOrderStatus(orderID, "paid"); err != nil {
			zap.L().Error("订单状态更新失败", 
				zap.String("order_id", orderID),
				zap.Error(err))
			ctx.JSON(500, map[string]interface{}{"error": err.Error()})
			return
		}
		ctx.JSON(200, map[string]interface{}{"status": "success"})
	})

	// 创建支付记录
	h.POST("/payment/create", func(c context.Context, ctx *app.RequestContext) {
		var req struct {
			OrderID string  `json:"order_id"`
			Amount  float64 `json:"amount"`
			UserID  uint    `json:"user_id"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(400, map[string]interface{}{"error": "无效请求参数"})
			return
		}

		// 生成支付记录
		paymentID := uuid.New().String()
		record := PaymentRecord{
			OrderID:   req.OrderID,
			PaymentID: paymentID,
			Amount:    req.Amount,
			Status:    "pending",
			UserID:    req.UserID,
		}

		if err := dal.DB.Create(&record).Error; err != nil {
			ctx.JSON(500, map[string]interface{}{"error": "支付记录创建失败"})
			return
		}

		// 返回模拟支付链接
		ctx.JSON(200, map[string]interface{}{
			"payment_url": fmt.Sprintf("http://localhost:8084/payment/confirm?payment_id=%s", paymentID),
		})
	})
}

// UpdateOrderStatus 通过RPC更新订单状态
func UpdateOrderStatus(orderID string, status string) error {
	req := &order.UpdateReq{
		OrderID: orderID,
		Status:  status,
	}

	resp, err := orderClient.UpdateStatus(context.Background(), req)
	if err != nil {
		return fmt.Errorf("RPC调用失败: %w", err)
	}

	if !resp {
		return fmt.Errorf("订单状态更新失败")
	}

	zap.L().Info("订单状态已更新",
		zap.String("order_id", orderID),
		zap.String("status", status))
	return nil
}