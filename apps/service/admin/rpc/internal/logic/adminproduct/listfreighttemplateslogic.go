package adminproductlogic

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	producttypes "github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ListFreightTemplatesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFreightTemplatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFreightTemplatesLogic {
	return &ListFreightTemplatesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFreightTemplatesLogic) ListFreightTemplates(in *admin.AdminPageReq) (*admin.AdminListFreightTemplatesResp, error) {
	resp, err := l.svcCtx.ProductRpc.ListFreightTemplates(l.ctx, &producttypes.PageReq{
		Page: in.Page,
		Size: in.PageSize,
	})
	if err != nil {
		l.Logger.Errorf("查询运费模板列表失败,error=%v", err)
		return nil, err
	}

	var list []*admin.AdminFreightTemplateInfo
	for _, item := range resp.List {
		list = append(list, &admin.AdminFreightTemplateInfo{
			Id:                    item.Id,
			Name:                  item.Name,
			CalcType:              admin.FreightCalcType(item.Type),
			DefaultFee:            item.DefaultFee,
			DefaultQuantity:       item.DefaultQuantity,
			ExtraFee:              item.ExtraFee,
			FreeThresholdAmount:   item.FreeThresholdAmount,
			FreeThresholdQuantity: item.FreeThresholdQuantity,
			IsDefault:             item.IsDefault,
			Status:                admin.SwitchStatus(item.Status),
			CreatedAt:             timestamppb.New(time.Unix(item.CreatedAt, 0)),
		})
	}

	var total int64
	if resp.Page != nil {
		total = resp.Page.Total
	}

	return &admin.AdminListFreightTemplatesResp{
		Total: total,
		List:  list,
	}, nil
}
