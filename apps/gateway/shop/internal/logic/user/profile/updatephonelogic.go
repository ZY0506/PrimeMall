// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package profile

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/ZY0506/PrimeMall/common/utils"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePhoneLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdatePhoneLogic 修改手机号（幂等）
func NewUpdatePhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePhoneLogic {
	return &UpdatePhoneLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePhoneLogic) UpdatePhone(req *types.UpdatePhoneReq) error {
	var err error
	// 参数校验
	if !utils.ValidatePhone(req.NewPhone) {
		return response.NewBizError(response.ErrCodeInvalidParam, "手机号格式错误")
	}
	// 注入userId
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}

	_, err = l.svcCtx.UserRpc.UpdatePhone(l.ctx, &user.UpdatePhoneReq{
		OldPhoneCode:   req.OldPhoneCode,
		NewPhone:       req.NewPhone,
		NewPhoneCode:   req.NewPhoneCode,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc出错，error=%v", err)
		return err
	}

	return nil
}
