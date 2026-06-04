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

type DeleteProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductLogic {
	return &DeleteProductLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *DeleteProductLogic) DeleteProduct(req *types.IdReq) error {
	if req.Id == 0 {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "商品ID不能为空")
	}
	_, err := l.svcCtx.AdminProductRpc.DeleteProduct(l.ctx, &admin.IdReq{Id: req.Id})
	if err != nil {
		l.Logger.Errorf("删除商品失败 productId=%d: %v", req.Id, err)
		return err
	}
	l.Logger.Infof("删除商品成功 productId=%d", req.Id)
	return nil
}
