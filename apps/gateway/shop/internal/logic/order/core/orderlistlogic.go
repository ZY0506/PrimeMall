// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package core

import (
	"context"
	"fmt"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"time"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewOrderListLogic 订单列表
func NewOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderListLogic {
	return &OrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderListLogic) OrderList(req *types.OrderListReq) (resp *types.OrderListResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}
	// 调用rpc
	orderListResp, err := l.svcCtx.OrderRpc.OrderList(l.ctx, &order.OrderListRequest{
		Page:   req.Page,
		Size:   req.Size,
		Status: order.OrderStatus(req.Status),
	})
	if err != nil {
		l.Logger.Errorf("调用RPC获取订单列表失败，error=%v", err)
		return nil, err
	}

	// 封装响应
	list := make([]types.OrderListItem, 0)
	for _, orderInfo := range orderListResp.List {
		items := make([]types.OrderItemSnapshot, 0)
		for _, item := range orderInfo.Items {
			items = append(items, types.OrderItemSnapshot{
				FullName: fmt.Sprintf("%s %s", item.ProductName, item.SkuName),
				Pic:      item.Pic,
				Price:    item.Price,
				Quantity: item.Quantity,
			})
		}
		list = append(list, types.OrderListItem{
			OrderSn:    orderInfo.OrderSn,
			Status:     int64(orderInfo.Status),
			StatusDesc: constants.OrderStatusMap[int(orderInfo.Status)],
			PayAmount:  orderInfo.PayAmount,
			CreateTime: orderInfo.CreateTime.AsTime().Format(time.RFC3339),
			Items:      items,
		})
	}

	return &types.OrderListResp{
		List:  list,
		Total: orderListResp.Total,
	}, nil
}
