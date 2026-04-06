package database

import (
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDB 初始化数据库
func InitDB(dsn string, maxIdleConns int, maxOpenConns int) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 检查底层 DB 连接是否存活
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
		return nil
	}
	err = sqlDB.Ping()
	if err != nil {
		panic(err)
		return nil
	}

	// 设置连接池中的最大闲置连接数
	sqlDB.SetMaxIdleConns(maxIdleConns)
	// 设置数据库的最大打开连接数
	sqlDB.SetMaxOpenConns(maxOpenConns)

	return db
}
