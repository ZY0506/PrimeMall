package searchlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearSearchHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearSearchHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearSearchHistoryLogic {
	return &ClearSearchHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 清除搜索历史
func (l *ClearSearchHistoryLogic) ClearSearchHistory(in *search.ClearSearchHistoryReq) (*search.Empty, error) {
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("ClearSearchHistory GetUserIdFromCtx error: %v", err)
		return nil, err
	}

	err = l.svcCtx.SearchHistoryModel.DeleteByUserID(l.ctx, userId)
	if err != nil {
		l.Logger.Errorf("ClearSearchHistory DeleteByUserID error: %v", err)
		return nil, err
	}
	return &search.Empty{}, nil
}
