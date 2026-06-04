package internal_api

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UseCouponLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUseCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UseCouponLogic {
	return &UseCouponLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UseCouponLogic) UseCoupon(req *types.UseCouponReq) (resp *types.UseCouponResp, err error) {
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	rpcResp, err := l.svcCtx.MarketingRpc.UseCoupon(l.ctx, &marketing.UseCouponReq{
		UserCouponId: req.UserCouponId,
		OrderSn:      req.OrderSn,
	})
	if err != nil {
		l.Logger.Errorf("UseCoupon RPC error: %v", err)
		return nil, err
	}

	return &types.UseCouponResp{
		Success:        rpcResp.Success,
		DiscountAmount: rpcResp.DiscountAmount,
		ErrorMsg:       rpcResp.ErrorMsg,
	}, nil
}
