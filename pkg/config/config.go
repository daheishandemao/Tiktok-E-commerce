package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"
)

var Conf *Config

type Config struct {
	Redis  RedisConfig  `yaml:"redis"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Consul ConsulConfig `yaml:"consul"`
	JWT    JWTConfig    `yaml:"jwt"`
}

type RedisConfig struct {
	Addr         string `yaml:"addr"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

// MySQL配置（用于GORM初始化）
type MySQLConfig struct {
	DSN         string `yaml:"dsn"`           // 连接字符串
	MaxIdleConn int    `yaml:"max_idle_conn"` // 连接池参数
	MaxOpenConn int    `yaml:"max_open_conn"`
}

// Consul配置（用于服务注册发现）
type ConsulConfig struct {
	Address         string `yaml:"address"`          // 注册中心地址
	ServiceCheckTTL string `yaml:"check_ttl"`        // 健康检查间隔
	DeregisterAfter string `yaml:"deregister_after"` // 服务注销延迟
}

// JWT配置（用于令牌签发）
type JWTConfig struct {
	Secret      string `yaml:"secret"`       // 加密密钥
	ExpireHours int    `yaml:"expire_hours"` // 有效期
	Issuer      string `yaml:"issuer"`       // 签发机构
}

// 其他配置结构体...

func Init() error {
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(filename))

	file, err := os.Open(filepath.Join(rootDir, "config/config.yaml"))
	if err != nil {
		return fmt.Errorf("配置文件加载失败: %v", err)
	}
	defer file.Close()

	Conf = &Config{}
	if err := yaml.NewDecoder(file).Decode(Conf); err != nil {
		return fmt.Errorf("配置文件解析失败: %v", err)
	}

	return validateConfig()
}

func validateConfig() error {
	if Conf.Redis.Addr == "" {
		return fmt.Errorf("redis地址必须配置")
	}
	if Conf.MySQL.DSN == "" {
		return fmt.Errorf("MySQL DSN必须配置")
	}
	return nil
}
