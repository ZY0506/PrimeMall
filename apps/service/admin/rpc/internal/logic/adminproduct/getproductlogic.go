package adminproductlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	producttypes "github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProductLogic) GetProduct(in *admin.IdReq) (*admin.AdminProductDetailResp, error) {
	resp, err := l.svcCtx.ProductRpc.GetProductAdmin(l.ctx, &producttypes.IdReq{Id: in.Id})
	if err != nil {
		l.Logger.Errorf("获取商品详情失败,error=%v", err)
		return nil, err
	}

	var skus []*admin.AdminSkuDetailItem
	for _, s := range resp.Skus {
		var specs []*admin.AdminSpecItem
		for _, sp := range s.Specs {
			specs = append(specs, &admin.AdminSpecItem{
				Key:   sp.Key,
				Value: sp.Value,
			})
		}
		skus = append(skus, &admin.AdminSkuDetailItem{
			Id:          s.Id,
			SkuCode:     s.SkuCode,
			Price:       s.Price,
			MarketPrice: s.MarketPrice,
			CostPrice:   s.CostPrice,
			Stock:       s.Stock,
			LockedStock: s.LockedStock,
			Specs:       specs,
			Images:      s.Images,
			Weight:      s.Weight,
			Status:      admin.SwitchStatus(s.Status),
		})
	}

	return &admin.AdminProductDetailResp{
		Id:                resp.Id,
		CategoryId:        resp.CategoryId,
		Name:              resp.Name,
		Brand:             resp.Brand,
		Description:       resp.Description,
		Content:           resp.Content,
		Cover:             resp.Cover,
		Images:            resp.Images,
		VideoUrl:          resp.VideoUrl,
		Status:            admin.AdminProductStatus(resp.Status),
		FreightTemplateId: resp.FreightTemplateId,
		SalesCount:        resp.SalesCount,
		VirtualSales:      resp.VirtualSales,
		Skus:              skus,
	}, nil
}
