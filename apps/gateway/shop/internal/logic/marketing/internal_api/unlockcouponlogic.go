package internal_api

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlockCouponLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnlockCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlockCouponLogic {
	return &UnlockCouponLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnlockCouponLogic) UnlockCoupon(req *types.UnlockCouponReq) (resp *types.UnlockCouponResp, err error) {
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	rpcResp, err := l.svcCtx.MarketingRpc.UnlockCoupon(l.ctx, &marketing.UnlockCouponReq{
		UserCouponId: req.UserCouponId,
		OrderSn:      req.OrderSn,
	})
	if err != nil {
		l.Logger.Errorf("UnlockCoupon RPC error: %v", err)
		return nil, err
	}

	return &types.UnlockCouponResp{
		Success:  rpcResp.Success,
		ErrorMsg: rpcResp.ErrorMsg,
	}, nil
}
