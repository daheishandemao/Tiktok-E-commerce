package service

import "errors"

func Settlement(userID uint) (string, error) {
	// 1. 获取购物车数据
	cartItems := GetCartItems(userID)

	// 2. 校验库存
	for _, item := range cartItems {
		if !CheckStock(item.ProductID, item.Quantity) {
			return "", errors.New("库存不足")
		}
	}

	// 3. 创建订单
	orderID := GenerateOrderID() // 使用雪花算法生成

	// 4. 扣减库存
	if err := DecreaseStock(cartItems); err != nil {
		return "", err
	}

	// 5. 清空购物车
	ClearCart(userID)

	return orderID, nil
}
