// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package aftersale

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AfterSaleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAfterSaleListLogic 获取我的售后单列表（分页）
func NewAfterSaleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AfterSaleListLogic {
	return &AfterSaleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AfterSaleListLogic) AfterSaleList(req *types.AfterSaleListReq) (resp *types.AfterSaleListResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}
	// 调用rpc
	listResp, err := l.svcCtx.OrderRpc.AfterSaleList(l.ctx, &order.AfterSaleListRequest{
		Page:   req.Page,
		Size:   req.Size,
		Status: order.AfterSaleStatus(req.Status),
	})
	if err != nil {
		l.Logger.Errorf("调用RPC获取售后列表失败，error=%v", err)
		return nil, err
	}
	list := make([]types.AfterSaleListItem, 0)
	for _, v := range listResp.List {
		list = append(list, types.AfterSaleListItem{
			AfterSaleId:  v.AfterSaleId,
			OrderSn:      v.OrderSn,
			Type:         int64(v.Type),
			Status:       int64(v.Status),
			StatusDesc:   v.StatusDesc,
			RefundAmount: v.RefundAmount,
			ProductName:  v.SpuName,
			SkuName:      v.SkuName,
			Pic:          v.Pic,
		})
	}
	return &types.AfterSaleListResp{
		List:  list,
		Total: listResp.Total,
	}, nil
}
