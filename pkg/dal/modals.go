package dal

import (
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB // 全局数据库实例
type OrderStatus string

const (
	OrderStatusUnpaid   OrderStatus = "unpaid"
	OrderStatusPaid     OrderStatus = "paid"
	OrderStatusCanceled OrderStatus = "canceled"
)

// User 用户模型
type User struct {
	gorm.Model        // 包含ID, CreatedAt等字段
	Username   string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Password   string `gorm:"type:varchar(100);not null"`
	LastLogin  *time.Time
}

// Product 商品模型
type Product struct {
	gorm.Model
	Name        string  `gorm:"type:varchar(100)"`
	Description string  `gorm:"type:text"`
	Price       float64 `gorm:"type:decimal(10,2)"`
	Stock       int     `gorm:"default:0"`
	Status      int     `gorm:"default:1"` // 1-上架 0-下架
}

// Cart 购物车模型
type Cart struct {
	UserID    uint `gorm:"primaryKey"`
	ProductID uint `gorm:"primaryKey"`
	Quantity  int
}

type Order struct {
	gorm.Model
	UserID  uint
	OrderNo string `gorm:"type:varchar(32);uniqueIndex"`
	Amount  float64
	Items   string      // JSON存储商品快照
	Status  OrderStatus `gorm:"type:varchar(20);index"`
}


type PaymentRecord struct {
    gorm.Model
    OrderID     string `gorm:"uniqueIndex"`
    PaymentID   string
    Amount      float64
    Status      string // pending/success/failed
    UserID      uint
}