package paymentadminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ListPaymentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPaymentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPaymentsLogic {
	return &ListPaymentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListPayments - 支付列表查询
func (l *ListPaymentsLogic) ListPayments(in *payment.AdminListPaymentsRequest) (*payment.AdminListPaymentsResponse, error) {
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 20
	}

	filter := &model.PaymentListFilter{
		Page:     in.Page,
		PageSize: in.PageSize,
		Status:   int64(in.Status),
		OrderSn:  in.OrderSn,
	}
	if in.StartTime != nil && in.StartTime.IsValid() {
		filter.StartTime = in.StartTime.AsTime()
	}
	if in.EndTime != nil && in.EndTime.IsValid() {
		filter.EndTime = in.EndTime.AsTime()
	}

	list, total, err := l.svcCtx.PaymentModel.FindListByPage(l.ctx, filter)
	if err != nil {
		l.Logger.Errorf("查询支付列表失败, error=%v", err)
		return nil, errorx.NewBizError(response.InternalError, "查询失败")
	}

	items := make([]*payment.AdminPaymentListItem, 0, len(list))
	for _, p := range list {
		var payTime *timestamppb.Timestamp
		if p.PayTime.Valid {
			payTime = timestamppb.New(p.PayTime.Time)
		}
		items = append(items, &payment.AdminPaymentListItem{
			PaymentSn: p.PaymentSn,
			OrderSn:   p.OrderSn,
			UserId:    p.UserId,
			Amount:    p.Amount,
			Status:    payment.PaymentStatus(p.Status),
			PayTime:   payTime,
		})
	}

	return &payment.AdminListPaymentsResponse{
		Total: total,
		List:  items,
	}, nil
}
