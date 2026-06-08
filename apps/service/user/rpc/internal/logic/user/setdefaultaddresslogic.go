package userlogic

import (
	"context"
	"errors"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDefaultAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDefaultAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultAddressLogic {
	return &SetDefaultAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetDefaultAddress 设置默认地址
func (l *SetDefaultAddressLogic) SetDefaultAddress(in *user.IdPathReq) (*user.EmptyResp, error) {
	// 获取当前账号信息
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 检查id
	addr, err := l.svcCtx.AddressModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Infof("地址不存在,id=%d", in.Id)
			return nil, errorx.NewBizError(response.ErrCodeAddressNotFound, "地址不存在")
		}
		l.Logger.Errorf("查询地址信息失败,id=%d,err=%v", in.Id, err)
		return nil, err
	}

	if addr.UserId != userId {
		l.Logger.Infof("用户无权限修改该地址,userId=%d,addressId=%d", userId, addr.Id)
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "用户无权限修改该地址")
	}

	// 设置默认地址
	err = l.svcCtx.AddressModel.SetDefaultAddress(l.ctx, userId, in.Id)
	if err != nil {
		l.Logger.Errorf("设置默认地址失败,userId=%d,addressId=%d,err=%v", userId, in.Id, err)
		return nil, err
	}
	l.Logger.Infof("设置默认地址成功,userId=%d,addressId=%d", userId, in.Id)

	return &user.EmptyResp{}, nil
}
