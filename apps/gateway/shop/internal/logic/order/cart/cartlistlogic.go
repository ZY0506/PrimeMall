// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package cart

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CartListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCartListLogic 获取购物车列表
func NewCartListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CartListLogic {
	return &CartListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CartListLogic) CartList() (resp *types.CartListResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}
	// 调用rpc
	cartListResp, err := l.svcCtx.CartRpc.CartList(l.ctx, &order.Empty{})
	if err != nil {
		l.Logger.Errorf("调用rpc获取购物车列表失败，error=%v", err)
		return nil, err
	}

	cartList := make([]types.CartItem, 0, len(cartListResp.Items))
	for _, item := range cartListResp.Items {
		cartList = append(cartList, types.CartItem{
			Id:          item.Id,
			Pic:         item.Pic,
			Price:       item.Price,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Selected:    item.Selected,
			SkuId:       item.SkuId,
			SkuName:     item.SkuName,
			SpuId:       item.SpuId,
			Status:      item.Status,
			Stock:       item.Stock,
		})
	}

	// 构建返回响应
	resp = &types.CartListResp{
		Items:       cartList,
		TotalAmount: cartListResp.TotalAmount,
	}
	return resp, nil
}
