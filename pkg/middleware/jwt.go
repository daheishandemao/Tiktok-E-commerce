package middleware

// import (
// 	"context"
// 	"github.com/golang-jwt/jwt/v5"
// 	"github.com/cloudwego/hertz/pkg/app"
// )

// // 必须与handlers中使用相同的密钥
// var JwtSecret = []byte("douyin_secret_2024@secure")

// func JwtMiddleware() app.HandlerFunc {
// 	return func(c context.Context, ctx *app.RequestContext) {
// 		token := ctx.GetHeader("Authorization")
// 		claims, err := jwt.ParseToken(token)

// 		if err != nil {
// 			ctx.JSON(401, "Invalid token")
// 			ctx.Abort()
// 			return
// 		}

// 		ctx.Set("userID", claims.UserID)
// 		ctx.Next(c)
// 	}
// }
