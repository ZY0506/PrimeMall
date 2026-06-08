package adminproductlogic

import (
	"context"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/logic/utils"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	producttypes "github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdjustStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdjustStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdjustStockLogic {
	return &AdjustStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 库存管理
func (l *AdjustStockLogic) AdjustStock(in *admin.AdminAdjustStockReq) (*admin.AdminStockChangeResp, error) {
	startTime := time.Now()

	// 从 ctx 中获取管理员 ID
	adminId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取管理员ID失败,err=%v", err)
		return nil, err
	}

	// 获取管理员信息
	adminInfo, err := l.svcCtx.AdminModel.FindOne(l.ctx, adminId)
	if err != nil {
		l.Logger.Errorf("查询管理员信息失败,adminId=%d,err=%v", adminId, err)
		return nil, err
	}

	operator := fmt.Sprintf("%s(%d)", adminInfo.RealName, adminId)

	// 转换 Items
	var items []*producttypes.SkuStockItem
	for _, item := range in.Items {
		items = append(items, &producttypes.SkuStockItem{
			SkuId:    item.SkuId,
			Quantity: item.Quantity,
		})
	}

	// 调用商品服务调整库存
	resp, err := l.svcCtx.ProductRpc.AdjustStock(l.ctx, &producttypes.AdjustStockReq{
		Items:    items,
		Reason:   in.Reason,
		Operator: operator,
	})
	if err != nil {
		l.Logger.Errorf("调整库存失败,err=%v", err)
		return nil, err
	}

	// 转换结果
	var results []*admin.AdminSkuStockResult
	for _, r := range resp.Results {
		results = append(results, &admin.AdminSkuStockResult{
			SkuId:         r.SkuId,
			Success:       r.Success,
			Message:       r.Message,
			CurrentStock:  r.CurrentStock,
			CurrentLocked: r.CurrentLocked,
		})
	}

	// 记录操作日志
	stockResp := &admin.AdminStockChangeResp{
		Success: resp.Success,
		Results: results,
	}
	durationMs := time.Since(startTime).Milliseconds()
	utils.RecordAdminLog(l.ctx, l.svcCtx, utils.LogModuleProduct, "调整库存", in, stockResp, durationMs)

	l.Logger.Infof("调整库存成功,operator=%s", operator)
	return stockResp, nil
}
