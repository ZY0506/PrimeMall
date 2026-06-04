package productinternallogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type CalculateFreightLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCalculateFreightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CalculateFreightLogic {
	return &CalculateFreightLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CalculateFreightLogic) CalculateFreight(in *product.CalculateFreightReq) (*product.CalculatePriceResp, error) {
	if len(in.Items) == 0 {
		l.Logger.Errorf("计算运费失败：商品列表为空")
		return nil, errorx.NewBizError(response.ErrCodeFreightCalculationFailed, "商品列表为空")
	}

	// 查询默认运费模板
	template, err := l.svcCtx.FreightTemplateModel.FindDefault(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询运费模板失败: %v", err)
		return nil, errorx.NewBizError(response.ErrCodeFreightTemplateNotFound, "运费模板不存在")
	}

	if template.Status != 1 {
		l.Logger.Errorf("运费模板已禁用, templateId=%d", template.Id)
		return nil, errorx.NewBizError(response.ErrCodeFreightTemplateNotFound, "运费模板不可用")
	}

	// 计算商品总数量和总金额
	var totalQuantity int64
	for _, item := range in.Items {
		totalQuantity += item.Quantity
	}

	// 满件包邮检查
	if template.FreeThresholdQuantity > 0 && totalQuantity >= template.FreeThresholdQuantity {
		l.Logger.Infof("满件包邮: quantity=%d, threshold=%d", totalQuantity, template.FreeThresholdQuantity)
		return &product.CalculatePriceResp{TotalPrice: 0}, nil
	}

	// TODO: 满额包邮检查（需要获取商品价格，当前SkuStockItem无价格字段，后续可扩展）

	// 基于模板计算运费
	// 计费方式：type=1按件数，type=2按重量（当前使用件数计算）
	var freight int64
	if totalQuantity <= template.DefaultQuantity {
		freight = template.DefaultFee
	} else {
		extraUnits := totalQuantity - template.DefaultQuantity
		freight = template.DefaultFee + extraUnits*template.ExtraFee
	}

	l.Logger.Infof("计算运费成功: quantity=%d, freight=%d, template=%s", totalQuantity, freight, template.Name)
	return &product.CalculatePriceResp{
		TotalPrice: freight,
	}, nil
}
