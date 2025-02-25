package registry

import (
	"context"
	"fmt"
	"log" // 新增日志模块

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hashicorp/consul/api"
)

// RegisterService 注册服务到Consul（含健康检查增强逻辑）
func RegisterService(serviceName string, port int) error {
	// 创建默认配置（连接到本地Consul）
	config := api.DefaultConfig()
	config.Address = "localhost:8500" // 显式指定地址，防止DNS问题

	// 创建Consul客户端实例
	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("[致命错误] 创建Consul客户端失败: %v", err)
		return fmt.Errorf("consul客户端初始化失败: %w", err)
	}

	// 构建服务注册对象
	registration := &api.AgentServiceRegistration{
		ID:   fmt.Sprintf("%s-%d", serviceName, port), // 唯一服务ID
		Name: serviceName,                             // 服务名称
		Port: port,                                    // 服务监听端口
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://localhost:%d/health", port), // 健康检查端点
			Interval:                       "15s",                                           // 检查间隔（从10s调整为15s）
			Timeout:                        "3s",                                            // 超时时间（从5s调整为3s）
			DeregisterCriticalServiceAfter: "1m",                                            // 新增：健康检查失败1分钟后注销服务
		},
		Tags: []string{"http"}, // 新增标签方便过滤
	}

	// 执行服务注册
	if err := client.Agent().ServiceRegister(registration); err != nil {
		log.Printf("[注册失败] 服务%s注册异常: %v", serviceName, err)
		return fmt.Errorf("服务注册失败: %w", err)
	}

	log.Printf("[成功] 服务%s:%d已注册", serviceName, port)
	return nil
}

// 新增健康检查端点处理函数
func AddHealthCheck(h *server.Hertz) {
	h.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(200, map[string]interface{}{
			"status":  "UP",
			"service": "user-service",
		})
	})
}
