package dal

import (
	"fmt"

	"github.com/daheishandemao/Tiktok-E-commerce/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() {
	// 配置MySQL连接（根据实际情况修改）
	// dsn := "root:123456@tcp(localhost:3306)/douyin?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := config.Conf.MySQL.DSN
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}

	// 自动迁移表结构
	if err := DB.AutoMigrate(&User{}, &Product{}, &Order{}); err != nil {
		panic(fmt.Sprintf("数据库迁移失败: %v", err))
	}

	// 设置连接池
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(config.Conf.MySQL.MaxIdleConn)
	sqlDB.SetMaxOpenConns(config.Conf.MySQL.MaxOpenConn)
}
