// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package address

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateAddressLogic 更新收货地址
func NewUpdateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAddressLogic {
	return &UpdateAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAddressLogic) UpdateAddress(req *types.UpdateAddressReq) error {
	// 注入userID
	var err error
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}
	_, err = l.svcCtx.UserRpc.UpdateAddress(l.ctx, &user.AddressItem{
		Id:            req.Id,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Address: &user.AddressDetail{
			Province:   req.Detail.Province,
			City:       req.Detail.City,
			District:   req.Detail.District,
			Detail:     req.Detail.DetailAddress,
			PostalCode: req.Detail.PostalCode,
		},
		IsDefault: req.IsDefault,
		Tag:       req.Tag,
	})
	if err != nil {
		l.Logger.Errorf("更新收货地址失败，error=%v", err)
		return err
	}

	return nil
}
