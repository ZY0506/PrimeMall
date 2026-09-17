package orderlogic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDetailLogic {
	return &OrderDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OrderDetailLogic) OrderDetail(in *order.OrderDetailRequest) (*order.OrderDetailResponse, error) {
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}
	// 查找订单信息
	orderInfo, err := l.svcCtx.OrderInfoModel.FindOneByOrderSn(l.ctx, in.OrderSn)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// 异步下单窗口：消息已发出但消费端尚未落库，DB 查不到属正常情况。
			// 回查"处理中"快照，返回 status=5，避免用户下完单立刻进详情看到"订单不存在"。
			return l.buildProcessingDetail(in.OrderSn, userId)
		}
		l.Logger.Errorf("获取订单信息失败,error=%v", err)
		return nil, err
	}
	if orderInfo.UserId != userId {
		l.Logger.Errorf("用户ID不匹配,error=%v", err)
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "用户无权限操作")
	}

	// 懒检查：如果订单是待支付且已过期，自动取消
	if orderInfo.Status == constants.ORDER_STATUS_PENDING_PAY && orderInfo.ExpireTime.Valid && time.Now().After(orderInfo.ExpireTime.Time) {
		l.Infof("懒检查：订单已过期，自动取消, order_sn=%s", in.OrderSn)
		if l.svcCtx.AutoCancelExpiredOrder(l.ctx, in.OrderSn) {
			// 刷新订单状态
			orderInfo.Status = constants.ORDER_STATUS_CANCELED
		}
	}

	// 查找订单项
	orderItems, err := l.svcCtx.OrderItemModel.FindListByOrderId(l.ctx, orderInfo.Id)
	if err != nil {
		l.Logger.Errorf("获取订单项失败,error=%v", err)
		return nil, err
	}
	// 反序列化地址快照
	addrSnap := order.AddressSnapshot{}
	err = json.Unmarshal([]byte(orderInfo.AddressSnap), &addrSnap)
	if err != nil {
		l.Logger.Errorf("反序列化地址快照失败,error=%v", err)
		return nil, err
	}
	// 待支付时间
	var expireTime *timestamppb.Timestamp
	if orderInfo.ExpireTime.Valid {
		expireTime = timestamppb.New(orderInfo.ExpireTime.Time)
	}
	// 支付时间
	var payTime *timestamppb.Timestamp
	if orderInfo.PayTime.Valid {
		payTime = timestamppb.New(orderInfo.PayTime.Time)
	}
	// 发货时间
	var deliveryTime *timestamppb.Timestamp
	if orderInfo.DeliveryTime.Valid {
		deliveryTime = timestamppb.New(orderInfo.DeliveryTime.Time)
	}
	// 取消时间
	var cancelTime *timestamppb.Timestamp
	if orderInfo.CancelTime.Valid {
		cancelTime = timestamppb.New(orderInfo.CancelTime.Time)
	}
	// 完成时间
	var finishTime *timestamppb.Timestamp
	if orderInfo.ReceiveTime.Valid {
		finishTime = timestamppb.New(orderInfo.ReceiveTime.Time)
	}
	var items []*order.OrderItem
	for _, v := range orderItems {
		items = append(items, &order.OrderItem{
			SkuId:       v.SkuId,
			SpuId:       v.SpuId,
			ProductName: v.SpuName,
			SkuName:     v.SkuName,
			Pic:         v.SkuPic,
			Price:       v.Price,
			Quantity:    v.Count,
			TotalAmount: v.TotalAmount,
		})
	}
	l.Logger.Info("获取订单详情成功")
	return &order.OrderDetailResponse{
		Base: &order.OrderListItem{
			OrderSn:    orderInfo.OrderSn,
			Status:     order.OrderStatus(orderInfo.Status),
			PayAmount:  orderInfo.PayAmount,
			CreateTime: timestamppb.New(orderInfo.CreatedAt),
			Items:      items,
		},
		Address: &order.AddressSnapshot{
			ReceiverName:  addrSnap.ReceiverName,
			ReceiverPhone: addrSnap.ReceiverPhone,
			Detail: &order.AddressDetail{
				Province:      addrSnap.Detail.Province,
				City:          addrSnap.Detail.City,
				District:      addrSnap.Detail.District,
				DetailAddress: addrSnap.Detail.DetailAddress,
				PostalCode:    addrSnap.Detail.PostalCode,
			},
		},
		FreightAmount: orderInfo.FreightAmount,
		CouponAmount:  orderInfo.CouponDiscount,
		DeliverySn:    orderInfo.DeliverySn,
		DeliveryCorp:  orderInfo.DeliveryCorp,
		ExpireTime:    expireTime,
		Remark:        orderInfo.Remark,
		PayTime:       payTime,
		DeliveryTime:  deliveryTime,
		CancelTime:    cancelTime,
		FinishTime:    finishTime,
	}, nil
}

