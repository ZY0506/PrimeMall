package adminproductlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	producttypes "github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 商品管理（SPU + SKU 一起创建）
func (l *CreateProductLogic) CreateProduct(in *admin.AdminCreateProductReq) (*admin.Empty, error) {
	var skus []*producttypes.SkuItem
	for _, s := range in.Skus {
		var specs []*producttypes.SpecItem
		for _, sp := range s.Specs {
			specs = append(specs, &producttypes.SpecItem{
				Key:   sp.Key,
				Value: sp.Value,
			})
		}
		skus = append(skus, &producttypes.SkuItem{
			SkuCode:     s.SkuCode,
			Price:       s.Price,
			MarketPrice: s.MarketPrice,
			CostPrice:   s.CostPrice,
			Stock:       s.Stock,
			Specs:       specs,
			Images:      s.Images,
			Weight:      s.Weight,
		})
	}

	_, err := l.svcCtx.ProductRpc.CreateProduct(l.ctx, &producttypes.CreateProductReq{
		CategoryId:        in.CategoryId,
		Name:              in.Name,
		Brand:             in.Brand,
		Description:       in.Description,
		Content:           in.Content,
		Cover:             in.Cover,
		Images:            in.Images,
		VideoUrl:          in.VideoUrl,
		FreightTemplateId: in.FreightTemplateId,
		Skus:              skus,
	})
	if err != nil {
		l.Logger.Errorf("创建商品失败,error=%v", err)
		return nil, err
	}

	return &admin.Empty{}, nil
}
