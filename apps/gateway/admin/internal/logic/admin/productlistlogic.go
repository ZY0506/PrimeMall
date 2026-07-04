// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package admin

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductListLogic {
	return &ProductListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProductListLogic) ProductList(req *types.AdminProductListReq) (resp *types.AdminProductListResp, err error) {
	rpcResp, err := l.svcCtx.AdminProductRpc.ListProducts(l.ctx, &admin.AdminListProductsReq{
		Page:       req.Page,
		PageSize:   req.Size,
		Keyword:    req.Keyword,
		CategoryId: req.CategoryId,
		Status:     admin.AdminProductStatus(req.Status),
	})
	if err != nil {
		l.Logger.Errorf("商品列表 RPC 调用失败: %v", err)
		return nil, err
	}
	list := make([]types.AdminProductItem, 0, len(rpcResp.List))
	for _, p := range rpcResp.List {
		createdAt := ""
		if p.CreatedAt != nil {
			createdAt = p.CreatedAt.AsTime().Format(time.RFC3339)
		}
		list = append(list, types.AdminProductItem{
			Id:         p.Id,
			Name:       p.Name,
			Brand:      p.Brand,
			DefaultPic: p.Cover,
			Price:      p.Price,
			Sales:      p.SalesCount,
			Status:     int64(p.Status),
			CreatedAt:  createdAt,
		})
	}
	return &types.AdminProductListResp{Total: rpcResp.Total, List: list}, nil
}
