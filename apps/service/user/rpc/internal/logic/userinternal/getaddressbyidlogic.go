package userinternallogic

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAddressByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAddressByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressByIdLogic {
	return &GetAddressByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetAddressById 获取地址信息
func (l *GetAddressByIdLogic) GetAddressById(in *user.GetAddressReq) (*user.AddressItem, error) {
	ret, err := l.svcCtx.AddressModel.FindOne(l.ctx, in.AddressId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("地址不存在,id=%d", in.AddressId)
			return nil, errorx.NewBizError(response.ErrCodeAddressNotFound, "地址不存在")
		}
		l.Logger.Infof("获取地址信息失败,id=%d,err=%v", in.AddressId, err)
		return nil, err
	}

	// 校验地址所属用户
	if ret.UserId != in.UserId {
		l.Logger.Errorf("地址不属于当前用户,id=%d,userId=%d", in.AddressId, in.UserId)
		return nil, errorx.NewBizError(response.ErrCodeAddressNotBelongUser, "地址不属于当前用户")
	}

	var addressDetail user.AddressDetail

	err = json.Unmarshal([]byte(ret.Info), &addressDetail)
	if err != nil {
		l.Logger.Errorf("解析地址信息失败,id=%d,err=%v", in.AddressId, err)
		return nil, err
	}

	return &user.AddressItem{
		Id:            ret.Id,
		ReceiverName:  ret.ReceiverName,
		ReceiverPhone: ret.ReceiverPhone,
		Address:       &addressDetail,
		Tag:           ret.Tag,
		IsDefault:     ret.IsDefault,
	}, nil
}
