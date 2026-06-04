package svc

import (
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/config"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/pkg/database"
	"github.com/ZY0506/PrimeMall/pkg/rdb"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config               config.Config
	DB                   sqlx.SqlConn
	Client               *redis.Client
	CategoryModel        model.CategoryModel
	ProductSpuModel      model.ProductSpuModel
	ProductSkuModel      model.ProductSkuModel
	FreightTemplateModel model.FreightTemplateModel
	StockLogModel        model.StockLogModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := database.InitDB(c.DB.DSN, c.DB.MaxIdleConns, c.DB.MaxOpenConns)
	client, err := rdb.InitRedis(c.RDB.RedisAddr, c.RDB.RedisPassword, c.RDB.RedisDB, c.RDB.RedisPoolSize, c.RDB.RedisMinIdleConns)
	if err != nil {
		logx.Errorf("redis init failed,error:%v", err.Error())
		panic(err)
	}
	return &ServiceContext{
		Config:               c,
		DB:                   db,
		Client:               client,
		CategoryModel:        model.NewCategoryModel(db),
		ProductSpuModel:      model.NewProductSpuModel(db),
		ProductSkuModel:      model.NewProductSkuModel(db),
		FreightTemplateModel: model.NewFreightTemplateModel(db),
		StockLogModel:        model.NewStockLogModel(db),
	}
}
