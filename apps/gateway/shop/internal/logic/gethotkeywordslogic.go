// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHotKeywordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 热搜词列表
func NewGetHotKeywordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHotKeywordsLogic {
	return &GetHotKeywordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHotKeywordsLogic) GetHotKeywords() (resp *types.HotKeywordsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
