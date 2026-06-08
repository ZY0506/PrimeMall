// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package profile

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/response"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateInfoLogic 更新个人资料
func NewUpdateInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateInfoLogic {
	return &UpdateInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateInfoLogic) UpdateInfo(req *types.UpdateUserInfoReq) error {
	// 获取当前用户信息
	var err error
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}
	// TODO: 校验头像url是否合法

	var birthday *timestamppb.Timestamp = nil
	if req.Birthday != "" {
		birthTime, err := time.Parse(time.RFC3339, req.Birthday)
		if err != nil {
			return response.NewBizError(response.ErrCodeInvalidTimestamp, "生日格式错误")
		}
		birthday = timestamppb.New(birthTime)
	}

	// 调用rpc
	_, err = l.svcCtx.UserRpc.UpdateUserInfo(l.ctx, &user.UpdateUserInfoReq{
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Gender:   req.Gender,
		Birthday: birthday,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc出错，error=%v", err)
		return err
	}

	return nil
}
