package middleware

// import (
// 	"context"

// 	"github.com/cloudwego/hertz/pkg/app"
// )

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
