package redis

import (
	"context"
	"time"

	"github.com/daheishandemao/Tiktok-E-commerce/pkg/config"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Redis连接初始化
var Client *redis.Client

func InitRedis() error {
	// 必须先初始化配置
	if config.Conf == nil {
		panic("配置未初始化！请先调用 config.Init()")
	}
	// 从统一配置获取参数
	conf := config.Conf.Redis
	Client = redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Password:     conf.Password,
		DB:           conf.DB,
		PoolSize:     conf.PoolSize,
		MinIdleConns: conf.MinIdleConns,
		DialTimeout:  3 * time.Second,
	})

	// 连接健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := Client.Ping(ctx).Result(); err != nil {
		zap.L().Error("Redis连接失败", zap.String("addr", conf.Addr),zap.Error(err))
		return err
	}

	zap.L().Info("Redis连接成功",
		zap.String("addr", conf.Addr),
		zap.Int("poolSize", Client.Options().PoolSize))
	return nil
}
