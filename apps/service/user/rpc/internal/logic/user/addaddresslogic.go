package userlogic

import (
	"context"
	"encoding/json"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAddressLogic {
	return &AddAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddAddress 添加收货地址
func (l *AddAddressLogic) AddAddress(in *user.AddAddressReq) (*user.EmptyResp, error) {
	// 获取当前账号信息
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 参数校验
	if in.Tag != constants.TAG_HOME && in.Tag != constants.TAG_OFFICE && in.Tag != constants.TAG_SCHOOL {
		l.Logger.Errorf("地址标签错误,tag=%s", in.Tag)
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "地址标签错误")
	}

	// 检查地址数量是否超过上限（最多20条）
	_, total, err := l.svcCtx.AddressModel.FindPageByUserId(l.ctx, userId, 1, 1)
	if err != nil {
		l.Logger.Errorf("查询地址数量失败,err=%v", err)
		return nil, err
	}
	if total >= 20 {
		return nil, errorx.NewBizError(response.ErrCodeAddressLimitExceeded, "地址数量已达上限（最多20条）")
	}

	// 处理AddressDetail
	addrDetail, err := json.Marshal(in.Address)
	if err != nil {
		l.Logger.Errorf("序列化地址详情失败,err=%v", err)
		return nil, err
	}

	addr := &model.UserAddress{
		UserId:        userId,
		Tag:           in.Tag,
		ReceiverName:  in.ReceiverName,
		ReceiverPhone: in.ReceiverPhone,
		Info:          string(addrDetail),
		IsDefault:     in.IsDefault,
	}

	_, err = l.svcCtx.AddressModel.InsertWithDefault(l.ctx, addr)
	if err != nil {
		l.Logger.Errorf("添加地址失败,err=%v", err)
		return nil, err
	}

	l.Logger.Info("添加地址成功")

	return &user.EmptyResp{}, nil
}
