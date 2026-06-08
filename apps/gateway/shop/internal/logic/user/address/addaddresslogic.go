// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package address

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddAddressLogic 添加收货地址（幂等）
func NewAddAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAddressLogic {
	return &AddAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddAddressLogic) AddAddress(req *types.AddAddressReq) error {
	// 注入userID
	var err error
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}

	// 添加收货地址
	_, err = l.svcCtx.UserRpc.AddAddress(l.ctx, &user.AddAddressReq{
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
		l.Logger.Errorf("调用rpc出错，error=%v", err)
		return err
	}
	return nil
}
