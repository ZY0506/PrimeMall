package paymentadminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentLogic {
	return &GetPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetPayment - 支付详情查询
func (l *GetPaymentLogic) GetPayment(in *payment.AdminGetPaymentRequest) (*payment.AdminPaymentDetailResponse, error) {
	// todo: add your logic here and delete this line

	return &payment.AdminPaymentDetailResponse{}, nil
}
