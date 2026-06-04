package searchlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSearchHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSearchHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSearchHistoryLogic {
	return &GetSearchHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户搜索历史
func (l *GetSearchHistoryLogic) GetSearchHistory(in *search.SearchHistoryReq) (*search.SearchHistoryResp, error) {
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("GetSearchHistory GetUserIdFromCtx error: %v", err)
		return nil, err
	}

	items, err := l.svcCtx.SearchHistoryModel.FindByUserID(l.ctx, userId)
	if err != nil {
		l.Logger.Errorf("GetSearchHistory FindByUserID error: %v", err)
		return nil, err
	}

	list := make([]*search.SearchHistoryItem, 0, len(items))
	for _, item := range items {
		list = append(list, &search.SearchHistoryItem{
			Keyword:   item.Keyword,
			CreatedAt: item.CreatedAt.Unix(),
		})
	}

	return &search.SearchHistoryResp{List: list}, nil
}
