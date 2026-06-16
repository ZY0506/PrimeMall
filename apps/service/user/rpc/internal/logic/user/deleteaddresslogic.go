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

type DeleteAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteAddress 删除收货地址
func (l *DeleteAddressLogic) DeleteAddress(in *user.IdPathReq) (*user.EmptyResp, error) {
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
		l.Logger.Infof("用户无权限删除该地址,userId=%d,addressId=%d", userId, in.Id)
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "用户无权限删除该地址")
	}

	// 检查是否为默认地址
	if addr.IsDefault == 1 {
		l.Logger.Infof("默认地址不可直接删除,addressId=%d", in.Id)
		return nil, errorx.NewBizError(response.ErrCodeDefaultAddressDelete, "默认地址不可直接删除，请先设置其他默认地址")
	}

	// 删除
	err = l.svcCtx.AddressModel.Delete(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("删除地址失败,id=%d,err=%v", in.Id, err)
		return nil, err
	}
	l.Logger.Info("删除地址成功,id=%d", in.Id)

	return &user.EmptyResp{}, nil
}
