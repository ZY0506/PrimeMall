// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package order

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func tsFormat(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339)
}

type OrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDetailLogic {
	return &OrderDetailLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *OrderDetailLogic) OrderDetail(req *types.OrderSnReq) (resp *types.AdminOrderDetailResp, err error) {
	if req.OrderSn == "" {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "订单号不能为空")
	}
	rpcResp, err := l.svcCtx.AdminOrderRpc.GetOrder(l.ctx, &admin.AdminGetOrderReq{OrderSn: req.OrderSn})
	if err != nil {
		return nil, err
	}
	addr := types.AdminOrderAddress{}
	if rpcResp.Address != nil {
		addr = types.AdminOrderAddress{
			ReceiverName: rpcResp.Address.ReceiverName, ReceiverPhone: rpcResp.Address.ReceiverPhone,
			Province: rpcResp.Address.Province, City: rpcResp.Address.City,
			District: rpcResp.Address.District, DetailAddress: rpcResp.Address.DetailAddress,
		}
	}
	items := make([]types.AdminOrderItem, 0, len(rpcResp.Items))
	for _, item := range rpcResp.Items {
		items = append(items, types.AdminOrderItem{
			SkuId: item.SkuId, SpuId: item.SpuId, ProductName: item.ProductName,
			SkuName: item.SkuName, Pic: item.Pic, Price: item.Price,
			Quantity: item.Quantity, TotalAmount: item.TotalAmount,
		})
	}
	return &types.AdminOrderDetailResp{
		OrderSn: rpcResp.OrderSn, UserId: rpcResp.UserId,
		UserPhone: rpcResp.UserPhone, Status: int64(rpcResp.Status),
		StatusDesc: rpcResp.StatusDesc, PayType: int64(rpcResp.PayType),
		PayTypeDesc: rpcResp.PayTypeDesc, PayAmount: rpcResp.PayAmount,
		TotalAmount: rpcResp.TotalAmount, FreightAmount: rpcResp.FreightAmount,
		CouponAmount: rpcResp.CouponAmount, Remark: rpcResp.Remark,
		Address: addr, Items: items,
		DeliverySn: rpcResp.DeliverySn, DeliveryCorp: rpcResp.DeliveryCompany,
		CreateTime: tsFormat(rpcResp.CreateTime), PayTime: tsFormat(rpcResp.PayTime),
		DeliveryTime: tsFormat(rpcResp.DeliveryTime), FinishTime: tsFormat(rpcResp.FinishTime),
		CancelTime: tsFormat(rpcResp.CancelTime),
	}, nil
}
