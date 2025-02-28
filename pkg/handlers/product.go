package handlers

import (
	"context"


	"github.com/cloudwego/hertz/pkg/app"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/dal"

)

//商品查询
func GetProduct(c context.Context, ctx *app.RequestContext) {
    id := ctx.Param("id")
    
    var product dal.Product
    if err := dal.DB.First(&product, id).Error; err != nil {
        ctx.JSON(404, "商品不存在")
        return
    }
    
    ctx.JSON(200, product)
}

//创造商品
func CreateProduct(c context.Context, ctx *app.RequestContext) {
    id := ctx.Param("id")
    
    var product dal.Product
    if err := dal.DB.First(&product, id).Error; err != nil {
        ctx.JSON(404, "商品不存在")
        return
    }
    
    ctx.JSON(200, product)
}