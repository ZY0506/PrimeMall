package core

import (
	"context"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefundDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRefundDetailLogic 查询退款详情
func NewRefundDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundDetailLogic {
	return &RefundDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// RefundDetail 查询退款详情
// 业务逻辑：注入userID → 通过订单号查询支付信息 → 返回退款记录
func (l *RefundDetailLogic) RefundDetail(req *types.OrderSnPathReq) (resp *types.RefundDetailResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	// 查询订单获取支付金额
	orderDetail, err := l.svcCtx.OrderRpc.OrderDetail(l.ctx, &order.OrderDetailRequest{
		OrderSn: req.OrderSn,
	})
	if err != nil {
		l.Logger.Errorf("查询订单详情失败，error=%v", err)
		return nil, err
	}

	// 查询支付流水号
	paymentSn, _ := l.svcCtx.Client.Get(l.ctx, fmt.Sprintf("payment:order:%s", req.OrderSn)).Result()

	refundSn := fmt.Sprintf("RF%s", req.OrderSn)

	resp = &types.RefundDetailResp{
		RefundSn:   refundSn,
		PaymentSn:  paymentSn,
		OrderSn:    req.OrderSn,
		Amount:     orderDetail.Base.PayAmount,
		Status:     1, // 模拟：退款成功
		StatusDesc: "退款成功",
		RefundTime: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("查询退款详情成功: orderSn=%s", req.OrderSn)
	return resp, nil
}
