package orderadminlogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleAfterSaleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleAfterSaleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleAfterSaleLogic {
	return &HandleAfterSaleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HandleAfterSale - 处理售后（审核同意/拒绝）
func (l *HandleAfterSaleLogic) HandleAfterSale(in *order.AdminHandleAfterSaleRequest) (*order.Empty, error) {
	if in.AfterSaleId == 0 {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "售后单ID不能为空")
	}

	if in.Action != order.AfterSaleAuditAction_AFTER_SALE_AUDIT_ACTION_APPROVE &&
		in.Action != order.AfterSaleAuditAction_AFTER_SALE_AUDIT_ACTION_REJECT {
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "审核动作不合法")
	}

	afterSaleInfo, err := l.svcCtx.AfterSaleModel.FindOne(l.ctx, in.AfterSaleId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("售后单不存在,after_sale_id=%d", in.AfterSaleId)
			return nil, errorx.NewBizError(response.ErrCodeAfterSaleNotFound, "售后单不存在")
		}
		l.Logger.Errorf("查询售后单失败,error=%v", err)
		return nil, err
	}

	if afterSaleInfo.Status != constants.AFTER_SALE_STATUS_PENDING {
		l.Logger.Errorf("售后单状态不正确,无法审核,after_sale_id=%d,status=%d", in.AfterSaleId, afterSaleInfo.Status)
		return nil, errorx.NewBizError(response.ErrCodeAfterSaleStatusInvalid, "售后单状态不正确，只有待审核的售后单才能审核")
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		now := time.Now()
		afterSaleInfo.AuditRemark = in.Remark
		afterSaleInfo.AuditTime = sql.NullTime{Time: now, Valid: true}

		if in.Action == order.AfterSaleAuditAction_AFTER_SALE_AUDIT_ACTION_APPROVE {
			afterSaleInfo.ApprovedAmount = afterSaleInfo.ApplyAmount

			if afterSaleInfo.Type == constants.AFTER_SALE_TYPE_REFUND_ONLY {
				afterSaleInfo.Status = constants.AFTER_SALE_STATUS_REFUNDING
				afterSaleInfo.RefundStatus = constants.REFUND_STATUS_REFUNDING
			} else {
				afterSaleInfo.Status = constants.AFTER_SALE_STATUS_RETURN
			}
		} else {
			afterSaleInfo.Status = constants.AFTER_SALE_STATUS_REFUSED
		}

		err := l.svcCtx.AfterSaleModel.UpdateTx(ctx, session, afterSaleInfo)
		if err != nil {
			l.Logger.Errorf("更新售后单失败,error=%v", err)
			return err
		}

		return nil
	})

	if err != nil {
		l.Logger.Errorf("处理售后失败,error=%v", err)
		return nil, err
	}

	actionDesc := "同意"
	if in.Action == order.AfterSaleAuditAction_AFTER_SALE_AUDIT_ACTION_REJECT {
		actionDesc = "拒绝"
	}
	l.Logger.Infof("处理售后成功,after_sale_id=%d,action=%s", in.AfterSaleId, actionDesc)
	return &order.Empty{}, nil
}
