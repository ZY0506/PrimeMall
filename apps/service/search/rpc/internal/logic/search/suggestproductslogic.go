package searchlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type SuggestProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSuggestProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SuggestProductsLogic {
	return &SuggestProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ES商品搜索建议（补全）
func (l *SuggestProductsLogic) SuggestProducts(in *search.SuggestReq) (*search.SuggestResp, error) {
	if l.svcCtx.ES == nil {
		return &search.SuggestResp{Suggestions: []string{}}, nil
	}

	suggestions, err := l.svcCtx.ES.Suggest(l.ctx, in.Keyword, int(in.Size))
	if err != nil {
		// 降级处理：ES查询失败时返回空结果
		l.Logger.Errorf("搜索建议：ES查询失败，已降级返回空结果，错误：%v", err)
		return &search.SuggestResp{Suggestions: []string{}}, nil
	}
	if suggestions == nil {
		suggestions = []string{}
	}

	return &search.SuggestResp{Suggestions: suggestions}, nil
}
