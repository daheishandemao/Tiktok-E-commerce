namespace go cart

struct CartItem {
    1: i64 user_id
    2: i64 product_id
    3: i32 quantity
}

struct CartRequest {
    1: i64 user_id
    2: optional i64 product_id
}

struct CartResponse {
    1: list<CartItem> items
    255: string message
}

service CartService {
    // 添加商品到购物车
    CartResponse AddItem(1: CartItem request)
    // 从购物车移除商品
    CartResponse RemoveItem(1: CartItem request)
    // 获取购物车内容
    CartResponse GetCart(1: CartRequest request)
}