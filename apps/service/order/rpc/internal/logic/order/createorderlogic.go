package orderlogic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateOrder 创建订单（异步版：写队列，返回"处理中"）
func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderRequest) (resp *order.CreateOrderResponse, err error) {
	// 1. 获取用户ID
	var userId uint64
	userId, err = ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败，error=%v", err)
		return nil, errorx.NewBizError(response.ErrCodeUnauthorized, "登录失效")
	}

	// 2. 幂等性校验
	var ok bool
	idempotencyKey := fmt.Sprintf("%s%s", constants.IDEMPOTENCY_KEY+constants.ORDER_SERVICE, in.IdempotencyKey)
	ok, err = l.svcCtx.Client.SetNX(l.ctx, idempotencyKey, "1", constants.IDEMPOTENCY_EXIRE).Result()
	if err != nil {
		l.Logger.Errorf("幂等性校验失败: %v", err)
		return nil, err
	}
	if !ok {
		return &order.CreateOrderResponse{
			OrderSn:   "",
			PayAmount: 0,
			Status:    constants.ORDER_STATUS_PROCESSING,
		}, nil
	}
	defer func() {
		if err != nil {
			_ = l.svcCtx.Client.Del(l.ctx, idempotencyKey).Err()
		}
	}()

	// 3. 解析缓存（快照数据，无任何查询）
	cacheKey := constants.SETTLEMENT_TOKEN_KEY + in.SettlementToken
	var cacheJson string
	cacheJson, err = l.svcCtx.Client.Get(l.ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var settlementData SettlementData
	if err = json.Unmarshal([]byte(cacheJson), &settlementData); err != nil {
		l.Logger.Errorf("解析缓存失败: %v", err)
		return nil, err
	}

	// 权限校验
	if settlementData.UserId != userId {
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "无权限")
	}

	// 4. 生成订单ID + 订单号
	orderId, _ := l.svcCtx.IDGenerator.NextID()
	orderSn, _ := l.svcCtx.IDGenerator.GenWithPrefix(constants.PREFIX_ORDER_SN)

	// 5. 构建消息体，发送到队列异步处理
	itemMsgs := make([]svc.OrderItemMessage, len(settlementData.ItemSnapshots))
	for i, s := range settlementData.ItemSnapshots {
		itemMsgs[i] = svc.OrderItemMessage{
			SkuId:       s.SkuId,
			SpuId:       s.SpuId,
			SpuName:     s.SpuName,
			SkuName:     s.SkuName,
			SkuPic:      s.SkuPic,
			Price:       s.Price,
			Count:       s.Count,
			TotalAmount: s.TotalAmount,
		}
	}
	msg := &svc.OrderCreateMessage{
		OrderId:         orderId,
		OrderSn:         orderSn,
		UserId:          userId,
		SettlementToken: in.SettlementToken,
		PayType:         int64(in.PayType),
		Remark:          in.Remark,
		IdempotencyKey:  in.IdempotencyKey,
		CartSkuIds:      in.CartSkuIds,
		CouponId:        settlementData.CouponId, // 使用预结算解析的 user_coupon_id
		AddressSnapshot: settlementData.AddressSnapshot,
		ItemSnapshots:   itemMsgs,
		TotalAmount:     settlementData.TotalAmount,
		FreightAmount:   settlementData.FreightAmount,
		CouponDiscount:  settlementData.CouponDiscount,
		PayAmount:       settlementData.PayAmount,
	}

	msgBytes, _ := json.Marshal(msg)
	err = l.svcCtx.MQClient.Publish(l.ctx, "", constants.ORDER_CREATE_ROUTING_KEY, msgBytes)
	if err != nil {
		l.Logger.Errorf("发送订单创建消息失败，error=%v", err)
		return nil, err
	}

	// 删除结算缓存
	_ = l.svcCtx.Client.Del(l.ctx, cacheKey).Err()

	// 延长幂等key过期时间
	_ = l.svcCtx.Client.Expire(l.ctx, idempotencyKey, 24*time.Hour).Err()

	l.Logger.Infof("订单创建消息已发送，order_sn=%s", orderSn)

	return &order.CreateOrderResponse{
		OrderSn:   orderSn,
		PayAmount: settlementData.PayAmount,
		Status:    constants.ORDER_STATUS_PROCESSING,
	}, nil
}
