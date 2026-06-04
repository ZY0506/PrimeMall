package productadminlogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProductLogic) UpdateProduct(in *product.UpdateProductReq) (*product.Empty, error) {
	if in.Id == 0 {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "商品ID不能为空")
	}

	spu, err := l.svcCtx.ProductSpuModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errorx.NewBizError(response.ErrCodeProductNotFound, "商品不存在")
		}
		l.Logger.Errorf("查询商品失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	if in.CategoryId > 0 {
		spu.CategoryId = in.CategoryId
	}
	if in.Name != "" {
		spu.Name = in.Name
	}
	if in.Brand != "" {
		spu.Brand = in.Brand
	}
	if in.Description != "" {
		spu.Desc = in.Description
	}
	if in.Content != "" {
		spu.Content = sql.NullString{String: in.Content, Valid: true}
	}
	if in.Cover != "" {
		spu.MainPic = in.Cover
	}
	if len(in.Images) > 0 {
		subPics, err := json.Marshal(in.Images)
		if err != nil {
			l.Logger.Errorf("序列化副图失败, images=%v, err=%v", in.Images, err)
			return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "商品图片数据格式错误")
		}
		spu.SubPics = sql.NullString{String: string(subPics), Valid: true}
	}
	if in.VideoUrl != "" {
		spu.VideoUrl = in.VideoUrl
	}
	if in.FreightTemplateId > 0 {
		spu.FreightTemplateId = in.FreightTemplateId
	}
	if in.Status != 0 {
		spu.Status = int64(in.Status)
	}

	if len(in.Skus) > 0 {
		for i, skuItem := range in.Skus {
			if skuItem.Price <= 0 {
				return nil, errorx.NewBizError(response.ErrCodeMissingParam, fmt.Sprintf("第%d个SKU的价格必须大于0", i+1))
			}
			if skuItem.Stock < 0 {
				return nil, errorx.NewBizError(response.ErrCodeMissingParam, fmt.Sprintf("第%d个SKU的库存不能为负数", i+1))
			}
			if skuItem.MarketPrice < 0 {
				return nil, errorx.NewBizError(response.ErrCodeMissingParam, fmt.Sprintf("第%d个SKU的市场价不能为负数", i+1))
			}
			if skuItem.CostPrice < 0 {
				return nil, errorx.NewBizError(response.ErrCodeMissingParam, fmt.Sprintf("第%d个SKU的成本价不能为负数", i+1))
			}
			if skuItem.Weight < 0 {
				return nil, errorx.NewBizError(response.ErrCodeMissingParam, fmt.Sprintf("第%d个SKU的重量不能为负数", i+1))
			}
		}
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		err := l.svcCtx.ProductSpuModel.UpdateTx(ctx, session, spu)
		if err != nil {
			logx.WithContext(ctx).Errorf("更新商品SPU失败, id=%d, err=%v", in.Id, err)
			return err
		}

		if len(in.Skus) == 0 {
			return nil
		}

		existingSkus, err := l.svcCtx.ProductSkuModel.FindListBySpuId(ctx, in.Id)
		if err != nil {
			logx.WithContext(ctx).Errorf("查询现有SKU失败, spuId=%d, err=%v", in.Id, err)
			return err
		}

		existingSkuMap := make(map[uint64]model.ProductSku)
		for _, sku := range *existingSkus {
			existingSkuMap[sku.Id] = sku
		}

		updatedSkuIds := make(map[uint64]bool)
		skusToInsert := make([]*model.ProductSku, 0)
		skusToUpdate := make([]*model.ProductSku, 0)

		for _, skuItem := range in.Skus {
			specsData, err := json.Marshal(skuItem.Specs)
			if err != nil {
				logx.WithContext(ctx).Errorf("序列化SKU规格失败, skuId=%d, specs=%v, err=%v", skuItem.Id, skuItem.Specs, err)
				return errorx.NewBizError(response.ErrCodeInvalidParam, fmt.Sprintf("SKU规格数据格式错误"))
			}

			imagesData, err := json.Marshal(skuItem.Images)
			if err != nil {
				logx.WithContext(ctx).Errorf("序列化SKU图片失败, skuId=%d, images=%v, err=%v", skuItem.Id, skuItem.Images, err)
				return errorx.NewBizError(response.ErrCodeInvalidParam, fmt.Sprintf("SKU图片数据格式错误"))
			}

			if skuItem.Id > 0 {
				if existingSku, ok := existingSkuMap[skuItem.Id]; ok {
					updatedSkuIds[skuItem.Id] = true
					existingSku.SpuName = spu.Name
					existingSku.SkuCode = skuItem.SkuCode
					existingSku.Price = skuItem.Price
					existingSku.MarketPrice = skuItem.MarketPrice
					existingSku.CostPrice = skuItem.CostPrice
					existingSku.Stock = skuItem.Stock
					existingSku.SpecData = string(specsData)
					existingSku.Images = sql.NullString{String: string(imagesData), Valid: len(skuItem.Images) > 0}
					existingSku.Weight = skuItem.Weight
					existingSku.Status = skuItem.Status
					existingSku.Version = existingSku.Version + 1
					skusToUpdate = append(skusToUpdate, &existingSku)
				} else {
					newSku := &model.ProductSku{
						SpuId:       in.Id,
						SpuName:     spu.Name,
						SkuCode:     skuItem.SkuCode,
						Price:       skuItem.Price,
						MarketPrice: skuItem.MarketPrice,
						CostPrice:   skuItem.CostPrice,
						Stock:       skuItem.Stock,
						LockedStock: 0,
						Version:     0,
						SpecData:    string(specsData),
						Images:      sql.NullString{String: string(imagesData), Valid: len(skuItem.Images) > 0},
						Weight:      skuItem.Weight,
						Status:      skuItem.Status,
						DeletedAt:   sql.NullTime{Time: time.Time{}, Valid: false},
					}
					skusToInsert = append(skusToInsert, newSku)
				}
			} else {
				newSku := &model.ProductSku{
					SpuId:       in.Id,
					SpuName:     spu.Name,
					SkuCode:     skuItem.SkuCode,
					Price:       skuItem.Price,
					MarketPrice: skuItem.MarketPrice,
					CostPrice:   skuItem.CostPrice,
					Stock:       skuItem.Stock,
					LockedStock: 0,
					Version:     0,
					SpecData:    string(specsData),
					Images:      sql.NullString{String: string(imagesData), Valid: len(skuItem.Images) > 0},
					Weight:      skuItem.Weight,
					Status:      skuItem.Status,
					DeletedAt:   sql.NullTime{Time: time.Time{}, Valid: false},
				}
				skusToInsert = append(skusToInsert, newSku)
			}
		}

		skuModel := l.svcCtx.ProductSkuModel.(interface {
			withSession(session sqlx.Session) model.ProductSkuModel
		}).withSession(session)

		for _, sku := range skusToUpdate {
			err := skuModel.UpdateTx(ctx, session, sku)
			if err != nil {
				logx.WithContext(ctx).Errorf("更新SKU失败, id=%d, err=%v", sku.Id, err)
				return err
			}
		}

		for _, sku := range skusToInsert {
			_, err := skuModel.InsertTx(ctx, session, sku)
			if err != nil {
				logx.WithContext(ctx).Errorf("创建SKU失败, err=%v", err)
				return err
			}
		}

		for skuId := range existingSkuMap {
			if !updatedSkuIds[skuId] {
				err := skuModel.DeleteTx(ctx, session, skuId)
				if err != nil {
					logx.WithContext(ctx).Errorf("删除SKU失败, id=%d, err=%v", skuId, err)
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		l.Logger.Errorf("更新商品失败（已回滚）, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	l.Logger.Infof("更新商品成功, id=%d", in.Id)
	return &product.Empty{}, nil
}
