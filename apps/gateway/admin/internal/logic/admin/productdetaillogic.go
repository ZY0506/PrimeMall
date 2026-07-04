package admin

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

type ProductDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductDetailLogic {
	return &ProductDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProductDetailLogic) ProductDetail(req *types.IdReq) (resp *types.AdminProductDetailResp, err error) {
	rpcResp, err := l.svcCtx.AdminProductRpc.GetProduct(l.ctx, &admin.IdReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	skus := make([]types.AdminSkuItem, 0, len(rpcResp.Skus))
	for _, s := range rpcResp.Skus {
		specs := make([]types.ProductSpecItem, 0, len(s.Specs))
		for _, sp := range s.Specs {
			specs = append(specs, types.ProductSpecItem{Name: sp.Key, Values: sp.Value})
		}
		skus = append(skus, types.AdminSkuItem{
			Id: s.Id, Code: s.SkuCode, Pic: s.Images, Price: s.Price,
			Stock: s.Stock, Weight: s.Weight, SpecData: specs,
		})
	}
	// 默认价格为所有 SKU 中的最低价
	var price int64
	if len(rpcResp.Skus) > 0 {
		price = rpcResp.Skus[0].Price
		for _, s := range rpcResp.Skus[1:] {
			if s.Price < price {
				price = s.Price
			}
		}
	}
	createdAt := ""
	if rpcResp.CreatedAt != nil {
		createdAt = rpcResp.CreatedAt.AsTime().Format(time.RFC3339)
	}
	return &types.AdminProductDetailResp{
		Id: rpcResp.Id, CategoryId: rpcResp.CategoryId, Brand: rpcResp.Brand,
		Name: rpcResp.Name, Price: price, Sales: rpcResp.SalesCount,
		Description: rpcResp.Description, Content: rpcResp.Content,
		DefaultPic: rpcResp.Cover, BannerPics: rpcResp.Images, VideoUrl: rpcResp.VideoUrl,
		FreightTemplateId: rpcResp.FreightTemplateId, Status: int64(rpcResp.Status),
		Skus: skus, CreatedAt: createdAt,
	}, nil
}
