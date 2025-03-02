// pkg/idl/product.thrift
namespace go product

struct ProductInfo {
    1: required i64 id
    2: required string name
    3: required double price
    4: required i32 stock
}

struct GetProductReq {
    1: required i64 product_id
}

struct DecreaseStockReq {
    1: required i64 product_id
    2: required i32 quantity
}

service ProductService {
    ProductInfo GetProduct(1: GetProductReq req)
    bool DecreaseStock(1: DecreaseStockReq req)
}