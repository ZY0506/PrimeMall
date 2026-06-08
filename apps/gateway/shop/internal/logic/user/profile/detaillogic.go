// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package profile

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type DetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDetailLogic 获取个人资料
func NewDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DetailLogic {
	return &DetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DetailLogic) Detail() (resp *types.UserInfoResp, err error) {
	// 注入user_id
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}
	res, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.EmptyReq{})
	if err != nil {
		l.Logger.Errorf("调用rpc出错，error=%v", err)
		return
	}
	return &types.UserInfoResp{
		Id:         res.Id,
		Phone:      res.Phone,
		Nickname:   res.Nickname,
		Avatar:     res.Avatar,
		Gender:     res.Gender,
		Birthday:   res.Birthday.AsTime().Format(time.RFC3339),
		Status:     int64(res.Status),
		StatusDesc: res.StatusDesc,
	}, nil
}
