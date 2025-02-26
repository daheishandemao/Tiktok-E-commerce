package dal

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() {
	// 配置MySQL连接（根据实际情况修改）
	dsn := "root:123456@tcp(localhost:3306)/douyin?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}

	// 自动迁移表结构
	if err := DB.AutoMigrate(&User{}, &Product{}); err != nil {
		panic(fmt.Sprintf("数据库迁移失败: %v", err))
	}
}
