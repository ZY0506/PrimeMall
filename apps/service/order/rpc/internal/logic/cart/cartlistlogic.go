package cartlogic

import (
	"context"
	"fmt"
	"strings"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CartListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCartListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CartListLogic {
	return &CartListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CartList 获取购物车列表
func (l *CartListLogic) CartList(in *order.Empty) (*order.CartListResponse, error) {
	// 获取当前用户ID
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}

	// 获取购物车列表
	list, err := l.svcCtx.CartModel.FindListByUserId(l.ctx, userId)
	if err != nil {
		l.Logger.Errorf("获取购物车列表失败，error=%v", err)
		return nil, err
	}
	var skuIds []uint64
	for _, v := range list {
		skuIds = append(skuIds, v.SkuId)
	}
	// 获取商品信息
	skusRet, err := l.svcCtx.ProductRpc.GetSkuListByIds(l.ctx, &product.SkuIdsReq{SkuIds: skuIds})
	if err != nil {
		l.Logger.Errorf("获取SKU列表失败，error=%v", err)
		return nil, err
	}
	// 构建map
	skuMap := make(map[uint64]*product.SkuItem)
	for _, sku := range skusRet.SkuItems {
		skuMap[sku.Id] = sku
	}

	// 构建返回响应
	var resp []*order.CartItem
	var totalAmount int64
	for _, v := range list {
		sku, ok := skuMap[v.SkuId]
		if !ok {
			l.Logger.Errorf("从skuMap中获取sku信息失败，skuId=%d", v.SkuId)
			return nil, errorx.NewBizError(response.ErrCodeProductNotFound, "商品不存在")
		}
		pic := ""
		if len(sku.Images) > 0 {
			pic = sku.Images[0]
		}
		skuName := ""
		if len(sku.Specs) > 0 {
			// 简单拼接规格值，实际根据业务调整
			var specStrs []string
			for _, spec := range sku.Specs {
				specStrs = append(specStrs, spec.Value)
			}
			skuName = fmt.Sprintf("%s", strings.Join(specStrs, ","))
		}
		resp = append(resp, &order.CartItem{
			Id:          v.Id,
			SkuId:       v.SkuId,
			SpuId:       sku.SpuId,
			ProductName: sku.SpuName,
			SkuName:     skuName,
			Pic:         pic,
			Price:       sku.Price,
			Quantity:    int64(v.Count),
			Selected:    v.Selected,
			Stock:       sku.Stock,
			Status:      sku.SpuStatus,
		})
		if v.Selected {
			totalAmount += sku.Price * int64(v.Count)
		}
	}
	l.Logger.Info("获取购物车列表成功")

	return &order.CartListResponse{
		Items:       resp,
		TotalAmount: totalAmount,
	}, nil
}
