package svc

import (
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/config"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/pkg/database"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config             config.Config
	DB                 sqlx.SqlConn
	ES                 *ESClient
	SearchHistoryModel model.SearchHistoryModel
	HotKeywordModel    model.HotKeywordModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := database.InitDB(c.DB.DSN, c.DB.MaxIdleConns, c.DB.MaxOpenConns)

	var es *ESClient
	if len(c.Elasticsearch.Addresses) > 0 {
		var err error
		es, err = NewESClient(c.Elasticsearch.Addresses, c.Elasticsearch.Username, c.Elasticsearch.Password)
		if err != nil {
			logx.Errorf("Elasticsearch初始化失败，搜索将仅使用数据库模式，错误：%v", err)
			es = nil
		}
	}

	return &ServiceContext{
		Config:             c,
		DB:                 db,
		ES:                 es,
		SearchHistoryModel: model.NewSearchHistoryModel(db),
		HotKeywordModel:    model.NewHotKeywordModel(db),
	}
}
