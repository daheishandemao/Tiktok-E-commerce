package dal

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB // 全局数据库实例

// User 用户模型
type User struct {
    gorm.Model  // 包含ID, CreatedAt等字段
    Username  string `gorm:"type:varchar(50);uniqueIndex;not null"`
    Password  string `gorm:"type:varchar(100);not null"`
    LastLogin *time.Time
}
// Product 商品模型（后续使用）
type Product struct {
    gorm.Model
    Name        string  `gorm:"type:varchar(100)"`
    Description string  `gorm:"type:text"`
    Price       float64 `gorm:"type:decimal(10,2)"`
    Stock       int     `gorm:"default:0"`
    Status      int     `gorm:"default:1"` // 1-上架 0-下架
}
func InitMySQL() {
    dsn := "root:123456@tcp(127.0.0.1:3306)/douyin_user?charset=utf8mb4&parseTime=True"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("数据库连接失败")
    }
    DB = db
    DB.AutoMigrate(&User{})  // 自动建表
}