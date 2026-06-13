package admin

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type CouponDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCouponDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CouponDetailLogic {
	return &CouponDetailLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *CouponDetailLogic) CouponDetail(req *types.IdReq) (resp *types.AdminCouponDetailResp, err error) {
	rpcResp, err := l.svcCtx.AdminCouponRpc.GetCouponDetail(l.ctx, &admin.IdReq{Id: req.Id})
	if err != nil {
		l.Logger.Errorf("优惠券详情 RPC 调用失败，id=%d, error=%v", req.Id, err)
		return nil, err
	}

	return &types.AdminCouponDetailResp{
		Id:                rpcResp.Id,
		Name:              rpcResp.Name,
		Type:              rpcResp.Type,
		ThresholdAmount:   rpcResp.ThresholdAmount,
		ReduceAmount:      rpcResp.ReduceAmount,
		DiscountRate:      rpcResp.DiscountRate,
		MaxDiscountAmount: rpcResp.MaxDiscountAmount,
		TotalQuantity:     int64(rpcResp.TotalQuantity),
		UsedQuantity:      int64(rpcResp.UsedQuantity),
		PerUserLimit:      int64(rpcResp.PerUserLimit),
		StartTime:         rpcResp.StartTime,
		EndTime:           rpcResp.EndTime,
		Status:            int64(rpcResp.Status),
		Description:       rpcResp.Description,
		CreatedAt:         rpcResp.CreatedAt,
	}, nil
}
