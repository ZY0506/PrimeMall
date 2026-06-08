package admin

import (
	"context"
	"strconv"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
)

type HandleRefundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleRefundLogic {
	return &HandleRefundLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *HandleRefundLogic) HandleRefund(req *types.HandleRefundReq) error {
	afterSaleId, err := strconv.ParseUint(req.RefundSn, 10, 64)
	if err != nil {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "售后单号格式错误")
	}
	if req.Action != 1 && req.Action != 2 {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "审核动作无效(1=同意 2=拒绝)")
	}
	_, err = l.svcCtx.AdminAfterSaleRpc.HandleAfterSale(l.ctx, &admin.AdminHandleAfterSaleReq{
		AfterSaleId: afterSaleId,
		Action:      admin.AfterSaleAuditAction(req.Action),
		Remark:      req.Remark,
	})
	if err != nil {
		l.Logger.Errorf("处理售后失败 afterSaleId=%d: %v", afterSaleId, err)
		return err
	}
	l.Logger.Infof("处理售后成功 afterSaleId=%d action=%d", afterSaleId, req.Action)
	return nil
}
