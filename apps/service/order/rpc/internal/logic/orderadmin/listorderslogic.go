package orderadminlogic

import (
	"context"
	"fmt"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/client/userinternal"
	"github.com/ZY0506/PrimeMall/common/constants"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListOrders - 订单列表查询
func (l *ListOrdersLogic) ListOrders(in *order.AdminListOrdersRequest) (*order.AdminListOrdersResponse, error) {
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 20
	}

	query := "select * from order_info where 1=1"
	var args []interface{}

	if in.Status != 0 {
		query += " and status = ?"
		args = append(args, int64(in.Status))
	}

	if in.OrderSn != "" {
		query += " and order_sn like ?"
		args = append(args, "%"+in.OrderSn+"%")
	}

	if in.StartTime != nil && in.StartTime.IsValid() {
		query += " and created_at >= ?"
		args = append(args, in.StartTime.AsTime())
	}

	if in.EndTime != nil && in.EndTime.IsValid() {
		query += " and created_at <= ?"
		args = append(args, in.EndTime.AsTime())
	}

	countQuery := fmt.Sprintf("select count(*) from (%s) as tmp", query)
	var total int64
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		l.Logger.Errorf("查询订单总数失败,error=%v", err)
		return nil, err
	}

	if total == 0 {
		return &order.AdminListOrdersResponse{
			Total: 0,
			List:  []*order.AdminOrderListItem{},
		}, nil
	}

	offset := (in.Page - 1) * in.PageSize
	query += " order by created_at desc limit ?, ?"
	args = append(args, offset, in.PageSize)

	var orderList []*model.OrderInfo
	err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &orderList, query, args...)
	if err != nil {
		l.Logger.Errorf("查询订单列表失败,error=%v", err)
		return nil, err
	}

	resultList := make([]*order.AdminOrderListItem, 0, len(orderList))
	for _, orderInfo := range orderList {
		userResp, err := l.svcCtx.UserRpc.GetUserById(l.ctx, &userinternal.GetUserByIdReq{UserId: orderInfo.UserId})
		if err != nil {
			l.Logger.Errorf("查询用户信息失败,userId=%d,error=%v", orderInfo.UserId, err)
			continue
		}

		resultList = append(resultList, &order.AdminOrderListItem{
			OrderSn:    orderInfo.OrderSn,
			UserId:     orderInfo.UserId,
			UserPhone:  userResp.Phone,
			PayAmount:  orderInfo.PayAmount,
			Status:     order.OrderStatus(orderInfo.Status),
			StatusDesc: constants.OrderStatusMap[int(orderInfo.Status)],
			CreatedAt:  timestamppb.New(orderInfo.CreatedAt),
		})
	}

	l.Logger.Infof("查询订单列表成功,total=%d", total)
	return &order.AdminListOrdersResponse{
		Total: total,
		List:  resultList,
	}, nil
}
