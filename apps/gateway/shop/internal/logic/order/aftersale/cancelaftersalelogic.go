// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package aftersale

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelAfterSaleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 取消售后申请（仅待审核状态可取消）
func NewCancelAfterSaleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelAfterSaleLogic {
	return &CancelAfterSaleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelAfterSaleLogic) CancelAfterSale(req *types.IdPathReq) error {
	// 注入userID
	var err error
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}
	// 调用rpc
	_, err = l.svcCtx.OrderRpc.CancelAfterSale(l.ctx, &order.CancelAfterSaleRequest{
		AfterSaleId: req.Id,
	})
	if err != nil {
		l.Logger.Errorf("调用RPC取消售后申请失败，error=%v", err)
		return err
	}

	return nil
}
