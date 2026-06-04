package delivery

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type TrackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewTrackLogic 查询物流轨迹
func NewTrackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TrackLogic {
	return &TrackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Track 查询物流轨迹
// 业务逻辑：注入userID → 调用OrderRpc查询订单详情获取物流信息 → 构造模拟物流轨迹
func (l *TrackLogic) Track(req *types.OrderSnPathReq) (resp *types.DeliveryTrackResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	// 调用订单RPC查询订单详情
	orderDetail, err := l.svcCtx.OrderRpc.OrderDetail(l.ctx, &order.OrderDetailRequest{
		OrderSn: req.OrderSn,
	})
	if err != nil {
		l.Logger.Errorf("调用RPC获取订单详情失败，error=%v", err)
		return nil, err
	}

	// 仅已发货或已完成订单有物流信息
	if orderDetail.Base.Status != order.OrderStatus_ORDER_STATUS_SHIPPED &&
		orderDetail.Base.Status != order.OrderStatus_ORDER_STATUS_COMPLETED {
		return nil, errorx.NewBizError(response.ErrCodeOrderStatusInvalid, "订单未发货，暂无物流信息")
	}

	// 获取物流信息
	deliverySn := orderDetail.DeliverySn
	deliveryCorp := orderDetail.DeliveryCorp

	// 如果没有物流单号，返回默认信息
	if deliverySn == "" {
		deliverySn = "暂无物流单号"
	}
	if deliveryCorp == "" {
		deliveryCorp = "暂无物流公司"
	}

	// 构造模拟物流轨迹
	// 实际项目中，此处应调用第三方物流API查询真实轨迹
	var tracks []types.DeliveryTrackItem

	// 根据订单状态返回不同层级的轨迹
	if orderDetail.Base.Status == order.OrderStatus_ORDER_STATUS_SHIPPED {
		tracks = []types.DeliveryTrackItem{
			{
				Time:        orderDetail.DeliveryTime.AsTime().Format("2006-01-02 15:04:05"),
				Station:     "发货仓库",
				Status:      "已揽收",
				Description: "快递已揽收，等待运输",
			},
			{
				Time:        orderDetail.DeliveryTime.AsTime().Add(2 * 60 * 60).Format("2006-01-02 15:04:05"),
				Station:     "中转中心",
				Status:      "运输中",
				Description: "快件已到达中转中心，准备发往目的地",
			},
		}
	} else if orderDetail.Base.Status == order.OrderStatus_ORDER_STATUS_COMPLETED {
		tracks = []types.DeliveryTrackItem{
			{
				Time:        orderDetail.DeliveryTime.AsTime().Format("2006-01-02 15:04:05"),
				Station:     "发货仓库",
				Status:      "已揽收",
				Description: "快递已揽收，等待运输",
			},
			{
				Time:        orderDetail.DeliveryTime.AsTime().Add(2 * 60 * 60).Format("2006-01-02 15:04:05"),
				Station:     "中转中心",
				Status:      "运输中",
				Description: "快件已到达中转中心，准备发往目的地",
			},
			{
				Time:        orderDetail.DeliveryTime.AsTime().Add(24 * 60 * 60).Format("2006-01-02 15:04:05"),
				Station:     "配送站",
				Status:      "派送中",
				Description: "快件已到达配送站，配送员正在派送",
			},
			{
				Time:        orderDetail.FinishTime.AsTime().Format("2006-01-02 15:04:05"),
				Station:     "",
				Status:      "已签收",
				Description: "快件已签收，感谢使用",
			},
		}
	}

	l.Logger.Infof("查询物流轨迹成功: orderSn=%s, corp=%s, sn=%s", req.OrderSn, deliveryCorp, deliverySn)

	return &types.DeliveryTrackResp{
		OrderSn:      req.OrderSn,
		DeliverySn:   deliverySn,
		DeliveryCorp: deliveryCorp,
		Status:       200, // 默认运输中
		StatusDesc:   "运输中",
		Tracks:       tracks,
	}, nil
}
