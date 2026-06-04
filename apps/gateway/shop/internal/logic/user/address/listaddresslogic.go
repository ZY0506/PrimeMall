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

type ListAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListAddressLogic 获取地址列表（分页）
func NewListAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAddressLogic {
	return &ListAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAddressLogic) ListAddress(req *types.AddressListReq) (resp *types.AddressListResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}
	ret, err := l.svcCtx.UserRpc.ListAddress(l.ctx, &user.AddressListReq{
		Page: req.Page,
		Size: req.Size,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc出错，error=%v", err)
		return nil, err
	}

	if ret != nil {
		resp = new(types.AddressListResp)
		for i := range ret.List {
			resp.List = append(resp.List, types.AddressItem{
				Id:            ret.List[i].Id,
				ReceiverName:  ret.List[i].ReceiverName,
				ReceiverPhone: ret.List[i].ReceiverPhone,
				Detail: types.AddressDetail{
					Province:      ret.List[i].Address.Province,
					City:          ret.List[i].Address.City,
					District:      ret.List[i].Address.District,
					DetailAddress: ret.List[i].Address.Detail,
					PostalCode:    ret.List[i].Address.PostalCode,
				},
				IsDefault: ret.List[i].IsDefault,
				Tag:       ret.List[i].Tag,
			})
		}
		resp.Total = ret.Total
	}
	return resp, nil
}
