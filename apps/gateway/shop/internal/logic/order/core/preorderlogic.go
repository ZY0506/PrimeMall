// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package core

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPreOrderLogic 预下单（确认订单页，生成结算令牌）
func NewPreOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreOrderLogic {
	return &PreOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreOrderLogic) PreOrder(req *types.PreOrderReq) (resp *types.PreOrderResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}
	// 调用rpc
	items := make([]*order.OrderItemSimple, 0)
	for _, item := range req.Items {
		items = append(items, &order.OrderItemSimple{
			SkuId:    item.SkuId,
			Quantity: item.Quantity,
		})
	}
	preOrderResp, err := l.svcCtx.OrderRpc.PreOrder(l.ctx, &order.PreOrderRequest{
		AddressId: req.AddressId,
		CouponId:  req.CouponId,
		Items:     items,
	})
	if err != nil {
		l.Logger.Errorf("调用RPC预下单失败，error=%v", err)
		return nil, err
	}

	// 封装响应数据
	orderItems := make([]types.OrderItem, 0)
	for _, item := range preOrderResp.Items {
		orderItems = append(orderItems, types.OrderItem{
			SkuId:       item.SkuId,
			SpuId:       item.SpuId,
			ProductName: item.ProductName,
			SkuName:     item.SkuName,
			Pic:         item.Pic,
			Price:       item.Price,
			Quantity:    item.Quantity,
			TotalAmount: item.TotalAmount,
		})
	}
	resp = &types.PreOrderResp{
		Token: preOrderResp.SettlementToken,
		Items: orderItems,
		Address: types.AddressSnapshot{
			ReceiverName:  preOrderResp.Address.ReceiverName,
			ReceiverPhone: preOrderResp.Address.ReceiverPhone,
			Detail: types.AddressDetail{
				Province:      preOrderResp.Address.Detail.Province,
				City:          preOrderResp.Address.Detail.City,
				DetailAddress: preOrderResp.Address.Detail.DetailAddress,
				District:      preOrderResp.Address.Detail.District,
				PostalCode:    preOrderResp.Address.Detail.PostalCode,
			},
		},
		TotalAmount:   preOrderResp.TotalAmount,
		FreightAmount: preOrderResp.FreightAmount,
		CouponAmount:  preOrderResp.CouponAmount,
		PayAmount:     preOrderResp.PayAmount,
	}

	return resp, nil
}
