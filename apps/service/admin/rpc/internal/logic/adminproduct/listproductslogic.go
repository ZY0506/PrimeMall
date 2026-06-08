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

type ListProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListProductsLogic) ListProducts(in *admin.AdminListProductsReq) (*admin.AdminListProductsResp, error) {
	resp, err := l.svcCtx.ProductRpc.ListProductsAdmin(l.ctx, &producttypes.ListProductsAdminReq{
		CategoryId: in.CategoryId,
		Keyword:    in.Keyword,
		Status:     producttypes.ProductStatus(in.Status),
		Page: &producttypes.PageReq{
			Page: in.Page,
			Size: in.PageSize,
		},
	})
	if err != nil {
		l.Logger.Errorf("查询商品列表失败,error=%v", err)
		return nil, err
	}

	var list []*admin.AdminProductItem
	for _, item := range resp.List {
		list = append(list, &admin.AdminProductItem{
			Id:         item.Id,
			Name:       item.Name,
			Brand:      item.Brand,
			Cover:      item.Cover,
			Price:      item.Price,
			SalesCount: item.SalesCount,
			Status:     admin.AdminProductStatus(item.Status),
			CreatedAt:  timestamppb.New(time.Unix(item.CreatedAt, 0)),
		})
	}

	var total int64
	if resp.Page != nil {
		total = resp.Page.Total
	}

	return &admin.AdminListProductsResp{
		Total: total,
		List:  list,
	}, nil
}
