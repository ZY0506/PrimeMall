package searchlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHotKeywordsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHotKeywordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHotKeywordsLogic {
	return &GetHotKeywordsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 热搜词列表
func (l *GetHotKeywordsLogic) GetHotKeywords(in *search.Empty) (*search.HotKeywordsResp, error) {
	items, err := l.svcCtx.HotKeywordModel.FindActiveList(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取热搜词：查询失败，错误：%v", err)
		return nil, err
	}

	list := make([]*search.HotKeywordItem, 0, len(items))
	for _, item := range items {
		list = append(list, &search.HotKeywordItem{
			Id:          item.Id,
			Keyword:     item.Keyword,
			SearchCount: item.SearchCount,
			Sort:        int32(item.Sort),
		})
	}

	return &search.HotKeywordsResp{List: list}, nil
}
