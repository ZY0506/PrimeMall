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
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

func (l *CreateProductLogic) CreateProduct(in *product.CreateProductReq) (*product.Empty, error) {
	if in.Name == "" {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "商品名称不能为空")
	}
	if in.CategoryId == 0 {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "商品分类不能为空")
	}

	if len(in.Skus) == 0 {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "商品至少需要包含一个SKU")
	}

	if in.FreightTemplateId > 0 {
		_, err := l.svcCtx.FreightTemplateModel.FindOne(l.ctx, in.FreightTemplateId)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, errorx.NewBizError(response.ErrCodeFreightTemplateNotFound, "运费模板不存在")
			}
			l.Logger.Errorf("查询运费模板失败, id=%d, err=%v", in.FreightTemplateId, err)
			return nil, err
		}
	}

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

	subPics, err := json.Marshal(in.Images)
	if err != nil {
		l.Logger.Errorf("序列化副图失败, images=%v, err=%v", in.Images, err)
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "商品图片数据格式错误")
	}

	skusData := make([]*model.ProductSku, 0, len(in.Skus))
	for i, skuItem := range in.Skus {
		specsData, err := json.Marshal(skuItem.Specs)
		if err != nil {
			l.Logger.Errorf("序列化SKU规格失败, index=%d, specs=%v, err=%v", i, skuItem.Specs, err)
			return nil, errorx.NewBizError(response.ErrCodeInvalidParam, fmt.Sprintf("第%d个SKU规格数据格式错误", i+1))
		}

		imagesData, err := json.Marshal(skuItem.Images)
		if err != nil {
			l.Logger.Errorf("序列化SKU图片失败, index=%d, images=%v, err=%v", i, skuItem.Images, err)
			return nil, errorx.NewBizError(response.ErrCodeInvalidParam, fmt.Sprintf("第%d个SKU图片数据格式错误", i+1))
		}

		sku := &model.ProductSku{
			SpuId:       0,
			SpuName:     in.Name,
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
			Status:      1,
			DeletedAt:   sql.NullTime{Time: time.Time{}, Valid: false},
		}
		skusData = append(skusData, sku)
	}

	spu := &model.ProductSpu{
		CategoryId:        in.CategoryId,
		Name:              in.Name,
		Brand:             in.Brand,
		Desc:              in.Description,
		Content:           sql.NullString{String: in.Content, Valid: in.Content != ""},
		MainPic:           in.Cover,
		SubPics:           sql.NullString{String: string(subPics), Valid: len(in.Images) > 0},
		VideoUrl:          in.VideoUrl,
		Status:            1,
		SalesCount:        0,
		VirtualSales:      0,
		FreightTemplateId: in.FreightTemplateId,
		DeleteAt:          sql.NullTime{Time: time.Time{}, Valid: false},
	}

	var spuId int64

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		result, err := l.svcCtx.ProductSpuModel.InsertTx(ctx, session, spu)
		if err != nil {
			logx.WithContext(ctx).Errorf("创建商品SPU失败, err=%v", err)
			return err
		}

		spuId, err = result.LastInsertId()
		if err != nil {
			logx.WithContext(ctx).Errorf("获取SPU ID失败, err=%v", err)
			return err
		}

		for _, sku := range skusData {
			sku.SpuId = uint64(spuId)
		}

		err = l.svcCtx.ProductSkuModel.BatchInsertTx(ctx, session, skusData)
		if err != nil {
			logx.WithContext(ctx).Errorf("批量创建SKU失败, err=%v", err)
			return err
		}

		return nil
	})

	if err != nil {
		l.Logger.Errorf("创建商品失败（已回滚）, name=%s, err=%v", in.Name, err)
		return nil, err
	}

	l.Logger.Infof("创建商品成功, id=%d, name=%s, skuCount=%d", spuId, in.Name, len(in.Skus))

	// 同步商品到ES搜索引擎（异步非阻塞）
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("同步商品到ES panic: %v", r)
			}
		}()
		if _, err := l.svcCtx.SearchRpc.SyncProductToES(context.Background(), &search.SyncProductReq{ProductId: uint64(spuId)}); err != nil {
			logx.Errorf("同步商品到ES失败, productId=%d, err=%v", spuId, err)
		}
	}()

	return &product.Empty{}, nil
}
