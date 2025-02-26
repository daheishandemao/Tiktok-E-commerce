package middleware

import (
	"context"
	"log"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/config"  // 替换为实际模块路径

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/golang-jwt/jwt/v5"
)

var (
	JwtSecret     = []byte("douyin_secret_2024")
	initialized   = false  // 配置初始化标志
)

// 初始化时加载配置
var whitelist map[string]bool

// 初始化配置（需显式调用）
func InitAuthMiddleware(configPath string) {
	if err := config.InitAuthConfig(configPath); err != nil {
		log.Fatalf("认证配置加载失败: %v", err)
	}
	initialized = true
}

func JWTAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !initialized {
			c.JSON(500, map[string]string{"error": "认证模块未初始化"})
			c.Abort()
			return
		}
		
		currentPath := c.FullPath()

		// 黑名单拦截（最高优先级）
		if config.BlacklistMap[currentPath] {
			c.JSON(403, map[string]string{"error": "禁止访问"})
			c.Abort()
			return
		}
		
		// 白名单放行
		if config.WhitelistMap[currentPath] {
			c.Next(ctx)
			return
		}
		
	
		// 关键修改点：将 []byte 转换为 string
		tokenString := string(c.GetHeader("Authorization"))
		if tokenString == "" {
			c.JSON(401, map[string]string{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return JwtSecret, nil
		})


		if err != nil || !token.Valid {
			c.JSON(401, map[string]string{"error": "无效令牌"})
			c.Abort()
			return
		}

		
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, exists := claims["userID"]; exists {
				c.Set("userID", userID)
				c.Next(ctx)
				return
			}
		}
		c.JSON(401, map[string]string{"error": "令牌解析失败"})
		c.Abort()
	}
}