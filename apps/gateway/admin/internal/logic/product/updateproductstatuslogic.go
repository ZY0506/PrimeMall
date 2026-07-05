// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package product

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductStatusLogic {
	return &UpdateProductStatusLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *UpdateProductStatusLogic) UpdateProductStatus(req *types.UpdateStatusReq) error {
	if req.Id == 0 {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "商品ID不能为空")
	}
	if req.Status != 1 && req.Status != 2 {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "状态值无效(1=上架 2=下架)")
	}
	_, err := l.svcCtx.AdminProductRpc.UpdateProductStatus(l.ctx, &admin.AdminUpdateProductStatusReq{
		Id: req.Id, Status: admin.AdminProductStatus(req.Status),
	})
	if err != nil {
		l.Logger.Errorf("更新商品状态失败 productId=%d status=%d: %v", req.Id, req.Status, err)
		return err
	}
	l.Logger.Infof("更新商品状态成功 productId=%d status=%d", req.Id, req.Status)
	return nil
}
