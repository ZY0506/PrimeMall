package admin

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
)

type SaveProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveProductLogic {
	return &SaveProductLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *SaveProductLogic) SaveProduct(req *types.SaveProductReq) error {
	if req.Name == "" {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "商品名称不能为空")
	}
	if req.CategoryId == 0 {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "所属分类不能为空")
	}
	skus := make([]*admin.AdminSkuItem, 0, len(req.Skus))
	for _, s := range req.Skus {
		specs := make([]*admin.AdminSpecItem, 0, len(s.SpecData))
		for _, sp := range s.SpecData {
			specs = append(specs, &admin.AdminSpecItem{Key: sp.Name, Value: sp.Value})
		}
		skus = append(skus, &admin.AdminSkuItem{
			SkuCode: s.Code, Price: s.Price, Stock: s.Stock,
			Weight: s.Weight, Images: s.Pic, Specs: specs,
		})
	}
	if req.Id > 0 {
		_, err := l.svcCtx.AdminProductRpc.UpdateProduct(l.ctx, &admin.AdminUpdateProductReq{
			Id: req.Id, CategoryId: req.CategoryId, Name: req.Name, Brand: req.Brand,
			Description: req.Description, Content: req.Content, Cover: req.DefaultPic,
			Images: req.BannerPics, VideoUrl: req.VideoUrl,
			FreightTemplateId: req.FreightTemplateId, Status: admin.AdminProductStatus(req.Status),
			Skus: skus,
		})
		if err != nil {
			l.Logger.Errorf("更新商品失败 productId=%d: %v", req.Id, err)
			return err
		}
		l.Logger.Infof("更新商品成功 productId=%d name=%s", req.Id, req.Name)
		return nil
	}
	_, err := l.svcCtx.AdminProductRpc.CreateProduct(l.ctx, &admin.AdminCreateProductReq{
		CategoryId: req.CategoryId, Name: req.Name, Brand: req.Brand,
		Description: req.Description, Content: req.Content, Cover: req.DefaultPic,
		Images: req.BannerPics, VideoUrl: req.VideoUrl,
		FreightTemplateId: req.FreightTemplateId, Skus: skus,
	})
	if err != nil {
		l.Logger.Errorf("创建商品失败: %v", err)
		return err
	}
	l.Logger.Infof("创建商品成功 name=%s", req.Name)
	return nil
}
