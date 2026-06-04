package orderlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderListLogic {
	return &OrderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// OrderList 订单列表
func (l *OrderListLogic) OrderList(in *order.OrderListRequest) (*order.OrderListResponse, error) {
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}
	// 查询订单信息
	list, total, err := l.svcCtx.OrderInfoModel.FindList(l.ctx, userId, int64(in.Status), in.Page, in.Size)
	if err != nil {
		l.Logger.Errorf("查询订单信息失败,error=%v", err)
		return nil, err
	}

	// 拼接订单项
	orderList := make([]*order.OrderListItem, 0, total)
	for _, v := range list {
		// TODO:可以优化为批量查询
		// 查找订单项
		orderItems, err := l.svcCtx.OrderItemModel.FindListByOrderId(l.ctx, v.Id)
		if err != nil {
			l.Logger.Errorf("获取订单项失败,error=%v", err)
			return nil, err
		}
		var items []*order.OrderItem
		for _, item := range orderItems {
			items = append(items, &order.OrderItem{
				SkuId:       item.SkuId,
				SpuId:       item.SpuId,
				ProductName: item.SpuName,
				SkuName:     item.SkuName,
				Pic:         item.SkuPic,
				Price:       item.Price,
				Quantity:    item.Count,
				TotalAmount: item.TotalAmount,
			})
		}
		orderList = append(orderList, &order.OrderListItem{
			OrderSn:    v.OrderSn,
			Status:     order.OrderStatus(v.Status),
			PayAmount:  v.PayAmount,
			CreateTime: timestamppb.New(v.CreatedAt),
			Items:      items,
		})
	}
	l.Logger.Info("获取订单列表成功")
	return &order.OrderListResponse{
		Total: total,
		List:  orderList,
	}, nil
}
