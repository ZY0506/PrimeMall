package adminproductlogic

import (
	"context"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	producttypes "github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ListStockLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListStockLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListStockLogsLogic {
	return &ListStockLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListStockLogsLogic) ListStockLogs(in *admin.AdminListStockLogsReq) (*admin.AdminListStockLogsResp, error) {
	// 转换 change_type: string → StockOpType enum
	var changeType producttypes.StockOpType
	if in.ChangeType != "" {
		if v, ok := producttypes.StockOpType_value[in.ChangeType]; ok {
			changeType = producttypes.StockOpType(v)
		}
	}

	// 转换 start_time/end_time: Timestamp → string (Unix秒)
	startTimeStr := ""
	if in.StartTime != nil && in.StartTime.IsValid() {
		startTimeStr = fmt.Sprintf("%d", in.StartTime.AsTime().Unix())
	}
	endTimeStr := ""
	if in.EndTime != nil && in.EndTime.IsValid() {
		endTimeStr = fmt.Sprintf("%d", in.EndTime.AsTime().Unix())
	}

	resp, err := l.svcCtx.ProductRpc.ListStockLogs(l.ctx, &producttypes.ListStockLogsReq{
		SkuId:      in.SkuId,
		ChangeType: changeType,
		StartTime:  startTimeStr,
		EndTime:    endTimeStr,
		Page: &producttypes.PageReq{
			Page: in.Page,
			Size: in.PageSize,
		},
	})
	if err != nil {
		l.Logger.Errorf("查询库存日志失败,error=%v", err)
		return nil, err
	}

	var list []*admin.AdminStockLogItem
	for _, item := range resp.List {
		list = append(list, &admin.AdminStockLogItem{
			Id:          item.Id,
			SkuId:       item.SkuId,
			OrderSn:     item.OrderSn,
			ChangeType:  item.ChangeType.String(),
			Quantity:    item.Quantity,
			BeforeStock: item.BeforeStock,
			AfterStock:  item.AfterStock,
			Remark:      item.Remark,
			CreatedAt:   timestamppb.New(time.Unix(item.CreatedAt, 0)),
		})
	}

	var total int64
	if resp.Page != nil {
		total = resp.Page.Total
	}

	return &admin.AdminListStockLogsResp{
		Total: total,
		List:  list,
	}, nil
}
