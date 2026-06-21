package productlogic

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strconv"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/go-redis/redis/v8"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductDetailLogic {
	return &ProductDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProductDetailLogic) ProductDetail(in *product.IdReq) (*product.ProductDetailResp, error) {
	// ===================== 缓存查询（Cache-Aside） =====================
	cacheKey := constants.ProductDetailKey + strconv.FormatUint(in.Id, 10)
	cached, err := l.svcCtx.Client.Get(l.ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var cachedResp product.ProductDetailResp
		if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
			l.Logger.Infof("商品详情缓存命中，id=%v", in.Id)
			return &cachedResp, nil
		}
	}
	if err != nil && !errors.Is(err, redis.Nil) {
		l.Logger.Errorf("查询商品详情缓存失败，error=%v", err)
	}

	// 查询SPU
	spu, err := l.svcCtx.ProductSpuModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Infof("商品不存在，id=%v", in.Id)
			return nil, errorx.NewBizError(response.ErrCodeProductNotFound, "商品不存在")
		}
		l.Logger.Errorf("查询商品详情失败，error=%v", err)
		return nil, err
	}
	if spu.DeleteAt.Valid {
		l.Logger.Infof("商品不存在，id=%v", in.Id)
		return nil, errorx.NewBizError(response.ErrCodeProductNotFound, "商品不存在")
	} else if spu.Status != 1 {
		l.Logger.Infof("商品已下架，id=%v", in.Id)
		return nil, errorx.NewBizError(response.ErrCodeProductOffline, "商品已下架")
	}

	// 查询SKU
	skus, err := l.svcCtx.ProductSkuModel.FindListBySpuId(l.ctx, spu.Id)
	if err != nil {
		l.Logger.Errorf("查询商品sku失败，error=%v", err)
		return nil, err
	}
	if len(*skus) == 0 {
		l.Logger.Errorf("商品sku不存在，id=%v", in.Id)
		return nil, errorx.NewBizError(response.ErrCodeSkuNotFound, "商品sku不存在")
	}
	// 运费模板 ID 直接使用 SPU 上的值，无需额外查表（前端仅用于下单时计算运费）
	// 若 SPU 未设置模板（FreightTemplateId=0），下单时按无模板处理

	var skuList []*product.SkuItem
	attributesMap := make(map[string][]string)
	var attributes []*product.AttributeItem
	// 计算最低价
	minPrice := (*skus)[0].Price
	for _, sku := range *skus {
		if sku.Price < minPrice {
			minPrice = sku.Price
		}
		var specsData []*product.SpecItem
		err := json.Unmarshal([]byte(sku.SpecData), &specsData)
		if err != nil {
			l.Logger.Errorf("反序列化sku.SpecData失败，error=%v", err)
			return nil, err
		}
		for _, spec := range specsData {
			// 去重：检查该 key 下是否已有该 value
			if !slices.Contains(attributesMap[spec.Key], spec.Value) {
				attributesMap[spec.Key] = append(attributesMap[spec.Key], spec.Value)
			}
		}
		// 反序列化照片
		var images []string
		if sku.Images.Valid {
			err = json.Unmarshal([]byte(sku.Images.String), &images)
			if err != nil {
				l.Logger.Errorf("反序列化sku.Images失败，error=%v", err)
				return nil, err
			}
		}
		skuList = append(skuList, &product.SkuItem{
			Id:          sku.Id,
			SpuId:       sku.SpuId,
			SpuName:     sku.SpuName,
			SkuCode:     sku.SkuCode,
			Price:       sku.Price,
			MarketPrice: sku.MarketPrice,
			Stock:       sku.Stock,
			Specs:       specsData,
			Weight:      sku.Weight,
			Images:      images,
		})
	}
	for key, values := range attributesMap {
		attributes = append(attributes, &product.AttributeItem{
			Name:   key,
			Values: values,
		})
	}

	// 解析副图列表
	var subImages []string
	if spu.SubPics.Valid {
		err = json.Unmarshal([]byte(spu.SubPics.String), &subImages)
		if err != nil {
			l.Logger.Errorf("反序列化spu.SubPics失败，error=%v", err)
			return nil, err
		}
	}

	resp := &product.ProductDetailResp{
		Id:                spu.Id,
		CategoryId:        spu.CategoryId,
		Name:              spu.Name,
		Brand:             spu.Brand,
		Description:       spu.Desc,
		Cover:             spu.MainPic,
		Images:            subImages,
		VideoUrl:          spu.VideoUrl,
		Price:             minPrice,
		Sales:             spu.SalesCount,
		ShowSales:         spu.VirtualSales,
		Status:            product.ProductStatus(spu.Status),
		FreightTemplateId: spu.FreightTemplateId,
		Skus:              skuList,
		Attributes:        attributes,
	}

	// ===================== 写入缓存 =====================
	if jsonBytes, marshalErr := json.Marshal(resp); marshalErr == nil {
		l.svcCtx.Client.Set(l.ctx, cacheKey, string(jsonBytes), constants.ProductDetailTTL)
	}

	return resp, nil
}
