package orderlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type AfterSaleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAfterSaleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AfterSaleListLogic {
	return &AfterSaleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AfterSaleListLogic) AfterSaleList(in *order.AfterSaleListRequest) (*order.AfterSaleListResponse, error) {
	// 1. 获取当前登录用户ID（仅查询自己的售后单）
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败, error=%v", err)
		return nil, err
	}

	// 2. 分页参数校验与默认值处理
	page := in.Page
	size := in.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 { // 限制最大分页数量，防止全表查询
		size = 10
	}
	offset := (page - 1) * size

	// 3. 分页查询【当前用户】的售后单列表
	afterSaleList, total, err := l.svcCtx.AfterSaleModel.FindPageByUserId(l.ctx, userId, int64(in.Status), offset, size)
	if err != nil {
		l.Logger.Errorf("分页查询售后列表失败, error=%v", err)
		return nil, err
	}

	// 5. 无数据直接返回空列表
	if len(afterSaleList) == 0 {
		return &order.AfterSaleListResponse{
			Total: total,
			List:  []*order.AfterSaleListItem{},
		}, nil
	}

	// 6. 批量提取售后单号，批量查询商品快照（高效，避免循环查库）
	var afterSaleSnList []string
	for _, item := range afterSaleList {
		afterSaleSnList = append(afterSaleSnList, item.AfterSaleSn)
	}
	// 批量查询售后商品（1对1关系）
	afterSaleItemMap, err := l.svcCtx.AfterSaleItemModel.FindMapByAfterSaleSns(l.ctx, afterSaleSnList)
	if err != nil {
		l.Logger.Errorf("批量查询售后商品失败, error=%v", err)
		return nil, err
	}

	// 7. 组装列表数据
	var respList []*order.AfterSaleListItem
	for _, sale := range afterSaleList {
		// 构建基础列表项
		item := &order.AfterSaleListItem{
			AfterSaleId:  sale.Id,
			OrderSn:      sale.OrderSn,
			Type:         order.AfterSaleType(sale.Type),
			Status:       order.AfterSaleStatus(sale.Status),
			StatusDesc:   constants.OrderStatusMap[int(sale.Status)],
			RefundAmount: sale.RealRefundAmount, // 实际退款金额
		}

		// 关联商品信息（从批量map中获取）
		if saleItem, ok := afterSaleItemMap[sale.AfterSaleSn]; ok {
			item.SpuName = saleItem.ProductName
			item.SkuName = saleItem.SkuName
			item.Pic = saleItem.SkuPic
		}

		respList = append(respList, item)
	}

	// 8. 返回最终响应
	l.Logger.Infof("查询售后列表成功, userId=%d, total=%d", userId, total)
	return &order.AfterSaleListResponse{
		Total: total,
		List:  respList,
	}, nil
}
