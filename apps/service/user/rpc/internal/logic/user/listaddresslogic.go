package userlogic

import (
	"context"
	"encoding/json"

	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAddressLogic {
	return &ListAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListAddress 获取地址列表
func (l *ListAddressLogic) ListAddress(in *user.AddressListReq) (*user.AddressListResp, error) {
	// 参数校验
	if in.Page <= 0 || in.Size <= 0 {
		l.Logger.Errorf("参数错误,page=%d,size=%d", in.Page, in.Size)
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "参数错误")
	}

	// 获取当前账号信息
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 查询数据
	list, total, err := l.svcCtx.AddressModel.FindPageByUserId(l.ctx, userId, in.Page, in.Size)
	if err != nil {
		l.Logger.Errorf("查询地址列表失败,userId=%d,page=%d,size=%d,err=%v", userId, in.Page, in.Size, err)
		return nil, err
	}
	var addressList []*user.AddressItem
	for _, addr := range list {
		item := &user.AddressItem{
			Id:            addr.Id,
			Tag:           addr.Tag,
			ReceiverName:  addr.ReceiverName,
			ReceiverPhone: addr.ReceiverPhone,
			IsDefault:     addr.IsDefault,
		}
		var detail user.AddressDetail
		if err := json.Unmarshal([]byte(addr.Info), &detail); err != nil {
			l.Logger.Errorf("解析地址详情失败, id=%d, info=%s, err=%v", addr.Id, addr.Info, err)
			return nil, errorx.NewBizError(response.InternalError, "解析地址信息失败")
		}
		item.Address = &detail
		addressList = append(addressList, item)
	}

	l.Logger.Info("获取地址列表成功")

	return &user.AddressListResp{
		List:  addressList,
		Total: total,
	}, nil
}
