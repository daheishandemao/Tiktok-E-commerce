package handlers

import (
	"context"
	// "errors"
	// "encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/dal"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/util"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 订单相关错误定义
var (
	ErrInvalidParams     = NewOrderError("参数错误")
	ErrProductNotFound   = NewOrderError("商品不存在")
	ErrStockInsufficient = NewOrderError("库存不足")
	ErrOrderCreateFailed = NewOrderError("订单创建失败")
)

type OrderHandler struct {
	db          *gorm.DB
	redisClient *redis.Client
	orderNoGen  util.OrderNoGenerator
}

func NewOrderHandler(db *gorm.DB, redisClient *redis.Client) *OrderHandler {
	return &OrderHandler{
		db:          db,
		redisClient: redisClient,
		orderNoGen:  util.NewSonyflakeGenerator(),
	}
}

type CartItem struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

// CreateOrder 创建订单
// @Summary 创建新订单
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c context.Context, ctx *app.RequestContext) {
	// 正确获取用户ID
	userIDVal, exists := ctx.Get("userID")
	if !exists {
		respondError(ctx, 401, NewOrderError("用户未认证").WithCode(401))
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		respondError(ctx, 401, NewOrderError("用户认证信息异常").WithCode(401))
		return
	}

	// 正确获取请求体
	ctx.Request.Body()
	var cartItems []CartItem
	if err := ctx.BindAndValidate(&cartItems); err != nil { // 使用Hertz的验证功能
		zap.L().Warn("参数校验失败",
			zap.Error(err),
			zap.ByteString("raw_body", ctx.Request.Body()))
		respondError(ctx, 400, ErrInvalidParams.WithDetail(err.Error()))
		return
	}

	order, err := h.createOrderTransaction(userID, cartItems)
	if err != nil {
		respondError(ctx, err.Code, err)
		return
	}

	go h.cleanCartAsync(userID, c)
	ctx.JSON(200, order)
}

// 事务性订单创建
func (h *OrderHandler) createOrderTransaction(userID uint, items []CartItem) (*dal.Order, *OrderError) {
	tx := h.db.Begin()
	if tx.Error != nil {
		return nil, NewOrderError("事务启动失败").WithCode(500)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 扣减库存
	// if err := h.deductStock(tx, items); err != nil {
	// 	tx.Rollback()
	// 	return nil, err
	// }

	// 创建订单
	order, err := h.createOrderRecord(tx, userID, items)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, NewOrderError("事务提交失败").WithCode(500)
	}

	return order, nil
}

// Redis分布式锁示例
// func acquireLock(key string, ttl time.Duration) bool {
// 	result := redis.Client.SetNX(key, "locked", ttl)
// 	return result.Val()
// }

// 扣减库存（带悲观锁）
func (h *OrderHandler) deductStock(tx *gorm.DB, items []CartItem) *OrderError {
	// 在扣库存前加锁
	// if !acquireLock("product_lock:"+productID, 10*time.Second) {
	// 	return errors.New("系统繁忙，请重试")
	// }
	// defer redis.Client.Del("product_lock:" + productID)
	for _, item := range items {
		var product dal.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&product, item.ProductID).Error; err != nil {

			if err == gorm.ErrRecordNotFound {
				return ErrProductNotFound.WithDetail(fmt.Sprintf("productID: %d", item.ProductID))
			}
			return NewOrderError("库存查询失败").WithCode(500)
		}

		if product.Stock < item.Quantity {
			return ErrStockInsufficient.WithDetail(fmt.Sprintf("productID: %d, stock: %d", item.ProductID, product.Stock))
		}

		if err := tx.Model(&product).
			Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {

			zap.L().Error("库存扣减失败",
				zap.Uint("productID", item.ProductID),
				zap.Int("quantity", item.Quantity),
				zap.Error(err))
			return NewOrderError("库存更新失败").WithCode(500)
		}
	}
	return nil
}

// 创建订单记录
func (h *OrderHandler) createOrderRecord(tx *gorm.DB, userID uint, items []CartItem) (*dal.Order, *OrderError) {
	total, err := calculateTotal(items)
	if err != nil {
		return nil, NewOrderError("金额计算失败").WithCode(400)
	}

	order := &dal.Order{
		UserID:  userID,
		OrderNo: h.orderNoGen.Generate(),
		Status:  dal.OrderStatusUnpaid,
		Amount:  total,
		Items:   marshalItems(items),
	}

	if err := tx.Create(order).Error; err != nil {
		zap.L().Error("订单创建失败",
			zap.Uint("userID", userID),
			zap.Any("items", items),
			zap.Error(err))
		return nil, ErrOrderCreateFailed.WithCode(500)
	}

	return order, nil
}

// 异步清理购物车
func (h *OrderHandler) cleanCartAsync(userID uint, c context.Context) {
	const maxRetry = 3
	key := fmt.Sprintf("cart:%d", userID)

	for i := 0; i < maxRetry; i++ {
		if err := h.redisClient.Del(c).Err(); err == nil {
			return
		}
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}

	zap.L().Warn("购物车清理失败",
		zap.Uint("userID", userID),
		zap.String("key", key))
}

// OrderError 订单业务错误
type OrderError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func NewOrderError(msg string) *OrderError {
	return &OrderError{Message: msg}
}

func (e *OrderError) WithCode(code int) *OrderError {
	e.Code = code
	return e
}

func (e *OrderError) WithDetail(detail string) *OrderError {
	e.Detail = detail
	return e
}

// 辅助函数
func calculateTotal(items []CartItem) (float64, error) {
	// 实现金额计算逻辑
	return 0.0, nil
}

func marshalItems(items []CartItem) string {
	// 实现序列化逻辑
	return ""
}

// 统一错误响应方法
func respondError(ctx *app.RequestContext, code int, err *OrderError) {
	ctx.JSON(code, map[string]interface{}{
		"code":    code,
		"message": err,
		"detail":  err.Detail,
	})
}
func UpdateOrderStatus(orderID string, status string) error {
    // 实际应调用order服务的接口
    return db.Model(&Order{}).Where("order_id = ?", orderID).Update("status", status).Error
}