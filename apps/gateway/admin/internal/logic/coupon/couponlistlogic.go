// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package coupon

import (
	"context"
	"strconv"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type CouponListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCouponListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CouponListLogic {
	return &CouponListLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *CouponListLogic) CouponList(req *types.AdminCouponListReq) (resp *types.AdminCouponListResp, err error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 10
	}
	rpcResp, err := l.svcCtx.AdminCouponRpc.ListCoupons(l.ctx, &admin.AdminListCouponsReq{
		Page:     req.Page,
		PageSize: req.Size,
		Status:   req.Status,
		Name:     req.Name,
	})
	if err != nil {
		l.Logger.Errorf("优惠券列表 RPC 调用失败: %v", err)
		return nil, err
	}
	list := make([]types.AdminCouponItem, 0, len(rpcResp.List))
	for _, c := range rpcResp.List {
		list = append(list, types.AdminCouponItem{
			Id:            c.Id,
			Name:          c.Name,
			Type:          c.Type,
			TotalQuantity: int64(c.TotalQuantity),
			UsedQuantity:  int64(c.UsedQuantity),
			Status:        int64(c.Status),
			StartTime:     strconv.FormatInt(c.StartTime, 10),
			EndTime:       strconv.FormatInt(c.EndTime, 10),
		})
	}
	return &types.AdminCouponListResp{Total: rpcResp.Total, List: list}, nil
}