// buildProcessingDetail 在订单尚未落库时回查"处理中"快照（下单请求已受理、消费端尚在处理），
// 返回 status=5 的订单详情。消费端落库成功后快照即被删除，此后一律以 DB 为准。
// 快照缺失（消息丢失、消费失败、TTL 过期）时仍返回"订单不存在"，保持原有的降级行为。
func (l *OrderDetailLogic) buildProcessingDetail(orderSn string, userId uint64) (*order.OrderDetailResponse, error) {
	notFound := errorx.NewBizError(response.ErrCodeOrderNotFound, "订单不存在")

	snapshot, err := l.svcCtx.Client.Get(l.ctx, constants.OrderProcessingKey+orderSn).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			l.Logger.Errorf("查询订单处理中快照失败, order_sn=%s, error=%v", orderSn, err)
		}
		l.Logger.Errorf("订单号：%s 不存在", orderSn)
		return nil, notFound
	}

	var msg svc.OrderCreateMessage
	if err = json.Unmarshal([]byte(snapshot), &msg); err != nil {
		l.Logger.Errorf("解析订单处理中快照失败, order_sn=%s, error=%v", orderSn, err)
		return nil, notFound
	}

	// 权限校验必须在返回任何数据之前：快照内含结算令牌、幂等键等敏感字段
	if msg.UserId != userId {
		l.Logger.Errorf("用户ID不匹配（处理中快照）, order_sn=%s", orderSn)
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "用户无权限操作")
	}

	var addrSnap order.AddressSnapshot
	if err = json.Unmarshal([]byte(msg.AddressSnapshot), &addrSnap); err != nil {
		l.Logger.Errorf("反序列化地址快照失败, order_sn=%s, error=%v", orderSn, err)
		return nil, err
	}
	addr := &order.AddressSnapshot{
		ReceiverName:  addrSnap.ReceiverName,
		ReceiverPhone: addrSnap.ReceiverPhone,
	}
	if addrSnap.Detail != nil {
		addr.Detail = &order.AddressDetail{
			Province:      addrSnap.Detail.Province,
			City:          addrSnap.Detail.City,
			District:      addrSnap.Detail.District,
			DetailAddress: addrSnap.Detail.DetailAddress,
			PostalCode:    addrSnap.Detail.PostalCode,
		}
	}

	items := make([]*order.OrderItem, 0, len(msg.ItemSnapshots))
	for _, v := range msg.ItemSnapshots {
		items = append(items, &order.OrderItem{
			SkuId:       v.SkuId,
			SpuId:       v.SpuId,
			ProductName: v.SpuName,
			SkuName:     v.SkuName,
			Pic:         v.SkuPic,
			Price:       v.Price,
			Quantity:    v.Count,
			TotalAmount: v.TotalAmount,
		})
	}

	l.Logger.Infof("订单处理中，返回下单快照, order_sn=%s", orderSn)
	return &order.OrderDetailResponse{
		Base: &order.OrderListItem{
			OrderSn:    msg.OrderSn,
			Status:     order.OrderStatus(constants.ORDER_STATUS_PROCESSING),
			PayAmount:  msg.PayAmount,
			CreateTime: timestamppb.Now(),
			Items:      items,
		},
		Address:       addr,
		FreightAmount: msg.FreightAmount,
		CouponAmount:  msg.CouponDiscount,
		Remark:        msg.Remark,
	}, nil
}
