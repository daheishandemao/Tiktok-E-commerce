package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/golang-jwt/jwt/v5"
)

var JwtSecret = []byte("douyin_secret_2024")

func JWTAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 白名单路径检查
		if c.FullPath() == "/login" || c.FullPath() == "/register" {
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
			return JwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, map[string]string{"error": "无效令牌"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("userID", claims["userID"])
		c.Next(ctx)
	}
}