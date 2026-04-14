package database

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// InitDB 初始化数据库
func InitDB(dsn string, maxIdleConns int, maxOpenConns int) sqlx.SqlConn {
	// 创建sql连接
	conn := sqlx.NewMysql(dsn)

	// 设置数据库连接池
	rawDB, err := conn.RawDB()
	if err != nil {
		panic(err)
	}
	rawDB.SetMaxIdleConns(maxIdleConns)
	rawDB.SetMaxOpenConns(maxOpenConns)

	return conn
}
