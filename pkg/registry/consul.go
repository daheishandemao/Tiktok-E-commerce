package registry

import (
	"context"
	"fmt"
	"log" // 新增日志模块
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hashicorp/consul/api"
)

// RegisterService 注册服务到Consul（含健康检查增强逻辑）
func RegisterService(serviceName string, port int) (string, error) {
	// 创建默认配置（连接到本地Consul）
	config := api.DefaultConfig()
	config.Address = "localhost:8500" // 显式指定地址，防止DNS问题

	// 创建Consul客户端实例
	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("[致命错误] 创建Consul客户端失败: %v", err)
		return "", fmt.Errorf("consul客户端初始化失败: %w", err)
	}

	// 构建服务注册对象
	serviceID := fmt.Sprintf("%s-%d", serviceName, port)
	registration := &api.AgentServiceRegistration{
		ID:   serviceID,   // 唯一服务ID
		Name: serviceName, // 服务名称
		Port: port,
		Meta: map[string]string{"env": "dev"}, // 服务监听端口
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://localhost:%d/health", port), // 健康检查端点
			Interval:                       "15s",                                           // 检查间隔（从10s调整为15s）
			Timeout:                        "3s",                                            // 超时时间（从5s调整为3s）
			DeregisterCriticalServiceAfter: "1m",                                            // 新增：健康检查失败1分钟后注销服务
		},
		Tags: []string{"http"}, // 新增标签方便过滤
	}

	// 带重试的注册逻辑
	for retry := 0; retry < 3; retry++ {
		if err = client.Agent().ServiceRegister(registration); err == nil {
			hlog.Infof("服务注册成功 [%s:%d]", serviceName, port)
			return serviceID, nil
		}
		hlog.Warnf("注册尝试 %d/3 失败，2秒后重试...", retry+1)
		time.Sleep(2 * time.Second)
	}

	return "", fmt.Errorf("服务注册失败: %w", err)
}

// 取消服务注册
func DeregisterService(serviceID string) error {
	config := api.DefaultConfig()
	client, err := api.NewClient(config)
	if err != nil {
		hlog.Error("Consul客户端创建失败", err)
		return err
	}

	if err := client.Agent().ServiceDeregister(serviceID); err != nil {
		hlog.Errorf("服务注销失败 [ID:%s]", serviceID, err)
		return err
	}

	hlog.Infof("服务已注销 [ID:%s]", serviceID)
	return nil
}

// 新增健康检查端点处理函数
func AddHealthCheck(h *server.Hertz, serviceName string) {
	h.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(200, map[string]interface{}{
			"status":    "UP",
			"service":   serviceName, // 动态显示服务名称
			"timestamp": time.Now().Unix(),
		})
	})
}
