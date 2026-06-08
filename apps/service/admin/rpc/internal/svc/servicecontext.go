package svc

import (
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/config"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/client/orderadmin"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/client/productadmin"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/client/admin"
	"github.com/ZY0506/PrimeMall/common/snowflakes"
	"github.com/ZY0506/PrimeMall/pkg/database"
	"github.com/ZY0506/PrimeMall/pkg/rdb"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config              config.Config
	DB                  sqlx.SqlConn
	Client              *redis.Client
	AdminModel          model.AdminModel
	RoleModel           model.RoleModel
	PermissionModel     model.PermissionModel
	RolePermissionModel model.RolePermissionModel
	AdminLogModel       model.AdminLogModel
	IDGenerator         *snowflakes.Generator
	ProductRpc          productadmin.ProductAdmin
	UserRpc             admin.Admin
	OrderRpc            orderadmin.OrderAdmin
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化db
	db := database.InitDB(c.DB.DSN, c.DB.MaxIdleConns, c.DB.MaxOpenConns)
	// 初始化redis
	client, err := rdb.InitRedis(c.RDB.RedisAddr, c.RDB.RedisPassword, c.RDB.RedisDB, c.RDB.RedisPoolSize, c.RDB.RedisMinIdleConns)
	if err != nil {
		logx.Errorf("redis init failed,error:%v", err.Error())
		panic(err)
	}
	IDGenerator, err := snowflakes.NewGenerator(c.Snowflake.NodeID)
	if err != nil {
		logx.Errorf("snowflake init failed,error:%v", err.Error())
		panic(err)
	}
	return &ServiceContext{
		Config:              c,
		DB:                  db,
		Client:              client,
		AdminModel:          model.NewAdminModel(db),
		RoleModel:           model.NewRoleModel(db),
		PermissionModel:     model.NewPermissionModel(db),
		RolePermissionModel: model.NewRolePermissionModel(db),
		AdminLogModel:       model.NewAdminLogModel(db),
		IDGenerator:         IDGenerator,
		ProductRpc:          productadmin.NewProductAdmin(zrpc.MustNewClient(c.ProductRpc)),
		UserRpc:             admin.NewAdmin(zrpc.MustNewClient(c.UserRpc)),
		OrderRpc:            orderadmin.NewOrderAdmin(zrpc.MustNewClient(c.OrderRpc)),
	}
}
