package userlogic

import (
	"context"
	"encoding/json"

	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAddressLogic {
	return &UpdateAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateAddress 更新收货地址
func (l *UpdateAddressLogic) UpdateAddress(in *user.AddressItem) (*user.EmptyResp, error) {
	// 获取当前用户id
	userID, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 查询要修改的地址
	address, err := l.svcCtx.AddressModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Infof("地址不存在,id=%d", in.Id)
			return nil, errorx.NewBizError(response.ErrCodeAddressNotFound, "地址不存在")
		}
		l.Logger.Errorf("查询地址信息失败,id=%d,err=%v", in.Id, err)
		return nil, err
	}
	if address.UserId != userID {
		l.Logger.Errorf("用户无权限修改该地址,userId=%d,addressId=%d", userID, in.Id)
		return nil, errorx.NewBizError(response.ErrCodeAddressNotBelongUser, "该地址不属于当前用户")
	}
	if in.ReceiverName != "" {
		address.ReceiverName = in.ReceiverName
	}
	if in.ReceiverPhone != "" {
		address.ReceiverPhone = in.ReceiverPhone
	}
	if in.Tag != "" {
		address.Tag = in.Tag
	}
	if in.Address != nil {
		detail, err := json.Marshal(in.Address)
		if err != nil {
			l.Logger.Errorf("地址序列化失败,err=%v", err)
			return nil, err
		}
		address.Info = string(detail)
	}
	err = l.svcCtx.AddressModel.UpdateAddress(l.ctx, address)
	if err != nil {
		l.Logger.Errorf("更新地址失败,err=%v", err)
		return nil, err
	}

	l.Logger.Info("更新地址成功")

	return &user.EmptyResp{}, nil
}
