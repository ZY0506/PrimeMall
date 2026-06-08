package coupon

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/types/marketing"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClaimCouponLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClaimCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClaimCouponLogic {
	return &ClaimCouponLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClaimCouponLogic) ClaimCoupon(req *types.ClaimCouponReq) (resp *types.ClaimCouponResp, err error) {
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	rpcResp, err := l.svcCtx.MarketingRpc.ClaimCoupon(l.ctx, &marketing.ClaimCouponReq{
		CouponId: req.CouponId,
	})
	if err != nil {
		l.Logger.Errorf("ClaimCoupon RPC error: %v", err)
		return nil, err
	}

	return &types.ClaimCouponResp{
		UserCouponId: rpcResp.UserCouponId,
		ExpireTime:   rpcResp.ExpireTime,
	}, nil
}
