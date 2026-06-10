package svc

import (
	"context"
	"strconv"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/config"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/client/search"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/pkg/database"
	"github.com/ZY0506/PrimeMall/pkg/rdb"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
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
	SearchRpc            search.Search
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := database.InitDB(c.DB.DSN, c.DB.MaxIdleConns, c.DB.MaxOpenConns)
	client, err := rdb.InitRedis(c.RDB.RedisAddr, c.RDB.RedisPassword, c.RDB.RedisDB, c.RDB.RedisPoolSize, c.RDB.RedisMinIdleConns)
	if err != nil {
		logx.Errorf("redis init failed,error:%v", err.Error())
		panic(err)
	}
	sc := &ServiceContext{
		Config:               c,
		DB:                   db,
		Client:               client,
		CategoryModel:        model.NewCategoryModel(db),
		ProductSpuModel:      model.NewProductSpuModel(db),
		ProductSkuModel:      model.NewProductSkuModel(db),
		FreightTemplateModel: model.NewFreightTemplateModel(db),
		StockLogModel:        model.NewStockLogModel(db),
		SearchRpc:            search.NewSearch(zrpc.MustNewClient(c.SearchRpc)),
	}

	// 异步预热热销SKU库存到Redis
	go sc.warmUpStockCache()

	return sc
}

// warmUpStockCache 启动时预热热销商品库存到Redis缓存
func (s *ServiceContext) warmUpStockCache() {
	ctx := context.Background()
	logx.Info("开始预热商品库存缓存...")

	skus, err := s.ProductSkuModel.FindHotSkus(ctx, 200)
	if err != nil {
		logx.Errorf("预热库存失败：查询热销SKU出错，%v", err)
		return
	}

	pipe := s.Client.Pipeline()
	for _, sku := range *skus {
		key := constants.ProductStockKey + strconv.FormatUint(sku.Id, 10)
		pipe.Set(ctx, key, sku.Stock, constants.ProductStockTTL)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		logx.Errorf("预热库存写入Redis失败，%v", err)
		return
	}

	logx.Infof("库存缓存预热完成，共加载 %d 个SKU", len(*skus))
}
