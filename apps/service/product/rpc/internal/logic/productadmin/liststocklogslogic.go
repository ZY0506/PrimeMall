package productadminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *ListStockLogsLogic) ListStockLogs(in *product.ListStockLogsReq) (*product.ListStockLogsResp, error) {
	if in.Page == nil {
		in.Page = &product.PageReq{Page: 1, Size: 10}
	}
	if in.Page.Page <= 0 {
		in.Page.Page = 1
	}
	if in.Page.Size <= 0 {
		in.Page.Size = 10
	}

	logs, total, err := l.svcCtx.StockLogModel.FindList(l.ctx, in.SkuId, in.Page.Page, in.Page.Size)
	if err != nil {
		l.Logger.Errorf("查询库存日志列表失败, err=%v", err)
		return nil, err
	}

	var list []*product.StockLog
	for _, log := range logs {
		createdAt := int64(0)
		if !log.CreatedAt.IsZero() {
			createdAt = log.CreatedAt.Unix()
		}
		list = append(list, &product.StockLog{
			Id:          log.Id,
			SkuId:       log.SkuId,
			OrderSn:     log.OrderSn,
			ChangeType:  product.StockOpType(log.ChangeType),
			Quantity:    log.Quantity,
			BeforeStock: log.BeforeStock,
			AfterStock:  log.AfterStock,
			Remark:      log.Remark,
			CreatedAt:   createdAt,
		})
	}

	l.Logger.Infof("查询库存日志列表成功, total=%d", total)
	return &product.ListStockLogsResp{
		Page: &product.PageResp{
			Total: total,
			Page:  in.Page.Page,
			Size:  in.Page.Size,
		},
		List: list,
	}, nil
}
