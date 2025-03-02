package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/redis"
)

func AddToCart(c context.Context, ctx *app.RequestContext) {
	userID := ctx.GetUint("userID")
	productIDStr := ctx.Query("product_id")

	// 参数校验
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		ctx.JSON(400, "商品ID格式错误")
		return
	}

	// 校验商品是否存在
	// var p dal.Product
	// if err := dal.DB.First(&p, productID).Error; err != nil {
	// 	ctx.JSON(404, "商品不存在")
	// 	return
	// }

	// 使用HSET存储 {用户ID: {商品ID:数量}}
	key := fmt.Sprintf("cart:%d", userID)
	if err := redis.Client.HIncrBy(c,key, strconv.FormatUint(productID, 10), 1).Err(); err != nil {
		ctx.JSON(500, "系统错误,购物车操作失败")
		return
	}

	ctx.JSON(200, map[string]interface{}{
        "user_id":    userID,
        "product_id": productID,
        "new_count":  redis.Client.HGet(c, key, strconv.FormatUint(productID, 10)).Val(),
    })
}

// 带版本号的清空操作
func ClearCart(c context.Context, ctx *app.RequestContext) {
	version := ctx.Query("version")
	if version != "confirm-v1" {
		ctx.JSON(400, "需要确认版本号")
		return
	}
	key := fmt.Sprintf("cart:%d", ctx.GetUint("userID"))
	if err := redis.Client.Del(c, key).Err(); err != nil {
        ctx.JSON(500, "清空购物车失败")
        return
    }
	ctx.JSON(200, "购物车已清空")
}
func TestRedis(c context.Context, ctx *app.RequestContext) {
    err := redis.Client.Set(c, "test_key", "hello", 10*time.Second).Err()
    if err != nil {
        ctx.JSON(500, "写入失败")
        return
    }
    
    val, err := redis.Client.Get(c, "test_key").Result()
    ctx.JSON(200, map[string]interface{}{"value": val})
}